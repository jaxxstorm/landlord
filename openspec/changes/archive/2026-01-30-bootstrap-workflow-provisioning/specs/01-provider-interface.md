# Specification: Workflow Provider Interface

## Overview

This specification defines the `Provider` interface that all workflow providers must implement. The interface abstracts workflow lifecycle operations across different workflow engines (AWS Step Functions, Temporal, etc.).

## Provider Interface

```go
package workflow

import "context"

type Provider interface {
    // Name returns the unique provider identifier
    // Must be lowercase, alphanumeric with hyphens
    // Examples: "mock", "step-functions", "temporal"
    Name() string
    
    // CreateWorkflow creates a workflow definition
    // Must be idempotent - calling twice with same WorkflowID should succeed
    // Returns CreateWorkflowResult with provider-specific resource IDs
    CreateWorkflow(ctx context.Context, spec *WorkflowSpec) (*CreateWorkflowResult, error)
    
    // StartExecution starts a workflow execution
    // Returns immediately with ExecutionResult containing execution ID
    // Execution runs asynchronously - use GetExecutionStatus to monitor
    StartExecution(ctx context.Context, workflowID string, input *ExecutionInput) (*ExecutionResult, error)
    
    // GetExecutionStatus queries current execution state
    // Returns ExecutionStatus with current state, input, output, and history
    GetExecutionStatus(ctx context.Context, executionID string) (*ExecutionStatus, error)
    
    // StopExecution stops a running execution
    // Idempotent - calling on already-stopped execution succeeds
    // Reason is logged for audit trail
    StopExecution(ctx context.Context, executionID string, reason string) error
    
    // DeleteWorkflow removes a workflow definition
    // Idempotent - deleting non-existent workflow succeeds
    // May fail if there are running executions (provider-specific)
    DeleteWorkflow(ctx context.Context, workflowID string) error
    
    // Validate performs provider-specific validation on a workflow spec
    // Returns error describing validation failures
    // Manager calls this before CreateWorkflow
    Validate(ctx context.Context, spec *WorkflowSpec) error
}
```

## Method Contracts

### Name()

**Purpose**: Returns the unique identifier for this provider.

**Requirements**:
- MUST return a non-empty string
- MUST be lowercase alphanumeric with hyphens only
- MUST be unique across all registered providers
- SHOULD be human-readable (e.g., "step-functions", not "sfn1")

**Examples**:
```go
func (p *StepFunctionsProvider) Name() string {
    return "step-functions"
}

func (p *TemporalProvider) Name() string {
    return "temporal"
}
```

---

### CreateWorkflow()

**Purpose**: Creates a workflow definition in the provider's workflow engine.

**Parameters**:
- `ctx`: Context for cancellation and timeout
- `spec`: Complete workflow specification

**Returns**:
- `CreateWorkflowResult`: Contains provider-specific resource IDs
- `error`: Any creation errors

**Requirements**:
- MUST be idempotent - calling twice with same WorkflowID succeeds (no-op on second call)
- MUST validate spec.ProviderConfig if used
- MUST parse and validate spec.Definition according to provider's format
- MUST respect ctx cancellation
- MUST NOT start execution (use StartExecution for that)
- SHOULD return quickly (< 10 seconds)

**Errors**:
- ErrInvalidSpec: If spec.Definition is invalid for this provider
- Context errors: If ctx is cancelled or times out
- Provider-specific errors: Wrapped with context

**Example**:
```go
result, err := provider.CreateWorkflow(ctx, &workflow.WorkflowSpec{
    WorkflowID:   "tenant-provisioning",
    ProviderType: "step-functions",
    Name:         "Tenant Provisioning Workflow",
    Definition:   []byte(`{"StartAt": "Provision", ...}`),
    Timeout:      30 * time.Minute,
})
// Returns: CreateWorkflowResult{WorkflowID: "tenant-provisioning", ResourceIDs: {"arn": "arn:aws:..."}}
```

---

### StartExecution()

**Purpose**: Starts a workflow execution with provided input.

**Parameters**:
- `ctx`: Context for cancellation and timeout
- `workflowID`: ID of workflow to execute (from CreateWorkflow)
- `input`: Execution input including execution name and data

**Returns**:
- `ExecutionResult`: Contains execution ID and initial state
- `error`: Any start errors

**Requirements**:
- MUST return immediately (non-blocking)
- MUST return unique execution ID
- MUST validate workflowID exists
- MUST validate input.Input is valid JSON
- MUST set initial state to "running" or "pending"
- Execution continues asynchronously after return
- SHOULD generate unique execution ID if input.ExecutionName is empty

**Errors**:
- ErrWorkflowNotFound: If workflowID doesn't exist
- ErrInvalidSpec: If input.Input is invalid
- Context errors: If ctx is cancelled or times out
- Provider-specific errors: Wrapped with context

