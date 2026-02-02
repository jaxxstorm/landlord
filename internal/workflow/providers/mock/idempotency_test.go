package mock

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"github.com/jaxxstorm/landlord/internal/workflow"
)

// TestProviderIdempotency tests that duplicate execution names return existing execution
func TestProviderIdempotency(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	provider := New(logger)

	spec := &workflow.WorkflowSpec{
		WorkflowID:   "test-workflow",
		Name:         "test-workflow",
		ProviderType: "mock",
	}

	// Create workflow first
	_, err := provider.CreateWorkflow(context.Background(), spec)
	if err != nil {
		t.Fatalf("CreateWorkflow failed: %v", err)
	}

	// First execution
	input1 := &workflow.ExecutionInput{
		ExecutionName: "test-execution-idempotent",
		Input:         []byte(`{"test": "data"}`),
	}

	result1, err := provider.StartExecution(context.Background(), "test-workflow", input1)
	if err != nil {
		t.Fatalf("First StartExecution failed: %v", err)
	}

	// Second execution with same name - should return existing
	input2 := &workflow.ExecutionInput{
		ExecutionName: "test-execution-idempotent",
		Input:         []byte(`{"test": "data"}`),
	}

	result2, err := provider.StartExecution(context.Background(), "test-workflow", input2)
	if err != nil {
		t.Fatalf("Second StartExecution failed: %v", err)
	}

	// Should return same execution ID
	if result1.ExecutionID != result2.ExecutionID {
		t.Errorf("expected same execution ID, got %s and %s", result1.ExecutionID, result2.ExecutionID)
	}

	// Should indicate it's the same execution
	if result2.Message != "execution already started (idempotent result)" {
		t.Errorf("expected idempotent message, got %s", result2.Message)
	}
}

// TestProviderDifferentNamesCreateDifferentExecutions tests different names create different executions
func TestProviderDifferentNamesCreateDifferentExecutions(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	provider := New(logger)

	spec := &workflow.WorkflowSpec{
		WorkflowID:   "test-workflow",
		Name:         "test-workflow",
		ProviderType: "mock",
	}

	// Create workflow first
	_, err := provider.CreateWorkflow(context.Background(), spec)
	if err != nil {
		t.Fatalf("CreateWorkflow failed: %v", err)
	}

	// First execution
	input1 := &workflow.ExecutionInput{
		ExecutionName: "test-execution-1",
		Input:         []byte(`{"test": "data"}`),
	}

	result1, err := provider.StartExecution(context.Background(), "test-workflow", input1)
	if err != nil {
		t.Fatalf("First StartExecution failed: %v", err)
	}

	// Second execution with different name
	input2 := &workflow.ExecutionInput{
		ExecutionName: "test-execution-2",
		Input:         []byte(`{"test": "data"}`),
	}

	result2, err := provider.StartExecution(context.Background(), "test-workflow", input2)
	if err != nil {
		t.Fatalf("Second StartExecution failed: %v", err)
	}

	// Should return different execution IDs
	if result1.ExecutionID == result2.ExecutionID {
		t.Errorf("expected different execution IDs, got same: %s", result1.ExecutionID)
	}
}

// TestProviderWithoutExecutionName creates unique execution each time
func TestProviderWithoutExecutionName(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	provider := New(logger)

	spec := &workflow.WorkflowSpec{
		WorkflowID:   "test-workflow",
		Name:         "test-workflow",
		ProviderType: "mock",
	}

	// Create workflow first
	_, err := provider.CreateWorkflow(context.Background(), spec)
	if err != nil {
		t.Fatalf("CreateWorkflow failed: %v", err)
	}

	// Two executions without names - should be different
	input1 := &workflow.ExecutionInput{
		Input: []byte(`{"test": "data"}`),
	}

	result1, err := provider.StartExecution(context.Background(), "test-workflow", input1)
	if err != nil {
		t.Fatalf("First StartExecution failed: %v", err)
	}

	input2 := &workflow.ExecutionInput{
		Input: []byte(`{"test": "data"}`),
	}

	result2, err := provider.StartExecution(context.Background(), "test-workflow", input2)
	if err != nil {
		t.Fatalf("Second StartExecution failed: %v", err)
	}

	// Should return different execution IDs
	if result1.ExecutionID == result2.ExecutionID {
		t.Errorf("expected different execution IDs, got same: %s", result1.ExecutionID)
	}
}

// TestProviderStatusAfterIdempotentTrigger tests status is correct after idempotent trigger
func TestProviderStatusAfterIdempotentTrigger(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	provider := New(logger)

	spec := &workflow.WorkflowSpec{
		WorkflowID:   "test-workflow",
		Name:         "test-workflow",
		ProviderType: "mock",
	}

	// Create workflow
	_, err := provider.CreateWorkflow(context.Background(), spec)
	if err != nil {
		t.Fatalf("CreateWorkflow failed: %v", err)
	}

	// First execution
	input := &workflow.ExecutionInput{
		ExecutionName: "test-idempotent-status",
		Input:         []byte(`{"test": "data"}`),
	}

	result1, err := provider.StartExecution(context.Background(), "test-workflow", input)
	if err != nil {
		t.Fatalf("First StartExecution failed: %v", err)
	}

	// Get status after first execution
	status1, err := provider.GetExecutionStatus(context.Background(), result1.ExecutionID)
	if err != nil {
		t.Fatalf("GetExecutionStatus failed: %v", err)
	}

	// Second execution (idempotent)
	result2, err := provider.StartExecution(context.Background(), "test-workflow", input)
	if err != nil {
		t.Fatalf("Second StartExecution failed: %v", err)
	}

	// Get status after second execution - should be same
	status2, err := provider.GetExecutionStatus(context.Background(), result2.ExecutionID)
	if err != nil {
		t.Fatalf("GetExecutionStatus failed: %v", err)
	}

	// States should match (showing it's the same execution)
	if status1.State != status2.State {
		t.Errorf("expected same state after idempotent trigger, got %s and %s", status1.State, status2.State)
	}
}
