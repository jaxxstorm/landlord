package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/jaxxstorm/landlord/internal/compute"
	"go.uber.org/zap"
)

type mockProvider struct {
	name                   string
	createWorkflowFunc     func(ctx context.Context, spec *WorkflowSpec) (*CreateWorkflowResult, error)
	invokeFunc             func(ctx context.Context, workflowID string, request *ProvisionRequest) (*ExecutionResult, error)
	startExecutionFunc     func(ctx context.Context, workflowID string, input *ExecutionInput) (*ExecutionResult, error)
	getWorkflowStatusFunc  func(ctx context.Context, executionID string) (*WorkflowStatus, error)
	getExecutionStatusFunc func(ctx context.Context, executionID string) (*ExecutionStatus, error)
	stopExecutionFunc      func(ctx context.Context, executionID string, reason string) error
	deleteWorkflowFunc     func(ctx context.Context, workflowID string) error
	validateFunc           func(ctx context.Context, spec *WorkflowSpec) error
}

func (m *mockProvider) Name() string { return m.name }

func (m *mockProvider) CreateWorkflow(ctx context.Context, spec *WorkflowSpec) (*CreateWorkflowResult, error) {
	if m.createWorkflowFunc != nil {
		return m.createWorkflowFunc(ctx, spec)
	}
	return &CreateWorkflowResult{
		WorkflowID:   spec.WorkflowID,
		ProviderType: spec.ProviderType,
		ResourceIDs:  map[string]string{"workflow": spec.WorkflowID},
		CreatedAt:    time.Now(),
	}, nil
}

func (m *mockProvider) Invoke(ctx context.Context, workflowID string, request *ProvisionRequest) (*ExecutionResult, error) {
	if m.invokeFunc != nil {
		return m.invokeFunc(ctx, workflowID, request)
	}
	return &ExecutionResult{
		ExecutionID:  "exec-123",
		WorkflowID:   workflowID,
		ProviderType: m.name,
		State:        StateRunning,
		StartedAt:    time.Now(),
	}, nil
}

func (m *mockProvider) GetWorkflowStatus(ctx context.Context, executionID string) (*WorkflowStatus, error) {
	if m.getWorkflowStatusFunc != nil {
		return m.getWorkflowStatusFunc(ctx, executionID)
	}
	return &WorkflowStatus{
		ExecutionID: executionID,
		State:       StateSucceeded,
		Output:      json.RawMessage(`{}`),
	}, nil
}

func (m *mockProvider) StartExecution(ctx context.Context, workflowID string, input *ExecutionInput) (*ExecutionResult, error) {
	if m.startExecutionFunc != nil {
		return m.startExecutionFunc(ctx, workflowID, input)
	}
	return &ExecutionResult{
		ExecutionID:  "exec-123",
		WorkflowID:   workflowID,
		ProviderType: m.name,
		State:        StateRunning,
		StartedAt:    time.Now(),
	}, nil
}

func (m *mockProvider) GetExecutionStatus(ctx context.Context, executionID string) (*ExecutionStatus, error) {
	if m.getExecutionStatusFunc != nil {
		return m.getExecutionStatusFunc(ctx, executionID)
	}
	return &ExecutionStatus{
		ExecutionID:  executionID,
		WorkflowID:   "test-workflow",
		ProviderType: m.name,
		State:        StateSucceeded,
		StartTime:    time.Now(),
		Input:        json.RawMessage(`{}`),
	}, nil
}

func (m *mockProvider) StopExecution(ctx context.Context, executionID string, reason string) error {
	if m.stopExecutionFunc != nil {
		return m.stopExecutionFunc(ctx, executionID, reason)
	}
	return nil
}

func (m *mockProvider) DeleteWorkflow(ctx context.Context, workflowID string) error {
	if m.deleteWorkflowFunc != nil {
		return m.deleteWorkflowFunc(ctx, workflowID)
	}
	return nil
}

func (m *mockProvider) Validate(ctx context.Context, spec *WorkflowSpec) error {
	if m.validateFunc != nil {
		return m.validateFunc(ctx, spec)
	}
	return nil
}

func (m *mockProvider) PostComputeCallback(ctx context.Context, executionID string, payload *compute.CallbackPayload, opts *compute.CallbackOptions) error {
	// Stub implementation for tests
	return nil
}

func TestManagerCreateWorkflow(t *testing.T) {
	registry := NewRegistry(zap.NewNop())
	provider := &mockProvider{name: "test"}
	registry.Register(provider)

	manager := New(registry, zap.NewNop())

	spec := &WorkflowSpec{
		WorkflowID:   "test-workflow",
		ProviderType: "test",
		Name:         "Test Workflow",
		Definition:   json.RawMessage(`{"test": true}`),
	}

	result, err := manager.CreateWorkflow(context.Background(), spec)
	if err != nil {
		t.Fatalf("CreateWorkflow failed: %v", err)
	}

	if result.WorkflowID != "test-workflow" {
		t.Errorf("expected workflow_id test-workflow, got %s", result.WorkflowID)
	}
}

