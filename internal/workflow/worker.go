package workflow

import "context"

// WorkerEngine defines the interface for workflow worker engines.
type WorkerEngine interface {
	// Name returns the unique worker engine identifier.
	Name() string

	// Register performs any required worker registration with the workflow backend.
	Register(ctx context.Context) error

	// Start runs the worker engine until the context is canceled.
	Start(ctx context.Context, addr string) error
}

// WorkerJobPayload is the standard job payload passed to workflow workers.
type WorkerJobPayload = ProvisionRequest

// ComputeProviderResolver resolves the compute provider for a tenant.
type ComputeProviderResolver interface {
	ResolveProvider(ctx context.Context, tenantID, tenantUUID string) (string, error)
}
