package controller

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jaxxstorm/landlord/internal/config"
	"github.com/jaxxstorm/landlord/internal/tenant"
	"github.com/jaxxstorm/landlord/internal/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// TestReconciler_StateTransitionsDuringConfigChangeRestart verifies that state transitions
// during config change restart are correct and don't have unexpected side effects
func TestReconciler_StateTransitionsDuringConfigChangeRestart(t *testing.T) {
	testCases := []struct {
		name           string
		initialStatus  tenant.Status
		expectedStatus tenant.Status
		description    string
	}{
		{
			name:           "provisioning stays provisioning",
			initialStatus:  tenant.StatusProvisioning,
			expectedStatus: tenant.StatusProvisioning,
			description:    "Workflow restart should not change tenant from provisioning",
		},
		{
			name:           "updating stays updating",
			initialStatus:  tenant.StatusUpdating,
			expectedStatus: tenant.StatusUpdating,
			description:    "Workflow restart should not change tenant from updating",
		},
		{
			name:           "deleting stays deleting",
			initialStatus:  tenant.StatusDeleting,
			expectedStatus: tenant.StatusDeleting,
			description:    "Workflow restart should not change tenant from deleting",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newMemoryTenantRepo()
			logger := zaptest.NewLogger(t)

			// Create tenant with backing-off workflow
			originalConfig := map[string]interface{}{
				"env": "test",
			}
			originalHash, _ := tenant.ComputeConfigHash(originalConfig)

			execID := "exec-backing-off"
			tn := &tenant.Tenant{
				ID:                  uuid.New(),
				Name:                "test-tenant",
				Status:              tc.initialStatus,
				DesiredConfig:       originalConfig,
				WorkflowExecutionID: &execID,
				WorkflowConfigHash:  &originalHash,
			}
			err := repo.CreateTenant(context.Background(), tn)
			require.NoError(t, err)

			stopCalled := false
			newExecID := "exec-new"

			wfClient := &mockWorkflowClientForController{
				getStatusFunc: func(ctx context.Context, executionID string) (*workflow.ExecutionStatus, error) {
					if stopCalled && executionID == execID {
						// After stop, return cancelled
						return &workflow.ExecutionStatus{
							ExecutionID: executionID,
							State:       workflow.StateCancelled,
						}, nil
					}
					return &workflow.ExecutionStatus{
						ExecutionID: executionID,
						State:       workflow.StateRunning,
						Metadata: map[string]string{
							"retry_state": string(workflow.SubStateBackingOff),
						},
					}, nil
				},
				stopExecutionFunc: func(ctx context.Context, t *tenant.Tenant, executionID string, reason string) error {
					stopCalled = true
					return nil
				},
				triggerWithSourceFunc: func(ctx context.Context, tn *tenant.Tenant, action, source string) (string, error) {
					// Verify tenant status hasn't changed
					assert.Equal(t, tc.initialStatus, tn.Status, "Tenant status should not change before triggering new workflow")
					return newExecID, nil
				},
				determineActionFunc: func(status tenant.Status) (string, error) {
					switch status {
					case tenant.StatusProvisioning:
						return "provision", nil
					case tenant.StatusUpdating:
						return "update", nil
					case tenant.StatusDeleting:
						return "delete", nil
					default:
						return "", nil
					}
				},
			}

			cfg := config.ControllerConfig{
				Enabled:                true,
				ReconciliationInterval: 100 * time.Millisecond,
				StatusPollInterval:     100 * time.Millisecond,
				Workers:                1,
				WorkflowTriggerTimeout: 5 * time.Second,
				ShutdownTimeout:        5 * time.Second,
				MaxRetries:             1,
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			reconciler := &Reconciler{
				tenantRepo:     repo,
				workflowClient: wfClient,
				config:         cfg,
				logger:         logger,
				retryCount:     make(map[string]int),
				queue:          NewRateLimitingQueue(),
				ctx:            ctx,
				cancel:         cancel,
			}

			// Update config to trigger restart
			newConfig := map[string]interface{}{
				"env": "test-v2",
			}
			tn.DesiredConfig = newConfig
			err = repo.UpdateTenant(context.Background(), tn)
			require.NoError(t, err)

			// Reconcile
			err = reconciler.reconcile(tn.ID.String())
			assert.NoError(t, err, tc.description)

			// Verify stop and trigger were called
			assert.True(t, stopCalled, "StopExecution should have been called")

			// Verify final tenant state
			updatedTenant, err := repo.GetTenantByID(context.Background(), tn.ID)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, updatedTenant.Status, "Tenant status should remain unchanged after restart")
			assert.Equal(t, newExecID, *updatedTenant.WorkflowExecutionID, "New execution ID should be set")

			// Verify config hash was updated
			newHash, _ := tenant.ComputeConfigHash(newConfig)
			assert.Equal(t, newHash, *updatedTenant.WorkflowConfigHash, "Config hash should be updated")
		})
	}
}

