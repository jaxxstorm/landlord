# Specification: Registry and Manager

## Overview

This specification defines the Registry (provider storage) and Manager (workflow orchestration facade) components of the workflow framework.

## Registry Component

### Purpose

The Registry provides thread-safe storage and lookup of workflow providers. It ensures providers can be registered at startup and retrieved safely during concurrent operations.

### Registry Struct

```go
type Registry struct {
    mu        sync.RWMutex
    providers map[string]Provider
    logger    *zap.Logger
}
```

**Fields**:
- `mu`: Read-write mutex for thread safety
- `providers`: Map from provider name to Provider instance
- `logger`: Structured logger for operations

### Methods

#### NewRegistry

```go
func NewRegistry(logger *zap.Logger) *Registry
```

Creates a new Registry instance.

**Requirements**:
- MUST initialize providers map
- MUST store logger reference
- MUST return ready-to-use Registry

#### Register

```go
func (r *Registry) Register(provider Provider) error
```

Registers a workflow provider.

**Parameters**:
- `provider`: Provider instance to register

**Returns**:
- `error`: ErrProviderConflict if provider name already registered, nil on success

**Requirements**:
- MUST acquire write lock
- MUST validate provider.Name() is non-empty
- MUST check for duplicate provider names
- MUST log successful registration
- MUST return ErrProviderConflict if name already exists
- MUST be thread-safe

**Example**:
```go
registry := workflow.NewRegistry(logger)
err := registry.Register(mock.New())
if err != nil {
    log.Fatal(err)
}
```

#### Get

```go
func (r *Registry) Get(providerType string) (Provider, error)
```

Retrieves a provider by name.

**Parameters**:
- `providerType`: Name of provider to retrieve

**Returns**:
- `Provider`: The requested provider
- `error`: ErrProviderNotFound if not registered

**Requirements**:
- MUST acquire read lock
- MUST return ErrProviderNotFound if provider doesn't exist
- MUST be thread-safe for concurrent reads

**Example**:
```go
provider, err := registry.Get("step-functions")
if err != nil {
    return err
}
```

#### List

```go
func (r *Registry) List() []string
```

Returns list of all registered provider names.

**Returns**:
- `[]string`: Sorted list of provider names

**Requirements**:
- MUST acquire read lock
- MUST return empty slice if no providers registered
- SHOULD return sorted list for consistency
- MUST be thread-safe

**Example**:
```go
providers := registry.List()
// Returns: ["mock", "restate", "step-functions"]
```

#### Has

```go
func (r *Registry) Has(providerType string) bool
```

Checks if a provider is registered.

**Parameters**:
- `providerType`: Name to check

**Returns**:
- `bool`: true if registered, false otherwise

**Requirements**:
- MUST acquire read lock
- MUST be thread-safe

**Example**:
```go
if registry.Has("temporal") {
    // Use temporal provider
}
```

## Manager Component

### Purpose

The Manager provides a high-level facade for workflow operations. It handles validation, provider routing, error handling, and logging.

### Manager Struct

```go
type Manager struct {
    registry *Registry
    logger   *zap.Logger
}
```

**Fields**:
- `registry`: Provider registry for lookup
- `logger`: Structured logger with component context

### Methods

#### New

```go
func New(registry *Registry, logger *zap.Logger) *Manager
```

Creates a new Manager instance.

**Requirements**:
- MUST store registry reference
- MUST create logger with "component": "workflow-manager"
- MUST return ready-to-use Manager

#### CreateWorkflow

```go
func (m *Manager) CreateWorkflow(ctx context.Context, spec *WorkflowSpec) (*CreateWorkflowResult, error)
```

Creates a workflow definition.

**Parameters**:
- `ctx`: Context for cancellation and timeout
- `spec`: Workflow specification

**Returns**:
- `CreateWorkflowResult`: Creation result with resource IDs
- `error`: Any errors during creation

