# Tasks: Bootstrap Compute Provisioning Framework

## Group 1: Core Types and Interfaces (7 tasks)

### 1.1 Create compute package structure
**File**: `internal/compute/` (directory)

Create the package structure:
- `internal/compute/`
- `internal/compute/providers/`
- `internal/compute/common/`

### 1.2 Define Provider interface
**File**: `internal/compute/provider.go`

Implement the Provider interface with all method signatures:
- Name() string
- Provision()
- Update()
- Destroy()
- GetStatus()
- Validate()

### 1.3 Define TenantComputeSpec types
**File**: `internal/compute/types.go`

Implement all spec-related types:
- TenantComputeSpec
- ContainerSpec
- PortMapping
- HealthCheckConfig
- ResourceRequirements
- NetworkConfig
- DNSConfig
- SecretReference

### 1.4 Define result types
**File**: `internal/compute/types.go`

Implement result types:
- ProvisionResult
- ProvisionStatus constants
- UpdateResult
- UpdateStatus constants
- Endpoint

### 1.5 Define status types
**File**: `internal/compute/types.go`

Implement status types:
- ComputeStatus
- ComputeState constants
- ContainerStatus
- HealthStatus constants

### 1.6 Define error types
**File**: `internal/compute/errors.go`

Define all error variables:
- ErrProviderNotFound
- ErrProviderConflict
- ErrInvalidSpec
- ErrProvisionFailed
- ErrUpdateFailed
- ErrTenantNotFound

### 1.7 Implement validation functions
**File**: `internal/compute/validation.go`

Implement ValidateComputeSpec() with all validation rules:
- TenantID validation
- Container name uniqueness
- Resource limits
- Port ranges

## Group 2: Provider Registry (5 tasks)

### 2.1 Create Registry struct
**File**: `internal/compute/registry.go`

Define Registry with:
- providers map
- sync.RWMutex
- logger field

### 2.2 Implement Registry.Register()
**File**: `internal/compute/registry.go`

Implement provider registration:
- Thread-safe with mutex
- Check for duplicates
- Log registration
- Return ErrProviderConflict on duplicate

### 2.3 Implement Registry.Get()
**File**: `internal/compute/registry.go`

Implement provider lookup:
- Thread-safe read lock
- Return ErrProviderNotFound if missing

### 2.4 Implement Registry.List() and Has()
**File**: `internal/compute/registry.go`

Implement utility methods:
- List() returns provider names
- Has() checks existence

### 2.5 Write Registry unit tests
**File**: `internal/compute/registry_test.go`

Test:
- Register success
- Register duplicate rejection
- Get existing/non-existing
- List providers
- Concurrent access safety

## Group 3: Compute Manager (6 tasks)

### 3.1 Create Manager struct
**File**: `internal/compute/manager.go`

Define Manager with:
- Registry reference
- Logger field
- New() constructor

### 3.2 Implement Manager.ProvisionTenant()
**File**: `internal/compute/manager.go`

Implement provisioning:
- Validate spec
- Get provider from registry
- Delegate to provider
- Log start/completion
- Handle errors

### 3.3 Implement Manager.UpdateTenant()
**File**: `internal/compute/manager.go`

Implement updates:
- Validate spec
- Get provider
- Delegate to provider
- Log changes

### 3.4 Implement Manager.DestroyTenant()
**File**: `internal/compute/manager.go`

Implement destroy:
- Get provider
- Delegate to provider
- Log destruction
- Handle errors

### 3.5 Implement Manager.GetTenantStatus() and ValidateTenantSpec()
**File**: `internal/compute/manager.go`

Implement utility methods:
- GetTenantStatus() queries provider
- ValidateTenantSpec() validates without provisioning
- ListProviders() returns available providers

### 3.6 Write Manager unit tests
**File**: `internal/compute/manager_test.go`

Test:
- Routing to correct provider
- Provider not found handling
- Spec validation before delegation
- Error propagation
- Logging behavior

## Group 4: Example Provider Skeleton (4 tasks)

### 4.1 Create mock provider for testing
**File**: `internal/compute/providers/mock/provider.go`

Implement a mock provider:
- In-memory storage
- Implements Provider interface
- Configurable success/failure
- For testing purposes only

### 4.2 Implement mock provider methods
**File**: `internal/compute/providers/mock/provider.go`

Implement all Provider methods:
- Name() returns "mock"
- Provision() stores tenant in memory
- Update() modifies stored tenant
- Destroy() removes from memory
- GetStatus() returns stored state
- Validate() checks basic spec

### 4.3 Add mock provider types
**File**: `internal/compute/providers/mock/types.go`

Define mock-specific types:
- MockConfig
- Internal storage structures

### 4.4 Write mock provider tests
**File**: `internal/compute/providers/mock/provider_test.go`

Test mock provider:
- Interface compliance
- Provision/destroy lifecycle
- Update idempotency
- Status queries

## Group 5: Integration and Documentation (3 tasks)

### 5.1 Add compute configuration to main config
**File**: `internal/config/config.go`

Add ComputeConfig struct:
```go
type ComputeConfig struct {
    DefaultProvider string `env:"COMPUTE_DEFAULT_PROVIDER" default:"mock"`
}
```

Add to main Config struct

### 5.2 Wire compute manager in main.go
**File**: `cmd/landlord/main.go`

Initialize compute system:
- Create registry
- Register mock provider (for now)
- Create manager
- Add to application context

### 5.3 Create provider development guide
**File**: `docs/compute-providers.md`

Document:
- How to implement a new provider
- Required interface methods
- Testing requirements
- Registration process
- Example provider code

## Acceptance Criteria

- [x] Provider interface defined with 6 methods
- [x] All types and constants defined
- [x] Registry supports register/get/list operations
- [x] Registry is thread-safe
- [x] Manager validates specs before delegation
- [x] Manager routes to correct providers
- [x] Mock provider fully implements interface
- [x] Mock provider has unit tests
- [x] All error types defined
- [x] Validation functions work correctly
- [x] Manager has unit tests
- [x] Integration wired in main.go
- [x] Provider development guide written
- [x] `go test -short ./...` passes
- [x] No real provider implementations (deferred)

## Notes

- No actual cloud provider implementations in this change
- Focus on framework and abstractions
- Mock provider for testing only
- ECS/Kubernetes/Nomad providers come later
- No database persistence yet (in-memory only)
