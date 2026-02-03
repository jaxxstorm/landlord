package restate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jaxxstorm/landlord/internal/compute"
	"github.com/jaxxstorm/landlord/internal/config"
	"github.com/jaxxstorm/landlord/internal/workflow"
	restate "github.com/restatedev/sdk-go"
	"github.com/restatedev/sdk-go/server"
	"go.uber.org/zap"
)

// TenantProvisioningService defines the Restate service for tenant lifecycle workflows.
type TenantProvisioningService struct {
	computeRegistry        *compute.Registry
	defaultComputeProvider string
	computeResolver        workflow.ComputeProviderResolver
	logger                 *zap.Logger
}

// ProvisioningRequest represents the input for lifecycle operations.
type ProvisioningRequest = workflow.ProvisionRequest

// NewTenantProvisioningService creates a new tenant lifecycle service.
func NewTenantProvisioningService(
	computeRegistry *compute.Registry,
	defaultComputeProvider string,
	computeResolver workflow.ComputeProviderResolver,
	logger *zap.Logger,
) *TenantProvisioningService {
	return &TenantProvisioningService{
		computeRegistry:        computeRegistry,
		defaultComputeProvider: defaultComputeProvider,
		computeResolver:        computeResolver,
		logger:                 logger.With(zap.String("component", "tenant-provisioning-service")),
	}
}

// Execute handles tenant lifecycle operations.
func (s *TenantProvisioningService) Execute(ctx context.Context, req *ProvisioningRequest) (*workflow.ExecutionStatus, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}

	tenantID := req.TenantUUID
	if tenantID == "" {
		tenantID = req.TenantID
	}
	if tenantID == "" {
		return nil, fmt.Errorf("tenant identifier is required")
	}

	operation := req.Operation
	if operation == "" {
		operation = "provision"
	}

	s.logger.Info("executing tenant workflow",
		zap.String("tenant_id", tenantID),
		zap.String("tenant_name", req.TenantID),
		zap.String("operation", operation),
	)

	switch operation {
	case "plan":
		return s.plan(tenantID)
	case "create", "apply", "provision":
		return s.provision(ctx, tenantID, req)
	case "destroy", "delete":
		return s.destroy(ctx, tenantID, req)
	case "update":
		return s.update(ctx, tenantID, req)
	default:
		return nil, fmt.Errorf("unknown operation: %s", req.Operation)
	}
}

func (s *TenantProvisioningService) plan(tenantID string) (*workflow.ExecutionStatus, error) {
	output, err := json.Marshal(map[string]string{
		"status":    "planned",
		"tenant_id": tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal output: %w", err)
	}

	return &workflow.ExecutionStatus{
		ExecutionID:  fmt.Sprintf("plan-%s", tenantID),
		ProviderType: "restate",
		State:        workflow.StateSucceeded,
		Output:       output,
	}, nil
}

func (s *TenantProvisioningService) provision(ctx context.Context, tenantID string, req *ProvisioningRequest) (*workflow.ExecutionStatus, error) {
	computeProvider, providerType, err := s.resolveComputeProvider(ctx, req)
	if err != nil {
		return nil, err
	}

	spec := buildComputeSpec(tenantID, providerType, req.DesiredConfig)
	result, err := computeProvider.Provision(ctx, spec)
	if err != nil {
		if status, statusErr := computeProvider.GetStatus(ctx, tenantID); statusErr == nil {
			output, marshalErr := json.Marshal(status)
			if marshalErr != nil {
				return nil, fmt.Errorf("marshal output: %w", marshalErr)
			}
			return &workflow.ExecutionStatus{
				ExecutionID:  fmt.Sprintf("provision-%s", tenantID),
				ProviderType: "restate",
				State:        workflow.StateSucceeded,
				Output:       output,
			}, nil
		}
		s.logger.Error("compute provisioning failed", zap.Error(err))
		return nil, fmt.Errorf("compute provisioning failed: %w", err)
	}

	output, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("marshal output: %w", err)
	}

	return &workflow.ExecutionStatus{
		ExecutionID:  fmt.Sprintf("provision-%s", tenantID),
		ProviderType: "restate",
		State:        workflow.StateSucceeded,
		Output:       output,
	}, nil
}

