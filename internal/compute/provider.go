package compute

import (
	"context"
	"encoding/json"
)

// Provider defines the interface for compute provisioning implementations
type Provider interface {
	// Name returns the unique identifier for this provider
	// Examples: "ecs", "kubernetes", "nomad"
	Name() string

	// Provision creates new compute resources for a tenant
	// Returns ProvisionResult with resource IDs and endpoints
	// Returns error if provisioning fails
	Provision(ctx context.Context, spec *TenantComputeSpec) (*ProvisionResult, error)

	// Update modifies existing compute resources for a tenant
	// Must be idempotent - calling with same spec should not cause changes
	// Returns UpdateResult with changes made
	Update(ctx context.Context, tenantID string, spec *TenantComputeSpec) (*UpdateResult, error)

	// Destroy removes all compute resources for a tenant
	// Must be idempotent - calling on non-existent tenant should not error
	// Returns error only for unexpected failures
	Destroy(ctx context.Context, tenantID string) error

	// GetStatus queries the current state of tenant compute resources
	// Returns ComputeStatus with current state
	// Returns ErrTenantNotFound if tenant doesn't exist
	GetStatus(ctx context.Context, tenantID string) (*ComputeStatus, error)

	// Validate checks if a compute spec is valid for this provider
	// Does not provision - only validates the specification
	// Returns error describing what's invalid
	Validate(ctx context.Context, spec *TenantComputeSpec) error

	// ValidateConfig validates provider-specific configuration
	// Used by API layer to validate compute_config before storage
	// Returns detailed error with field names and constraints if invalid
	ValidateConfig(config json.RawMessage) error

	// ConfigSchema returns the JSON Schema (draft 2020-12) for compute_config.
	// Returned schema should be a valid JSON object.
	ConfigSchema() json.RawMessage

	// ConfigDefaults returns an optional default compute_config payload.
	// Return nil when no defaults are available.
	ConfigDefaults() json.RawMessage
}
