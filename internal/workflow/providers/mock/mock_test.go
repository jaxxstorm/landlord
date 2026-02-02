package mock

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/jaxxstorm/landlord/internal/workflow"
	"go.uber.org/zap"
)

func TestProvider_Name(t *testing.T) {
	p := New(zap.NewNop())
	if p.Name() != "mock" {
		t.Errorf("expected name 'mock', got %s", p.Name())
	}
}

func TestProvider_CreateWorkflow(t *testing.T) {
	p := New(zap.NewNop())
	ctx := context.Background()

	spec := &workflow.WorkflowSpec{
		WorkflowID:   "test-workflow",
		ProviderType: "mock",
		Name:         "Test Workflow",
		Description:  "A test workflow",
		Definition:   json.RawMessage(`{"test": true}`),
	}

	result, err := p.CreateWorkflow(ctx, spec)
	if err != nil {
		t.Fatalf("CreateWorkflow failed: %v", err)
	}

	if result.WorkflowID != spec.WorkflowID {
		t.Errorf("expected workflow_id %s, got %s", spec.WorkflowID, result.WorkflowID)
	}
	if result.ProviderType != "mock" {
		t.Errorf("expected provider_type 'mock', got %s", result.ProviderType)
	}

	result2, err := p.CreateWorkflow(ctx, spec)
	if err != nil {
		t.Fatalf("Second CreateWorkflow failed: %v", err)
	}

	if result2.WorkflowID != spec.WorkflowID {
		t.Errorf("expected workflow_id %s on second create, got %s", spec.WorkflowID, result2.WorkflowID)
	}
}

func TestProvider_StartExecution(t *testing.T) {
	p := New(zap.NewNop())
	ctx := context.Background()

	spec := &workflow.WorkflowSpec{
		WorkflowID:   "test-workflow",
		ProviderType: "mock",
		Name:         "Test Workflow",
		Definition:   json.RawMessage(`{"test": true}`),
	}

	_, err := p.CreateWorkflow(ctx, spec)
	if err != nil {
		t.Fatalf("CreateWorkflow failed: %v", err)
	}

	input := &workflow.ExecutionInput{
		Input: json.RawMessage(`{"param": "value"}`),
	}

	result, err := p.StartExecution(ctx, "test-workflow", input)
	if err != nil {
		t.Fatalf("StartExecution failed: %v", err)
	}

	if result.WorkflowID != "test-workflow" {
		t.Errorf("expected workflow_id 'test-workflow', got %s", result.WorkflowID)
	}
	if result.State != workflow.StateSucceeded {
		t.Errorf("expected state 'succeeded', got %s", result.State)
	}

	_, err = p.StartExecution(ctx, "nonexistent", input)
	if err != workflow.ErrWorkflowNotFound {
		t.Errorf("expected ErrWorkflowNotFound, got %v", err)
	}
}

func TestProvider_GetExecutionStatus(t *testing.T) {
	p := New(zap.NewNop())
	ctx := context.Background()

	spec := &workflow.WorkflowSpec{
		WorkflowID:   "test-workflow",
		ProviderType: "mock",
		Name:         "Test Workflow",
		Definition:   json.RawMessage(`{"test": true}`),
	}

	_, err := p.CreateWorkflow(ctx, spec)
	if err != nil {
		t.Fatalf("CreateWorkflow failed: %v", err)
	}

	input := &workflow.ExecutionInput{
		ExecutionName: "test-execution",
		Input:         json.RawMessage(`{"param": "value"}`),
	}

	execResult, err := p.StartExecution(ctx, "test-workflow", input)
	if err != nil {
		t.Fatalf("StartExecution failed: %v", err)
	}

	status, err := p.GetExecutionStatus(ctx, execResult.ExecutionID)
	if err != nil {
		t.Fatalf("GetExecutionStatus failed: %v", err)
	}

	if status.ExecutionID != execResult.ExecutionID {
		t.Errorf("expected execution_id %s, got %s", execResult.ExecutionID, status.ExecutionID)
	}
	if status.State != workflow.StateSucceeded {
		t.Errorf("expected state 'succeeded', got %s", status.State)
	}
	if status.StopTime == nil {
		t.Error("expected stop_time to be set")
	}

	_, err = p.GetExecutionStatus(ctx, "nonexistent")
	if err != workflow.ErrExecutionNotFound {
		t.Errorf("expected ErrExecutionNotFound, got %v", err)
	}
}

func TestProvider_StopExecution(t *testing.T) {
	p := New(zap.NewNop())
	ctx := context.Background()

	spec := &workflow.WorkflowSpec{
		WorkflowID:   "test-workflow",
		ProviderType: "mock",
		Name:         "Test Workflow",
		Definition:   json.RawMessage(`{"test": true}`),
	}

	_, err := p.CreateWorkflow(ctx, spec)
	if err != nil {
		t.Fatalf("CreateWorkflow failed: %v", err)
	}

	input := &workflow.ExecutionInput{
		Input: json.RawMessage(`{"param": "value"}`),
	}

	execResult, err := p.StartExecution(ctx, "test-workflow", input)
	if err != nil {
		t.Fatalf("StartExecution failed: %v", err)
	}

	err = p.StopExecution(ctx, execResult.ExecutionID, "user requested")
	if err != nil {
		t.Fatalf("StopExecution failed: %v", err)
	}

	status, err := p.GetExecutionStatus(ctx, execResult.ExecutionID)
	if err != nil {
		t.Fatalf("GetExecutionStatus failed: %v", err)
	}

	if status.State != workflow.StateCancelled {
		t.Errorf("expected state 'cancelled', got %s", status.State)
	}

	err = p.StopExecution(ctx, "nonexistent", "test")
	if err != workflow.ErrExecutionNotFound {
		t.Errorf("expected ErrExecutionNotFound, got %v", err)
	}
}