**Requirements**:
- MUST log workflow creation start
- MUST validate spec using ValidateWorkflowSpec
- MUST return ErrInvalidSpec if validation fails
- MUST get provider from registry using spec.ProviderType
- MUST return provider lookup errors
- MUST delegate to provider.CreateWorkflow
- MUST log workflow creation completion
- MUST log errors on failure

**Example**:
```go
result, err := manager.CreateWorkflow(ctx, &workflow.WorkflowSpec{
    WorkflowID:   "tenant-provisioning",
    ProviderType: "step-functions",
    Name:         "Tenant Provisioning",
    Definition:   definition,
})
```

#### StartExecution

```go
func (m *Manager) StartExecution(ctx context.Context, workflowID string, input *ExecutionInput) (*ExecutionResult, error)
```

Starts a workflow execution.

**Parameters**:
- `ctx`: Context for cancellation and timeout
- `workflowID`: ID of workflow to execute
- `input`: Execution input data

**Returns**:
- `ExecutionResult`: Execution result with execution ID
- `error`: Any errors during start

**Requirements**:
- MUST log execution start
- MUST validate input using ValidateExecutionInput
- MUST get provider from spec.ProviderType in execution context
- MUST delegate to provider.StartExecution
- MUST log execution started
- MUST log errors on failure

**Note**: This requires tracking which provider created which workflow, or accepting providerType parameter.

**Example**:
```go
result, err := manager.StartExecution(ctx, "tenant-provisioning", &workflow.ExecutionInput{
    ExecutionName: "tenant-123-provision",
    Input:         []byte(`{"tenant_id": "tenant-123"}`),
})
```

#### GetExecutionStatus

```go
func (m *Manager) GetExecutionStatus(ctx context.Context, executionID string, providerType string) (*ExecutionStatus, error)
```

Queries execution status.

**Parameters**:
- `ctx`: Context for cancellation and timeout
- `executionID`: ID of execution to query
- `providerType`: Provider managing this execution

**Returns**:
- `ExecutionStatus`: Current execution status
- `error`: Any query errors

**Requirements**:
- MUST get provider from registry
- MUST delegate to provider.GetExecutionStatus
- MUST return provider errors

**Example**:
```go
status, err := manager.GetExecutionStatus(ctx, "exec-abc123", "step-functions")
```

#### StopExecution

```go
func (m *Manager) StopExecution(ctx context.Context, executionID string, providerType string, reason string) error
```

Stops a running execution.

**Parameters**:
- `ctx`: Context for cancellation and timeout
- `executionID`: ID of execution to stop
- `providerType`: Provider managing this execution
- `reason`: Human-readable reason for stopping

**Returns**:
- `error`: Any stop errors

**Requirements**:
- MUST log stop request with reason
- MUST get provider from registry
- MUST delegate to provider.StopExecution
- MUST log completion or error

**Example**:
```go
err := manager.StopExecution(ctx, "exec-abc123", "step-functions", "User requested cancellation")
```

#### DeleteWorkflow

```go
func (m *Manager) DeleteWorkflow(ctx context.Context, workflowID string, providerType string) error
```

Deletes a workflow definition.

**Parameters**:
- `ctx`: Context for cancellation and timeout
- `workflowID`: ID of workflow to delete
- `providerType`: Provider managing this workflow

**Returns**:
- `error`: Any deletion errors

**Requirements**:
- MUST log deletion
- MUST get provider from registry
- MUST delegate to provider.DeleteWorkflow
- MUST log completion or error

**Example**:
```go
err := manager.DeleteWorkflow(ctx, "tenant-provisioning", "step-functions")
```

#### ValidateWorkflowSpec

```go
func (m *Manager) ValidateWorkflowSpec(ctx context.Context, spec *WorkflowSpec) error
```

Validates a workflow spec without creating it.

**Parameters**:
- `ctx`: Context for cancellation and timeout
- `spec`: Workflow specification to validate

**Returns**:
- `error`: Validation error or nil if valid

