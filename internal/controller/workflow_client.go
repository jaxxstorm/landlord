package controller

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/jaxxstorm/landlord/internal/tenant"
	"github.com/jaxxstorm/landlord/internal/workflow"
)

// WorkflowClient wraps the workflow manager and compute integration for controller use
type WorkflowClient struct {
	manager       *workflow.Manager
	computeClient *workflow.ComputeWorkflowClient
	logger        *zap.Logger
	timeout       time.Duration
	providerType  string
}

// NewWorkflowClient creates a workflow client
func NewWorkflowClient(manager *workflow.Manager, logger *zap.Logger, timeout time.Duration, providerType string) *WorkflowClient {
	return &WorkflowClient{
		manager:      manager,
		logger:       logger.With(zap.String("component", "workflow-client")),
		timeout:      timeout,
		providerType: providerType,
	}
}

// NewWorkflowClientWithCompute creates a workflow client with compute integration
func NewWorkflowClientWithCompute(manager *workflow.Manager, computeClient *workflow.ComputeWorkflowClient, logger *zap.Logger, timeout time.Duration, providerType string) *WorkflowClient {
	return &WorkflowClient{
		manager:       manager,
		computeClient: computeClient,
		logger:        logger.With(zap.String("component", "workflow-client")),
		timeout:       timeout,
		providerType:  providerType,
	}
}

// TriggerWorkflow triggers a workflow based on tenant status
// Returns execution ID and error
func (wc *WorkflowClient) TriggerWorkflow(ctx context.Context, t *tenant.Tenant, action string) (string, error) {
	return wc.TriggerWorkflowWithSource(ctx, t, action, "controller")
}

// TriggerWorkflowWithSource triggers a workflow with specified trigger source
func (wc *WorkflowClient) TriggerWorkflowWithSource(ctx context.Context, t *tenant.Tenant, action, triggerSource string) (string, error) {
	if wc.manager == nil {
		return "", fmt.Errorf("workflow manager not initialized")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, wc.timeout)
	defer cancel()

	wc.logger.Info("triggering workflow",
		zap.String("tenant_name", t.Name),
		zap.String("action", action),
		zap.String("status", string(t.Status)),
		zap.String("trigger_source", triggerSource))

	// Determine workflow ID based on action
	workflowID := fmt.Sprintf("tenant-%s-%s", t.ID.String(), action)
	if wc.providerType == "restate" {
		workflowID = "tenant-provisioning"
	}

	request := &workflow.ProvisionRequest{
		TenantID:      t.Name,
		TenantUUID:    t.ID.String(),
		Operation:     action,
		DesiredConfig: t.DesiredConfig,
	}
	if provider, ok := t.DesiredConfig["compute_provider"]; ok {
		if value, ok := provider.(string); ok {
			request.ComputeProvider = value
		}
	}

	// Start workflow execution
	// TODO: Get provider type from configuration or tenant
	providerType := wc.providerType // Use configured provider type

	result, err := wc.manager.Invoke(ctx, workflowID, providerType, request)
	if err != nil {
		wc.logger.Error("workflow trigger failed",
			zap.String("tenant_name", t.Name),
			zap.String("action", action),
			zap.Error(err))
		return "", err
	}

	wc.logger.Info("workflow triggered",
		zap.String("tenant_name", t.Name),
		zap.String("execution_id", result.ExecutionID))

	return result.ExecutionID, nil
}

// GetExecutionStatus queries the status of a workflow execution
func (wc *WorkflowClient) GetExecutionStatus(ctx context.Context, executionID string) (*workflow.ExecutionStatus, error) {
	if wc.manager == nil {
		return nil, fmt.Errorf("workflow manager not initialized")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, wc.timeout)
	defer cancel()

	providerType := wc.providerType

	// Get provider directly to access ExecutionStatus (which includes metadata)
	provider, err := wc.manager.GetProvider(providerType)
	if err != nil {
		wc.logger.Error("failed to get provider",
			zap.String("provider_type", providerType),
			zap.Error(err))
		return nil, err
	}

	// Call GetExecutionStatus which returns full ExecutionStatus with metadata
	status, err := provider.GetExecutionStatus(ctx, executionID)
	if err != nil {
		wc.logger.Error("failed to get execution status",
			zap.String("execution_id", executionID),
			zap.Error(err))
		return nil, err
	}

	if status == nil {
		return nil, fmt.Errorf("workflow status is nil")
	}

	return status, nil
}

