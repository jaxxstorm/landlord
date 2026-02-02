package workflow

import (
	"encoding/json"
	"time"

	"github.com/jaxxstorm/landlord/internal/compute"
)

// CallbackPayload represents a callback from compute operation completion
type CallbackPayload struct {
	// ExecutionID is the compute execution ID that completed
	ExecutionID string `json:"execution_id"`

	// TenantID is the associated tenant
	TenantID string `json:"tenant_id"`

	// WorkflowExecutionID links back to the workflow
	WorkflowExecutionID string `json:"workflow_execution_id"`

	// OperationType is the compute operation type
	OperationType compute.ComputeOperationType `json:"operation_type"`

	// Status is the final status of the compute operation
	Status compute.ComputeExecutionStatus `json:"status"`

	// ResourceIDs contains created resources (for successful operations)
	ResourceIDs json.RawMessage `json:"resource_ids,omitempty"`

	// ErrorCode is populated for failed operations
	ErrorCode *string `json:"error_code,omitempty"`

	// ErrorMessage is populated for failed operations
	ErrorMessage *string `json:"error_message,omitempty"`

	// IsRetriable indicates if the operation can be retried
	IsRetriable bool `json:"is_retriable"`

	// Timestamp of callback generation
	Timestamp time.Time `json:"timestamp"`
}

// CallbackOptions controls callback delivery behavior
type CallbackOptions struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int

	// RetryBackoffSeconds is the base backoff duration
	RetryBackoffSeconds int

	// TimeoutSeconds is the timeout for callback delivery
	TimeoutSeconds int
}

// DefaultCallbackOptions returns reasonable defaults for callback delivery
func DefaultCallbackOptions() CallbackOptions {
	return CallbackOptions{
		MaxRetries:          3,
		RetryBackoffSeconds: 2,
		TimeoutSeconds:      30,
	}
}
