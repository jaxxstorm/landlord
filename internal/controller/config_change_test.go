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

// TestReconciler_RestartsWorkflowOnConfigChange tests that config changes trigger workflow restart
func TestReconciler_RestartsWorkflowOnConfigChange(t *testing.T) {
	repo := newMemoryTenantRepo()
	logger := zaptest.NewLogger(t)

	// Create tenant with workflow execution in backing-off state
	originalConfig := map[string]interface{}{
		"env":  "dev",
		"size": "small",
	}
	originalHash, _ := tenant.ComputeConfigHash(originalConfig)

	execID := "exec-backing-off"
	tn := &tenant.Tenant{
		ID:                  uuid.New(),
		Name:                "test-tenant-backoff",
		Status:              tenant.StatusProvisioning,
		DesiredConfig:       originalConfig,
		WorkflowExecutionID: &execID,
		WorkflowConfigHash:  &originalHash,
	}
	err := repo.CreateTenant(context.Background(), tn)
	require.NoError(t, err)

	// Track calls to StopExecution
	stopCalled := false
	var stopExecutionID string
	var stopReason string

	// Track new workflow trigger
	newWorkflowTriggered := false
	var newExecutionID string

	// Mock workflow client that returns backing-off status initially,
	// then stopped after StopExecution is called
	wfClient := &mockWorkflowClientForController{
		getStatusFunc: func(ctx context.Context, executionID string) (*workflow.ExecutionStatus, error) {
			if stopCalled && executionID == stopExecutionID {
				// After stop, return cancelled state
				return &workflow.ExecutionStatus{
					ExecutionID: executionID,
					State:       workflow.StateCancelled,
				}, nil
			}
			// Initially return backing-off
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
			stopExecutionID = executionID
			stopReason = reason
			return nil
		},
		triggerWithSourceFunc: func(ctx context.Context, t *tenant.Tenant, action, source string) (string, error) {
			newWorkflowTriggered = true
			newExecutionID = "exec-new-workflow"
			return newExecutionID, nil
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

	// Update tenant config to trigger change detection
	newConfig := map[string]interface{}{
		"env":  "dev",
		"size": "large", // Changed
	}
	tn.DesiredConfig = newConfig

	err = repo.UpdateTenant(context.Background(), tn)
	require.NoError(t, err)

	// Reconcile - should detect config change + degraded state, stop and restart workflow
	err = reconciler.reconcile(tn.ID.String())
	require.NoError(t, err)

	// Verify StopExecution was called
	assert.True(t, stopCalled, "StopExecution should be called")
	assert.Equal(t, "exec-backing-off", stopExecutionID, "Should stop the backing-off execution")
	assert.Equal(t, "Configuration updated", stopReason, "Should provide config update reason")

	// Verify new workflow was triggered
	assert.True(t, newWorkflowTriggered, "New workflow should be triggered")
	assert.NotEmpty(t, newExecutionID, "New execution ID should be set")

	// Verify tenant was updated with new execution ID and config hash
	updatedTenant, err := repo.GetTenantByID(context.Background(), tn.ID)
	require.NoError(t, err)
	assert.NotNil(t, updatedTenant.WorkflowExecutionID, "WorkflowExecutionID should be set")
	assert.Equal(t, newExecutionID, *updatedTenant.WorkflowExecutionID, "Should have new execution ID")

	// Verify config hash was updated
	expectedHash, _ := tenant.ComputeConfigHash(newConfig)
	assert.NotNil(t, updatedTenant.WorkflowConfigHash, "WorkflowConfigHash should be set")
	assert.Equal(t, expectedHash, *updatedTenant.WorkflowConfigHash, "Config hash should be updated")
}

// TestReconciler_DoesNotRestartHealthyWorkflowOnConfigChange tests that healthy workflows are NOT restarted
func TestReconciler_DoesNotRestartHealthyWorkflowOnConfigChange(t *testing.T) {
	repo := newMemoryTenantRepo()
	logger := zaptest.NewLogger(t)

	// Create tenant with workflow execution in healthy running state
	originalConfig := map[string]interface{}{
		"env": "prod",
	}
	originalHash, _ := tenant.ComputeConfigHash(originalConfig)

	execID := "exec-healthy"
	tn := &tenant.Tenant{
		ID:                  uuid.New(),
		Name:                "test-tenant-healthy",
		Status:              tenant.StatusProvisioning,
		DesiredConfig:       originalConfig,
		WorkflowExecutionID: &execID,
		WorkflowConfigHash:  &originalHash,
	}
	err := repo.CreateTenant(context.Background(), tn)
	require.NoError(t, err)

	// Track calls to StopExecution
	stopCalled := false

	// Mock workflow client that returns healthy running status
	wfClient := &mockWorkflowClientForController{
		getStatusFunc: func(ctx context.Context, executionID string) (*workflow.ExecutionStatus, error) {
			// Return healthy running state (no retry_state metadata)
			return &workflow.ExecutionStatus{
				ExecutionID: executionID,
				State:       workflow.StateRunning,
				Metadata:    map[string]string{},
			}, nil
		},
		stopExecutionFunc: func(ctx context.Context, t *tenant.Tenant, executionID string, reason string) error {
			stopCalled = true
			return nil
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

	// Update tenant config
	newConfig := map[string]interface{}{
		"env": "prod-v2", // Changed
	}
	tn.DesiredConfig = newConfig

	err = repo.UpdateTenant(context.Background(), tn)
	require.NoError(t, err)

	// Reconcile - should detect config change but NOT restart because workflow is healthy
	err = reconciler.reconcile(tn.ID.String())
	require.NoError(t, err)

	// Verify StopExecution was NOT called
	assert.False(t, stopCalled, "StopExecution should NOT be called for healthy workflows")

	// Verify tenant execution ID unchanged
	updatedTenant, err := repo.GetTenantByID(context.Background(), tn.ID)
	require.NoError(t, err)
	assert.Equal(t, execID, *updatedTenant.WorkflowExecutionID, "Execution ID should remain unchanged")
}

// TestReconciler_DoesNotRestartOnConfigChangeIfNoConfigHash tests backward compatibility
func TestReconciler_DoesNotRestartOnConfigChangeIfNoConfigHash(t *testing.T) {
	repo := newMemoryTenantRepo()
	logger := zaptest.NewLogger(t)

	// Create tenant WITHOUT config hash (simulates old tenant)
	originalConfig := map[string]interface{}{
		"env": "staging",
	}

	execID := "exec-old-tenant"
	tn := &tenant.Tenant{
		ID:                  uuid.New(),
		Name:                "test-tenant-no-hash",
		Status:              tenant.StatusProvisioning,
		DesiredConfig:       originalConfig,
		WorkflowExecutionID: &execID,
		// No WorkflowConfigHash set (backward compatibility)
	}
	err := repo.CreateTenant(context.Background(), tn)
	require.NoError(t, err)

	stopCalled := false

	// Mock workflow client returning backing-off state
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
			stopCalled = true
			return nil
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

	// Update config (but no hash to compare)
	newConfig := map[string]interface{}{
		"env": "staging-v2",
	}
	tn.DesiredConfig = newConfig

	err = repo.UpdateTenant(context.Background(), tn)
	require.NoError(t, err)

	// Reconcile - should NOT restart because no config hash to compare
	err = reconciler.reconcile(tn.ID.String())
	require.NoError(t, err)

	// Verify StopExecution was NOT called (backward compatibility)
	assert.False(t, stopCalled, "StopExecution should NOT be called when no stored config hash exists")
}

// TestReconciler_DoesNotRestartCompletedWorkflowOnConfigChange tests that completed workflows stay completed
func TestReconciler_DoesNotRestartCompletedWorkflowOnConfigChange(t *testing.T) {
	repo := newMemoryTenantRepo()
	logger := zaptest.NewLogger(t)

	// Create tenant with completed (succeeded) workflow
	originalConfig := map[string]interface{}{
		"env": "prod",
	}
	originalHash, _ := tenant.ComputeConfigHash(originalConfig)

	tn := &tenant.Tenant{
		ID:                  uuid.New(),
		Name:                "test-tenant-completed",
		Status:              tenant.StatusReady, // Workflow completed successfully
		DesiredConfig:       originalConfig,
		WorkflowExecutionID: nil, // Execution ID cleared after completion
		WorkflowConfigHash:  &originalHash,
	}
	err := repo.CreateTenant(context.Background(), tn)
	require.NoError(t, err)

	stopCalled := false

	// Mock workflow client - should not be called for completed workflows
	wfClient := &mockWorkflowClientForController{
		stopExecutionFunc: func(ctx context.Context, t *tenant.Tenant, executionID string, reason string) error {
			stopCalled = true
			return nil
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

	// Update tenant config (even though workflow is complete)
	newConfig := map[string]interface{}{
		"env": "prod-v2", // Changed
	}
	tn.DesiredConfig = newConfig

	err = repo.UpdateTenant(context.Background(), tn)
	require.NoError(t, err)

	// Reconcile - should NOT restart because workflow is completed (no execution ID)
	err = reconciler.reconcile(tn.ID.String())
	require.NoError(t, err)

	// Verify StopExecution was NOT called
	assert.False(t, stopCalled, "StopExecution should NOT be called for completed workflows")

	// Verify tenant status unchanged (stays ready)
	updatedTenant, err := repo.GetTenantByID(context.Background(), tn.ID)
	require.NoError(t, err)
	assert.Equal(t, tenant.StatusReady, updatedTenant.Status, "Status should remain ready")
	assert.Nil(t, updatedTenant.WorkflowExecutionID, "Execution ID should remain nil")
}