func TestProvider_DeleteWorkflow(t *testing.T) {
	p := New(zap.NewNop())
	ctx := context.Background()

	spec := &workflow.WorkflowSpec{
		WorkflowID:   "test-workflow",
		ProviderType: "mock",
		Name:         "Test Workflow",
		Definition:   json.RawMessage(`{"test": true}`),
	}

	_, err := p.CreateWorkflow(ctx, spec)
	if err != nil {
		t.Fatalf("CreateWorkflow failed: %v", err)
	}

	err = p.DeleteWorkflow(ctx, "test-workflow")
	if err != nil {
		t.Fatalf("DeleteWorkflow failed: %v", err)
	}

	err = p.DeleteWorkflow(ctx, "test-workflow")
	if err != workflow.ErrWorkflowNotFound {
		t.Errorf("expected ErrWorkflowNotFound after deletion, got %v", err)
	}
}

func TestProvider_Invoke(t *testing.T) {
	p := New(zap.NewNop())
	ctx := context.Background()

	spec := &workflow.WorkflowSpec{
		WorkflowID:   "provision",
		ProviderType: "mock",
		Name:         "Provision",
		Definition:   json.RawMessage(`{"test": true}`),
	}

	_, err := p.CreateWorkflow(ctx, spec)
	if err != nil {
		t.Fatalf("CreateWorkflow failed: %v", err)
	}

	request := &workflow.ProvisionRequest{
		TenantID:      "tenant-123",
		TenantUUID:    "uuid-123",
		Operation:     "provision",
		DesiredImage:  "nginx:latest",
		DesiredConfig: map[string]interface{}{"replicas": "1"},
	}

	result, err := p.Invoke(ctx, "provision", request)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	expectedID := "tenant-tenant-123-provision"
	if result.ExecutionID != expectedID {
		t.Errorf("expected execution_id %s, got %s", expectedID, result.ExecutionID)
	}
	if result.State != workflow.StateSucceeded {
		t.Errorf("expected state succeeded, got %s", result.State)
	}
}

func TestProvider_GetWorkflowStatus(t *testing.T) {
	p := New(zap.NewNop())
	ctx := context.Background()

	spec := &workflow.WorkflowSpec{
		WorkflowID:   "provision",
		ProviderType: "mock",
		Name:         "Provision",
		Definition:   json.RawMessage(`{"test": true}`),
	}

	_, err := p.CreateWorkflow(ctx, spec)
	if err != nil {
		t.Fatalf("CreateWorkflow failed: %v", err)
	}

	request := &workflow.ProvisionRequest{
		TenantID:      "tenant-abc",
		TenantUUID:    "uuid-abc",
		Operation:     "provision",
		DesiredImage:  "nginx:latest",
		DesiredConfig: map[string]interface{}{"replicas": "1"},
	}

	result, err := p.Invoke(ctx, "provision", request)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	status, err := p.GetWorkflowStatus(ctx, result.ExecutionID)
	if err != nil {
		t.Fatalf("GetWorkflowStatus failed: %v", err)
	}

	if status.ExecutionID != result.ExecutionID {
		t.Errorf("expected execution_id %s, got %s", result.ExecutionID, status.ExecutionID)
	}
	if status.State != workflow.StateSucceeded {
		t.Errorf("expected state succeeded, got %s", status.State)
	}
	if len(status.Output) == 0 {
		t.Error("expected workflow status output to be set")
	}
}

func TestProvider_Validate(t *testing.T) {
	p := New(zap.NewNop())
	ctx := context.Background()

	tests := []struct {
		name    string
		spec    *workflow.WorkflowSpec
		wantErr bool
	}{
		{
			name: "valid JSON definition",
			spec: &workflow.WorkflowSpec{
				WorkflowID:   "test",
				ProviderType: "mock",
				Name:         "Test",
				Definition:   json.RawMessage(`{"test": true}`),
			},
			wantErr: false,
		},
		{
			name: "invalid JSON definition",
			spec: &workflow.WorkflowSpec{
				WorkflowID:   "test",
				ProviderType: "mock",
				Name:         "Test",
				Definition:   json.RawMessage(`{invalid json`),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.Validate(ctx, tt.spec)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProvider_ConcurrentOperations(t *testing.T) {
	p := New(zap.NewNop())
	ctx := context.Background()

	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			spec := &workflow.WorkflowSpec{
				WorkflowID:   "test-workflow",
				ProviderType: "mock",
				Name:         "Test Workflow",
				Definition:   json.RawMessage(`{"test": true}`),
			}

			_, err := p.CreateWorkflow(ctx, spec)
			if err != nil {
				t.Errorf("CreateWorkflow failed: %v", err)
			}
		}(i)
	}

	wg.Wait()

	p.mu.RLock()
	if len(p.workflows) != 1 {
		t.Errorf("expected 1 workflow after concurrent creates, got %d", len(p.workflows))
	}
	p.mu.RUnlock()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			input := &workflow.ExecutionInput{
				Input: json.RawMessage(`{"param": "value"}`),
			}

			_, err := p.StartExecution(ctx, "test-workflow", input)
			if err != nil {
				t.Errorf("StartExecution failed: %v", err)
			}
		}(i)
	}

	wg.Wait()

	p.mu.RLock()
	if len(p.executions) != numGoroutines {
		t.Errorf("expected %d executions after concurrent starts, got %d", numGoroutines, len(p.executions))
	}
	p.mu.RUnlock()
}
