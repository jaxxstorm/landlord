package workflow

import (
	"context"
	"encoding/json"

	"github.com/jaxxstorm/landlord/internal/compute"
)

// Provider defines the interface for workflow providers
type Provider interface {
	// Name returns the unique provider identifier
	Name() string

	// Invoke starts a workflow execution using a simplified request payload
	Invoke(ctx context.Context, workflowID string, request *ProvisionRequest) (*ExecutionResult, error)

	// GetWorkflowStatus queries current execution status with simplified response
	GetWorkflowStatus(ctx context.Context, executionID string) (*WorkflowStatus, error)

	// CreateWorkflow creates a workflow definition
	CreateWorkflow(ctx context.Context, spec *WorkflowSpec) (*CreateWorkflowResult, error)

	// StartExecution starts a workflow execution
	// IMPORTANT: Implementations MUST be idempotent - if called multiple times with the same
	// ExecutionInput.ExecutionName, the provider should return the existing execution result
	// instead of creating a duplicate. This ensures that network retries and API retries don't
	// create duplicate workflow executions.
	StartExecution(ctx context.Context, workflowID string, input *ExecutionInput) (*ExecutionResult, error)

	// GetExecutionStatus queries current execution state
	GetExecutionStatus(ctx context.Context, executionID string) (*ExecutionStatus, error)

	// StopExecution stops a running execution
	StopExecution(ctx context.Context, executionID string, reason string) error

	// DeleteWorkflow removes a workflow definition
	DeleteWorkflow(ctx context.Context, workflowID string) error

	// Validate performs provider-specific validation
	Validate(ctx context.Context, spec *WorkflowSpec) error

	// PostComputeCallback sends a compute execution completion callback to an execution
	// Used to notify running workflows of compute provisioning completion
	PostComputeCallback(ctx context.Context, executionID string, payload *compute.CallbackPayload, opts *compute.CallbackOptions) error
}

// ProvisionRequest is a simplified execution request for workflow providers
type ProvisionRequest struct {
	TenantID        string                 `json:"tenant_id"`
	TenantUUID      string                 `json:"tenant_uuid,omitempty"`
	Operation       string                 `json:"operation,omitempty"`
	DesiredConfig   map[string]interface{} `json:"desired_config,omitempty"`
	ComputeProvider string                 `json:"compute_provider,omitempty"`
	APIBaseURL      string                 `json:"api_base_url,omitempty"`
	Metadata        map[string]string      `json:"metadata,omitempty"` // Metadata like config_hash
}

// WorkflowStatus is a simplified execution status response
type WorkflowStatus struct {
	ExecutionID string          `json:"execution_id"`
	State       ExecutionState  `json:"state"`
	Output      json.RawMessage `json:"output,omitempty"`
	Error       *ExecutionError `json:"error,omitempty"`
}
