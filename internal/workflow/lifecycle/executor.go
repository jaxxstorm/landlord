// Package lifecycle contains payload-driven tenant lifecycle execution shared
// by workflow backends.
package lifecycle

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jaxxstorm/landlord/internal/compute"
	"github.com/jaxxstorm/landlord/internal/workflow"
	"go.uber.org/zap"
)

// Executor performs tenant lifecycle operations through registered compute
// providers without requiring tenant state from a database.
type Executor struct {
	computeRegistry *compute.Registry
	computeResolver workflow.ComputeProviderResolver
	computeManager  workflow.ComputeManager
	mutationGuard   MutationGuard
	providerType    string
	logger          *zap.Logger
}

// MutationGuard checks whether a lifecycle mutation may proceed.
type MutationGuard interface {
	BeforeMutation(context.Context, *workflow.ProvisionRequest) error
}

// WithComputeManager enables tracked compute execution for requests carrying a
// workflow_execution_id in their metadata.
func (e *Executor) WithComputeManager(computeManager workflow.ComputeManager) *Executor {
	e.computeManager = computeManager
	return e
}

// WithMutationGuard checks for cooperative cancellation before each mutation.
func (e *Executor) WithMutationGuard(mutationGuard MutationGuard) *Executor {
	e.mutationGuard = mutationGuard
	return e
}

// NewExecutor creates a lifecycle executor for one workflow backend.
func NewExecutor(computeRegistry *compute.Registry, computeResolver workflow.ComputeProviderResolver, providerType string, logger *zap.Logger) (*Executor, error) {
	if computeRegistry == nil {
		return nil, fmt.Errorf("compute registry is required")
	}
	if providerType == "" {
		return nil, fmt.Errorf("workflow provider type is required")
	}
	return &Executor{
		computeRegistry: computeRegistry,
		computeResolver: computeResolver,
		providerType:    providerType,
		logger:          logger.With(zap.String("component", "lifecycle-executor"), zap.String("provider", providerType)),
	}, nil
}

// Execute runs a provision, update, delete, or plan lifecycle operation from
// the request payload.
func (e *Executor) Execute(ctx context.Context, req *workflow.ProvisionRequest) (*workflow.ExecutionStatus, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	tenantID := req.TenantUUID
	if tenantID == "" {
		tenantID = req.TenantID
	}
	if tenantID == "" {
		return nil, fmt.Errorf("tenant identifier is required")
	}

	op := req.Operation
	if op == "" {
		op = "provision"
	}

	switch op {
	case "plan":
		return e.succeededStatus("plan", tenantID, map[string]string{"status": "planned", "tenant_id": tenantID})
	case "create", "apply", "provision":
		return e.provision(ctx, tenantID, req)
	case "update":
		return e.update(ctx, tenantID, req)
	case "delete", "destroy":
		return e.destroy(ctx, tenantID, req)
	default:
		return nil, fmt.Errorf("unknown operation: %s", op)
	}
}

func (e *Executor) provision(ctx context.Context, tenantID string, req *workflow.ProvisionRequest) (*workflow.ExecutionStatus, error) {
	provider, providerType, err := e.resolveComputeProvider(ctx, req)
	if err != nil {
		return nil, err
	}
	spec := buildComputeSpec(tenantID, providerType, req.DesiredConfig)
	if err := e.checkMutation(ctx, req); err != nil {
		return nil, err
	}
	if workflowExecutionID := trackedWorkflowExecutionID(req); e.computeManager != nil && workflowExecutionID != "" {
		execution, err := e.computeManager.ProvisionTenantWithTracking(ctx, spec, workflowExecutionID)
		if err != nil {
			return nil, fmt.Errorf("tracked compute provisioning failed: %w", err)
		}
		return e.trackedStatus("provision", tenantID, execution)
	}

	result, err := provider.Provision(ctx, spec)
	if err != nil {
		if status, statusErr := provider.GetStatus(ctx, tenantID); statusErr == nil {
			return e.succeededStatus("provision", tenantID, status)
		}
		return nil, fmt.Errorf("compute provisioning failed: %w", err)
	}
	return e.succeededStatus("provision", tenantID, result)
}

