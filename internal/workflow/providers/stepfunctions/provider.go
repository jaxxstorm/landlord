package stepfunctions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/jaxxstorm/landlord/internal/compute"
	"github.com/jaxxstorm/landlord/internal/workflow"
)

type Provider struct {
	region  string
	roleARN string
	acct    string
	logger  *zap.Logger
}

// Config holds configuration for the Step Functions provider
type Config struct {
	Region  string
	RoleARN string
}

// New creates a new Step Functions provider (compatible signature)
func New(ctx context.Context, cfg Config, logger *zap.Logger) (*Provider, error) {
	if cfg.RoleARN == "" {
		return nil, fmt.Errorf("role ARN is required")
	}
	if cfg.Region == "" {
		return nil, fmt.Errorf("region is required")
	}
	// Minimal/no-op provider for tests and local runs
	return &Provider{region: cfg.Region, roleARN: cfg.RoleARN, logger: logger.With(zap.String("provider", "step-functions"))}, nil
}

func (p *Provider) Name() string { return "step-functions" }

// Invoke starts a workflow execution using a simplified request payload
func (p *Provider) Invoke(ctx context.Context, workflowID string, request *workflow.ProvisionRequest) (*workflow.ExecutionResult, error) {
	if request == nil {
		return nil, fmt.Errorf("provision request is required")
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	executionName := fmt.Sprintf("tenant-%s-%s", request.TenantID, workflowID)
	input := &workflow.ExecutionInput{
		ExecutionName: executionName,
		Input:         payload,
		Tags: map[string]string{
			"tenant_id": request.TenantID,
		},
		TriggerSource: "reconciler",
	}

	return p.StartExecution(ctx, workflowID, input)
}

// GetWorkflowStatus returns a simplified workflow status for an execution
func (p *Provider) GetWorkflowStatus(ctx context.Context, executionID string) (*workflow.WorkflowStatus, error) {
	status, err := p.GetExecutionStatus(ctx, executionID)
	if err != nil {
		return nil, err
	}

	if status == nil {
		return nil, fmt.Errorf("execution status is nil")
	}

	return &workflow.WorkflowStatus{
		ExecutionID: status.ExecutionID,
		State:       status.State,
		Output:      status.Output,
		Error:       status.Error,
	}, nil
}

func (p *Provider) CreateWorkflow(ctx context.Context, spec *workflow.WorkflowSpec) (*workflow.CreateWorkflowResult, error) {
	arn := buildStateMachineARN(spec.WorkflowID, p.region, p.acct)
	time.Sleep(10 * time.Millisecond)
	return &workflow.CreateWorkflowResult{
		WorkflowID:   spec.WorkflowID,
		ProviderType: "step-functions",
		ResourceIDs:  map[string]string{"arn": arn},
		CreatedAt:    time.Now(),
		Message:      "created (simulated)",
	}, nil
}

func (p *Provider) DeleteWorkflow(ctx context.Context, workflowARN string) error {
	time.Sleep(5 * time.Millisecond)
	return nil
}

func (p *Provider) StartExecution(ctx context.Context, workflowID string, input *workflow.ExecutionInput) (*workflow.ExecutionResult, error) {
	b, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshal input: %w", err)
	}
	_ = b
	name := fmt.Sprintf("exec-%d", time.Now().UnixNano())
	executionARN := buildExecutionARN(workflowID, name, p.region, p.acct)
	return &workflow.ExecutionResult{
		ExecutionID:  executionARN,
		WorkflowID:   workflowID,
		ProviderType: "step-functions",
		State:        workflow.StateRunning,
		StartedAt:    time.Now(),
		Message:      "started (simulated)",
	}, nil
}

func (p *Provider) StopExecution(ctx context.Context, executionID string, reason string) error {
	_ = reason
	time.Sleep(5 * time.Millisecond)
	return nil
}

func (p *Provider) GetExecutionStatus(ctx context.Context, executionARN string) (*workflow.ExecutionStatus, error) {
	return &workflow.ExecutionStatus{
		ExecutionID:  executionARN,
		WorkflowID:   "",
		ProviderType: "step-functions",
		State:        workflow.StateRunning,
		StartTime:    time.Now(),
	}, nil
}

func (p *Provider) Validate(ctx context.Context, spec *workflow.WorkflowSpec) error {
	// Basic validation: ensure workflow ID and definition exist
	if spec == nil || spec.WorkflowID == "" || len(spec.Definition) == 0 {
		return workflow.ErrInvalidSpec
	}
	return nil
}

// PostComputeCallback sends a compute execution callback to a running Step Functions execution
func (p *Provider) PostComputeCallback(ctx context.Context, executionID string, payload *compute.CallbackPayload, opts *compute.CallbackOptions) error {
	_ = executionID
	_ = opts
	// Marshal the callback payload to JSON for logging/debug
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal callback payload: %w", err)
	}
	p.logger.Debug("posting compute callback",
		zap.String("execution_id", executionID),
		zap.String("tenant_id", payload.TenantID),
		zap.ByteString("payload", payloadJSON),
	)
	return nil
}
