package compute

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jaxxstorm/landlord/internal/logger"
)

// MockWorkflowProvider is a mock implementation of WorkflowProvider for testing
type MockWorkflowProvider struct {
	mock.Mock
}

func (m *MockWorkflowProvider) PostComputeCallback(ctx context.Context, executionID string, payload *CallbackPayload, opts *CallbackOptions) error {
	args := m.Called(ctx, executionID, payload, opts)
	return args.Error(0)
}

// MockComputeProvider is a minimal compute provider used for tests
type MockComputeProvider struct {
	mock.Mock
	ProvisionError error
	UpdateError    error
	DestroyError   error
}

func (m *MockComputeProvider) Name() string { return "docker" }

func (m *MockComputeProvider) Provision(ctx context.Context, spec *TenantComputeSpec) (*ProvisionResult, error) {
	if m.ProvisionError != nil {
		return nil, m.ProvisionError
	}
	args := m.Called(ctx, spec)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ProvisionResult), args.Error(1)
}

func (m *MockComputeProvider) Update(ctx context.Context, tenantID string, spec *TenantComputeSpec) (*UpdateResult, error) {
	if m.UpdateError != nil {
		return nil, m.UpdateError
	}
	args := m.Called(ctx, tenantID, spec)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UpdateResult), args.Error(1)
}

func (m *MockComputeProvider) Destroy(ctx context.Context, tenantID string) error {
	if m.DestroyError != nil {
		return m.DestroyError
	}
	args := m.Called(ctx, tenantID)
	return args.Error(0)
}

func (m *MockComputeProvider) GetStatus(ctx context.Context, tenantID string) (*ComputeStatus, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ComputeStatus), args.Error(1)
}

func (m *MockComputeProvider) Validate(ctx context.Context, spec *TenantComputeSpec) error {
	args := m.Called(ctx, spec)
	return args.Error(0)
}

func (m *MockComputeProvider) ValidateConfig(config json.RawMessage) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockComputeProvider) ConfigSchema() json.RawMessage {
	return json.RawMessage(`{}`)
}

func (m *MockComputeProvider) ConfigDefaults() json.RawMessage {
	return nil
}

// TestPostCallbackOnSuccess verifies callbacks are posted when compute succeeds
func TestPostCallbackOnSuccess(t *testing.T) {
	ctx := context.Background()
	log, _ := logger.New("development", "debug")
	defer log.Sync()

	// Create mocks
	registry := NewRegistry(log)
	mockRepo := &MockExecutionRepository{}
	mockProvider := &MockWorkflowProvider{}

	// Register a simple compute provider and stub provision call
	prov := &MockComputeProvider{}
	prov.On("Provision", mock.Anything, mock.Anything).Return(&ProvisionResult{
		TenantID:     "tenant-123",
		ProviderType: "docker",
		Status:       ProvisionStatusSuccess,
		ResourceIDs:  map[string]string{"container": "container-id-001"},
	}, nil)
	registry.Register(prov)
	mockProvider.On("PostComputeCallback", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockRepo.On("CreateComputeExecution", ctx, mock.Anything).Return(nil)
	mockRepo.On("UpdateComputeExecution", ctx, mock.Anything).Return(nil)
	mockRepo.On("AddExecutionHistory", ctx, mock.Anything).Return(nil)

	// Create manager with callback provider
	manager := NewWithTracking(registry, mockRepo, log)
	manager.SetWorkflowProvider(mockProvider)

	// Create compute spec
	spec := &TenantComputeSpec{
		TenantID:     "tenant-123",
		ProviderType: "docker",
		Containers: []ContainerSpec{
			{Name: "app", Image: "nginx:latest"},
		},
		Resources: ResourceRequirements{
			CPU:    256,
			Memory: 512,
		},
		ProviderConfig: json.RawMessage(`{"image": "test"}`),
	}

	// Provision tenant
	exec, err := manager.ProvisionTenantWithTracking(ctx, spec, "workflow-456")

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, exec)
	assert.Equal(t, ExecutionStatusSucceeded, exec.Status)

	// Verify callback was posted
	mockProvider.AssertCalled(t, "PostComputeCallback", mock.Anything, mock.Anything, mock.MatchedBy(func(payload *CallbackPayload) bool {
		return payload.Status == ExecutionStatusSucceeded && payload.TenantID == "tenant-123"
	}), mock.Anything)
}

