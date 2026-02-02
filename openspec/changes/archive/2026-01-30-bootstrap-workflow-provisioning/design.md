# Design: Bootstrap Workflow Provisioning Framework

## Overview

This document describes the architecture for a pluggable workflow provisioning framework that enables tenant provisioning operations to be orchestrated via different workflow engines (AWS Step Functions, Temporal, etc.).

## Architecture

### Component Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                     Application Layer                        │
│  (Tenant Provisioning, State Management, API Handlers)       │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                    Workflow Manager                          │
│  • CreateWorkflow()    • StartExecution()                    │
│  • GetExecutionStatus()  • StopExecution()                   │
│  • DeleteWorkflow()    • ListWorkflows()                     │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                   Workflow Registry                          │
│  • Register(provider)  • Get(providerType)                   │
│  • List()             • Thread-safe                          │
└────────────────────────┬────────────────────────────────────┘
                         │
          ┌──────────────┼──────────────┐
          ▼              ▼              ▼
    ┌──────────┐  ┌──────────┐  ┌──────────┐
    │   Mock   │  │   Step   │  │ Temporal │
    │ Provider │  │Functions │  │ Provider │
    │          │  │ Provider │  │  (future)│
    └──────────┘  └──────────┘  └──────────┘
         │             │              │
         ▼             ▼              ▼
    In-Memory      AWS Step      Temporal
                   Functions       Cloud
```

### Provider Interface

Every workflow provider implements:

```go
type Provider interface {
    // Name returns the unique provider identifier
    Name() string
    
    // CreateWorkflow creates a workflow definition
    CreateWorkflow(ctx context.Context, spec *WorkflowSpec) (*CreateWorkflowResult, error)
    
    // StartExecution starts a workflow execution
    StartExecution(ctx context.Context, workflowID string, input *ExecutionInput) (*ExecutionResult, error)
    
    // GetExecutionStatus queries current execution state
    GetExecutionStatus(ctx context.Context, executionID string) (*ExecutionStatus, error)
    
    // StopExecution stops a running execution
    StopExecution(ctx context.Context, executionID string, reason string) error
    
    // DeleteWorkflow removes a workflow definition
    DeleteWorkflow(ctx context.Context, workflowID string) error
    
    // Validate performs provider-specific validation
    Validate(ctx context.Context, spec *WorkflowSpec) error
}
```

### Type System

#### WorkflowSpec
Defines a workflow to be created:

```go
type WorkflowSpec struct {
    WorkflowID     string          // Unique workflow identifier
    ProviderType   string          // Provider name
    Name           string          // Human-readable name
    Description    string          // Workflow description
    Definition     json.RawMessage // Provider-specific definition
    Timeout        time.Duration   // Overall workflow timeout
    RetryPolicy    *RetryPolicy    // Retry configuration
    Tags           map[string]string
    ProviderConfig json.RawMessage // Provider-specific config
}
```

#### ExecutionInput
Input for starting a workflow:

```go
type ExecutionInput struct {
    ExecutionName string          // Unique execution identifier
    Input         json.RawMessage // Workflow input data
    Tags          map[string]string
}
```

#### ExecutionStatus
Current state of a workflow execution:

```go
type ExecutionStatus struct {
    ExecutionID   string
    WorkflowID    string
    State         ExecutionState // pending, running, succeeded, failed, etc.
    StartTime     time.Time
    StopTime      *time.Time
    Input         json.RawMessage
    Output        json.RawMessage
    Error         *ExecutionError
    History       []ExecutionEvent
}

type ExecutionState string
const (
    StatePending   ExecutionState = "pending"
    StateRunning   ExecutionState = "running"
    StateSucceeded ExecutionState = "succeeded"
    StateFailed    ExecutionState = "failed"
    StateTimedOut  ExecutionState = "timed_out"
    StateCancelled ExecutionState = "cancelled"
)
```

### Manager Component

The Manager provides a high-level facade:

```go
type Manager struct {
    registry *Registry
    logger   *zap.Logger
}

func (m *Manager) CreateWorkflow(ctx context.Context, spec *WorkflowSpec) (*CreateWorkflowResult, error)
func (m *Manager) StartExecution(ctx context.Context, workflowID string, input *ExecutionInput) (*ExecutionResult, error)
func (m *Manager) GetExecutionStatus(ctx context.Context, executionID string, providerType string) (*ExecutionStatus, error)
func (m *Manager) StopExecution(ctx context.Context, executionID string, providerType string, reason string) error
func (m *Manager) DeleteWorkflow(ctx context.Context, workflowID string, providerType string) error
func (m *Manager) ListProviders() []string
```

### Registry Component

Thread-safe provider storage:

```go
type Registry struct {
    mu        sync.RWMutex
    providers map[string]Provider
    logger    *zap.Logger
}

