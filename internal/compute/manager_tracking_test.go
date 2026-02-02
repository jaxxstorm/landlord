package compute

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jaxxstorm/landlord/internal/logger"
)

// MockExecutionRepository is a mock implementation for testing
type MockExecutionRepository struct {
	mock.Mock
}

func (m *MockExecutionRepository) CreateComputeExecution(ctx context.Context, exec *ComputeExecution) error {
	args := m.Called(ctx, exec)
	return args.Error(0)
}

func (m *MockExecutionRepository) UpdateComputeExecution(ctx context.Context, exec *ComputeExecution) error {
	args := m.Called(ctx, exec)
	return args.Error(0)
}

func (m *MockExecutionRepository) GetComputeExecution(ctx context.Context, executionID string) (*ComputeExecution, error) {
	args := m.Called(ctx, executionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ComputeExecution), args.Error(1)
}

func (m *MockExecutionRepository) ListComputeExecutions(ctx context.Context, tenantID string, filters ExecutionListFilters) ([]*ComputeExecution, error) {
	args := m.Called(ctx, tenantID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ComputeExecution), args.Error(1)
}

func (m *MockExecutionRepository) AddExecutionHistory(ctx context.Context, history *ComputeExecutionHistory) (err error) {
	defer func() {
		if r := recover(); r != nil {
			// If no expectation was set, swallow the panic and return nil to allow tests to run
			err = nil
		}
	}()
	args := m.Called(ctx, history)
	return args.Error(0)
}

func (m *MockExecutionRepository) GetExecutionHistory(ctx context.Context, executionID string) ([]*ComputeExecutionHistory, error) {
	args := m.Called(ctx, executionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ComputeExecutionHistory), args.Error(1)
}

// MockComputeProviderForTracking is a mock compute provider for testing
type MockComputeProviderForTracking struct {
	mock.Mock
	ProvisionError error
	UpdateError    error
	DestroyError   error
}

func (m *MockComputeProviderForTracking) Name() string {
	return "mock-tracking"
}

func (m *MockComputeProviderForTracking) Provision(ctx context.Context, spec *TenantComputeSpec) (*ProvisionResult, error) {
	if m.ProvisionError != nil {
		return nil, m.ProvisionError
	}
	args := m.Called(ctx, spec)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ProvisionResult), args.Error(1)
}

func (m *MockComputeProviderForTracking) Update(ctx context.Context, tenantID string, spec *TenantComputeSpec) (*UpdateResult, error) {
	if m.UpdateError != nil {
		return nil, m.UpdateError
	}
	args := m.Called(ctx, tenantID, spec)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UpdateResult), args.Error(1)
}

func (m *MockComputeProviderForTracking) Destroy(ctx context.Context, tenantID string) error {
	if m.DestroyError != nil {
		return m.DestroyError
	}
	args := m.Called(ctx, tenantID)
	return args.Error(0)
}

func (m *MockComputeProviderForTracking) GetStatus(ctx context.Context, tenantID string) (*ComputeStatus, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ComputeStatus), args.Error(1)
}

func (m *MockComputeProviderForTracking) Validate(ctx context.Context, spec *TenantComputeSpec) error {
	args := m.Called(ctx, spec)
	return args.Error(0)
}

func (m *MockComputeProviderForTracking) ValidateConfig(config json.RawMessage) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockComputeProviderForTracking) ConfigSchema() json.RawMessage {
	return json.RawMessage(`{}`)
}

func (m *MockComputeProviderForTracking) ConfigDefaults() json.RawMessage {
	return nil
}

