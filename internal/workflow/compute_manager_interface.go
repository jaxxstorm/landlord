package workflow

import (
	"context"

	"github.com/jaxxstorm/landlord/internal/compute"
)

// ComputeManager defines the interface for compute operations accessible to workflows
type ComputeManager interface {
	// ProvisionTenantWithTracking provisions compute resources with execution tracking
	ProvisionTenantWithTracking(ctx context.Context, spec *compute.TenantComputeSpec, workflowExecutionID string) (*compute.ComputeExecution, error)

	// UpdateTenantWithTracking updates compute resources with execution tracking
	UpdateTenantWithTracking(ctx context.Context, tenantID string, spec *compute.TenantComputeSpec, workflowExecutionID string) (*compute.ComputeExecution, error)

	// DeleteTenantWithTracking deletes compute resources with execution tracking
	DeleteTenantWithTracking(ctx context.Context, tenantID, providerType string, workflowExecutionID string) (*compute.ComputeExecution, error)

	// GetComputeExecution retrieves execution details
	GetComputeExecution(ctx context.Context, executionID string) (*compute.ComputeExecution, error)

	// ListProviders returns the list of available compute providers
	ListProviders() []string

	// Health performs a health check
	Health(ctx context.Context) error

	// MapProviderErrorToComputeError maps provider errors to standard compute errors
	MapProviderErrorToComputeError(err error) *compute.ComputeError
}