func TestManagerCreateWorkflowInvalidSpec(t *testing.T) {
	registry := NewRegistry(zap.NewNop())
	provider := &mockProvider{name: "test"}
	registry.Register(provider)

	manager := New(registry, zap.NewNop())

	spec := &WorkflowSpec{
		WorkflowID:   "INVALID",
		ProviderType: "test",
		Name:         "Test",
		Definition:   json.RawMessage(`{}`),
	}

	_, err := manager.CreateWorkflow(context.Background(), spec)
	if err == nil {
		t.Fatal("expected error for invalid spec")
	}

	if !errors.Is(err, ErrInvalidSpec) {
		t.Errorf("expected ErrInvalidSpec, got %v", err)
	}
}

func TestManagerCreateWorkflowProviderNotFound(t *testing.T) {
	registry := NewRegistry(zap.NewNop())
	manager := New(registry, zap.NewNop())

	spec := &WorkflowSpec{
		WorkflowID:   "test-workflow",
		ProviderType: "nonexistent",
		Name:         "Test",
		Definition:   json.RawMessage(`{}`),
	}

	_, err := manager.CreateWorkflow(context.Background(), spec)
	if err == nil {
		t.Fatal("expected error for missing provider")
	}

	if !errors.Is(err, ErrProviderNotFound) {
		t.Errorf("expected ErrProviderNotFound, got %v", err)
	}
}

func TestManagerStartExecution(t *testing.T) {
	registry := NewRegistry(zap.NewNop())
	provider := &mockProvider{name: "test"}
	registry.Register(provider)

	manager := New(registry, zap.NewNop())

	input := &ExecutionInput{
		ExecutionName: "test-execution",
		Input:         json.RawMessage(`{"key": "value"}`),
	}

	result, err := manager.StartExecution(context.Background(), "test-workflow", "test", input)
	if err != nil {
		t.Fatalf("StartExecution failed: %v", err)
	}

	if result.ExecutionID == "" {
		t.Error("expected non-empty execution_id")
	}

	if result.State != StateRunning {
		t.Errorf("expected state running, got %s", result.State)
	}
}

func TestManagerStartExecutionInvalidInput(t *testing.T) {
	registry := NewRegistry(zap.NewNop())
	provider := &mockProvider{name: "test"}
	registry.Register(provider)

	manager := New(registry, zap.NewNop())

	input := &ExecutionInput{
		ExecutionName: "test",
		Input:         json.RawMessage(`invalid json`),
	}

	_, err := manager.StartExecution(context.Background(), "test-workflow", "test", input)
	if err == nil {
		t.Fatal("expected error for invalid input")
	}

	if !errors.Is(err, ErrInvalidSpec) {
		t.Errorf("expected ErrInvalidSpec, got %v", err)
	}
}

func TestManagerGetExecutionStatus(t *testing.T) {
	registry := NewRegistry(zap.NewNop())
	provider := &mockProvider{name: "test"}
	registry.Register(provider)

	manager := New(registry, zap.NewNop())

	status, err := manager.GetExecutionStatus(context.Background(), "exec-123", "test")
	if err != nil {
		t.Fatalf("GetExecutionStatus failed: %v", err)
	}

	if status.State != StateSucceeded {
		t.Errorf("expected state succeeded, got %s", status.State)
	}
}

func TestManagerStopExecution(t *testing.T) {
	registry := NewRegistry(zap.NewNop())
	provider := &mockProvider{name: "test"}
	registry.Register(provider)

	manager := New(registry, zap.NewNop())

	err := manager.StopExecution(context.Background(), "exec-123", "test", "User requested")
	if err != nil {
		t.Fatalf("StopExecution failed: %v", err)
	}
}

func TestManagerDeleteWorkflow(t *testing.T) {
	registry := NewRegistry(zap.NewNop())
	provider := &mockProvider{name: "test"}
	registry.Register(provider)

	manager := New(registry, zap.NewNop())

	err := manager.DeleteWorkflow(context.Background(), "test-workflow", "test")
	if err != nil {
		t.Fatalf("DeleteWorkflow failed: %v", err)
	}
}

func TestManagerValidateWorkflowSpec(t *testing.T) {
	registry := NewRegistry(zap.NewNop())
	provider := &mockProvider{name: "test"}
	registry.Register(provider)

	manager := New(registry, zap.NewNop())

	spec := &WorkflowSpec{
		WorkflowID:   "test-workflow",
		ProviderType: "test",
		Name:         "Test",
		Definition:   json.RawMessage(`{}`),
	}

	err := manager.ValidateWorkflowSpec(context.Background(), spec)
	if err != nil {
		t.Fatalf("ValidateWorkflowSpec failed: %v", err)
	}
}

func TestManagerListProviders(t *testing.T) {
	registry := NewRegistry(zap.NewNop())
	registry.Register(&mockProvider{name: "provider1"})
	registry.Register(&mockProvider{name: "provider2"})

	manager := New(registry, zap.NewNop())

	providers := manager.ListProviders()
	if len(providers) != 2 {
		t.Errorf("expected 2 providers, got %d", len(providers))
	}
}
