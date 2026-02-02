# Spec: Provider Registry and Compute Manager

## Overview

Defines the registry for managing providers and the manager facade for coordinating compute operations.

## Provider Registry

### Purpose
- Register available compute providers
- Lookup providers by type
- Prevent duplicate registrations
- List available providers

### Interface

```go
package compute

import (
    "fmt"
    "sync"
    
    "go.uber.org/zap"
)

// Registry manages registered compute providers
type Registry struct {
    providers map[string]Provider
    mu        sync.RWMutex
    logger    *zap.Logger
}

// NewRegistry creates a new provider registry
func NewRegistry(logger *zap.Logger) *Registry {
    return &Registry{
        providers: make(map[string]Provider),
        logger:    logger.With(zap.String("component", "compute-registry")),
    }
}

// Register adds a provider to the registry
// Returns ErrProviderConflict if provider name already registered
func (r *Registry) Register(provider Provider) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    name := provider.Name()
    if _, exists := r.providers[name]; exists {
        return fmt.Errorf("%w: %s", ErrProviderConflict, name)
    }
    
    r.providers[name] = provider
    r.logger.Info("registered compute provider", zap.String("provider", name))
    return nil
}

// Get retrieves a provider by name
// Returns ErrProviderNotFound if provider not registered
func (r *Registry) Get(providerType string) (Provider, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    provider, exists := r.providers[providerType]
    if !exists {
        return nil, fmt.Errorf("%w: %s", ErrProviderNotFound, providerType)
    }
    
    return provider, nil
}

// List returns names of all registered providers
func (r *Registry) List() []string {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    names := make([]string, 0, len(r.providers))
    for name := range r.providers {
        names = append(names, name)
    }
    return names
}

// Has checks if a provider is registered
func (r *Registry) Has(providerType string) bool {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    _, exists := r.providers[providerType]
    return exists
}
```

### Behavior Requirements

#### Register()
- **MUST** be thread-safe
- **MUST** reject duplicate provider names
- **SHOULD** log registration events
- **MUST** validate provider.Name() is non-empty

#### Get()
- **MUST** be thread-safe
- **MUST** return error for unregistered providers
- **SHOULD** be fast (read-only lock)

#### List()
- **MUST** be thread-safe
- **MUST** return copy, not reference to internal map
- **MAY** return providers in any order

## Compute Manager

### Purpose
- Facade for compute operations
- Route requests to appropriate providers
- Handle cross-cutting concerns (logging, metrics)
- Validate specs before delegation

### Interface

