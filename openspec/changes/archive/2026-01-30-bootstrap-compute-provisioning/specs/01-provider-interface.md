# Spec: Compute Provider Interface

## Overview

Defines the contract that all compute providers must implement to integrate with the landlord control plane.

## Interface Definition

```go
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
}
```

## Method Contracts

### Name()
- **MUST** return a lowercase, alphanumeric identifier
- **MUST** be unique across all registered providers
- **MUST** match the value in `TenantComputeSpec.ProviderType`

### Provision()
- **MUST** be idempotent when called with same spec
- **MUST** create all resources atomically or roll back on failure
- **MUST** return resource identifiers for cleanup
- **SHOULD** validate spec before provisioning
- **MAY** take extended time (minutes) to complete
- Context cancellation **SHOULD** trigger cleanup

### Update()
- **MUST** be idempotent - same spec = no changes
- **MUST** not cause downtime if possible
- **MUST** roll back on failure to previous state
- **SHOULD** return list of changes made
- **MAY** be implemented as destroy-then-provision

### Destroy()
- **MUST** be idempotent - calling multiple times safe
- **MUST** clean up ALL resources created during Provision
- **MUST NOT** error if tenant doesn't exist
- **SHOULD** log but not fail if resources partially cleaned
- Context cancellation **SHOULD NOT** stop cleanup

### GetStatus()
- **MUST** return current state, not cached state
- **MUST** return ErrTenantNotFound if tenant doesn't exist
- **SHOULD** be fast (< 5 seconds typical)
- **SHOULD** include health information

### Validate()
- **MUST** check spec without side effects
- **MUST** validate provider-specific config
- **SHOULD** check resource limits
- **SHOULD** validate image names/availability
- **MAY** make read-only API calls to provider

## Error Handling

Providers **MUST** return these error types:

```go
var (
    ErrInvalidSpec     = errors.New("invalid compute specification")
    ErrProvisionFailed = errors.New("provisioning failed")
    ErrTenantNotFound  = errors.New("tenant compute resources not found")
    ErrUpdateFailed    = errors.New("update failed")
)
```

Providers **MAY** wrap errors with additional context:
```go
return fmt.Errorf("%w: %s", ErrProvisionFailed, details)
```

## Context Handling

All methods receive a `context.Context`:
- **MUST** respect context cancellation
- **SHOULD** respect context timeouts
- **SHOULD** propagate context to downstream calls

## Thread Safety

Provider implementations:
- **MUST** be safe for concurrent use
- **MAY** use internal locking if needed
- **SHOULD** allow parallel operations on different tenants

## Logging

Providers **SHOULD**:
- Log at INFO level for successful operations
- Log at WARN level for retryable failures
- Log at ERROR level for fatal errors
- Include tenantID in all log entries
- Never log secrets or credentials

## Testing Requirements

Each provider implementation **MUST** include:
- Unit tests for Validate()
- Interface compliance test
- Error handling tests
- Idempotency tests

## Example Implementation Skeleton

```go
package example

import (
    "context"
    "github.com/jaxxstorm/landlord/internal/compute"
    "go.uber.org/zap"
)

type Provider struct {
    config *Config
    logger *zap.Logger
}

func NewProvider(config *Config, logger *zap.Logger) *Provider {
    return &Provider{
        config: config,
        logger: logger.With(zap.String("provider", "example")),
    }
}

func (p *Provider) Name() string {
    return "example"
}

func (p *Provider) Provision(ctx context.Context, spec *compute.TenantComputeSpec) (*compute.ProvisionResult, error) {
    p.logger.Info("provisioning tenant", zap.String("tenant_id", spec.TenantID))
    
    if err := p.Validate(ctx, spec); err != nil {
        return nil, err
    }
    
    // Provider-specific provisioning logic here
    
    return &compute.ProvisionResult{
        TenantID:     spec.TenantID,
        ProviderType: p.Name(),
        Status:       compute.ProvisionStatusSuccess,
    }, nil
}

// ... implement other methods
```
