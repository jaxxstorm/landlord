# Specification: Workflow Types

## Overview

This specification defines all data structures for the workflow provisioning framework: workflow specifications, execution inputs, results, and status types.

## Workflow Specification Types

### WorkflowSpec

Defines a workflow to be created.

```go
type WorkflowSpec struct {
    // WorkflowID uniquely identifies this workflow
    // Format: lowercase alphanumeric with hyphens, 1-128 chars
    // Example: "tenant-provisioning", "data-migration-v2"
    WorkflowID string `json:"workflow_id"`
    
    // ProviderType specifies which workflow provider to use
    // Must match a registered provider name
    ProviderType string `json:"provider_type"`
    
    // Name is human-readable workflow name
    Name string `json:"name"`
    
    // Description explains what this workflow does
    Description string `json:"description,omitempty"`
    
    // Definition contains provider-specific workflow definition
    // Step Functions: JSON state machine
    // Temporal: Serialized workflow code reference
    Definition json.RawMessage `json:"definition"`
    
    // Timeout is maximum execution duration
    // Zero means use provider default
    Timeout time.Duration `json:"timeout,omitempty"`
    
    // RetryPolicy defines retry behavior for the workflow
    RetryPolicy *RetryPolicy `json:"retry_policy,omitempty"`
    
    // Tags are key-value metadata
    Tags map[string]string `json:"tags,omitempty"`
    
    // ProviderConfig contains provider-specific configuration
    // Validated by the specific provider
    ProviderConfig json.RawMessage `json:"provider_config,omitempty"`
}
```

**Validation Rules**:
- WorkflowID: `^[a-z0-9-]{1,128}$`
- ProviderType: non-empty
- Name: non-empty, max 256 chars
- Definition: non-empty valid JSON
- Timeout: if set, must be > 0

### RetryPolicy

Defines retry behavior for workflow execution.

```go
type RetryPolicy struct {
    // MaxAttempts is maximum number of execution attempts
    // 0 means no retries (single attempt)
    MaxAttempts int `json:"max_attempts"`
    
    // BackoffRate multiplier for exponential backoff
    // Must be >= 1.0
    BackoffRate float64 `json:"backoff_rate"`
    
    // InitialInterval is the initial retry delay
    InitialInterval time.Duration `json:"initial_interval"`
    
    // MaxInterval is the maximum retry delay
    MaxInterval time.Duration `json:"max_interval"`
}
```

## Execution Types

### ExecutionInput

Input for starting a workflow execution.

```go
type ExecutionInput struct {
    // ExecutionName uniquely identifies this execution
    // Format: alphanumeric with hyphens/underscores, 1-128 chars
    // If empty, provider generates unique ID
    ExecutionName string `json:"execution_name,omitempty"`
    
    // Input is JSON data passed to workflow
    // Must be valid JSON object or array
    Input json.RawMessage `json:"input"`
    
    // Tags are key-value metadata for this execution
    Tags map[string]string `json:"tags,omitempty"`
}
```

**Validation Rules**:
- ExecutionName: if set, `^[a-zA-Z0-9-_]{1,128}$`
- Input: non-empty valid JSON

## Result Types

### CreateWorkflowResult

Result from creating a workflow.

```go
type CreateWorkflowResult struct {
    // WorkflowID that was created
    WorkflowID string `json:"workflow_id"`
    
    // ProviderType used
    ProviderType string `json:"provider_type"`
    
    // ResourceIDs maps resource types to provider-specific IDs
    // Examples:
    //   Step Functions: {"arn": "arn:aws:states:..."}
    //   Temporal: {"namespace": "default", "workflow_type": "TenantProvisioning"}
    ResourceIDs map[string]string `json:"resource_ids"`
    
    // CreatedAt timestamp
    CreatedAt time.Time `json:"created_at"`
    
    // Message provides additional details
    Message string `json:"message,omitempty"`
}
```

### ExecutionResult

Result from starting an execution.

```go
type ExecutionResult struct {
    // ExecutionID uniquely identifies this execution
    ExecutionID string `json:"execution_id"`
    
    // WorkflowID that is executing
    WorkflowID string `json:"workflow_id"`
    
    // ProviderType used
    ProviderType string `json:"provider_type"`
    
    // State of execution (typically "running" or "pending")
    State ExecutionState `json:"state"`
    
    // StartedAt timestamp
    StartedAt time.Time `json:"started_at"`
    
    // Message provides additional details
    Message string `json:"message,omitempty"`
}
```

## Status Types

### ExecutionStatus

Current status of a workflow execution.

