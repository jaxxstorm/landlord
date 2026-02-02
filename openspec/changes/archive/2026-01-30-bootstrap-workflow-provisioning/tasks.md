# Tasks: Bootstrap Workflow Provisioning Framework

## Group 1: Core Types and Interfaces (7 tasks)

### 1.1 Create workflow package structure
**File**: `internal/workflow/` (directory)

Create the package structure:
- `internal/workflow/`
- `internal/workflow/providers/`

### 1.2 Define Provider interface
**File**: `internal/workflow/provider.go`

Implement the Provider interface with all method signatures:
- Name() string
- CreateWorkflow()
- StartExecution()
- GetExecutionStatus()
- StopExecution()
- DeleteWorkflow()
- Validate()

### 1.3 Define WorkflowSpec types
**File**: `internal/workflow/types.go`

Implement spec-related types:
- WorkflowSpec
- RetryPolicy
- ExecutionInput

### 1.4 Define result types
**File**: `internal/workflow/types.go`

Implement result types:
- CreateWorkflowResult
- ExecutionResult

### 1.5 Define status types
**File**: `internal/workflow/types.go`

Implement status types:
- ExecutionStatus
- ExecutionState constants
- ExecutionError
- ExecutionEvent

### 1.6 Define error types
**File**: `internal/workflow/errors.go`

Define all error variables:
- ErrProviderNotFound
- ErrProviderConflict
- ErrInvalidSpec
- ErrWorkflowNotFound
- ErrExecutionNotFound
- ErrExecutionFailed

### 1.7 Implement validation functions
**File**: `internal/workflow/validation.go`

Implement validation functions:
- ValidateWorkflowSpec() with all validation rules
- ValidateExecutionInput() with input validation

## Group 2: Provider Registry (5 tasks)

### 2.1 Create Registry struct
**File**: `internal/workflow/registry.go`

Define Registry with:
- providers map
- sync.RWMutex
- logger field

### 2.2 Implement Registry.Register()
**File**: `internal/workflow/registry.go`

Implement provider registration:
- Thread-safe with mutex
- Check for duplicates
- Log registration
- Return ErrProviderConflict on duplicate

### 2.3 Implement Registry.Get()
**File**: `internal/workflow/registry.go`

Implement provider lookup:
- Thread-safe read lock
- Return ErrProviderNotFound if missing

### 2.4 Implement Registry.List() and Has()
**File**: `internal/workflow/registry.go`

Implement utility methods:
- List() returns provider names
- Has() checks existence

### 2.5 Write Registry unit tests
**File**: `internal/workflow/registry_test.go`

Test:
- Register success
- Register duplicate rejection
- Get existing/non-existing
- List providers
- Concurrent access safety

## Group 3: Workflow Manager (6 tasks)

### 3.1 Create Manager struct
**File**: `internal/workflow/manager.go`

Define Manager with:
- Registry reference
- Logger field
- New() constructor

### 3.2 Implement Manager.CreateWorkflow()
**File**: `internal/workflow/manager.go`

Implement workflow creation:
- Validate spec
- Get provider from registry
- Delegate to provider
- Log start/completion
- Handle errors

### 3.3 Implement Manager.StartExecution()
**File**: `internal/workflow/manager.go`

Implement execution start:
- Validate input
- Get provider
- Delegate to provider
- Log execution start

### 3.4 Implement Manager.GetExecutionStatus()
**File**: `internal/workflow/manager.go`

Implement status query:
- Get provider
- Delegate to provider
- Return execution status

### 3.5 Implement Manager.StopExecution() and DeleteWorkflow()
**File**: `internal/workflow/manager.go`

Implement stop and delete:
- StopExecution() stops running execution
- DeleteWorkflow() removes workflow definition
- Log operations
- Handle errors

### 3.6 Implement Manager.ValidateWorkflowSpec() and ListProviders()
**File**: `internal/workflow/manager.go`

Implement utility methods:
- ValidateWorkflowSpec() validates without creating
- ListProviders() returns available providers

### 3.7 Write Manager unit tests
**File**: `internal/workflow/manager_test.go`

Test:
- Routing to correct provider
- Provider not found handling
- Spec validation before delegation
- Error propagation
- Logging behavior

## Group 4: Mock Provider (4 tasks)

### 4.1 Create mock provider structure
**File**: `internal/workflow/providers/mock/mock.go`

Create mock provider with:
- In-memory workflow storage
- In-memory execution storage
- Thread-safe mutexes

### 4.2 Implement mock provider lifecycle methods
**File**: `internal/workflow/providers/mock/mock.go`

Implement workflow lifecycle:
- Name() returns "mock"
- CreateWorkflow() stores in memory
- DeleteWorkflow() removes from memory
- Validate() checks basic spec

### 4.3 Implement mock provider execution methods
**File**: `internal/workflow/providers/mock/mock.go`

Implement execution:
- StartExecution() creates execution, sets to "succeeded" immediately
- GetExecutionStatus() returns stored state
- StopExecution() marks as "cancelled"

### 4.4 Write mock provider tests
**File**: `internal/workflow/providers/mock/mock_test.go`

Test mock provider:
- CreateWorkflow success and idempotency
- StartExecution and immediate completion
- GetExecutionStatus for various states
- StopExecution cancellation
- DeleteWorkflow removal
- Concurrent operations

## Group 5: Integration and Documentation (3 tasks)

### 5.1 Add workflow configuration to main config
**File**: `internal/config/config.go`

Add WorkflowConfig struct:
```go
type WorkflowConfig struct {
    DefaultProvider string `env:"WORKFLOW_DEFAULT_PROVIDER" default:"mock"`
}
```

Add to main Config struct

### 5.2 Wire workflow manager in main.go
**File**: `cmd/landlord/main.go`

Initialize workflow system:
- Create registry
- Register mock provider
- Create manager
- Log registered providers

### 5.3 Create provider development guide
**File**: `docs/workflow-providers.md`

Document:
- How to implement a new provider
- Required interface methods
- Testing requirements
- Registration process
- Example provider code
- Step Functions provider overview (for future)

## Acceptance Criteria

- [x] Provider interface defined with 7 methods
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
- [x] No real provider implementations (deferred to next change)

## Notes

- No AWS Step Functions provider in this change
- Focus on framework and abstractions
- Mock provider for testing only
- Mock provider immediately completes executions (no async simulation yet)
- Step Functions provider comes in next change
- No database persistence yet (in-memory only)
