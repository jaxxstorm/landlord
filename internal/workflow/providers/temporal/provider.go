package temporal

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jaxxstorm/landlord/internal/compute"
	"github.com/jaxxstorm/landlord/internal/config"
	"github.com/jaxxstorm/landlord/internal/workflow"
	"go.temporal.io/api/serviceerror"
	workflowservice "go.temporal.io/api/workflowservice/v1"
	temporalclient "go.temporal.io/sdk/client"
	temporal "go.temporal.io/sdk/temporal"
	"go.uber.org/zap"
)

// TemporalClient captures the subset of Temporal client methods the provider uses.
type TemporalClient interface {
	ExecuteWorkflow(ctx context.Context, options temporalclient.StartWorkflowOptions, workflow interface{}, args ...interface{}) (temporalclient.WorkflowRun, error)
	GetWorkflow(ctx context.Context, workflowID string, runID string) temporalclient.WorkflowRun
	CancelWorkflow(ctx context.Context, workflowID string, runID string) error
	TerminateWorkflow(ctx context.Context, workflowID string, runID string, reason string, details ...interface{}) error
	DescribeWorkflowExecution(ctx context.Context, workflowID string, runID string) (*workflowservice.DescribeWorkflowExecutionResponse, error)
	Close()
}

// ClientFactory creates a Temporal API client.
type ClientFactory func(options temporalclient.Options) (TemporalClient, error)

// ProviderOption configures provider construction.
type ProviderOption func(*Provider)

// WithClient injects a custom client (used by tests).
func WithClient(client TemporalClient) ProviderOption {
	return func(p *Provider) {
		p.client = client
	}
}

// WithClientFactory injects a custom client factory.
func WithClientFactory(factory ClientFactory) ProviderOption {
	return func(p *Provider) {
		p.clientFactory = factory
	}
}

type workflowRecord struct {
	createdAt time.Time
}

type executionRecord struct {
	executionID string
	workflowID  string
	runID       string
	input       *workflow.ExecutionInput
	state       workflow.ExecutionState
	startTime   time.Time
	stopTime    *time.Time
}

// Provider implements the workflow.Provider contract for Temporal.
type Provider struct {
	config config.TemporalConfig
	logger *zap.Logger

	clientFactory ClientFactory

	mu        sync.RWMutex
	workflows map[string]workflowRecord
	executions map[string]*executionRecord

	clientMux  sync.Mutex
	clientInit bool
	client     TemporalClient
}

var _ workflow.Provider = (*Provider)(nil)

// New creates a Temporal provider.
func New(cfg config.TemporalConfig, logger *zap.Logger, opts ...ProviderOption) (*Provider, error) {
	if err := ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid temporal configuration: %w", err)
	}

	p := &Provider{
		config:       cfg,
		logger:       logger.With(zap.String("component", "temporal-provider")),
		clientFactory: defaultClientFactory,
		workflows:    make(map[string]workflowRecord),
		executions:   make(map[string]*executionRecord),
	}
	for _, opt := range opts {
		opt(p)
	}
	p.registerDefaultWorkflows()

	p.logger.Info("temporal provider created",
		zap.String("host_port", cfg.HostPort),
		zap.String("namespace", cfg.Namespace),
		zap.String("task_queue", cfg.TaskQueue),
	)

	return p, nil
}

func defaultClientFactory(options temporalclient.Options) (TemporalClient, error) {
	return temporalclient.Dial(options)
}

func (p *Provider) registerDefaultWorkflows() {
	now := time.Now()
	for _, workflowID := range defaultWorkflowIDs() {
		p.workflows[workflowID] = workflowRecord{createdAt: now}
	}
}

func (p *Provider) ensureClient(ctx context.Context) (TemporalClient, error) {
	if p.client != nil {
		return p.client, nil
	}

	p.clientMux.Lock()
	defer p.clientMux.Unlock()

	if p.client != nil {
		return p.client, nil
	}

	client, err := p.clientFactory(temporalclient.Options{
		HostPort:  p.config.HostPort,
		Namespace: p.config.Namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrClientUnavailable, err)
	}

	p.client = client
	p.clientInit = true
	_ = ctx
	return p.client, nil
}