// TestReconciler_NoStateChangeForNonRestartableStatus verifies that tenants in terminal states
// (ready, failed, archived) don't trigger config change restart and maintain their status
func TestReconciler_NoStateChangeForNonRestartableStatus(t *testing.T) {
	testCases := []struct {
		name          string
		status        tenant.Status
		hasExecID     bool
		shouldRestart bool
	}{
		{
			name:          "ready without execution ID",
			status:        tenant.StatusReady,
			hasExecID:     false,
			shouldRestart: false,
		},
		{
			name:          "failed without execution ID",
			status:        tenant.StatusFailed,
			hasExecID:     false,
			shouldRestart: false,
		},
		{
			name:          "archived without execution ID",
			status:        tenant.StatusArchived,
			hasExecID:     false,
			shouldRestart: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newMemoryTenantRepo()
			logger := zaptest.NewLogger(t)

			originalConfig := map[string]interface{}{
				"env": "test",
			}
			originalHash, _ := tenant.ComputeConfigHash(originalConfig)

			var execID *string
			if tc.hasExecID {
				id := "exec-completed"
				execID = &id
			}

			tn := &tenant.Tenant{
				ID:                  uuid.New(),
				Name:                "test-tenant",
				Status:              tc.status,
				DesiredConfig:       originalConfig,
				WorkflowExecutionID: execID,
				WorkflowConfigHash:  &originalHash,
			}
			err := repo.CreateTenant(context.Background(), tn)
			require.NoError(t, err)

			stopCalled := false

			wfClient := &mockWorkflowClientForController{
				getStatusFunc: func(ctx context.Context, executionID string) (*workflow.ExecutionStatus, error) {
					return &workflow.ExecutionStatus{
						ExecutionID: executionID,
						State:       workflow.StateSucceeded,
					}, nil
				},
				stopExecutionFunc: func(ctx context.Context, t *tenant.Tenant, executionID string, reason string) error {
					stopCalled = true
					return nil
				},
				determineActionFunc: func(status tenant.Status) (string, error) {
					return "", nil // No action for terminal states
				},
			}

			cfg := config.ControllerConfig{
				Enabled:                true,
				ReconciliationInterval: 100 * time.Millisecond,
				StatusPollInterval:     100 * time.Millisecond,
				Workers:                1,
				WorkflowTriggerTimeout: 5 * time.Second,
				ShutdownTimeout:        5 * time.Second,
				MaxRetries:             1,
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			reconciler := &Reconciler{
				tenantRepo:     repo,
				workflowClient: wfClient,
				config:         cfg,
				logger:         logger,
				retryCount:     make(map[string]int),
				queue:          NewRateLimitingQueue(),
				ctx:            ctx,
				cancel:         cancel,
			}

			// Update config
			newConfig := map[string]interface{}{
				"env": "test-v2",
			}
			tn.DesiredConfig = newConfig
			err = repo.UpdateTenant(context.Background(), tn)
			require.NoError(t, err)

			// Reconcile
			err = reconciler.reconcile(tn.ID.String())
			assert.NoError(t, err)

			// Verify no restart occurred
			assert.False(t, stopCalled, "StopExecution should not be called for terminal states")

			// Verify status unchanged
			updatedTenant, err := repo.GetTenantByID(context.Background(), tn.ID)
			require.NoError(t, err)
			assert.Equal(t, tc.status, updatedTenant.Status, "Status should remain unchanged")
			assert.Equal(t, execID, updatedTenant.WorkflowExecutionID, "Execution ID should remain unchanged")
		})
	}
}