// TestPostCallbackOnFailure verifies callbacks are posted when compute fails
func TestPostCallbackOnFailure(t *testing.T) {
	ctx := context.Background()
	log, _ := logger.New("development", "debug")
	defer log.Sync()

	// Create mocks
	registry := NewRegistry(log)
	mockRepo := &MockExecutionRepository{}
	mockProvider := &MockWorkflowProvider{}

	// Register provider that fails provisioning
	provFail := &MockComputeProvider{}
	provFail.On("Provision", mock.Anything, mock.Anything).Return(nil, errors.New("compute engine unavailable"))
	registry.Register(provFail)
	mockProvider.On("PostComputeCallback", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockRepo.On("CreateComputeExecution", ctx, mock.Anything).Return(nil)
	mockRepo.On("UpdateComputeExecution", ctx, mock.Anything).Return(nil)
	mockRepo.On("AddExecutionHistory", ctx, mock.Anything).Return(nil)

	// Create manager with callback provider
	manager := NewWithTracking(registry, mockRepo, log)
	manager.SetWorkflowProvider(mockProvider)

	// Create compute spec
	spec := &TenantComputeSpec{
		TenantID:     "tenant-789",
		ProviderType: "docker",
		Containers: []ContainerSpec{
			{Name: "app", Image: "nginx:latest"},
		},
		Resources: ResourceRequirements{
			CPU:    256,
			Memory: 512,
		},
		ProviderConfig: json.RawMessage(`{"image": "test"}`),
	}

	// Provision tenant (should fail)
	exec, err := manager.ProvisionTenantWithTracking(ctx, spec, "workflow-999")

	// Assertions
	require.Error(t, err)
	require.NotNil(t, exec)
	assert.Equal(t, ExecutionStatusFailed, exec.Status)
	require.NotNil(t, exec.ErrorCode)
	assert.Equal(t, "PROVISIONING_FAILED", *exec.ErrorCode)

	// Verify callback was posted with failure status
	mockProvider.AssertCalled(t, "PostComputeCallback", mock.Anything, mock.Anything, mock.MatchedBy(func(payload *CallbackPayload) bool {
		return payload.Status == ExecutionStatusFailed && payload.TenantID == "tenant-789" && payload.ErrorCode != ""
	}), mock.Anything)
}

// TestCallbackNotPostedWithoutProvider verifies no crash if provider not configured
func TestCallbackNotPostedWithoutProvider(t *testing.T) {
	ctx := context.Background()
	log, _ := logger.New("development", "debug")
	defer log.Sync()

	// Create mocks without workflow provider
	registry := NewRegistry(log)
	mockRepo := &MockExecutionRepository{}

	provNoCallback := &MockComputeProvider{}
	provNoCallback.On("Provision", mock.Anything, mock.Anything).Return(&ProvisionResult{
		TenantID:     "tenant-111",
		ProviderType: "docker",
		Status:       ProvisionStatusSuccess,
	}, nil)
	registry.Register(provNoCallback)
	mockRepo.On("CreateComputeExecution", ctx, mock.Anything).Return(nil)
	mockRepo.On("UpdateComputeExecution", ctx, mock.Anything).Return(nil)
	mockRepo.On("AddExecutionHistory", ctx, mock.Anything).Return(nil)

	// Create manager WITHOUT callback provider
	manager := NewWithTracking(registry, mockRepo, log)
	// Don't call SetWorkflowProvider

	// Create compute spec
	spec := &TenantComputeSpec{
		TenantID:     "tenant-111",
		ProviderType: "docker",
		Containers: []ContainerSpec{
			{Name: "app", Image: "nginx:latest"},
		},
		Resources: ResourceRequirements{
			CPU:    256,
			Memory: 512,
		},
		ProviderConfig: json.RawMessage(`{"image": "test"}`),
	}

	// Provision tenant should succeed even without callback provider
	exec, err := manager.ProvisionTenantWithTracking(ctx, spec, "workflow-222")

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, exec)
	assert.Equal(t, ExecutionStatusSucceeded, exec.Status)
	// No callback should be posted, but no error should occur
}

