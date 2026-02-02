package workflow

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/jaxxstorm/landlord/internal/compute"
	"go.uber.org/zap"
)

// testProvider is a minimal mock for registry tests
type testProvider struct {
	name string
}

func (t *testProvider) Name() string {
	return t.name
}

func (t *testProvider) CreateWorkflow(ctx context.Context, spec *WorkflowSpec) (*CreateWorkflowResult, error) {
	return nil, nil
}

func (t *testProvider) Invoke(ctx context.Context, workflowID string, request *ProvisionRequest) (*ExecutionResult, error) {
	return nil, nil
}

func (t *testProvider) GetWorkflowStatus(ctx context.Context, executionID string) (*WorkflowStatus, error) {
	return nil, nil
}

func (t *testProvider) StartExecution(ctx context.Context, workflowID string, input *ExecutionInput) (*ExecutionResult, error) {
	return nil, nil
}

func (t *testProvider) GetExecutionStatus(ctx context.Context, executionID string) (*ExecutionStatus, error) {
	return nil, nil
}

func (t *testProvider) StopExecution(ctx context.Context, executionID string, reason string) error {
	return nil
}

func (t *testProvider) DeleteWorkflow(ctx context.Context, workflowID string) error {
	return nil
}

func (t *testProvider) Validate(ctx context.Context, spec *WorkflowSpec) error {
	return nil
}

func (t *testProvider) PostComputeCallback(ctx context.Context, executionID string, payload *compute.CallbackPayload, opts *compute.CallbackOptions) error {
	// Stub implementation for tests
	return nil
}

func TestRegistryRegister(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	registry := NewRegistry(logger)

	provider := &testProvider{name: "test"}
	err := registry.Register(provider)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if !registry.Has("test") {
		t.Error("provider not registered")
	}
}

func TestRegistryRegisterDuplicate(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	registry := NewRegistry(logger)

	provider := &testProvider{name: "test"}
	err := registry.Register(provider)
	if err != nil {
		t.Fatalf("First register failed: %v", err)
	}

	err = registry.Register(provider)
	if err == nil {
		t.Fatal("expected error for duplicate provider")
	}

	if !errors.Is(err, ErrProviderConflict) {
		t.Errorf("expected ErrProviderConflict, got %v", err)
	}
}

func TestRegistryGet(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	registry := NewRegistry(logger)

	provider := &testProvider{name: "test"}
	registry.Register(provider)

	retrieved, err := registry.Get("test")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.Name() != "test" {
		t.Errorf("expected name 'test', got %s", retrieved.Name())
	}
}

func TestRegistryGetNotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	registry := NewRegistry(logger)

	_, err := registry.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent provider")
	}

	if !errors.Is(err, ErrProviderNotFound) {
		t.Errorf("expected ErrProviderNotFound, got %v", err)
	}
}

func TestRegistryList(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	registry := NewRegistry(logger)

	// Empty list
	if len(registry.List()) != 0 {
		t.Error("expected empty list")
	}

	// Add providers
	registry.Register(&testProvider{name: "provider1"})
	registry.Register(&testProvider{name: "provider2"})
	registry.Register(&testProvider{name: "provider3"})

	providers := registry.List()
	if len(providers) != 3 {
		t.Errorf("expected 3 providers, got %d", len(providers))
	}

	// Should be sorted
	expected := []string{"provider1", "provider2", "provider3"}
	for i, name := range providers {
		if name != expected[i] {
			t.Errorf("expected %s at position %d, got %s", expected[i], i, name)
		}
	}
}

func TestRegistryHas(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	registry := NewRegistry(logger)

	provider := &testProvider{name: "test"}
	registry.Register(provider)

	if !registry.Has("test") {
		t.Error("Has returned false for registered provider")
	}

	if registry.Has("nonexistent") {
		t.Error("Has returned true for nonexistent provider")
	}
}

func TestRegistryConcurrency(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	registry := NewRegistry(logger)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			provider := &testProvider{name: fmt.Sprintf("provider%d", id)}
			registry.Register(provider)
		}(i)
	}

	wg.Wait()

	if len(registry.List()) != 10 {
		t.Errorf("expected 10 providers, got %d", len(registry.List()))
	}
}

func TestRegistryEmptyProviderName(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	registry := NewRegistry(logger)

	provider := &testProvider{name: ""}
	err := registry.Register(provider)
	if err == nil {
		t.Fatal("expected error for empty provider name")
	}
}