**Requirements**:
- MUST validate structure using ValidateWorkflowSpec
- MUST get provider from registry
- MUST delegate to provider.Validate for provider-specific validation
- MUST return descriptive errors

**Example**:
```go
err := manager.ValidateWorkflowSpec(ctx, spec)
if err != nil {
    return fmt.Errorf("invalid workflow: %w", err)
}
```

#### ListProviders

```go
func (m *Manager) ListProviders() []string
```

Returns list of registered providers.

**Returns**:
- `[]string`: Provider names

**Requirements**:
- MUST delegate to registry.List()

**Example**:
```go
providers := manager.ListProviders()
log.Info("Available providers", zap.Strings("providers", providers))
```

## Design Patterns

### Provider Lookup Pattern

All Manager methods that interact with providers follow this pattern:

```go
func (m *Manager) SomeOperation(ctx context.Context, providerType string, ...) error {
    // 1. Log operation start
    m.logger.Info("operation starting", zap.String("provider", providerType))
    
    // 2. Get provider from registry
    provider, err := m.registry.Get(providerType)
    if err != nil {
        m.logger.Error("provider not found", zap.Error(err))
        return err
    }
    
    // 3. Delegate to provider
    result, err := provider.SomeMethod(ctx, ...)
    if err != nil {
        m.logger.Error("operation failed", zap.Error(err))
        return err
    }
    
    // 4. Log success
    m.logger.Info("operation completed")
    return nil
}
```

### Validation Pattern

Manager validates before delegating:

```go
func (m *Manager) CreateWorkflow(ctx context.Context, spec *WorkflowSpec) (*CreateWorkflowResult, error) {
    // Structural validation first
    if err := ValidateWorkflowSpec(spec); err != nil {
        return nil, fmt.Errorf("%w: %s", ErrInvalidSpec, err)
    }
    
    // Get provider
    provider, err := m.registry.Get(spec.ProviderType)
    if err != nil {
        return nil, err
    }
    
    // Provider-specific validation
    // (implicitly called by provider.CreateWorkflow)
    
    // Delegate
    return provider.CreateWorkflow(ctx, spec)
}
```

## Thread Safety

- **Registry**: All methods use read/write locks appropriately
- **Manager**: Stateless, safe for concurrent use (registry handles synchronization)

## Logging Strategy

### Registry Logs

- Provider registration: INFO level with provider name
- Duplicate registration attempts: ERROR level
- Provider lookup failures: DEBUG level (not logged, returned as error)

### Manager Logs

- Operation start: INFO level with operation, workflow ID, provider
- Operation success: INFO level with result summary
- Operation failure: ERROR level with error details
- Validation failures: DEBUG level (error returned to caller)

## Error Handling

### Registry Errors

- `ErrProviderConflict`: Duplicate provider name during Register
- `ErrProviderNotFound`: Provider not found during Get

### Manager Errors

- Returns registry errors as-is
- Wraps validation errors with ErrInvalidSpec
- Propagates provider errors with added context

## Testing Requirements

### Registry Tests

- TestRegistryRegister: Successful registration
- TestRegistryRegisterDuplicate: Conflict detection
- TestRegistryGet: Successful lookup
- TestRegistryGetNotFound: Missing provider
- TestRegistryList: Empty and populated lists
- TestRegistryHas: Existence checks
- TestRegistryConcurrency: Thread safety

### Manager Tests

- TestManagerCreateWorkflow: Success path
- TestManagerCreateWorkflowInvalidSpec: Validation failure
- TestManagerCreateWorkflowProviderNotFound: Missing provider
- TestManagerStartExecution: Success path
- TestManagerStartExecutionInvalidInput: Validation failure
- TestManagerGetExecutionStatus: Success path
- TestManagerStopExecution: Success path
- TestManagerDeleteWorkflow: Success path
- TestManagerListProviders: Provider enumeration
- TestManagerValidateWorkflowSpec: Validation delegation