```go
type ExecutionStatus struct {
    // ExecutionID being queried
    ExecutionID string `json:"execution_id"`
    
    // WorkflowID that is executing
    WorkflowID string `json:"workflow_id"`
    
    // ProviderType managing this execution
    ProviderType string `json:"provider_type"`
    
    // State of execution
    State ExecutionState `json:"state"`
    
    // StartTime when execution started
    StartTime time.Time `json:"start_time"`
    
    // StopTime when execution finished (nil if still running)
    StopTime *time.Time `json:"stop_time,omitempty"`
    
    // Input provided to the execution
    Input json.RawMessage `json:"input"`
    
    // Output from the execution (only if succeeded)
    Output json.RawMessage `json:"output,omitempty"`
    
    // Error details if execution failed
    Error *ExecutionError `json:"error,omitempty"`
    
    // History of execution events
    History []ExecutionEvent `json:"history,omitempty"`
    
    // Metadata provider-specific status information
    Metadata map[string]string `json:"metadata,omitempty"`
}
```

### ExecutionState

Represents current execution state.

```go
type ExecutionState string

const (
    // StatePending: execution queued but not started
    StatePending ExecutionState = "pending"
    
    // StateRunning: execution in progress
    StateRunning ExecutionState = "running"
    
    // StateSucceeded: execution completed successfully
    StateSucceeded ExecutionState = "succeeded"
    
    // StateFailed: execution failed with error
    StateFailed ExecutionState = "failed"
    
    // StateTimedOut: execution exceeded timeout
    StateTimedOut ExecutionState = "timed_out"
    
    // StateCancelled: execution was stopped
    StateCancelled ExecutionState = "cancelled"
)
```

**State Transitions**:
```
pending → running → succeeded
               ├──→ failed
               ├──→ timed_out
               └──→ cancelled
```

### ExecutionError

Error details for failed executions.

```go
type ExecutionError struct {
    // Code is provider-specific error code
    Code string `json:"code"`
    
    // Message describes the error
    Message string `json:"message"`
    
    // Cause is the underlying error (if available)
    Cause string `json:"cause,omitempty"`
    
    // FailedAt which step/state failed
    FailedAt string `json:"failed_at,omitempty"`
}
```

### ExecutionEvent

Single event in execution history.

```go
type ExecutionEvent struct {
    // Timestamp when event occurred
    Timestamp time.Time `json:"timestamp"`
    
    // Type of event
    // Examples: "ExecutionStarted", "StateEntered", "StateExited", "ExecutionSucceeded"
    Type string `json:"type"`
    
    // Details provider-specific event data
    Details json.RawMessage `json:"details,omitempty"`
}
```

## Validation Patterns

### Workflow ID Pattern
```go
const WorkflowIDPattern = `^[a-z0-9-]{1,128}$`
```

### Execution Name Pattern
```go
const ExecutionNamePattern = `^[a-zA-Z0-9-_]{1,128}$`
```

## JSON Examples

### WorkflowSpec Example

```json
{
  "workflow_id": "tenant-provisioning",
  "provider_type": "step-functions",
  "name": "Tenant Provisioning Workflow",
  "description": "Provisions compute and database for new tenant",
  "definition": {
    "Comment": "Tenant provisioning state machine",
    "StartAt": "ProvisionCompute",
    "States": {
      "ProvisionCompute": {
        "Type": "Task",
        "Resource": "arn:aws:lambda:...",
        "Next": "ConfigureDatabase"
      },
      "ConfigureDatabase": {
        "Type": "Task",
        "Resource": "arn:aws:lambda:...",
        "End": true
      }
    }
  },
  "timeout": "30m",
  "retry_policy": {
    "max_attempts": 3,
    "backoff_rate": 2.0,
    "initial_interval": "1s",
    "max_interval": "60s"
  },
  "tags": {
    "environment": "production",
    "team": "platform"
  }
}
```

### ExecutionStatus Example

```json
{
  "execution_id": "exec-abc123",
  "workflow_id": "tenant-provisioning",
  "provider_type": "step-functions",
  "state": "succeeded",
  "start_time": "2026-01-30T10:00:00Z",
  "stop_time": "2026-01-30T10:05:00Z",
  "input": {
    "tenant_id": "tenant-123",
    "image": "app:v1.0.0"
  },
  "output": {
    "compute_id": "ecs-task-123",
    "database_id": "rds-instance-456"
  },
  "history": [
    {
      "timestamp": "2026-01-30T10:00:00Z",
      "type": "ExecutionStarted",
      "details": {}
    },
    {
      "timestamp": "2026-01-30T10:02:00Z",
      "type": "StateEntered",
      "details": {"state": "ProvisionCompute"}
    },
    {
      "timestamp": "2026-01-30T10:04:00Z",
      "type": "StateEntered",
      "details": {"state": "ConfigureDatabase"}
    },
    {
      "timestamp": "2026-01-30T10:05:00Z",
      "type": "ExecutionSucceeded",
      "details": {}
    }
  ]
}
```

## Type Size Limits

- WorkflowID: max 128 bytes
- ExecutionName: max 128 bytes
- Name: max 256 bytes
- Description: max 1024 bytes
- Definition: max 256 KB
- Input: max 256 KB
- Output: max 256 KB
- Tags: max 50 per resource
- Tag key/value: max 128 bytes each