// Name returns provider identifier.
func (p *Provider) Name() string {
	return "temporal"
}

// Invoke starts a workflow execution using a simplified request payload.
func (p *Provider) Invoke(ctx context.Context, workflowID string, request *workflow.ProvisionRequest) (*workflow.ExecutionResult, error) {
	if request == nil {
		return nil, fmt.Errorf("provision request is required")
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	tenantIdentifier := request.TenantUUID
	if tenantIdentifier == "" {
		tenantIdentifier = request.TenantID
	}
	if tenantIdentifier == "" {
		return nil, fmt.Errorf("tenant identifier is required")
	}

	operation := request.Operation
	if operation == "" {
		operation = "provision"
	}

	executionName := fmt.Sprintf("tenant-%s-%s-%s", tenantIdentifier, workflowID, operation)
	input := &workflow.ExecutionInput{
		ExecutionName: executionName,
		Input:         payload,
		Tags: map[string]string{
			"tenant_id":   request.TenantID,
			"tenant_uuid": request.TenantUUID,
			"operation":   operation,
		},
		Metadata:      request.Metadata,
		TriggerSource: "reconciler",
	}

	return p.StartExecution(ctx, workflowID, input)
}

// GetWorkflowStatus returns simplified workflow status.
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

// Validate validates workflow specification.
func (p *Provider) Validate(ctx context.Context, spec *workflow.WorkflowSpec) error {
	_ = ctx
	if spec == nil {
		return fmt.Errorf("workflow spec cannot be nil")
	}
	if spec.WorkflowID == "" {
		return fmt.Errorf("workflow ID is required")
	}
	if len(spec.Definition) == 0 {
		return fmt.Errorf("workflow definition is required")
	}
	if !json.Valid(spec.Definition) {
		return fmt.Errorf("workflow definition must be valid JSON")
	}
	if len(spec.ProviderConfig) > 0 && !json.Valid(spec.ProviderConfig) {
		return fmt.Errorf("provider config must be valid JSON")
	}
	return nil
}

// CreateWorkflow records workflow metadata and behaves idempotently.
func (p *Provider) CreateWorkflow(ctx context.Context, spec *workflow.WorkflowSpec) (*workflow.CreateWorkflowResult, error) {
	if err := p.Validate(ctx, spec); err != nil {
		return nil, fmt.Errorf("%w: %s", workflow.ErrInvalidSpec, err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if existing, ok := p.workflows[spec.WorkflowID]; ok {
		return &workflow.CreateWorkflowResult{
			WorkflowID:   spec.WorkflowID,
			ProviderType: p.Name(),
			ResourceIDs: map[string]string{
				"namespace":     p.config.Namespace,
				"task_queue":    p.config.TaskQueue,
				"workflow_type": workflowTypeForID(spec.WorkflowID),
			},
			CreatedAt: existing.createdAt,
			Message:   "workflow already exists",
		}, nil
	}

	now := time.Now()
	p.workflows[spec.WorkflowID] = workflowRecord{createdAt: now}

	return &workflow.CreateWorkflowResult{
		WorkflowID:   spec.WorkflowID,
		ProviderType: p.Name(),
		ResourceIDs: map[string]string{
			"namespace":     p.config.Namespace,
			"task_queue":    p.config.TaskQueue,
			"workflow_type": workflowTypeForID(spec.WorkflowID),
		},
		CreatedAt: now,
		Message:   "workflow registered",
	}, nil
}

// StartExecution starts Temporal workflow execution with idempotent execution IDs.
func (p *Provider) StartExecution(ctx context.Context, workflowID string, input *workflow.ExecutionInput) (*workflow.ExecutionResult, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("workflow ID is required")
	}
	if input == nil {
		return nil, fmt.Errorf("execution input is required")
	}
	if len(input.Input) == 0 {
		return nil, fmt.Errorf("execution input payload is required")
	}
	if !json.Valid(input.Input) {
		return nil, fmt.Errorf("execution input must be valid JSON")
	}

	p.mu.RLock()
	_, exists := p.workflows[workflowID]
	p.mu.RUnlock()
	if !exists {
		return nil, workflow.ErrWorkflowNotFound
	}

	executionID := input.ExecutionName
	if executionID == "" {
		executionID = deterministicExecutionName(workflowID, input)
	}

	p.mu.RLock()
	if existing, ok := p.executions[executionID]; ok {
		p.mu.RUnlock()
		return &workflow.ExecutionResult{
			ExecutionID:  existing.executionID,
			WorkflowID:   existing.workflowID,
			ProviderType: p.Name(),
			State:        existing.state,
			StartedAt:    existing.startTime,
			Message:      "execution already started (idempotent result)",
		}, nil
	}
	p.mu.RUnlock()

	client, err := p.ensureClient(ctx)
	if err != nil {
		return nil, err
	}

	opts := temporalclient.StartWorkflowOptions{
		ID:                       executionID,
		TaskQueue:                p.config.TaskQueue,
		WorkflowExecutionTimeout: p.config.Timeout,
		WorkflowRunTimeout:       p.config.Timeout,
	}
	if p.config.RetryAttempts > 0 {
		opts.RetryPolicy = &temporal.RetryPolicy{MaximumAttempts: int32(p.config.RetryAttempts)}
	}

	workflowType := workflowTypeForID(workflowID)
	workflowArg, err := workflowArgumentForID(workflowID, input.Input)
	if err != nil {
		return nil, fmt.Errorf("decode execution input for %s: %w", workflowID, err)
	}

	run, err := client.ExecuteWorkflow(ctx, opts, workflowType, workflowArg)
	if err != nil {
		var alreadyStarted *serviceerror.WorkflowExecutionAlreadyStarted
		if errors.As(err, &alreadyStarted) {
			run = client.GetWorkflow(ctx, executionID, "")
		} else {
			return nil, fmt.Errorf("start temporal execution: %w", err)
		}
	}

	now := time.Now()
	rec := &executionRecord{
		executionID: executionID,
		workflowID:  workflowID,
		runID:       run.GetRunID(),
		input:       input,
		state:       workflow.StateRunning,
		startTime:   now,
	}

	p.mu.Lock()
	p.executions[executionID] = rec
	p.mu.Unlock()

	return &workflow.ExecutionResult{
		ExecutionID:  executionID,
		WorkflowID:   workflowID,
		ProviderType: p.Name(),
		State:        workflow.StateRunning,
		StartedAt:    now,
		Message:      "execution started",
	}, nil
}

// GetExecutionStatus fetches execution status from Temporal and maps it to canonical states.
func (p *Provider) GetExecutionStatus(ctx context.Context, executionID string) (*workflow.ExecutionStatus, error) {
	if executionID == "" {
		return nil, fmt.Errorf("execution ID is required")
	}

	p.mu.RLock()
	rec := p.executions[executionID]
	p.mu.RUnlock()
	if rec == nil {
		return nil, workflow.ErrExecutionNotFound
	}

	client, err := p.ensureClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := client.DescribeWorkflowExecution(ctx, executionID, rec.runID)
	if err != nil {
		if isTemporalNotFound(err) {
			return nil, workflow.ErrExecutionNotFound
		}
		return nil, fmt.Errorf("describe temporal execution: %w", err)
	}

	state := workflow.StateRunning
	startTime := rec.startTime
	stopTime := rec.stopTime
	metadata := buildStatusMetadata(resp, rec)
	if info := resp.GetWorkflowExecutionInfo(); info != nil {
		state = mapExecutionState(info.GetStatus())
		if info.StartTime != nil {
			startTime = info.StartTime.AsTime()
		}
		if info.CloseTime != nil {
			closeTime := info.CloseTime.AsTime()
			stopTime = &closeTime
		}
	}

	var execErr *workflow.ExecutionError
	if state == workflow.StateFailed || state == workflow.StateTimedOut || state == workflow.StateCancelled {
		msg := "temporal execution reached terminal error state"
		if info := resp.GetWorkflowExecutionInfo(); info != nil {
			msg = fmt.Sprintf("temporal status: %s", info.GetStatus().String())
		}
		execErr = &workflow.ExecutionError{Code: string(state), Message: msg}
	}

	p.mu.Lock()
	rec.state = state
	rec.startTime = startTime
	rec.stopTime = stopTime
	p.mu.Unlock()

	return &workflow.ExecutionStatus{
		ExecutionID:  executionID,
		WorkflowID:   rec.workflowID,
		ProviderType: p.Name(),
		State:        state,
		StartTime:    startTime,
		StopTime:     stopTime,
		Input:        rec.input.Input,
		Error:        execErr,
		Metadata:     metadata,
	}, nil
}

// StopExecution requests cancellation/termination and is idempotent.
func (p *Provider) StopExecution(ctx context.Context, executionID string, reason string) error {
	if executionID == "" {
		return fmt.Errorf("execution ID is required")
	}

	p.mu.RLock()
	rec := p.executions[executionID]
	p.mu.RUnlock()
	if rec == nil {
		return nil
	}

	client, err := p.ensureClient(ctx)
	if err != nil {
		return err
	}

	cancelErr := client.CancelWorkflow(ctx, executionID, rec.runID)
	if cancelErr != nil && !isTemporalNotFound(cancelErr) && !isTemporalAlreadyCompleted(cancelErr) {
		terminateErr := client.TerminateWorkflow(ctx, executionID, rec.runID, reason)
		if terminateErr != nil && !isTemporalNotFound(terminateErr) && !isTemporalAlreadyCompleted(terminateErr) {
			return fmt.Errorf("stop temporal execution: cancel=%v terminate=%w", cancelErr, terminateErr)
		}
	}

	now := time.Now()
	p.mu.Lock()
	rec.state = workflow.StateCancelled
	rec.stopTime = &now
	p.mu.Unlock()
	return nil
}

// DeleteWorkflow removes local workflow metadata and is idempotent.
func (p *Provider) DeleteWorkflow(ctx context.Context, workflowID string) error {
	_ = ctx
	if workflowID == "" {
		return nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.workflows, workflowID)
	for id, rec := range p.executions {
		if rec.workflowID == workflowID {
			delete(p.executions, id)
		}
	}
	return nil
}

// PostComputeCallback is a no-op callback hook for Temporal provider.
func (p *Provider) PostComputeCallback(ctx context.Context, executionID string, payload *compute.CallbackPayload, opts *compute.CallbackOptions) error {
	_ = ctx
	_ = opts
	if executionID == "" {
		return fmt.Errorf("execution ID is required")
	}
	if payload == nil {
		return fmt.Errorf("callback payload is required")
	}

	p.mu.RLock()
	_, exists := p.executions[executionID]
	p.mu.RUnlock()
	if !exists {
		return workflow.ErrExecutionNotFound
	}

	p.logger.Debug("received compute callback",
		zap.String("execution_id", executionID),
		zap.String("tenant_id", payload.TenantID),
		zap.String("status", string(payload.Status)),
	)
	return nil
}

func deterministicExecutionName(workflowID string, input *workflow.ExecutionInput) string {
	h := sha1.New() //nolint:gosec // deterministic naming, not security-sensitive
	h.Write([]byte(workflowID))
	h.Write(input.Input)
	for k, v := range input.Metadata {
		h.Write([]byte(k))
		h.Write([]byte("="))
		h.Write([]byte(v))
	}
	return fmt.Sprintf("%s-%s", workflowID, hex.EncodeToString(h.Sum(nil))[:12])
}

func workflowTypeForID(workflowID string) string {
	if workflowID == "" {
		return "tenant-lifecycle"
	}
	return workflowID
}

func workflowArgumentForID(workflowID string, payload json.RawMessage) (interface{}, error) {
	if workflowID == tenantProvisioningWorkflowID {
		var req workflow.ProvisionRequest
		if err := json.Unmarshal(payload, &req); err != nil {
			return nil, err
		}
		return req, nil
	}

	var decoded interface{}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return nil, err
	}
	return decoded, nil
}

func isTemporalNotFound(err error) bool {
	var notFound *serviceerror.NotFound
	return errors.As(err, &notFound)
}

func isTemporalAlreadyCompleted(err error) bool {
	var alreadyRequested *serviceerror.CancellationAlreadyRequested
	if errors.As(err, &alreadyRequested) {
		return true
	}
	var notReady *serviceerror.WorkflowNotReady
	if errors.As(err, &notReady) {
		return true
	}
	var failedPrecondition *serviceerror.FailedPrecondition
	return errors.As(err, &failedPrecondition)
}
