package controller

import (
	"context"
	"errors"
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

// TestReconciler_StopExecutionFailure tests error handling when StopExecution fails
func TestReconciler_StopExecutionFailure(t *testing.T) {
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
		Status:              tenant.StatusProvisioning,
		DesiredConfig:       originalConfig,
		WorkflowExecutionID: &execID,
		WorkflowConfigHash:  &originalHash,
	}
	err := repo.CreateTenant(context.Background(), tn)
	require.NoError(t, err)

	stopError := errors.New("provider unavailable")

	wfClient := &mockWorkflowClientForController{
		getStatusFunc: func(ctx context.Context, executionID string) (*workflow.ExecutionStatus, error) {
			return &workflow.ExecutionStatus{
				ExecutionID: executionID,
				State:       workflow.StateRunning,
				Metadata: map[string]string{
					"retry_state": string(workflow.SubStateBackingOff),
				},
			}, nil
		},
		stopExecutionFunc: func(ctx context.Context, t *tenant.Tenant, executionID string, reason string) error {
			return stopError
		},
		determineActionFunc: func(status tenant.Status) (string, error) {
			return "provision", nil
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

	// Reconcile - should fail with stop error
	err = reconciler.reconcile(tn.ID.String())
	assert.Error(t, err, "Should return error when StopExecution fails")
	assert.Contains(t, err.Error(), "failed to stop workflow execution")

	// Verify tenant state unchanged (will retry on next reconciliation)
	updatedTenant, err := repo.GetTenantByID(context.Background(), tn.ID)
	require.NoError(t, err)
	assert.Equal(t, execID, *updatedTenant.WorkflowExecutionID, "Execution ID should remain unchanged after error")
}

// TestReconciler_TriggerWorkflowFailureAfterStop tests error handling when new workflow trigger fails
func TestReconciler_TriggerWorkflowFailureAfterStop(t *testing.T) {
	repo := newMemoryTenantRepo()
	logger := zaptest.NewLogger(t)

	originalConfig := map[string]interface{}{
		"env": "test",
	}
	originalHash, _ := tenant.ComputeConfigHash(originalConfig)

	execID := "exec-backing-off"
	tn := &tenant.Tenant{
		ID:                  uuid.New(),
		Name:                "test-tenant",
		Status:              tenant.StatusProvisioning,
		DesiredConfig:       originalConfig,
		WorkflowExecutionID: &execID,
		WorkflowConfigHash:  &originalHash,
	}
	err := repo.CreateTenant(context.Background(), tn)
	require.NoError(t, err)

	stopCalled := false
	triggerError := errors.New("workflow provider overloaded")

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
		triggerWithSourceFunc: func(ctx context.Context, t *tenant.Tenant, action, source string) (string, error) {
			return "", triggerError
		},
		determineActionFunc: func(status tenant.Status) (string, error) {
			return "provision", nil
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

	// Reconcile - should stop successfully but fail on trigger
	err = reconciler.reconcile(tn.ID.String())
	assert.Error(t, err, "Should return error when TriggerWorkflow fails")
	assert.Contains(t, err.Error(), "failed to trigger new workflow")

	// Verify stop was called
	assert.True(t, stopCalled, "StopExecution should have been called")

	// Verify tenant has no execution ID (stop succeeded, trigger failed)
	updatedTenant, err := repo.GetTenantByID(context.Background(), tn.ID)
	require.NoError(t, err)
	assert.Nil(t, updatedTenant.WorkflowExecutionID, "Execution ID should be cleared after stop")
}

// TestReconciler_GetStatusErrorDuringPolling tests error handling when status polling fails
func TestReconciler_GetStatusErrorDuringPolling(t *testing.T) {
	repo := newMemoryTenantRepo()
	logger := zaptest.NewLogger(t)

	originalConfig := map[string]interface{}{
		"env": "test",
	}
	originalHash, _ := tenant.ComputeConfigHash(originalConfig)

	execID := "exec-backing-off"
	tn := &tenant.Tenant{
		ID:                  uuid.New(),
		Name:                "test-tenant",
		Status:              tenant.StatusProvisioning,
		DesiredConfig:       originalConfig,
		WorkflowExecutionID: &execID,
		WorkflowConfigHash:  &originalHash,
	}
	err := repo.CreateTenant(context.Background(), tn)
	require.NoError(t, err)

	stopCalled := false
	pollCount := 0

	wfClient := &mockWorkflowClientForController{
		getStatusFunc: func(ctx context.Context, executionID string) (*workflow.ExecutionStatus, error) {
			if stopCalled {
				pollCount++
				// Fail on first few polls after stop
				if pollCount < 3 {
					return nil, errors.New("network timeout")
				}
				// Eventually succeed
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
		triggerWithSourceFunc: func(ctx context.Context, t *tenant.Tenant, action, source string) (string, error) {
			return "exec-new", nil
		},
		determineActionFunc: func(status tenant.Status) (string, error) {
			return "provision", nil
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

	// Reconcile - should handle polling errors gracefully and eventually succeed
	err = reconciler.reconcile(tn.ID.String())
	assert.NoError(t, err, "Should eventually succeed despite polling errors")

	// Verify stop was called and workflow restarted
	assert.True(t, stopCalled, "StopExecution should have been called")
	assert.GreaterOrEqual(t, pollCount, 3, "Should have retried polling after errors")
}