func (s *TenantProvisioningService) destroy(ctx context.Context, tenantID string, req *ProvisioningRequest) (*workflow.ExecutionStatus, error) {
	computeProvider, _, err := s.resolveComputeProvider(ctx, req)
	if err != nil {
		return nil, err
	}

	if err := computeProvider.Destroy(ctx, tenantID); err != nil {
		if errors.Is(err, compute.ErrTenantNotFound) {
			s.logger.Info("compute resources already removed", zap.String("tenant_id", tenantID))
		} else {
			s.logger.Error("compute deprovisioning failed", zap.Error(err))
			return nil, fmt.Errorf("compute deprovisioning failed: %w", err)
		}
	}

	output, err := json.Marshal(map[string]string{
		"status":    "archived",
		"tenant_id": tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal output: %w", err)
	}

	return &workflow.ExecutionStatus{
		ExecutionID:  fmt.Sprintf("delete-%s", tenantID),
		ProviderType: "restate",
		State:        workflow.StateSucceeded,
		Output:       output,
	}, nil
}

func (s *TenantProvisioningService) update(ctx context.Context, tenantID string, req *ProvisioningRequest) (*workflow.ExecutionStatus, error) {
	computeProvider, providerType, err := s.resolveComputeProvider(ctx, req)
	if err != nil {
		return nil, err
	}

	spec := buildComputeSpec(tenantID, providerType, req.DesiredConfig)
	result, err := computeProvider.Update(ctx, tenantID, spec)
	if err != nil {
		s.logger.Error("compute update failed", zap.Error(err))
		return nil, fmt.Errorf("compute update failed: %w", err)
	}

	output, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("marshal output: %w", err)
	}

	return &workflow.ExecutionStatus{
		ExecutionID:  fmt.Sprintf("update-%s", tenantID),
		ProviderType: "restate",
		State:        workflow.StateSucceeded,
		Output:       output,
	}, nil
}

func (s *TenantProvisioningService) resolveComputeProvider(ctx context.Context, req *ProvisioningRequest) (compute.Provider, string, error) {
	providerType := req.ComputeProvider
	if providerType == "" && s.computeResolver != nil {
		resolved, err := s.computeResolver.ResolveProvider(ctx, req.TenantID, req.TenantUUID)
		if err != nil {
			s.logger.Warn("failed to resolve compute provider", zap.Error(err))
		} else if resolved != "" {
			providerType = resolved
		}
	}
	if providerType == "" {
		providerType = s.defaultComputeProvider
	}
	if providerType == "" {
		return nil, "", fmt.Errorf("compute provider not specified")
	}
	provider, err := s.computeRegistry.Get(providerType)
	if err != nil {
		return nil, "", fmt.Errorf("compute provider lookup failed: %w", err)
	}
	return provider, providerType, nil
}

func buildComputeSpec(tenantID, providerType string, desiredConfig map[string]interface{}) *compute.TenantComputeSpec {
	spec := &compute.TenantComputeSpec{
		TenantID:     tenantID,
		ProviderType: providerType,
	}

	if providerType == "docker" {
		spec.Containers = []compute.ContainerSpec{
			{
				Name: "app",
			},
		}
	}

	if len(desiredConfig) > 0 {
		if raw, err := json.Marshal(desiredConfig); err == nil {
			spec.ProviderConfig = raw
		}
	}

	return spec
}

// RegisterService registers the tenant provisioning service with Restate.
func (s *TenantProvisioningService) RegisterService(ctx context.Context, client *Client, serviceName string) error {
	if serviceName == "" {
		serviceName = workflowServiceName(config.RestateConfig{}, tenantProvisioningWorkflowID)
	}
	return client.RegisterService(ctx, serviceName)
}

// Bind registers the tenant provisioning handlers with a Restate server.
func (s *TenantProvisioningService) Bind(server *server.Restate, serviceName string) {
	if serviceName == "" {
		serviceName = workflowServiceName(config.RestateConfig{}, tenantProvisioningWorkflowID)
	}

	server.Bind(
		restate.NewService(serviceName).
			Handler("execute", restate.NewServiceHandler(func(_ restate.Context, req ProvisioningRequest) (workflow.ExecutionStatus, error) {
				status, err := s.Execute(context.Background(), &req)
				if err != nil {
					return workflow.ExecutionStatus{}, err
				}
				if status == nil {
					return workflow.ExecutionStatus{}, nil
				}
				return *status, nil
			})),
	)
}
