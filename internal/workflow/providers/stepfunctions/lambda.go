package stepfunctions

import (
	"context"
	"fmt"

	"github.com/jaxxstorm/landlord/internal/workflow"
	"go.uber.org/zap"
)

// LifecycleExecutor is the payload-driven lifecycle operation invoked by Lambda.
type LifecycleExecutor interface {
	Execute(context.Context, *workflow.ProvisionRequest) (*workflow.ExecutionStatus, error)
}

// LifecycleHandler adapts a lifecycle executor to the AWS Lambda handler shape.
type LifecycleHandler struct {
	executor LifecycleExecutor
	logger   *zap.Logger
}

// LifecycleRequest is the synchronous Lambda task payload emitted by the
// Step Functions definition. Request remains the original provider input.
type LifecycleRequest struct {
	Request             *workflow.ProvisionRequest `json:"request"`
	WorkflowExecutionID string                     `json:"workflow_execution_id"`
}

// ExecutionStatusReader retrieves the current Step Functions execution status.
type ExecutionStatusReader interface {
	GetExecutionStatus(context.Context, string) (*workflow.ExecutionStatus, error)
}

// CancellationChecker prevents new mutations after a Step Functions execution
// has been stopped. It is cooperative because a mutation already in progress
// cannot be atomically rolled back by Step Functions.
type CancellationChecker struct {
	statusReader ExecutionStatusReader
}

// NewCancellationChecker creates a guard backed by Step Functions execution status.
func NewCancellationChecker(statusReader ExecutionStatusReader) (*CancellationChecker, error) {
	if statusReader == nil {
		return nil, fmt.Errorf("execution status reader is required")
	}
	return &CancellationChecker{statusReader: statusReader}, nil
}

// BeforeMutation implements lifecycle.MutationGuard.
func (c *CancellationChecker) BeforeMutation(ctx context.Context, request *workflow.ProvisionRequest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	executionARN := trackedWorkflowExecutionID(request)
	if executionARN == "" {
		return nil
	}
	status, err := c.statusReader.GetExecutionStatus(ctx, executionARN)
	if err != nil {
		return fmt.Errorf("get Step Functions execution status: %w", err)
	}
	if status == nil {
		return fmt.Errorf("get Step Functions execution status: empty response")
	}
	if status.State == workflow.StateCancelled {
		return fmt.Errorf("%w: Step Functions execution %s was stopped", context.Canceled, executionARN)
	}
	return nil
}

// NewLifecycleHandler creates a Lambda handler for Step Functions lifecycle requests.
func NewLifecycleHandler(executor LifecycleExecutor) (*LifecycleHandler, error) {
	if executor == nil {
		return nil, fmt.Errorf("lifecycle executor is required")
	}
	return &LifecycleHandler{executor: executor, logger: zap.NewNop()}, nil
}

// WithLogger configures structured invocation logging.
func (h *LifecycleHandler) WithLogger(logger *zap.Logger) *LifecycleHandler {
	if logger != nil {
		h.logger = logger.With(zap.String("component", "step-functions-lambda"))
	}
	return h
}

// Handle executes one lifecycle request from a synchronous Step Functions task.
func (h *LifecycleHandler) Handle(ctx context.Context, input *LifecycleRequest) (*workflow.ExecutionStatus, error) {
	if input == nil || input.Request == nil {
		return nil, fmt.Errorf("Step Functions request is required")
	}
	request := input.Request
	if input.WorkflowExecutionID != "" {
		if request.Metadata == nil {
			request.Metadata = make(map[string]string)
		}
		request.Metadata["workflow_execution_id"] = input.WorkflowExecutionID
	}
	h.logger.Info("executing Step Functions lifecycle request",
		zap.String("tenant_id", request.TenantID),
		zap.String("execution_arn", trackedWorkflowExecutionID(request)),
	)
	status, err := h.executor.Execute(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("execute lifecycle request: %w", err)
	}
	return status, nil
}

func trackedWorkflowExecutionID(request *workflow.ProvisionRequest) string {
	if request.Metadata == nil {
		return ""
	}
	return request.Metadata["workflow_execution_id"]
}