func (e *Executor) update(ctx context.Context, tenantID string, req *workflow.ProvisionRequest) (*workflow.ExecutionStatus, error) {
	provider, providerType, err := e.resolveComputeProvider(ctx, req)
	if err != nil {
		return nil, err
	}
	spec := buildComputeSpec(tenantID, providerType, req.DesiredConfig)
	if err := e.checkMutation(ctx, req); err != nil {
		return nil, err
	}
	if workflowExecutionID := trackedWorkflowExecutionID(req); e.computeManager != nil && workflowExecutionID != "" {
		execution, err := e.computeManager.UpdateTenantWithTracking(ctx, tenantID, spec, workflowExecutionID)
		if err != nil {
			return nil, fmt.Errorf("tracked compute update failed: %w", err)
		}
		return e.trackedStatus("update", tenantID, execution)
	}

	result, err := provider.Update(ctx, tenantID, spec)
	if err != nil {
		return nil, fmt.Errorf("compute update failed: %w", err)
	}
	return e.succeededStatus("update", tenantID, result)
}

func (e *Executor) destroy(ctx context.Context, tenantID string, req *workflow.ProvisionRequest) (*workflow.ExecutionStatus, error) {
	provider, providerType, err := e.resolveComputeProvider(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := e.checkMutation(ctx, req); err != nil {
		return nil, err
	}
	if workflowExecutionID := trackedWorkflowExecutionID(req); e.computeManager != nil && workflowExecutionID != "" {
		execution, err := e.computeManager.DeleteTenantWithTracking(ctx, tenantID, providerType, workflowExecutionID)
		if err != nil {
			return nil, fmt.Errorf("tracked compute deprovisioning failed: %w", err)
		}
		return e.trackedStatus("delete", tenantID, execution)
	}

	if err := provider.Destroy(ctx, tenantID); err != nil && !errors.Is(err, compute.ErrTenantNotFound) {
		return nil, fmt.Errorf("compute deprovisioning failed: %w", err)
	}
	return e.succeededStatus("delete", tenantID, map[string]string{"status": "archived", "tenant_id": tenantID})
}

func (e *Executor) checkMutation(ctx context.Context, req *workflow.ProvisionRequest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if e.mutationGuard != nil {
		return e.mutationGuard.BeforeMutation(ctx, req)
	}
	return nil
}

func trackedWorkflowExecutionID(req *workflow.ProvisionRequest) string {
	if req.Metadata == nil {
		return ""
	}
	return req.Metadata["workflow_execution_id"]
}

func (e *Executor) resolveComputeProvider(ctx context.Context, req *workflow.ProvisionRequest) (compute.Provider, string, error) {
	providerType := req.ComputeProvider
	if providerType == "" && e.computeResolver != nil {
		resolved, err := e.computeResolver.ResolveProvider(ctx, req.TenantID, req.TenantUUID)
		if err != nil {
			e.logger.Warn("failed to resolve compute provider", zap.Error(err))
		} else if resolved != "" {
			providerType = resolved
		}
	}
	if providerType == "" {
		providers := e.computeRegistry.List()
		if len(providers) == 1 {
			providerType = providers[0]
		}
	}
	if providerType == "" {
		return nil, "", fmt.Errorf("compute provider not specified")
	}

	provider, err := e.computeRegistry.Get(providerType)
	if err != nil {
		return nil, "", fmt.Errorf("compute provider lookup failed: %w", err)
	}
	return provider, providerType, nil
}

func buildComputeSpec(tenantID, providerType string, desiredConfig map[string]interface{}) *compute.TenantComputeSpec {
	spec := &compute.TenantComputeSpec{TenantID: tenantID, ProviderType: providerType}
	if providerType == "docker" {
		spec.Containers = []compute.ContainerSpec{{Name: "app"}}
	}
	if len(desiredConfig) > 0 {
		if raw, err := json.Marshal(desiredConfig); err == nil {
			spec.ProviderConfig = raw
		}
	}
	return spec
}

func (e *Executor) succeededStatus(operation, tenantID string, payload interface{}) (*workflow.ExecutionStatus, error) {
	output, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal output: %w", err)
	}
	return &workflow.ExecutionStatus{
		ExecutionID:  fmt.Sprintf("%s-%s", operation, tenantID),
		ProviderType: e.providerType,
		State:        workflow.StateSucceeded,
		Output:       output,
	}, nil
}

func (e *Executor) trackedStatus(operation, tenantID string, execution *compute.ComputeExecution) (*workflow.ExecutionStatus, error) {
	if execution == nil {
		return nil, fmt.Errorf("tracked compute operation returned no execution")
	}
	payload := map[string]interface{}{
		"compute_execution_id": execution.ExecutionID,
		"compute_status":       execution.Status,
		"tenant_id":            tenantID,
	}
	if len(execution.ResourceIDs) > 0 {
		var resourceIDs map[string]interface{}
		if err := json.Unmarshal(execution.ResourceIDs, &resourceIDs); err != nil {
			return nil, fmt.Errorf("unmarshal tracked resource IDs: %w", err)
		}
		payload["resource_ids"] = resourceIDs
	}
	return e.succeededStatus(operation, tenantID, payload)
}