// TestCallbackPayloadContainsResourceIDs verifies resource IDs are included
func TestCallbackPayloadContainsResourceIDs(t *testing.T) {
	ctx := context.Background()
	log, _ := logger.New("development", "debug")
	defer log.Sync()

	// Create mocks
	registry := NewRegistry(log)
	mockRepo := &MockExecutionRepository{}

	capturedPayload := &CallbackPayload{}
	mockProvider := &MockWorkflowProvider{}
	mockProvider.On("PostComputeCallback", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			payload := args.Get(2).(*CallbackPayload)
			*capturedPayload = *payload
		}).
		Return(nil)

	provResource := &MockComputeProvider{}
	provResource.On("Provision", mock.Anything, mock.Anything).Return(&ProvisionResult{
		TenantID:     "tenant-333",
		ProviderType: "docker",
		Status:       ProvisionStatusSuccess,
		ResourceIDs:  map[string]string{"container": "ctr-333"},
	}, nil)
	registry.Register(provResource)
	mockRepo.On("CreateComputeExecution", ctx, mock.Anything).Return(nil)
	mockRepo.On("UpdateComputeExecution", ctx, mock.Anything).Return(nil)
	mockRepo.On("AddExecutionHistory", ctx, mock.Anything).Return(nil)

	// Create manager with callback provider
	manager := NewWithTracking(registry, mockRepo, log)
	manager.SetWorkflowProvider(mockProvider)

	// Create compute spec
	spec := &TenantComputeSpec{
		TenantID:     "tenant-333",
		ProviderType: "docker",
		Containers: []ContainerSpec{
			{Name: "app", Image: "nginx:latest"},
		},
		Resources: ResourceRequirements{
			CPU:    256,
			Memory: 512,
		},
		ProviderConfig: json.RawMessage(`{"image": "test"}`),
	}

	// Provision tenant
	exec, err := manager.ProvisionTenantWithTracking(ctx, spec, "workflow-444")

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, exec)

	// Verify callback payload has resource IDs
	assert.Equal(t, ExecutionStatusSucceeded, capturedPayload.Status)
	assert.Equal(t, "tenant-333", capturedPayload.TenantID)
	assert.NotNil(t, capturedPayload.ResourceIDs)
}

// TestCallbackPayloadContainsErrorDetails verifies error details are included on failure
func TestCallbackPayloadContainsErrorDetails(t *testing.T) {
	ctx := context.Background()
	log, _ := logger.New("development", "debug")
	defer log.Sync()

	// Create mocks
	registry := NewRegistry(log)
	mockRepo := &MockExecutionRepository{}

	capturedPayload := &CallbackPayload{}
	mockProvider := &MockWorkflowProvider{}
	mockProvider.On("PostComputeCallback", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			payload := args.Get(2).(*CallbackPayload)
			*capturedPayload = *payload
		}).
		Return(nil)

	providerErr := errors.New("provider error: timeout occurred")
	provError := &MockComputeProvider{}
	provError.On("Provision", mock.Anything, mock.Anything).Return(nil, providerErr)
	registry.Register(provError)
	mockRepo.On("CreateComputeExecution", ctx, mock.Anything).Return(nil)
	mockRepo.On("UpdateComputeExecution", ctx, mock.Anything).Return(nil)
	mockRepo.On("AddExecutionHistory", ctx, mock.Anything).Return(nil)

	// Create manager with callback provider
	manager := NewWithTracking(registry, mockRepo, log)
	manager.SetWorkflowProvider(mockProvider)

	// Create compute spec
	spec := &TenantComputeSpec{
		TenantID:     "tenant-555",
		ProviderType: "docker",
		Containers: []ContainerSpec{
			{Name: "app", Image: "nginx:latest"},
		},
		Resources: ResourceRequirements{
			CPU:    256,
			Memory: 512,
		},
		ProviderConfig: json.RawMessage(`{"image": "test"}`),
	}

	// Provision tenant (will fail)
	exec, err := manager.ProvisionTenantWithTracking(ctx, spec, "workflow-666")

	// Assertions
	require.Error(t, err)
	require.NotNil(t, exec)

	// Verify callback payload has error details
	assert.Equal(t, ExecutionStatusFailed, capturedPayload.Status)
	assert.Equal(t, "tenant-555", capturedPayload.TenantID)
	assert.NotEmpty(t, capturedPayload.ErrorCode)
	assert.NotEmpty(t, capturedPayload.ErrorMessage)
	// Timeouts should be retriable
	assert.True(t, capturedPayload.IsRetriable)
}