func (r *Registry) Register(provider Provider) error
func (r *Registry) Get(providerType string) (Provider, error)
func (r *Registry) List() []string
func (r *Registry) Has(providerType string) bool
```

### Validation

Structural validation before provider delegation:

```go
func ValidateWorkflowSpec(spec *WorkflowSpec) error {
    // Workflow ID format
    // Provider type non-empty
    // Name non-empty
    // Timeout > 0
    // Definition non-empty
}

func ValidateExecutionInput(input *ExecutionInput) error {
    // Execution name format
    // Input valid JSON
}
```

### Error Handling

Standard error types:

```go
var (
    ErrProviderNotFound   = errors.New("workflow provider not found")
    ErrProviderConflict   = errors.New("workflow provider already registered")
    ErrInvalidSpec        = errors.New("invalid workflow specification")
    ErrWorkflowNotFound   = errors.New("workflow not found")
    ErrExecutionNotFound  = errors.New("execution not found")
    ErrExecutionFailed    = errors.New("workflow execution failed")
)
```

## Provider Organization

### Directory Structure

```
internal/workflow/
├── provider.go          # Provider interface
├── types.go             # All type definitions
├── errors.go            # Error types
├── validation.go        # Validation functions
├── registry.go          # Provider registry
├── registry_test.go     # Registry tests
├── manager.go           # Manager facade
├── manager_test.go      # Manager tests
└── providers/
    ├── mock/
    │   ├── mock.go      # Mock provider
    │   └── mock_test.go # Mock tests
    └── stepfunctions/   # (future change)
        └── ...
```

### Mock Provider

In-memory implementation for testing:

```go
type Provider struct {
    mu         sync.RWMutex
    workflows  map[string]*workflowState
    executions map[string]*executionState
}

type workflowState struct {
    Spec      *workflow.WorkflowSpec
    CreatedAt time.Time
}

type executionState struct {
    WorkflowID string
    Input      *workflow.ExecutionInput
    State      workflow.ExecutionState
    StartTime  time.Time
    StopTime   *time.Time
    Output     json.RawMessage
    Error      *workflow.ExecutionError
}
```

Mock provider simulates workflow execution:
- CreateWorkflow stores definition in memory
- StartExecution immediately transitions to "succeeded" (configurable)
- GetExecutionStatus returns stored state
- StopExecution marks execution as cancelled
- DeleteWorkflow removes from memory

## Integration with Application

### Configuration

Add to `internal/config/config.go`:

```go
type WorkflowConfig struct {
    DefaultProvider string `env:"WORKFLOW_DEFAULT_PROVIDER" default:"mock"`
}
```

### Initialization

In `cmd/landlord/main.go`:

```go
// Initialize workflow manager
workflowRegistry := workflow.NewRegistry(log)
workflowRegistry.Register(mock.New())
workflowManager := workflow.New(workflowRegistry, log)

log.Info("registered workflow providers",
    zap.Strings("providers", workflowManager.ListProviders()),
)
```

## Data Flow Examples

### Creating and Executing a Workflow

```
1. Application → Manager.CreateWorkflow(spec)
2. Manager → ValidateWorkflowSpec(spec)
3. Manager → Registry.Get(spec.ProviderType)
4. Manager → Provider.CreateWorkflow(spec)
5. Provider → [Create workflow in engine]
6. Provider → Return CreateWorkflowResult
7. Manager → Return result to application

8. Application → Manager.StartExecution(workflowID, input)
9. Manager → ValidateExecutionInput(input)
10. Manager → Provider.StartExecution(workflowID, input)
11. Provider → [Start execution in engine]
12. Provider → Return ExecutionResult
13. Manager → Return result to application
```

### Monitoring Execution

```
1. Application → Manager.GetExecutionStatus(executionID, providerType)
2. Manager → Registry.Get(providerType)
3. Manager → Provider.GetExecutionStatus(executionID)
4. Provider → [Query engine]
5. Provider → Return ExecutionStatus
6. Manager → Return status to application
```

## Design Decisions

### Why Provider Interface Pattern?

Same rationale as compute framework:
- Enables testing without cloud dependencies
- Supports multiple workflow engines
- Clean separation of concerns
- Future-proof for new engines

### Why JSON for Definitions?

- Step Functions uses JSON state machines
- Temporal uses code but can be represented as JSON
- Flexible for different workflow engines
- Easy to store and transmit

### Why Manager Facade?

- Centralizes validation
- Provides consistent logging
- Simplifies error handling
- Single integration point for application

### Why Separate Execution from Workflow?

- Workflows are templates (definition)
- Executions are instances (runtime)
- Mirrors Step Functions and Temporal models
- Enables workflow reuse with different inputs

## Testing Strategy

1. **Registry Tests**: Thread-safety, registration, lookup
2. **Manager Tests**: Validation, provider routing, error handling
3. **Mock Provider Tests**: Full Provider interface implementation
4. **Integration Tests**: End-to-end workflow lifecycle

## Future Enhancements

- Database persistence of execution history
- Webhook notifications on state changes
- Workflow versioning
- Scheduled executions
- Parallel execution support in mock provider
