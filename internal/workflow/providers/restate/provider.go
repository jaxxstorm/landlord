package restate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/jaxxstorm/landlord/internal/compute"
	"github.com/jaxxstorm/landlord/internal/config"
	"github.com/jaxxstorm/landlord/internal/workflow"
	"go.uber.org/zap"
)

// Provider is a Restate.dev workflow provider
type Provider struct {
	config        config.RestateConfig
	logger        *zap.Logger
	client        *Client
	clientMux     sync.Mutex
	clientInit    bool
	initialized   bool
	registeredMux sync.RWMutex
	registered    map[string]bool
}

// New creates a new Restate provider
func New(cfg config.RestateConfig, logger *zap.Logger) (*Provider, error) {
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid restate configuration: %w", err)
	}

	p := &Provider{
		config:      cfg,
		logger:      logger.With(zap.String("component", "restate-provider")),
		initialized: true,
		registered:  make(map[string]bool),
	}

	p.logger.Info("restate provider created",
		zap.String("endpoint", cfg.Endpoint),
		zap.String("execution_mechanism", cfg.ExecutionMechanism),
		zap.String("auth_type", cfg.AuthType),
	)

	if err := p.registerWorkflows(context.Background()); err != nil {
		p.logger.Warn("workflow registration incomplete",
			zap.Error(err),
		)
	}

	return p, nil
}

// Name returns the provider identifier
func (p *Provider) Name() string {
	return "restate"
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
		Metadata:      request.Metadata, // Pass through metadata (e.g., config_hash)
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

// Validate performs provider-specific validation on a workflow spec
func (p *Provider) Validate(ctx context.Context, spec *workflow.WorkflowSpec) error {
	if spec == nil {
		return fmt.Errorf("workflow spec cannot be nil")
	}

	if spec.WorkflowID == "" {
		return fmt.Errorf("workflow ID is required")
	}

	if len(spec.Definition) == 0 {
		return fmt.Errorf("workflow definition is required")
	}

	p.logger.Debug("workflow spec validated", zap.String("workflow_id", spec.WorkflowID))
	return nil
}

// PostComputeCallback sends a compute execution callback to a Restate service
func (p *Provider) PostComputeCallback(ctx context.Context, executionID string, payload *compute.CallbackPayload, opts *compute.CallbackOptions) error {
	if executionID == "" {
		return fmt.Errorf("execution ID is required")
	}

	if payload == nil {
		return fmt.Errorf("callback payload is required")
	}

	// Restate callback mechanism would typically involve:
	// 1. Calling a Restate service endpoint to deliver the callback
	// 2. The service would be running as a long-lived process listening for callbacks
	// 3. Or using Restate's built-in callback handling if available

	p.logger.Debug("posting compute callback",
		zap.String("execution_id", executionID),
		zap.String("tenant_id", payload.TenantID),
		zap.String("status", string(payload.Status)),
	)

	return nil
}

// CreateWorkflow creates a workflow definition in Restate
func (p *Provider) CreateWorkflow(ctx context.Context, spec *workflow.WorkflowSpec) (*workflow.CreateWorkflowResult, error) {
	if err := p.Validate(ctx, spec); err != nil {
		return nil, fmt.Errorf("%w: %s", workflow.ErrInvalidSpec, err)
	}

	// Ensure client is initialized
	client, err := p.ensureClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize restate client: %w", err)
	}

	// Normalize service name
	serviceName := workflowServiceName(p.config, spec.WorkflowID)

	// Check if service already exists (idempotency)
	existing, err := client.GetService(ctx, serviceName)
	if err == nil && existing != nil {
		p.logger.Debug("workflow already exists (idempotent)",
			zap.String("workflow_id", spec.WorkflowID),
			zap.String("service_name", serviceName),
		)
		p.markRegistered(serviceName)
		return &workflow.CreateWorkflowResult{
			WorkflowID:   spec.WorkflowID,
			ProviderType: "restate",
			ResourceIDs: map[string]string{
				"service_name": serviceName,
				"endpoint":     p.config.Endpoint,
			},
			Message: "service already registered",
		}, nil
	}

	if err != nil && !errors.Is(err, workflow.ErrWorkflowNotFound) {
		return nil, fmt.Errorf("failed to query restate service: %w", err)
	}

	if err := client.RegisterService(ctx, serviceName); err != nil {
		if isAlreadyExistsError(err) {
			p.logger.Info("workflow registration already completed",
				zap.String("workflow_id", spec.WorkflowID),
				zap.String("service_name", serviceName),
			)
			p.markRegistered(serviceName)
		} else {
			p.logger.Error("workflow registration failed",
				zap.String("workflow_id", spec.WorkflowID),
				zap.String("service_name", serviceName),
				zap.Error(err),
			)
			return nil, fmt.Errorf("failed to register workflow: %w", err)
		}
	} else {
		p.logger.Info("workflow registration completed",
			zap.String("workflow_id", spec.WorkflowID),
			zap.String("service_name", serviceName),
		)
		p.markRegistered(serviceName)
	}

	// Register service with Restate
	resourceIDs := map[string]string{
		"service_name": serviceName,
		"endpoint":     p.config.Endpoint,
	}

	if p.config.ExecutionMechanism != "" {
		resourceIDs["execution_mechanism"] = p.config.ExecutionMechanism
	}

	p.logger.Info("workflow created",
		zap.String("workflow_id", spec.WorkflowID),
		zap.String("service_name", serviceName),
		zap.String("endpoint", p.config.Endpoint),
	)

	return &workflow.CreateWorkflowResult{
		WorkflowID:   spec.WorkflowID,
		ProviderType: "restate",
		ResourceIDs:  resourceIDs,
	}, nil
}