// DetermineAction determines the workflow action based on tenant status
func (wc *WorkflowClient) DetermineAction(status tenant.Status) (string, error) {
	switch status {
	case tenant.StatusRequested, tenant.StatusPlanning:
		return "provision", nil
	case tenant.StatusUpdating:
		return "update", nil
	case tenant.StatusDeleting:
		return "delete", nil
	case tenant.StatusArchiving:
		return "delete", nil
	case tenant.StatusProvisioning:
		return "provision", nil
	case tenant.StatusReady, tenant.StatusArchived, tenant.StatusFailed:
		return "", fmt.Errorf("no action for terminal status: %s", status)
	default:
		return "", fmt.Errorf("unknown status: %s", status)
	}
}

// IsRetryableError classifies workflow errors as retryable or fatal
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for known retryable errors
	// TODO: Expand this list based on actual workflow provider errors
	switch {
	case err == context.DeadlineExceeded:
		return true
	case err == context.Canceled:
		return false // Explicit cancellation is not retryable
	default:
		// Default to retryable for unknown errors
		return true
	}
}

// ProvisionTenant provisions compute resources (callable from workflows)
func (wc *WorkflowClient) ProvisionTenant(ctx context.Context, spec *workflow.ProvisionTenantInput, workflowExecutionID string) (*workflow.ProvisionTenantOutput, error) {
	if wc.computeClient == nil {
		return nil, fmt.Errorf("compute client not configured")
	}

	wc.logger.Debug("provisioning tenant via compute",
		zap.String("tenant_id", spec.Spec.TenantID),
		zap.String("workflow_execution_id", spec.WorkflowExecutionID),
	)

	return wc.computeClient.ProvisionTenant(ctx, spec)
}

// UpdateTenant updates compute resources (callable from workflows)
func (wc *WorkflowClient) UpdateTenant(ctx context.Context, spec *workflow.UpdateTenantInput, workflowExecutionID string) (*workflow.UpdateTenantOutput, error) {
	if wc.computeClient == nil {
		return nil, fmt.Errorf("compute client not configured")
	}

	wc.logger.Debug("updating tenant via compute",
		zap.String("tenant_id", spec.TenantID),
		zap.String("workflow_execution_id", spec.WorkflowExecutionID),
	)

	return wc.computeClient.UpdateTenant(ctx, spec)
}

// DeleteTenant deletes compute resources (callable from workflows)
func (wc *WorkflowClient) DeleteTenant(ctx context.Context, spec *workflow.DeleteTenantInput, workflowExecutionID string) (*workflow.DeleteTenantOutput, error) {
	if wc.computeClient == nil {
		return nil, fmt.Errorf("compute client not configured")
	}

	wc.logger.Debug("deleting tenant via compute",
		zap.String("tenant_id", spec.TenantID),
		zap.String("workflow_execution_id", spec.WorkflowExecutionID),
	)

	return wc.computeClient.DeleteTenant(ctx, spec)
}

// GetComputeExecutionStatus queries the status of a compute execution
func (wc *WorkflowClient) GetComputeExecutionStatus(ctx context.Context, executionID string) (*workflow.ComputeExecutionStatusOutput, error) {
	if wc.computeClient == nil {
		return nil, fmt.Errorf("compute client not configured")
	}

	wc.logger.Debug("querying compute execution status",
		zap.String("execution_id", executionID),
	)

	return wc.computeClient.GetComputeExecutionStatus(ctx, &workflow.GetExecutionStatusInput{
		ExecutionID: executionID,
	})
}
