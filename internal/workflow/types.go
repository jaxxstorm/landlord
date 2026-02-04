package workflow

import (
	"encoding/json"
	"time"
)

// WorkflowSpec defines a workflow to be created
type WorkflowSpec struct {
	WorkflowID     string            `json:"workflow_id"`
	ProviderType   string            `json:"provider_type"`
	Name           string            `json:"name"`
	Description    string            `json:"description,omitempty"`
	Definition     json.RawMessage   `json:"definition"`
	Timeout        time.Duration     `json:"timeout,omitempty"`
	RetryPolicy    *RetryPolicy      `json:"retry_policy,omitempty"`
	Tags           map[string]string `json:"tags,omitempty"`
	ProviderConfig json.RawMessage   `json:"provider_config,omitempty"`
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxAttempts     int           `json:"max_attempts"`
	BackoffRate     float64       `json:"backoff_rate"`
	InitialInterval time.Duration `json:"initial_interval"`
	MaxInterval     time.Duration `json:"max_interval"`
}

// ExecutionInput is input for starting a workflow
type ExecutionInput struct {
	ExecutionName string            `json:"execution_name,omitempty"`
	Input         json.RawMessage   `json:"input"`
	Tags          map[string]string `json:"tags,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"` // Metadata like config_hash for change detection
	TriggerSource string            `json:"trigger_source,omitempty"` // "api" or "controller"
}

// CreateWorkflowResult is result from creating a workflow
type CreateWorkflowResult struct {
	WorkflowID   string            `json:"workflow_id"`
	ProviderType string            `json:"provider_type"`
	ResourceIDs  map[string]string `json:"resource_ids"`
	CreatedAt    time.Time         `json:"created_at"`
	Message      string            `json:"message,omitempty"`
}

// ExecutionResult is result from starting an execution
type ExecutionResult struct {
	ExecutionID  string         `json:"execution_id"`
	WorkflowID   string         `json:"workflow_id"`
	ProviderType string         `json:"provider_type"`
	State        ExecutionState `json:"state"`
	StartedAt    time.Time      `json:"started_at"`
	Message      string         `json:"message,omitempty"`
}

// ExecutionStatus represents current status of a workflow execution
type ExecutionStatus struct {
	ExecutionID  string            `json:"execution_id"`
	WorkflowID   string            `json:"workflow_id"`
	ProviderType string            `json:"provider_type"`
	State        ExecutionState    `json:"state"`
	StartTime    time.Time         `json:"start_time"`
	StopTime     *time.Time        `json:"stop_time,omitempty"`
	Input        json.RawMessage   `json:"input"`
	Output       json.RawMessage   `json:"output,omitempty"`
	Error        *ExecutionError   `json:"error,omitempty"`
	History      []ExecutionEvent  `json:"history,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// ExecutionState represents execution state
type ExecutionState string

const (
	StatePending   ExecutionState = "pending"
	StateRunning   ExecutionState = "running"
	StateSucceeded ExecutionState = "succeeded"
	StateFailed    ExecutionState = "failed"
	StateTimedOut  ExecutionState = "timed_out"
	StateCancelled ExecutionState = "cancelled"
)

// WorkflowSubState represents provider-agnostic execution sub-state
type WorkflowSubState string

const (
	SubStateRunning    WorkflowSubState = "running"
	SubStateWaiting    WorkflowSubState = "waiting"
	SubStateBackingOff WorkflowSubState = "backing-off"
	SubStateError      WorkflowSubState = "error"
	SubStateSucceeded  WorkflowSubState = "succeeded"
	SubStateFailed     WorkflowSubState = "failed"
)

// MapExecutionStateToSubState maps execution state to canonical workflow sub-state
func MapExecutionStateToSubState(state ExecutionState) WorkflowSubState {
	switch state {
	case StateSucceeded:
		return SubStateSucceeded
	case StateFailed, StateTimedOut, StateCancelled:
		return SubStateFailed
	case StatePending:
		return SubStateWaiting
	case StateRunning:
		return SubStateRunning
	default:
		return SubStateRunning
	}
}

// ExecutionError contains error details for failed executions
type ExecutionError struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Cause    string `json:"cause,omitempty"`
	FailedAt string `json:"failed_at,omitempty"`
}

// ExecutionEvent is a single event in execution history
type ExecutionEvent struct {
	Timestamp time.Time       `json:"timestamp"`
	Type      string          `json:"type"`
	Details   json.RawMessage `json:"details,omitempty"`
}