**Example**:
```go
result, err := provider.StartExecution(ctx, "tenant-provisioning", &workflow.ExecutionInput{
    ExecutionName: "tenant-123-provision",
    Input:         []byte(`{"tenantID": "tenant-123", "image": "app:v1"}`),
})
// Returns: ExecutionResult{ExecutionID: "exec-abc123", State: "running"}
```

---

### GetExecutionStatus()

**Purpose**: Queries current state of a workflow execution.

**Parameters**:
- `ctx`: Context for cancellation and timeout
- `executionID`: ID of execution to query (from StartExecution)

**Returns**:
- `ExecutionStatus`: Current execution state, input, output, history
- `error`: Any query errors

**Requirements**:
- MUST return current state accurately
- MUST include execution history if available
- MUST include output if execution succeeded
- MUST include error if execution failed
- SHOULD return quickly (< 5 seconds)
- Terminal states (succeeded, failed, cancelled, timed_out) MUST NOT change

**Errors**:
- ErrExecutionNotFound: If executionID doesn't exist
- Context errors: If ctx is cancelled or times out
- Provider-specific errors: Wrapped with context

**Example**:
```go
status, err := provider.GetExecutionStatus(ctx, "exec-abc123")
// Returns: ExecutionStatus{
//   ExecutionID: "exec-abc123",
//   State:       "succeeded",
//   StartTime:   <time>,
//   StopTime:    <time>,
//   Output:      []byte(`{"result": "success"}`),
// }
```

---

### StopExecution()

**Purpose**: Stops a running workflow execution.

**Parameters**:
- `ctx`: Context for cancellation and timeout
- `executionID`: ID of execution to stop
- `reason`: Human-readable reason for stopping (for audit)

**Returns**:
- `error`: Any stop errors

**Requirements**:
- MUST be idempotent - stopping already-stopped execution succeeds
- MUST log reason for audit trail
- MUST transition execution to "cancelled" state
- SHOULD stop gracefully (allow cleanup)
- SHOULD return quickly (< 5 seconds)

**Errors**:
- ErrExecutionNotFound: If executionID doesn't exist
- Context errors: If ctx is cancelled or times out
- Provider-specific errors: Wrapped with context

**Example**:
```go
err := provider.StopExecution(ctx, "exec-abc123", "User requested cancellation")
```

---

### DeleteWorkflow()

**Purpose**: Removes a workflow definition.

**Parameters**:
- `ctx`: Context for cancellation and timeout
- `workflowID`: ID of workflow to delete

**Returns**:
- `error`: Any deletion errors

**Requirements**:
- MUST be idempotent - deleting non-existent workflow succeeds
- MAY fail if there are running executions (provider-specific)
- MUST remove workflow definition from provider
- SHOULD return quickly (< 5 seconds)

**Errors**:
- Context errors: If ctx is cancelled or times out
- Provider-specific errors: Wrapped with context (e.g., "workflow has running executions")

**Example**:
```go
err := provider.DeleteWorkflow(ctx, "tenant-provisioning")
```

---

### Validate()

**Purpose**: Performs provider-specific validation on a workflow spec.

**Parameters**:
- `ctx`: Context for cancellation and timeout
- `spec`: Workflow specification to validate

**Returns**:
- `error`: Validation error describing what's invalid, or nil if valid

**Requirements**:
- MUST validate spec.Definition according to provider's format
- MUST validate spec.ProviderConfig if used
- MUST return descriptive error messages
- SHOULD validate quickly (< 1 second)
- MUST NOT create any resources (validation only)

**Errors**:
- ErrInvalidSpec: If any validation fails (wrapped with details)
- Context errors: If ctx is cancelled or times out

**Example**:
```go
err := provider.Validate(ctx, &workflow.WorkflowSpec{
    Definition: []byte(`{invalid json`),
})
// Returns: error wrapping ErrInvalidSpec with "invalid JSON in definition: ..."
```

## Implementation Guidelines

1. **Idempotency**: CreateWorkflow, DeleteWorkflow, StopExecution must be idempotent
2. **Context Handling**: All methods must respect ctx cancellation and timeout
3. **Error Wrapping**: Wrap provider-specific errors with context using fmt.Errorf
4. **Logging**: Log all operations for observability
5. **Thread Safety**: Providers should be safe for concurrent use
6. **Resource Cleanup**: DeleteWorkflow should clean up all provider resources

## Testing Requirements

All providers must have tests covering:
- Happy path for all methods
- Idempotency of CreateWorkflow, DeleteWorkflow, StopExecution
- Error cases: not found, invalid input, timeout
- Concurrent access to provider methods
- Context cancellation handling