// TestUpdateTenantPostsCallback verifies callbacks for update operations
func TestUpdateTenantPostsCallback(t *testing.T) {
	ctx := context.Background()
	log, _ := logger.New("development", "debug")
	defer log.Sync()

	// Create mocks
	registry := NewRegistry(log)
	mockRepo := &MockExecutionRepository{}
	mockProvider := &MockWorkflowProvider{}

	provUpdate := &MockComputeProvider{}
	provUpdate.On("Update", mock.Anything, "tenant-update", mock.Anything).Return(&UpdateResult{
		TenantID:     "tenant-update",
		ProviderType: "docker",
		Status:       UpdateStatusSuccess,
	}, nil)
	registry.Register(provUpdate)
	mockProvider.On("PostComputeCallback", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockRepo.On("CreateComputeExecution", ctx, mock.Anything).Return(nil)
	mockRepo.On("UpdateComputeExecution", ctx, mock.Anything).Return(nil)
	mockRepo.On("AddExecutionHistory", ctx, mock.Anything).Return(nil)

	// Create manager with callback provider
	manager := NewWithTracking(registry, mockRepo, log)
	manager.SetWorkflowProvider(mockProvider)

	// Create compute spec
	spec := &TenantComputeSpec{
		TenantID:     "tenant-update",
		ProviderType: "docker",
		Containers: []ContainerSpec{
			{Name: "app", Image: "nginx:latest"},
		},
		Resources: ResourceRequirements{
			CPU:    256,
			Memory: 512,
		},
		ProviderConfig: json.RawMessage(`{"image": "updated"}`),
	}

	// Update tenant
	exec, err := manager.UpdateTenantWithTracking(ctx, "tenant-update", spec, "workflow-update")

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, exec)
	assert.Equal(t, ExecutionStatusSucceeded, exec.Status)

	// Verify callback was posted for update
	mockProvider.AssertCalled(t, "PostComputeCallback", mock.Anything, mock.Anything, mock.MatchedBy(func(payload *CallbackPayload) bool {
		return payload.Status == ExecutionStatusSucceeded && payload.TenantID == "tenant-update"
	}), mock.Anything)
}

// TestDeleteTenantPostsCallback verifies callbacks for delete operations
func TestDeleteTenantPostsCallback(t *testing.T) {
	ctx := context.Background()
	log, _ := logger.New("development", "debug")
	defer log.Sync()

	// Create mocks
	registry := NewRegistry(log)
	mockRepo := &MockExecutionRepository{}
	mockProvider := &MockWorkflowProvider{}

	provDelete := &MockComputeProvider{}
	provDelete.On("Destroy", mock.Anything, "tenant-delete").Return(nil)
	registry.Register(provDelete)
	mockProvider.On("PostComputeCallback", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockRepo.On("CreateComputeExecution", ctx, mock.Anything).Return(nil)
	mockRepo.On("UpdateComputeExecution", ctx, mock.Anything).Return(nil)
	mockRepo.On("AddExecutionHistory", ctx, mock.Anything).Return(nil)

	// Create manager with callback provider
	manager := NewWithTracking(registry, mockRepo, log)
	manager.SetWorkflowProvider(mockProvider)

	// Delete tenant
	exec, err := manager.DeleteTenantWithTracking(ctx, "tenant-delete", "docker", "workflow-delete")

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, exec)
	assert.Equal(t, ExecutionStatusSucceeded, exec.Status)

	// Verify callback was posted for delete
	mockProvider.AssertCalled(t, "PostComputeCallback", mock.Anything, mock.Anything, mock.MatchedBy(func(payload *CallbackPayload) bool {
		return payload.Status == ExecutionStatusSucceeded && payload.TenantID == "tenant-delete"
	}), mock.Anything)
}

