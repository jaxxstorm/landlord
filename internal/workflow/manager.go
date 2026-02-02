package workflow

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// Manager coordinates workflow operations
type Manager struct {
	registry *Registry
	logger   *zap.Logger
}

// New creates a new workflow manager
func New(registry *Registry, logger *zap.Logger) *Manager {
	return &Manager{
		registry: registry,
		logger:   logger.With(zap.String("component", "workflow-manager")),
	}
}

// CreateWorkflow creates a workflow definition
func (m *Manager) CreateWorkflow(ctx context.Context, spec *WorkflowSpec) (*CreateWorkflowResult, error) {
	m.logger.Info("creating workflow",
		zap.String("workflow_id", spec.WorkflowID),
		zap.String("provider", spec.ProviderType),
	)

	// Validate spec
	if err := ValidateWorkflowSpec(spec); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidSpec, err)
	}

	// Get provider
	provider, err := m.registry.Get(spec.ProviderType)
	if err != nil {
		return nil, err
	}

	// Delegate to provider
	result, err := provider.CreateWorkflow(ctx, spec)
	if err != nil {
		m.logger.Error("workflow creation failed",
			zap.String("workflow_id", spec.WorkflowID),
			zap.Error(err),
		)
		return nil, err
	}

	m.logger.Info("workflow created",
		zap.String("workflow_id", spec.WorkflowID),
	)

	return result, nil
}

// StartExecution starts a workflow execution
func (m *Manager) StartExecution(ctx context.Context, workflowID string, providerType string, input *ExecutionInput) (*ExecutionResult, error) {
	m.logger.Info("starting execution",
		zap.String("workflow_id", workflowID),
		zap.String("provider", providerType),
	)

	// Validate input
	if err := ValidateExecutionInput(input); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidSpec, err)
	}

	// Get provider
	provider, err := m.registry.Get(providerType)
	if err != nil {
		return nil, err
	}

	// Delegate to provider
	result, err := provider.StartExecution(ctx, workflowID, input)
	if err != nil {
		m.logger.Error("execution start failed",
			zap.String("workflow_id", workflowID),
			zap.Error(err),
		)
		return nil, err
	}

	m.logger.Info("execution started",
		zap.String("execution_id", result.ExecutionID),
		zap.String("workflow_id", workflowID),
	)

	return result, nil
}

// Invoke starts a workflow execution using a simplified request payload
func (m *Manager) Invoke(ctx context.Context, workflowID string, providerType string, request *ProvisionRequest) (*ExecutionResult, error) {
	m.logger.Info("invoking workflow",
		zap.String("workflow_id", workflowID),
		zap.String("provider", providerType),
	)

	provider, err := m.registry.Get(providerType)
	if err != nil {
		return nil, err
	}

	result, err := provider.Invoke(ctx, workflowID, request)
	if err != nil {
		m.logger.Error("workflow invoke failed",
			zap.String("workflow_id", workflowID),
			zap.Error(err),
		)
		return nil, err
	}

	m.logger.Info("workflow invoked",
		zap.String("execution_id", result.ExecutionID),
		zap.String("workflow_id", workflowID),
	)

	return result, nil
}

// GetExecutionStatus queries execution status
func (m *Manager) GetExecutionStatus(ctx context.Context, executionID string, providerType string) (*ExecutionStatus, error) {
	// Get provider
	provider, err := m.registry.Get(providerType)
	if err != nil {
		return nil, err
	}

	// Delegate to provider
	status, err := provider.GetExecutionStatus(ctx, executionID)
	if err != nil {
		return nil, err
	}

	return status, nil
}

// GetWorkflowStatus queries simplified execution status
func (m *Manager) GetWorkflowStatus(ctx context.Context, executionID string, providerType string) (*WorkflowStatus, error) {
	provider, err := m.registry.Get(providerType)
	if err != nil {
		return nil, err
	}

	return provider.GetWorkflowStatus(ctx, executionID)
}

// StopExecution stops a running execution
func (m *Manager) StopExecution(ctx context.Context, executionID string, providerType string, reason string) error {
	m.logger.Info("stopping execution",
		zap.String("execution_id", executionID),
		zap.String("provider", providerType),
		zap.String("reason", reason),
	)

	// Get provider
	provider, err := m.registry.Get(providerType)
	if err != nil {
		return err
	}

	// Delegate to provider
	if err := provider.StopExecution(ctx, executionID, reason); err != nil {
		m.logger.Error("stop execution failed",
			zap.String("execution_id", executionID),
			zap.Error(err),
		)
		return err
	}

	m.logger.Info("execution stopped",
		zap.String("execution_id", executionID),
	)

	return nil
}

// DeleteWorkflow removes a workflow definition
func (m *Manager) DeleteWorkflow(ctx context.Context, workflowID string, providerType string) error {
	m.logger.Info("deleting workflow",
		zap.String("workflow_id", workflowID),
		zap.String("provider", providerType),
	)

	// Get provider
	provider, err := m.registry.Get(providerType)
	if err != nil {
		return err
	}

	// Delegate to provider
	if err := provider.DeleteWorkflow(ctx, workflowID); err != nil {
		m.logger.Error("workflow deletion failed",
			zap.String("workflow_id", workflowID),
			zap.Error(err),
		)
		return err
	}

	m.logger.Info("workflow deleted",
		zap.String("workflow_id", workflowID),
	)

	return nil
}

// ValidateWorkflowSpec validates a spec without creating it
func (m *Manager) ValidateWorkflowSpec(ctx context.Context, spec *WorkflowSpec) error {
	// Structural validation
	if err := ValidateWorkflowSpec(spec); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidSpec, err)
	}

	// Get provider for provider-specific validation
	provider, err := m.registry.Get(spec.ProviderType)
	if err != nil {
		return err
	}

	// Delegate to provider
	return provider.Validate(ctx, spec)
}

// ListProviders returns available provider types
func (m *Manager) ListProviders() []string {
	return m.registry.List()
}