// StartExecution starts a workflow execution
func (p *Provider) StartExecution(ctx context.Context, workflowID string, input *workflow.ExecutionInput) (*workflow.ExecutionResult, error) {
	if workflowID == "" {
		return nil, fmt.Errorf("workflow ID is required")
	}

	if input == nil {
		return nil, fmt.Errorf("execution input is required")
	}

	// Ensure client is initialized
	client, err := p.ensureClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize restate client: %w", err)
	}

	// Get service name - use fixed name for tenant provisioning
	serviceName := workflowServiceName(p.config, workflowID)

	if err := p.ensureServiceRegistered(ctx, serviceName); err != nil {
		return nil, err
	}

	// Generate execution name if not provided
	executionName := input.ExecutionName
	if executionName == "" {
		executionName = generateExecutionID(workflowID)
	}

	// Invoke service through Restate
	executionID, err := client.InvokeService(ctx, serviceName, executionName, input.Input)
	if err != nil {
		p.logger.Error("failed to start execution",
			zap.String("workflow_id", workflowID),
			zap.String("service_name", serviceName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to start execution: %w", err)
	}

	p.logger.Info("execution started",
		zap.String("workflow_id", workflowID),
		zap.String("execution_id", executionID),
		zap.String("service_name", serviceName),
	)

	return &workflow.ExecutionResult{
		ExecutionID:  executionID,
		WorkflowID:   workflowID,
		ProviderType: "restate",
		State:        "running",
	}, nil
}

// GetExecutionStatus queries current execution state
func (p *Provider) GetExecutionStatus(ctx context.Context, executionID string) (*workflow.ExecutionStatus, error) {
	if executionID == "" {
		return nil, fmt.Errorf("execution ID is required")
	}

	// Ensure client is initialized
	client, err := p.ensureClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize restate client: %w", err)
	}

	// Query execution status from Restate
	status, err := client.GetExecutionStatus(ctx, executionID)
	if err != nil {
		p.logger.Error("failed to get execution status",
			zap.String("execution_id", executionID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get execution status: %w", err)
	}

	p.logger.Info("provider received execution status from client",
		zap.String("execution_id", executionID),
		zap.String("state", string(status.State)),
		zap.Any("metadata", status.Metadata))

	return status, nil
}

// StopExecution stops a running execution
func (p *Provider) StopExecution(ctx context.Context, executionID string, reason string) error {
	if executionID == "" {
		return fmt.Errorf("execution ID is required")
	}

	// Ensure client is initialized
	client, err := p.ensureClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize restate client: %w", err)
	}

	// Stop execution in Restate
	if err := client.CancelExecution(ctx, executionID); err != nil {
		p.logger.Error("failed to stop execution",
			zap.String("execution_id", executionID),
			zap.String("reason", reason),
			zap.Error(err),
		)
		return fmt.Errorf("failed to stop execution: %w", err)
	}

	p.logger.Info("execution stopped",
		zap.String("execution_id", executionID),
		zap.String("reason", reason),
	)

	return nil
}

// DeleteWorkflow removes a workflow definition
func (p *Provider) DeleteWorkflow(ctx context.Context, workflowID string) error {
	if workflowID == "" {
		return fmt.Errorf("workflow ID is required")
	}

	// Ensure client is initialized
	client, err := p.ensureClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize restate client: %w", err)
	}

	// Get service name
	serviceName := workflowServiceName(p.config, workflowID)

	// Delete service from Restate (idempotent)
	if err := client.DeleteService(ctx, serviceName); err != nil {
		// Don't fail if service doesn't exist (idempotency)
		if !isNotFoundError(err) {
			p.logger.Error("failed to delete workflow",
				zap.String("workflow_id", workflowID),
				zap.String("service_name", serviceName),
				zap.Error(err),
			)
			return fmt.Errorf("failed to delete workflow: %w", err)
		}
	}

	p.logger.Info("workflow deleted",
		zap.String("workflow_id", workflowID),
		zap.String("service_name", serviceName),
	)

	return nil
}

