# Design: Compute Provisioning Framework

## Architecture Overview

```
┌─────────────────────────────────────────┐
│         Tenant Domain Layer             │
│  (owns tenant lifecycle & state)        │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│      Compute Manager (Facade)           │
│  - Routes requests to providers         │
│  - Manages provider registry            │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│       Provider Interface                │
│  Provision() Update() Destroy()         │
│  GetStatus() Validate()                 │
└──────────────┬──────────────────────────┘
               │
       ┌───────┴───────┬──────────┐
       ▼               ▼          ▼
  ┌────────┐     ┌─────────┐  ┌────────┐
  │  ECS   │     │   K8s   │  │ Nomad  │
  │Provider│     │Provider │  │Provider│
  └────────┘     └─────────┘  └────────┘
```

## Core Components

### 1. Provider Interface (`internal/compute/provider.go`)

Defines the contract all compute providers must implement:

```go
type Provider interface {
    Name() string
    
    Provision(ctx context.Context, spec *TenantComputeSpec) (*ProvisionResult, error)
    Update(ctx context.Context, tenantID string, spec *TenantComputeSpec) (*UpdateResult, error)
    Destroy(ctx context.Context, tenantID string) error
    GetStatus(ctx context.Context, tenantID string) (*ComputeStatus, error)
    Validate(ctx context.Context, spec *TenantComputeSpec) error
}
```

### 2. Tenant Compute Specification

Provider-agnostic specification for tenant compute requirements:

```go
type TenantComputeSpec struct {
    TenantID      string
    ProviderType  string  // "ecs", "kubernetes", "nomad"
    Containers    []ContainerSpec
    Resources     ResourceRequirements
    Networking    NetworkConfig
    Secrets       []SecretReference
    ProviderConfig json.RawMessage  // Provider-specific config
}

type ContainerSpec struct {
    Name       string
    Image      string
    Command    []string
    Args       []string
    Env        map[string]string
    Ports      []PortMapping
    HealthCheck *HealthCheckConfig
}
```

### 3. Provider Registry (`internal/compute/registry.go`)

Manages available providers and routes requests:

```go
type Registry struct {
    providers map[string]Provider
    logger    *zap.Logger
}

func NewRegistry(logger *zap.Logger) *Registry
func (r *Registry) Register(provider Provider) error
func (r *Registry) Get(providerType string) (Provider, error)
func (r *Registry) List() []string
```

### 4. Compute Manager (`internal/compute/manager.go`)

Facade that coordinates compute operations:

```go
type Manager struct {
    registry *Registry
    logger   *zap.Logger
}

func New(registry *Registry, logger *zap.Logger) *Manager
func (m *Manager) ProvisionTenant(ctx context.Context, spec *TenantComputeSpec) (*ProvisionResult, error)
func (m *Manager) UpdateTenant(ctx context.Context, tenantID string, spec *TenantComputeSpec) (*UpdateResult, error)
func (m *Manager) DestroyTenant(ctx context.Context, tenantID, providerType string) error
func (m *Manager) GetTenantStatus(ctx context.Context, tenantID, providerType string) (*ComputeStatus, error)
```

## Provider Organization

```
internal/compute/
├── provider.go           # Provider interface
├── registry.go           # Provider registry
├── manager.go            # Compute manager facade
├── types.go             # Shared types (specs, results, status)
├── providers/           # Provider implementations
│   ├── ecs/
│   │   ├── provider.go
│   │   ├── types.go
│   │   └── ecs_test.go
│   ├── kubernetes/
│   │   ├── provider.go
│   │   ├── types.go
│   │   └── kubernetes_test.go
│   └── nomad/
│       ├── provider.go
│       ├── types.go
│       └── nomad_test.go
└── common/              # Shared utilities
    ├── validation.go
    └── helpers.go
```

## Result Types

### Provision Result
```go
type ProvisionResult struct {
    TenantID      string
    ProviderType  string
    Status        ProvisionStatus
    ResourceIDs   map[string]string  // Provider-specific IDs
    Endpoints     []Endpoint
    Message       string
    ProvisionedAt time.Time
}
```

### Compute Status
```go
type ComputeStatus struct {
    TenantID     string
    ProviderType string
    State        ComputeState  // Running, Stopped, Failed, etc.
    Containers   []ContainerStatus
    Health       HealthStatus
    LastUpdated  time.Time
    Metadata     map[string]string
}
```

## Error Handling

Define common error types:
```go
var (
    ErrProviderNotFound    = errors.New("compute provider not found")
    ErrInvalidSpec        = errors.New("invalid compute specification")
    ErrProvisionFailed    = errors.New("provisioning failed")
    ErrTenantNotFound     = errors.New("tenant compute resources not found")
    ErrProviderConflict   = errors.New("provider already registered")
)
```

## Adding a New Provider

1. Create package under `internal/compute/providers/<name>/`
2. Implement the `Provider` interface
3. Define provider-specific types in `types.go`
4. Register provider in `cmd/landlord/main.go`:

```go
import "github.com/jaxxstorm/landlord/internal/compute/providers/ecs"

// In main setup
registry := compute.NewRegistry(logger)
registry.Register(ecs.NewProvider(ecsConfig, logger))
```

## Configuration

Provider configuration follows Kong pattern:

```go
type ComputeConfig struct {
    DefaultProvider string `env:"COMPUTE_DEFAULT_PROVIDER" default:"ecs"`
    
    // Provider-specific configs
    ECS        *ECSConfig        `embed:"" prefix:"ecs-"`
    Kubernetes *KubernetesConfig `embed:"" prefix:"k8s-"`
}
```

## Testing Strategy

- **Unit tests**: Test each provider implementation independently
- **Interface compliance**: Ensure all providers implement the interface
- **Registry tests**: Test provider registration and retrieval
- **Manager tests**: Test request routing and error handling
- **Integration tests**: Deferred to provider implementation changes

## Security Considerations

- Secrets should never be logged
- Provider credentials use environment variables
- Tenant isolation is provider-specific
- Audit logging for all provision/destroy operations

## Future Enhancements

- Provider health checks
- Provider capability discovery
- Multi-provider tenant deployments
- Provider-specific metrics and monitoring
- Out-of-tree provider plugin system
