package workflow

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/jaxxstorm/landlord/internal/compute"
)

// ComputeWorkflowClient provides compute provisioning operations accessible to workflows
type ComputeWorkflowClient struct {
	computeManager ComputeManager
	logger         *zap.Logger
}

// NewComputeWorkflowClient creates a new workflow-accessible compute client
func NewComputeWorkflowClient(computeManager ComputeManager, logger *zap.Logger) *ComputeWorkflowClient {
	return &ComputeWorkflowClient{
		computeManager: computeManager,
		logger:         logger.With(zap.String("component", "compute-workflow-client")),
	}
}

// ProvisionTenantInput represents input for a provision operation called from a workflow
type ProvisionTenantInput struct {
	Spec                *compute.TenantComputeSpec `json:"spec"`
	WorkflowExecutionID string                     `json:"workflow_execution_id"`
}

// ProvisionTenant provisions compute resources from a workflow context
func (c *ComputeWorkflowClient) ProvisionTenant(ctx context.Context, input *ProvisionTenantInput) (*ProvisionTenantOutput, error) {
	c.logger.Info("workflow provisioning tenant",
		zap.String("tenant_id", input.Spec.TenantID),
		zap.String("workflow_execution_id", input.WorkflowExecutionID),
	)

	exec, err := c.computeManager.ProvisionTenantWithTracking(ctx, input.Spec, input.WorkflowExecutionID)
	if err != nil {
		compErr := c.computeManager.MapProviderErrorToComputeError(err)
		return &ProvisionTenantOutput{
			ExecutionID: exec.ExecutionID,
			Error:       compErr,
		}, err
	}

	// Marshal resource IDs
	var resourceIDs map[string]interface{}
	if exec.ResourceIDs != nil {
		_ = json.Unmarshal(exec.ResourceIDs, &resourceIDs)
	}

	return &ProvisionTenantOutput{
		ExecutionID: exec.ExecutionID,
		Status:      exec.Status,
		ResourceIDs: resourceIDs,
	}, nil
}

// ProvisionTenantOutput represents the result of a provision operation
type ProvisionTenantOutput struct {
	ExecutionID string                         `json:"execution_id"`
	Status      compute.ComputeExecutionStatus `json:"status"`
	ResourceIDs map[string]interface{}         `json:"resource_ids,omitempty"`
	Error       *compute.ComputeError          `json:"error,omitempty"`
}

// UpdateTenantInput represents input for an update operation
type UpdateTenantInput struct {
	TenantID            string
	Spec                *compute.TenantComputeSpec
	WorkflowExecutionID string
}

// UpdateTenant updates compute resources from a workflow context
func (c *ComputeWorkflowClient) UpdateTenant(ctx context.Context, input *UpdateTenantInput) (*UpdateTenantOutput, error) {
	c.logger.Info("workflow updating tenant",
		zap.String("tenant_id", input.TenantID),
		zap.String("workflow_execution_id", input.WorkflowExecutionID),
	)

	exec, err := c.computeManager.UpdateTenantWithTracking(ctx, input.TenantID, input.Spec, input.WorkflowExecutionID)
	if err != nil {
		compErr := c.computeManager.MapProviderErrorToComputeError(err)
		return &UpdateTenantOutput{
			ExecutionID: exec.ExecutionID,
			Error:       compErr,
		}, err
	}

	var resourceIDs map[string]interface{}
	if exec.ResourceIDs != nil {
		_ = json.Unmarshal(exec.ResourceIDs, &resourceIDs)
	}

	return &UpdateTenantOutput{
		ExecutionID: exec.ExecutionID,
		Status:      exec.Status,
		ResourceIDs: resourceIDs,
	}, nil
}

// UpdateTenantOutput represents the result of an update operation
type UpdateTenantOutput struct {
	ExecutionID string                         `json:"execution_id"`
	Status      compute.ComputeExecutionStatus `json:"status"`
	ResourceIDs map[string]interface{}         `json:"resource_ids,omitempty"`
	Error       *compute.ComputeError          `json:"error,omitempty"`
}

// DeleteTenantInput represents input for a delete operation
type DeleteTenantInput struct {
	TenantID            string
	ProviderType        string
	WorkflowExecutionID string
}

// DeleteTenant deletes compute resources from a workflow context
func (c *ComputeWorkflowClient) DeleteTenant(ctx context.Context, input *DeleteTenantInput) (*DeleteTenantOutput, error) {
	c.logger.Info("workflow deleting tenant",
		zap.String("tenant_id", input.TenantID),
		zap.String("workflow_execution_id", input.WorkflowExecutionID),
	)

	exec, err := c.computeManager.DeleteTenantWithTracking(ctx, input.TenantID, input.ProviderType, input.WorkflowExecutionID)
	if err != nil {
		compErr := c.computeManager.MapProviderErrorToComputeError(err)
		return &DeleteTenantOutput{
			ExecutionID: exec.ExecutionID,
			Error:       compErr,
		}, err
	}

	return &DeleteTenantOutput{
		ExecutionID: exec.ExecutionID,
		Status:      exec.Status,
	}, nil
}

// DeleteTenantOutput represents the result of a delete operation
type DeleteTenantOutput struct {
	ExecutionID string                         `json:"execution_id"`
	Status      compute.ComputeExecutionStatus `json:"status"`
	Error       *compute.ComputeError          `json:"error,omitempty"`
}

// GetExecutionStatusInput represents input for a status query
type GetExecutionStatusInput struct {
	ExecutionID string `json:"execution_id"`
}

// GetComputeExecutionStatus retrieves the current status of a compute execution
func (c *ComputeWorkflowClient) GetComputeExecutionStatus(ctx context.Context, input *GetExecutionStatusInput) (*ComputeExecutionStatusOutput, error) {
	exec, err := c.computeManager.GetComputeExecution(ctx, input.ExecutionID)
	if err != nil {
		c.logger.Error("failed to get execution status",
			zap.String("execution_id", input.ExecutionID),
			zap.Error(err),
		)
		return nil, err
	}

	var resourceIDs map[string]interface{}
	if exec.ResourceIDs != nil {
		_ = json.Unmarshal(exec.ResourceIDs, &resourceIDs)
	}

	output := &ComputeExecutionStatusOutput{
		ExecutionID:   exec.ExecutionID,
		TenantID:      exec.TenantID,
		OperationType: exec.OperationType,
		Status:        exec.Status,
		ResourceIDs:   resourceIDs,
		CreatedAt:     exec.CreatedAt.Unix(),
		UpdatedAt:     exec.UpdatedAt.Unix(),
	}

	if exec.ErrorCode != nil {
		output.ErrorCode = *exec.ErrorCode
	}
	if exec.ErrorMessage != nil {
		output.ErrorMessage = *exec.ErrorMessage
	}

	return output, nil
}

// ComputeExecutionStatusOutput represents execution status
type ComputeExecutionStatusOutput struct {
	ExecutionID   string                         `json:"execution_id"`
	TenantID      string                         `json:"tenant_id"`
	OperationType compute.ComputeOperationType   `json:"operation_type"`
	Status        compute.ComputeExecutionStatus `json:"status"`
	ResourceIDs   map[string]interface{}         `json:"resource_ids,omitempty"`
	ErrorCode     string                         `json:"error_code,omitempty"`
	ErrorMessage  string                         `json:"error_message,omitempty"`
	CreatedAt     int64                          `json:"created_at"`
	UpdatedAt     int64                          `json:"updated_at"`
}