// ensureClient initializes the Restate SDK client on first use (lazy initialization)
func (p *Provider) ensureClient(ctx context.Context) (*Client, error) {
	p.clientMux.Lock()
	defer p.clientMux.Unlock()

	if p.clientInit {
		return p.client, nil
	}

	client, err := NewClient(ctx, p.config, p.logger)
	if err != nil {
		return nil, err
	}

	p.client = client
	p.clientInit = true
	return p.client, nil
}

func (p *Provider) registerWorkflows(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, p.config.Timeout)
	defer cancel()

	client, err := p.ensureClient(ctx)
	if err != nil {
		p.logger.Warn("failed to initialize restate client during registration", zap.Error(err))
		return err
	}

	var failures []error
	registeredCount := 0
	for _, workflowID := range defaultWorkflowIDs() {
		serviceName := workflowServiceName(p.config, workflowID)
		err := client.RegisterService(ctx, serviceName)
		if err != nil {
			if isAlreadyExistsError(err) {
				p.logger.Info("workflow already registered (idempotent)",
					zap.String("workflow_id", workflowID),
					zap.String("service_name", serviceName),
				)
				p.markRegistered(serviceName)
				continue
			}
			p.logger.Warn("workflow registration failed",
				zap.String("workflow_id", workflowID),
				zap.String("service_name", serviceName),
				zap.Error(err),
			)
			failures = append(failures, err)
			continue
		}

		p.logger.Info("workflow registered",
			zap.String("workflow_id", workflowID),
			zap.String("service_name", serviceName),
		)
		p.markRegistered(serviceName)
		registeredCount++
	}

	if err := p.registerWorkerService(ctx); err != nil {
		p.logger.Warn("worker service registration failed",
			zap.Error(err),
		)
		failures = append(failures, err)
	} else {
		registeredCount++
	}

	if len(failures) == 0 {
		p.logger.Info("workflow registration complete",
			zap.Int("registered_count", registeredCount),
		)
		return nil
	}

	return fmt.Errorf("workflow registration completed with %d failure(s)", len(failures))
}

func (p *Provider) registerWorkerService(ctx context.Context) error {
	if !p.config.WorkerRegisterOnStartup {
		return nil
	}

	clientCfg := p.config
	if p.config.WorkerAdminEndpoint != "" {
		clientCfg.AdminEndpoint = p.config.WorkerAdminEndpoint
	}

	client, err := NewClient(ctx, clientCfg, p.logger)
	if err != nil {
		return fmt.Errorf("init restate client: %w", err)
	}

	if p.config.WorkerAdvertisedURL == "" {
		return nil
	}

	if err := client.RegisterDeployment(ctx, p.config.WorkerAdvertisedURL); err != nil {
		if errors.Is(err, errAdminAPINotSupported) {
			p.logger.Warn("worker deployment registration not supported by restate admin api", zap.Error(err))
			return nil
		}
		return err
	}

	p.logger.Info("worker deployment registered",
		zap.String("uri", p.config.WorkerAdvertisedURL),
	)
	return nil
}

func (p *Provider) markRegistered(serviceName string) {
	p.registeredMux.Lock()
	defer p.registeredMux.Unlock()
	p.registered[serviceName] = true
}

func (p *Provider) isRegistered(serviceName string) bool {
	p.registeredMux.RLock()
	defer p.registeredMux.RUnlock()
	return p.registered[serviceName]
}

func (p *Provider) ensureServiceRegistered(ctx context.Context, serviceName string) error {
	if p.isRegistered(serviceName) {
		return nil
	}

	client, err := p.ensureClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize restate client: %w", err)
	}

	_, err = client.GetService(ctx, serviceName)
	if err == nil {
		p.markRegistered(serviceName)
		return nil
	}

	if errors.Is(err, workflow.ErrWorkflowNotFound) {
		return fmt.Errorf("%w: workflow service %s not registered yet; retry after startup", workflow.ErrWorkflowNotFound, serviceName)
	}

	return fmt.Errorf("failed to verify workflow registration: %w", err)
}
