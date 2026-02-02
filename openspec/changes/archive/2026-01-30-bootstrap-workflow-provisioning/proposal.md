# Proposal: Bootstrap Workflow Provisioning Framework

## Why

The tenant provisioning system needs to orchestrate complex, long-running operations with retries, partial failure handling, and state tracking. Currently, there is no abstraction for workflow execution - we need a pluggable framework that can work with multiple workflow engines (AWS Step Functions initially, Temporal in the future).

### Problems We're Solving

1. **No Workflow Orchestration**: Complex tenant provisioning requires coordinating multiple steps (provision compute, configure networking, initialize databases, run migrations). Without a workflow engine, we'd have to implement retry logic, state persistence, and error handling manually.

2. **Vendor Lock-in Risk**: Hard-coding AWS Step Functions throughout the codebase makes it difficult to migrate to other workflow engines (Temporal, Cadence, Conductor) as requirements evolve.

3. **No State Tracking**: We need to track workflow execution state, see what step failed, view execution history, and enable debugging of tenant provisioning failures.

4. **Failure Resilience**: Tenant provisioning operations can fail partway through. We need automatic retries, timeout handling, and the ability to resume from failure points without re-executing completed steps.

5. **Testing Complexity**: Without abstraction, testing workflow logic requires AWS credentials and real Step Functions, making unit tests slow and brittle.

### Success Criteria

- Pluggable workflow provider interface supporting multiple engines
- Create, execute, monitor, and delete workflows
- Track workflow state and execution history
- Handle failures, retries, and timeouts at the framework level
- Mock provider for testing without cloud dependencies
- AWS Step Functions provider as initial implementation
- Framework integrated into main application

## What

Create a workflow provisioning framework following the same patterns as the compute framework:

- **Provider Interface**: Abstract workflow operations (create, start, monitor, delete)
- **Type System**: Workflow specifications, execution results, and state tracking
- **Registry Pattern**: Thread-safe provider registration and lookup
- **Manager Facade**: High-level API for workflow orchestration
- **Mock Provider**: In-memory testing implementation
- **AWS Provider**: Step Functions integration (deferred to next change)

### Core Capabilities

1. **Workflow Lifecycle Management**
   - Create workflow definitions
   - Start workflow executions
   - Monitor execution status
   - Cancel/stop running executions
   - Delete workflow definitions

2. **State Tracking**
   - Track execution state (pending, running, succeeded, failed, timed_out)
   - Capture execution history and step details
   - Store input/output for each execution

3. **Error Handling**
   - Surface execution errors with context
   - Support retry policies
   - Handle timeouts gracefully

4. **Observability**
   - Log all workflow operations
   - Track execution metrics
   - Provide execution history

### Out of Scope (This Change)

- Actual AWS Step Functions provider (next change)
- Temporal/Cadence/Conductor providers
- Workflow definition language/DSL
- Database persistence of workflow history
- Workflow versioning
- Complex branching/parallel execution logic in mock

## Changes Required

### New Files

1. **internal/workflow/provider.go**: Provider interface with 6 methods
2. **internal/workflow/types.go**: Workflow specs, results, states, and execution details
3. **internal/workflow/errors.go**: Standard error types
4. **internal/workflow/validation.go**: Workflow spec validation
5. **internal/workflow/registry.go**: Provider registry
6. **internal/workflow/manager.go**: Workflow manager facade
7. **internal/workflow/providers/mock/mock.go**: Mock provider implementation
8. **Test files**: Complete test coverage for all components
9. **docs/workflow-providers.md**: Provider development guide

### Modified Files

1. **internal/config/config.go**: Add WorkflowConfig
2. **cmd/landlord/main.go**: Initialize workflow manager

### Architecture

```
internal/workflow/
├── provider.go          # Provider interface
├── types.go             # Workflow specs and results
├── errors.go            # Error types
├── validation.go        # Validation functions
├── registry.go          # Provider registry
├── manager.go           # Manager facade
└── providers/
    └── mock/
        ├── mock.go      # Mock provider
        └── mock_test.go # Tests
```

## Alignment with Project Goals

- **Pluggability**: Workflow engine is pluggable via Provider interface
- **Failure Is First-Class**: Framework explicitly handles retries, timeouts, failures
- **Observability**: All operations logged and traceable
- **Ports and Adapters**: Clean separation between domain (workflow orchestration) and adapters (Step Functions, Temporal)
- **Testing Strategy**: Mock provider enables fast unit tests

## Timeline

Single change to establish the framework and mock provider. AWS Step Functions provider in follow-up change.

## Dependencies

None - this is foundational infrastructure.

## Risks

- **API Mismatch**: Different workflow engines have different capabilities. The Provider interface must be abstract enough to support Step Functions, Temporal, and others while remaining simple.
  - *Mitigation*: Start with common operations (create, start, monitor, delete). Add engine-specific features via ProviderConfig field.

- **Execution Model Differences**: Step Functions uses state machines, Temporal uses code-as-workflow. Abstraction may be leaky.
  - *Mitigation*: Focus on execution control (start/stop/monitor) rather than definition format. Let providers handle their own definition languages.
