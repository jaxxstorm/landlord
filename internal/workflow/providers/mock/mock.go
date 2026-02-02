package mock

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jaxxstorm/landlord/internal/compute"
	"github.com/jaxxstorm/landlord/internal/workflow"
	"go.uber.org/zap"
)

// Provider is an in-memory mock workflow provider for testing
type Provider struct {
	mu          sync.RWMutex
	workflows   map[string]*workflowData
	executions  map[string]*executionData
	logger      *zap.Logger
	execCounter uint64
}

type workflowData struct {
	spec      *workflow.WorkflowSpec
	createdAt time.Time
}

type executionData struct {
	id         string
	workflowID string
	input      *workflow.ExecutionInput
	status     *workflow.ExecutionStatus
}

// New creates a new mock provider
func New(logger *zap.Logger) *Provider {
	return &Provider{
		workflows:  make(map[string]*workflowData),
		executions: make(map[string]*executionData),
		logger:     logger,
	}
}

// Name returns the provider identifier
func (p *Provider) Name() string {
	return "mock"
}

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

// CreateWorkflow stores a workflow in memory
func (p *Provider) CreateWorkflow(ctx context.Context, spec *workflow.WorkflowSpec) (*workflow.CreateWorkflowResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.workflows[spec.WorkflowID]; exists {
		p.logger.Debug("workflow already exists", zap.String("workflow_id", spec.WorkflowID))
		return &workflow.CreateWorkflowResult{
			WorkflowID:   spec.WorkflowID,
			ProviderType: "mock",
			ResourceIDs:  map[string]string{"workflow_id": spec.WorkflowID},
			CreatedAt:    p.workflows[spec.WorkflowID].createdAt,
			Message:      "workflow already exists",
		}, nil
	}

	now := time.Now()
	p.workflows[spec.WorkflowID] = &workflowData{
		spec:      spec,
		createdAt: now,
	}

	p.logger.Info("created mock workflow",
		zap.String("workflow_id", spec.WorkflowID),
		zap.String("name", spec.Name))

	return &workflow.CreateWorkflowResult{
		WorkflowID:   spec.WorkflowID,
		ProviderType: "mock",
		ResourceIDs:  map[string]string{"workflow_id": spec.WorkflowID},
		CreatedAt:    now,
		Message:      "mock workflow created successfully",
	}, nil
}

// StartExecution starts a workflow execution
func (p *Provider) StartExecution(ctx context.Context, workflowID string, input *workflow.ExecutionInput) (*workflow.ExecutionResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.workflows[workflowID]; !exists {
		return nil, workflow.ErrWorkflowNotFound
	}

	executionID := fmt.Sprintf("exec-%s-%d", workflowID, time.Now().UnixNano())
	if input.ExecutionName != "" {
		executionID = input.ExecutionName
	} else {
		counter := atomic.AddUint64(&p.execCounter, 1)
		executionID = fmt.Sprintf("exec-%s-%d-%d", workflowID, time.Now().UnixNano(), counter)
	}

	// Idempotency: Return existing execution if already created with same execution name
	if execData, exists := p.executions[executionID]; exists {
		return &workflow.ExecutionResult{
			ExecutionID:  execData.id,
			WorkflowID:   execData.workflowID,
			ProviderType: "mock",
			State:        execData.status.State,
			StartedAt:    execData.status.StartTime,
			Message:      "execution already started (idempotent result)",
		}, nil
	}

	now := time.Now()

	status := &workflow.ExecutionStatus{
		ExecutionID:  executionID,
		WorkflowID:   workflowID,
		ProviderType: "mock",
		State:        workflow.StateSucceeded,
		StartTime:    now,
		StopTime:     &now,
		Input:        input.Input,
		Output:       json.RawMessage(`{"result": "success", "mock": true}`),
		History: []workflow.ExecutionEvent{
			{
				Timestamp: now,
				Type:      "ExecutionStarted",
				Details:   json.RawMessage(`{}`),
			},
			{
				Timestamp: now,
				Type:      "ExecutionSucceeded",
				Details:   json.RawMessage(`{}`),
			},
		},
		Metadata: map[string]string{
			"mock": "true",
		},
	}

	p.executions[executionID] = &executionData{
		id:         executionID,
		workflowID: workflowID,
		input:      input,
		status:     status,
	}

	p.logger.Info("started mock execution",
		zap.String("execution_id", executionID),
		zap.String("workflow_id", workflowID),
		zap.String("state", string(workflow.StateSucceeded)))

	return &workflow.ExecutionResult{
		ExecutionID:  executionID,
		WorkflowID:   workflowID,
		ProviderType: "mock",
		State:        workflow.StateSucceeded,
		StartedAt:    now,
		Message:      "mock execution completed immediately",
	}, nil
}

// GetExecutionStatus returns the status of an execution
func (p *Provider) GetExecutionStatus(ctx context.Context, executionID string) (*workflow.ExecutionStatus, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	exec, exists := p.executions[executionID]
	if !exists {
		return nil, workflow.ErrExecutionNotFound
	}

	return exec.status, nil
}

// StopExecution cancels a running execution
func (p *Provider) StopExecution(ctx context.Context, executionID string, reason string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	exec, exists := p.executions[executionID]
	if !exists {
		return workflow.ErrExecutionNotFound
	}

	now := time.Now()
	exec.status.State = workflow.StateCancelled
	exec.status.StopTime = &now
	exec.status.History = append(exec.status.History, workflow.ExecutionEvent{
		Timestamp: now,
		Type:      "ExecutionCancelled",
		Details:   json.RawMessage(fmt.Sprintf(`{"reason": "%s"}`, reason)),
	})

	p.logger.Info("stopped mock execution",
		zap.String("execution_id", executionID),
		zap.String("reason", reason))

	return nil
}

// DeleteWorkflow removes a workflow from memory
func (p *Provider) DeleteWorkflow(ctx context.Context, workflowID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.workflows[workflowID]; !exists {
		return workflow.ErrWorkflowNotFound
	}

	delete(p.workflows, workflowID)

	p.logger.Info("deleted mock workflow", zap.String("workflow_id", workflowID))

	return nil
}

// Validate performs basic validation on the workflow spec
func (p *Provider) Validate(ctx context.Context, spec *workflow.WorkflowSpec) error {
	if len(spec.Definition) > 0 {
		if !json.Valid(spec.Definition) {
			return fmt.Errorf("definition must be valid JSON")
		}
	}

	return nil
}

// PostComputeCallback is a stub for the mock provider
func (p *Provider) PostComputeCallback(ctx context.Context, executionID string, payload *compute.CallbackPayload, opts *compute.CallbackOptions) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if _, exists := p.executions[executionID]; !exists {
		return fmt.Errorf("execution not found: %s", executionID)
	}

	p.logger.Debug("received compute callback",
		zap.String("execution_id", executionID),
		zap.String("tenant_id", payload.TenantID),
		zap.String("status", string(payload.Status)),
	)

	return nil
}