```go
package compute

import (
    "context"
    "fmt"
    
    "go.uber.org/zap"
)

// Manager coordinates compute provisioning operations
type Manager struct {
    registry *Registry
    logger   *zap.Logger
}

// New creates a new compute manager
func New(registry *Registry, logger *zap.Logger) *Manager {
    return &Manager{
        registry: registry,
        logger:   logger.With(zap.String("component", "compute-manager")),
    }
}

// ProvisionTenant provisions compute resources for a tenant
func (m *Manager) ProvisionTenant(ctx context.Context, spec *TenantComputeSpec) (*ProvisionResult, error) {
    m.logger.Info("provisioning tenant",
        zap.String("tenant_id", spec.TenantID),
        zap.String("provider", spec.ProviderType),
    )
    
    // Validate spec
    if err := ValidateComputeSpec(spec); err != nil {
        return nil, fmt.Errorf("%w: %s", ErrInvalidSpec, err)
    }
    
    // Get provider
    provider, err := m.registry.Get(spec.ProviderType)
    if err != nil {
        return nil, err
    }
    
    // Delegate to provider
    result, err := provider.Provision(ctx, spec)
    if err != nil {
        m.logger.Error("provisioning failed",
            zap.String("tenant_id", spec.TenantID),
            zap.Error(err),
        )
        return nil, err
    }
    
    m.logger.Info("provisioning completed",
        zap.String("tenant_id", spec.TenantID),
        zap.String("status", string(result.Status)),
    )
    
    return result, nil
}

// UpdateTenant updates existing compute resources
func (m *Manager) UpdateTenant(ctx context.Context, tenantID string, spec *TenantComputeSpec) (*UpdateResult, error) {
    m.logger.Info("updating tenant",
        zap.String("tenant_id", tenantID),
        zap.String("provider", spec.ProviderType),
    )
    
    // Validate spec
    if err := ValidateComputeSpec(spec); err != nil {
        return nil, fmt.Errorf("%w: %s", ErrInvalidSpec, err)
    }
    
    // Get provider
    provider, err := m.registry.Get(spec.ProviderType)
    if err != nil {
        return nil, err
    }
    
    // Delegate to provider
    result, err := provider.Update(ctx, tenantID, spec)
    if err != nil {
        m.logger.Error("update failed",
            zap.String("tenant_id", tenantID),
            zap.Error(err),
        )
        return nil, err
    }
    
    m.logger.Info("update completed",
        zap.String("tenant_id", tenantID),
        zap.String("status", string(result.Status)),
    )
    
    return result, nil
}

// DestroyTenant removes compute resources for a tenant
func (m *Manager) DestroyTenant(ctx context.Context, tenantID, providerType string) error {
    m.logger.Info("destroying tenant",
        zap.String("tenant_id", tenantID),
        zap.String("provider", providerType),
    )
    
    // Get provider
    provider, err := m.registry.Get(providerType)
    if err != nil {
        return err
    }
    
    // Delegate to provider
    if err := provider.Destroy(ctx, tenantID); err != nil {
        m.logger.Error("destroy failed",
            zap.String("tenant_id", tenantID),
            zap.Error(err),
        )
        return err
    }
    
    m.logger.Info("tenant destroyed",
        zap.String("tenant_id", tenantID),
    )
    
    return nil
}

// GetTenantStatus queries current status of tenant compute
func (m *Manager) GetTenantStatus(ctx context.Context, tenantID, providerType string) (*ComputeStatus, error) {
    // Get provider
    provider, err := m.registry.Get(providerType)
    if err != nil {
        return nil, err
    }
    
    // Delegate to provider
    status, err := provider.GetStatus(ctx, tenantID)
    if err != nil {
        return nil, err
    }
    
    return status, nil
}

// ValidateTenantSpec validates a spec without provisioning
func (m *Manager) ValidateTenantSpec(ctx context.Context, spec *TenantComputeSpec) error {
    // Validate structure
    if err := ValidateComputeSpec(spec); err != nil {
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
```

### Behavior Requirements

#### ProvisionTenant()
- **MUST** validate spec before delegating
- **MUST** log start and completion
- **MUST** return provider errors unchanged
- **SHOULD** include timing metrics

#### UpdateTenant()
- **MUST** validate spec before delegating  
- **MUST** ensure tenantID matches spec.TenantID
- **MUST** log changes

#### DestroyTenant()
- **MUST** log destruction events
- **MUST** propagate provider errors
- **SHOULD** not fail if tenant doesn't exist

#### GetTenantStatus()
- **SHOULD** cache status briefly (future enhancement)
- **MUST** return provider errors unchanged
- **SHOULD** be fast

## Validation Functions

```go
// ValidateComputeSpec performs structural validation
func ValidateComputeSpec(spec *TenantComputeSpec) error {
    if spec.TenantID == "" {
        return errors.New("tenant_id required")
    }
    
    if spec.ProviderType == "" {
        return errors.New("provider_type required")
    }
    
    if len(spec.Containers) == 0 {
        return errors.New("at least one container required")
    }
    
    // Validate container names are unique
    names := make(map[string]bool)
    for _, c := range spec.Containers {
        if c.Name == "" {
            return errors.New("container name required")
        }
        if names[c.Name] {
            return fmt.Errorf("duplicate container name: %s", c.Name)
        }
        names[c.Name] = true
    }
    
    // Validate resources
    if spec.Resources.CPU < 128 {
        return errors.New("cpu must be at least 128 millicores")
    }
    if spec.Resources.Memory < 128 {
        return errors.New("memory must be at least 128 MB")
    }
    
    return nil
}
```

## Error Types

```go
var (
    ErrProviderNotFound  = errors.New("compute provider not found")
    ErrProviderConflict  = errors.New("provider already registered")
    ErrInvalidSpec       = errors.New("invalid compute specification")
    ErrProvisionFailed   = errors.New("provisioning failed")
    ErrUpdateFailed      = errors.New("update failed")
    ErrTenantNotFound    = errors.New("tenant compute resources not found")
)
```

## Testing Requirements

### Registry Tests
- Register multiple providers
- Reject duplicate registrations
- Get non-existent provider
- List registered providers
- Thread safety under concurrent access

### Manager Tests  
- Route to correct provider
- Handle provider not found
- Validate spec before delegation
- Log operations correctly
- Propagate provider errors