// TestCallbackRetrySuccess verifies callback retry succeeds after initial failure
func TestCallbackRetrySuccess(t *testing.T) {
	ctx := context.Background()
	log, _ := logger.New("development", "debug")
	defer log.Sync()

	// Create mocks
	registry := NewRegistry(log)
	mockRepo := &MockExecutionRepository{}

	// Create custom workflow mock that tracks calls
	callCount := 0
	customWorkflow := &customWorkflowProvider{
		postCallback: func(ctx context.Context, execID string, payload *CallbackPayload, opts *CallbackOptions) error {
			callCount++
			if callCount <= 2 {
				return errors.New("temporary network error")
			}
			return nil
		},
	}

	// Register a simple compute provider and stub provision call
	prov := &MockComputeProvider{}
	prov.On("Provision", mock.Anything, mock.Anything).Return(&ProvisionResult{
		TenantID:     "tenant-retry",
		ProviderType: "docker",
		Status:       ProvisionStatusSuccess,
		ResourceIDs:  map[string]string{"container_id": "abc123"},
	}, nil)
	registry.Register(prov)

	mockRepo.On("CreateComputeExecution", mock.Anything, mock.Anything).Return(nil)
	mockRepo.On("UpdateComputeExecution", mock.Anything, mock.Anything).Return(nil)

	// Create manager
	manager := NewWithTracking(registry, mockRepo, log)
	manager.SetWorkflowProvider(customWorkflow)

	// Provision tenant
	spec := &TenantComputeSpec{
		TenantID:       "tenant-retry",
		ProviderType:   "docker",
		ProviderConfig: json.RawMessage(`{"image": "nginx:latest"}`),
		Containers: []ContainerSpec{
			{Name: "web", Image: "nginx:latest"},
		},
		Resources: ResourceRequirements{
			CPU:    256,
			Memory: 512,
		},
	}

	exec, err := manager.ProvisionTenantWithTracking(ctx, spec, "workflow-retry")

	// Assertions
	require.NoError(t, err)
	require.NotNil(t, exec)
	assert.Equal(t, ExecutionStatusSucceeded, exec.Status)

	// Verify callback was attempted 3 times
	assert.Equal(t, 3, callCount)

	// No failed callbacks should be stored (since it eventually succeeded)
	failedCallbacks := manager.GetFailedCallbacks()
	assert.Empty(t, failedCallbacks)
}

// customWorkflowProvider is a simple implementation for testing callback retries
type customWorkflowProvider struct {
	postCallback func(ctx context.Context, execID string, payload *CallbackPayload, opts *CallbackOptions) error
}

func (c *customWorkflowProvider) PostComputeCallback(ctx context.Context, execID string, payload *CallbackPayload, opts *CallbackOptions) error {
	return c.postCallback(ctx, execID, payload, opts)
}