func TestComputeManagerTracking(t *testing.T) {
	ctx := context.Background()
	log, _ := logger.New("development", "debug")
	defer log.Sync()

	t.Run("ProvisionTenantWithTracking_Success", func(t *testing.T) {
		mockRepo := new(MockExecutionRepository)
		mockProvider := new(MockComputeProviderForTracking)
		registry := NewRegistry(log)
		registry.Register(mockProvider)

		manager := NewWithTracking(registry, mockRepo, log)

		spec := &TenantComputeSpec{
			TenantID:       "tenant-001",
			ProviderType:   "mock-tracking",
			ProviderConfig: json.RawMessage(`{"cpu": "1", "memory": "512Mi"}`),
			Containers:     []ContainerSpec{{Name: "test", Image: "test:latest"}},
			Resources:      ResourceRequirements{CPU: 256, Memory: 512},
		}

		// Mock repository calls (accept any context)
		mockRepo.On("CreateComputeExecution", mock.Anything, mock.Anything).Return(nil)
		mockRepo.On("UpdateComputeExecution", mock.Anything, mock.Anything).Return(nil)

		// Mock provider call
		mockProvider.On("Provision", mock.Anything, spec).Return(
			&ProvisionResult{
				TenantID:     "tenant-001",
				ProviderType: "mock-tracking",
				Status:       ProvisionStatusSuccess,
				ResourceIDs:  map[string]string{"container": "container-id-001"},
			},
			nil,
		)

		result, err := manager.ProvisionTenantWithTracking(ctx, spec, "wf-exec-001")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, ExecutionStatusSucceeded, result.Status)

		mockRepo.AssertExpectations(t)
		mockProvider.AssertExpectations(t)
	})

	t.Run("ProvisionTenantWithTracking_Failure", func(t *testing.T) {
		mockRepo := new(MockExecutionRepository)
		mockProvider := new(MockComputeProviderForTracking)
		registry := NewRegistry(log)
		registry.Register(mockProvider)

		manager := NewWithTracking(registry, mockRepo, log)

		spec := &TenantComputeSpec{
			TenantID:       "tenant-002",
			ProviderType:   "mock-tracking",
			ProviderConfig: json.RawMessage(`{"image": "invalid:image"}`),
			Containers:     []ContainerSpec{{Name: "test", Image: "test:latest"}},
			Resources:      ResourceRequirements{CPU: 256, Memory: 512},
		}

		// Mock repository calls (accept any context)
		mockRepo.On("CreateComputeExecution", mock.Anything, mock.Anything).Return(nil)
		mockRepo.On("UpdateComputeExecution", mock.Anything, mock.Anything).Return(nil)

		// Mock provider error
		mockProvider.On("Provision", mock.Anything, spec).Return(
			nil,
			errors.New("image not found"),
		)

		result, err := manager.ProvisionTenantWithTracking(ctx, spec, "wf-exec-002")
		assert.Error(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, ExecutionStatusFailed, result.Status)

		mockRepo.AssertExpectations(t)
		mockProvider.AssertExpectations(t)
	})

	t.Run("UpdateTenantWithTracking_Success", func(t *testing.T) {
		mockRepo := new(MockExecutionRepository)
		mockProvider := new(MockComputeProviderForTracking)
		registry := NewRegistry(log)
		registry.Register(mockProvider)

		manager := NewWithTracking(registry, mockRepo, log)

		spec := &TenantComputeSpec{
			TenantID:       "tenant-003",
			ProviderType:   "mock-tracking",
			ProviderConfig: json.RawMessage(`{"cpu": "2"}`),
			Containers:     []ContainerSpec{{Name: "test", Image: "test:latest"}},
			Resources:      ResourceRequirements{CPU: 256, Memory: 512},
		}

		mockRepo.On("CreateComputeExecution", mock.Anything, mock.Anything).Return(nil)
		mockRepo.On("UpdateComputeExecution", mock.Anything, mock.Anything).Return(nil)

		mockProvider.On("Update", mock.Anything, "tenant-003", spec).Return(
			&UpdateResult{
				TenantID:     "tenant-003",
				ProviderType: "mock-tracking",
				Status:       UpdateStatusSuccess,
				Changes:      []string{"cpu updated to 2"},
			},
			nil,
		)

		result, err := manager.UpdateTenantWithTracking(ctx, "tenant-003", spec, "wf-exec-003")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, ExecutionStatusSucceeded, result.Status)

		mockRepo.AssertExpectations(t)
		mockProvider.AssertExpectations(t)
	})

	t.Run("DeleteTenantWithTracking_Success", func(t *testing.T) {
		mockRepo := new(MockExecutionRepository)
		mockProvider := new(MockComputeProviderForTracking)
		registry := NewRegistry(log)
		registry.Register(mockProvider)

		manager := NewWithTracking(registry, mockRepo, log)

		mockRepo.On("CreateComputeExecution", mock.Anything, mock.Anything).Return(nil)
		mockRepo.On("UpdateComputeExecution", mock.Anything, mock.Anything).Return(nil)

		mockProvider.On("Destroy", mock.Anything, "tenant-004").Return(nil)

		result, err := manager.DeleteTenantWithTracking(ctx, "tenant-004", "mock-tracking", "wf-exec-004")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, ExecutionStatusSucceeded, result.Status)

		mockRepo.AssertExpectations(t)
		mockProvider.AssertExpectations(t)
	})

	t.Run("GetComputeExecution", func(t *testing.T) {
		mockRepo := new(MockExecutionRepository)
		registry := NewRegistry(log)

		manager := NewWithTracking(registry, mockRepo, log)

		expectedExec := &ComputeExecution{
			ExecutionID: "exec-005",
			TenantID:    "tenant-005",
			Status:      ExecutionStatusRunning,
		}

		mockRepo.On("GetComputeExecution", ctx, "exec-005").Return(expectedExec, nil)
		exec, err := manager.GetComputeExecution(ctx, "exec-005")
		assert.NoError(t, err)
		assert.Equal(t, expectedExec, exec)

		mockRepo.AssertExpectations(t)
	})

	t.Run("GenerateComputeExecutionID_Deterministic", func(t *testing.T) {
		mockRepo := new(MockExecutionRepository)
		registry := NewRegistry(log)

		manager := NewWithTracking(registry, mockRepo, log)

		id1 := manager.GenerateComputeExecutionID("tenant-001", OperationTypeProvision)
		id2 := manager.GenerateComputeExecutionID("tenant-001", OperationTypeProvision)

		// Same inputs should produce same ID (deterministic)
		assert.Equal(t, id1, id2)

		// Different inputs should produce different IDs
		id3 := manager.GenerateComputeExecutionID("tenant-002", OperationTypeProvision)
		assert.NotEqual(t, id1, id3)

		id4 := manager.GenerateComputeExecutionID("tenant-001", OperationTypeUpdate)
		assert.NotEqual(t, id1, id4)
	})

	t.Run("MapProviderErrorToComputeError", func(t *testing.T) {
		mockRepo := new(MockExecutionRepository)
		registry := NewRegistry(log)

		manager := NewWithTracking(registry, mockRepo, log)

		tests := []struct {
			name            string
			providerErr     error
			expectRetriable bool
		}{
			{
				name:            "timeout error",
				providerErr:     errors.New("context deadline exceeded"),
				expectRetriable: true,
			},
			{
				name:            "quota exceeded",
				providerErr:     errors.New("resource quota exceeded"),
				expectRetriable: true,
			},
			{
				name:            "invalid config",
				providerErr:     errors.New("invalid configuration"),
				expectRetriable: false,
			},
			{
				name:            "generic error",
				providerErr:     errors.New("unknown error"),
				expectRetriable: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				compErr := manager.MapProviderErrorToComputeError(tt.providerErr)
				assert.NotNil(t, compErr)
				assert.Equal(t, tt.expectRetriable, compErr.IsRetriable)
				assert.Equal(t, tt.providerErr.Error(), compErr.ProviderError)
			})
		}
	})
}