// TestCallbackRetryExhausted verifies failed callbacks are stored after all retries
func TestCallbackRetryExhausted(t *testing.T) {
	ctx := context.Background()
	log, _ := logger.New("development", "debug")
	defer log.Sync()

	// Create mocks
	registry := NewRegistry(log)
	mockRepo := &MockExecutionRepository{}

	// Create custom workflow mock that always fails
	customWorkflow := &customWorkflowProvider{
		postCallback: func(ctx context.Context, execID string, payload *CallbackPayload, opts *CallbackOptions) error {
			return errors.New("persistent workflow error")
		},
	}

	// Register a simple compute provider and stub provision call
	prov := &MockComputeProvider{}
	prov.On("Provision", mock.Anything, mock.Anything).Return(&ProvisionResult{
		TenantID:     "tenant-fail",
		ProviderType: "docker",
		Status:       ProvisionStatusSuccess,
		ResourceIDs:  map[string]string{"container_id": "abc123"},
	}, nil)
	registry.Register(prov)

	mockRepo.On("CreateComputeExecution", mock.Anything, mock.Anything).Return(nil)
	mockRepo.On("UpdateComputeExecution", mock.Anything, mock.Anything).Return(nil)

	// Create manager
	manager := NewWithTracking(registry, mockRepo, log)
	manager.SetWorkflowProvider(customWorkflow)

	// Provision tenant
	spec := &TenantComputeSpec{
		TenantID:       "tenant-fail",
		ProviderType:   "docker",
		ProviderConfig: json.RawMessage(`{"image": "nginx:latest"}`),
		Containers: []ContainerSpec{
			{Name: "web", Image: "nginx:latest"},
		},
		Resources: ResourceRequirements{
			CPU:    256,
			Memory: 512,
		},
	}

	exec, err := manager.ProvisionTenantWithTracking(ctx, spec, "workflow-fail")

	// Assertions - operation succeeds even though callback failed
	require.NoError(t, err)
	require.NotNil(t, exec)
	assert.Equal(t, ExecutionStatusSucceeded, exec.Status)

	// Failed callback should be stored
	failedCallbacks := manager.GetFailedCallbacks()
	require.Len(t, failedCallbacks, 1)
	assert.Equal(t, exec.ExecutionID, failedCallbacks[0].ExecutionID)
	assert.Equal(t, "tenant-fail", failedCallbacks[0].Payload.TenantID)
	assert.Contains(t, failedCallbacks[0].Error, "persistent workflow error")
}

// TestManualCallbackRetry verifies manual retry of failed callbacks
func TestManualCallbackRetry(t *testing.T) {
	ctx := context.Background()
	log, _ := logger.New("development", "debug")
	defer log.Sync()

	// Create mocks
	registry := NewRegistry(log)
	mockRepo := &MockExecutionRepository{}

	// Create custom workflow mock that fails first 4 times, succeeds on 5th (manual retry)
	initialAttempts := 0
	customWorkflow := &customWorkflowProvider{
		postCallback: func(ctx context.Context, execID string, payload *CallbackPayload, opts *CallbackOptions) error {
			initialAttempts++
			if initialAttempts <= 4 { // 4 automatic attempts (1 initial + 3 retries)
				return errors.New("temporary error")
			}
			return nil // Manual retry succeeds
		},
	}

	// Register a simple compute provider and stub provision call
	prov := &MockComputeProvider{}
	prov.On("Provision", mock.Anything, mock.Anything).Return(&ProvisionResult{
		TenantID:     "tenant-manual",
		ProviderType: "docker",
		Status:       ProvisionStatusSuccess,
		ResourceIDs:  map[string]string{"container_id": "abc123"},
	}, nil)
	registry.Register(prov)

	mockRepo.On("CreateComputeExecution", mock.Anything, mock.Anything).Return(nil)
	mockRepo.On("UpdateComputeExecution", mock.Anything, mock.Anything).Return(nil)

	// Create manager
	manager := NewWithTracking(registry, mockRepo, log)
	manager.SetWorkflowProvider(customWorkflow)

	// Provision tenant (will fail callback)
	spec := &TenantComputeSpec{
		TenantID:       "tenant-manual",
		ProviderType:   "docker",
		ProviderConfig: json.RawMessage(`{"image": "nginx:latest"}`),
		Containers: []ContainerSpec{
			{Name: "web", Image: "nginx:latest"},
		},
		Resources: ResourceRequirements{
			CPU:    256,
			Memory: 512,
		},
	}

	exec, err := manager.ProvisionTenantWithTracking(ctx, spec, "workflow-manual")
	require.NoError(t, err)

	// Failed callback should be stored
	failedCallbacks := manager.GetFailedCallbacks()
	require.Len(t, failedCallbacks, 1)

	// Manual retry should succeed
	err = manager.RetryFailedCallback(exec.ExecutionID)
	require.NoError(t, err)

	// Failed callback should be removed
	failedCallbacks = manager.GetFailedCallbacks()
	assert.Empty(t, failedCallbacks)
}
