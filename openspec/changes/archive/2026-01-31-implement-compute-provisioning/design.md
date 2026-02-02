## Context

Currently, the system has:
- A pluggable compute provisioning layer (`internal/compute/manager.go`, `provider.go`) that can provision tenant compute resources
- A workflow engine layer (`internal/workflow/manager.go`) that triggers workflows via API or controller reconciliation
- A reconciliation controller that monitors tenant state and triggers workflows when needed
- A database layer tracking tenant state, workflow execution IDs, and state history

However, workflows execute independently of compute operations. Workflows have no way to invoke compute provisioning, nor do they receive feedback about whether compute resources were actually created. This creates a disconnect: workflows orchestrate "how" tenants should be configured, but have no connection to "what" infrastructure actually runs.

## Goals / Non-Goals

**Goals:**
- Enable workflows to directly invoke compute provisioning operations (create, update, delete)
- Provide workflows with queryable, trackable execution state for compute operations
- Allow workflows to react to compute operation outcomes via callbacks or polling
- Support structured error information enabling workflow-level retry logic
- Maintain the pluggable provider abstraction (different compute engines work identically)
- Enable idempotent compute operations (safe to retry with same ID)

**Non-Goals:**
- Implement specific compute provider logic (ECS, Kubernetes details belong in provider implementations)
- Create a unified "execution engine" merging workflow + compute into a single system
- Support transaction-like guarantees across workflow + compute (eventual consistency is acceptable)
- Build a full UI or dashboard for monitoring compute operations

## Decisions

### Decision 1: Extend WorkflowClient to provide compute operation access to workflows

**Rationale:**
Workflows are executed by providers (Step Functions, Temporal, etc.) and need a way to invoke compute operations. The cleanest approach is to extend `WorkflowClient` (or create a new `ComputeWorkflowClient`) that workflows can call as part of their state machine definitions.

**Implementation:**
- Add methods like `ProvisionTenant()`, `UpdateTenant()`, `DeleteTenant()` to the workflow client
- Workflows will invoke these methods as task states (in Step Functions) or as activity calls (in Temporal)
- The workflow client will delegate to the compute manager

**Alternatives considered:**
- Shared message queue: Too indirect; adds latency and complexity for request-response patterns
- Direct provider calls: Violates abstraction; ties workflows to specific compute provider APIs
- Webhook callbacks only: Cannot support synchronous response patterns workflows may need

### Decision 2: Create a ComputeExecution entity to track compute operation state

**Rationale:**
Unlike workflow executions (which are fully managed by the workflow provider), compute operations are managed by the landlord system. We need a durable, queryable record of each compute provisioning attempt.

**Implementation:**
- New database table: `compute_executions` (tenant_id, execution_id, operation_type, status, resource_ids, error_details, created_at, updated_at)
- New database table: `compute_execution_history` (execution_id, status, timestamp, details) for audit trail
- Methods: `CreateComputeExecution()`, `UpdateComputeExecutionStatus()`, `GetComputeExecution()`
- These operations are called by ComputeManager when workflows invoke compute operations

**Alternatives considered:**
- Store compute state only in workflow inputs/outputs: Too volatile; state is lost if workflow crashes
- Reuse workflow_execution_id for compute tracking: Conflates two different execution models; creates confusion

### Decision 3: Implement callback delivery for compute operation completion

**Rationale:**
Workflows need to know when compute operations finish without polling. Callbacks enable event-driven workflow progression.

**Implementation:**
- After compute operation completes (success or failure), system posts callback to workflow provider
- For Step Functions: Send task success/failure response
- For Temporal: Send signal to workflow or complete activity
- Callback includes execution ID, status, resource IDs (if success), error details (if failure)
- Callbacks use deterministic execution IDs for idempotency

**Alternatives considered:**
- Polling only: Simple to implement but adds latency and complexity for workflows
- SQS/message queue: Decouples systems but adds infrastructure complexity
- Long-polling HTTP: Works but inefficient for high-volume operations

### Decision 4: Database schema for compute execution tracking

**Rationale:**
Need efficient querying and audit trail for compute operations.

**Schema:**
```sql
CREATE TABLE compute_executions (
  id SERIAL PRIMARY KEY,
  execution_id VARCHAR(255) UNIQUE NOT NULL,
  tenant_id VARCHAR(255) NOT NULL,
  workflow_execution_id VARCHAR(255) NOT NULL,  -- link back to triggering workflow
  operation_type VARCHAR(50) NOT NULL,  -- 'provision', 'update', 'delete'
  status VARCHAR(50) NOT NULL,  -- 'pending', 'running', 'succeeded', 'failed'
  resource_ids JSONB,  -- provider-specific resource identifiers
  error_code VARCHAR(100),
  error_message TEXT,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);

CREATE TABLE compute_execution_history (
  id SERIAL PRIMARY KEY,
  compute_execution_id VARCHAR(255) NOT NULL,
  status VARCHAR(50) NOT NULL,
  details JSONB,
  timestamp TIMESTAMP DEFAULT NOW(),
  FOREIGN KEY (compute_execution_id) REFERENCES compute_executions(execution_id)
);
```

### Decision 5: ComputeManager orchestrates compute operations and tracks state

**Rationale:**
Existing ComputeManager already validates specs and routes to providers. Extend it to also create/update ComputeExecution records.

**Implementation:**
- Add `ProvisionTenantWithTracking()` method that:
  1. Creates ComputeExecution record with 'pending' status
  2. Calls provider's `Provision()` method
  3. Updates status to 'running'
  4. On completion: Updates status to 'succeeded'/'failed' + stores resource IDs/errors
  5. Posts callback to workflow system
- Same pattern for Update and Delete operations

### Decision 6: Standardized error handling across providers

**Rationale:**
Workflows need consistent error information regardless of provider. Compute providers (ECS, Kubernetes, etc.) have different error models.

**Implementation:**
- Define `ComputeError` struct with standardized fields:
  ```go
  type ComputeError struct {
    Code string  // e.g., "RESOURCE_EXHAUSTED", "INVALID_CONFIG"
    Message string
    IsRetriable bool
    ProviderError string  // raw provider error for debugging
  }
  ```
- Each provider maps its native errors to ComputeError types
- Workflows check `IsRetriable` to decide retry vs. abort

## Risks / Trade-offs

**Risk: Callback delivery failure → Workflow gets stuck**
- Mitigation: Implement callback retry logic with exponential backoff; workflow can also poll as fallback; store pending callbacks durably

**Risk: Compute operation completes but callback is lost → Inconsistent state**
- Mitigation: Use deterministic execution IDs; workflows can re-query compute status during retry; callback idempotency ensures no duplicate processing

**Risk: Database schema changes for compute tracking require migrations**
- Mitigation: Use auto-migration pattern already in place; make migrations backward-compatible; test with both PostgreSQL and SQLite

**Risk: Workflow providers have different callback mechanisms (Step Functions vs. Temporal)**
- Mitigation: Abstract callback delivery behind WorkflowProvider interface; each provider implements its callback posting logic

**Trade-off: Callback delivery adds latency vs. simplicity**
- Chosen callback route because event-driven is more reliable than polling; latency is acceptable (seconds, not milliseconds)

## Migration Plan

**Phase 1: Database and Data Models**
1. Add database migrations for `compute_executions` and `compute_execution_history` tables
2. Define `ComputeExecution` and `ComputeError` types in `internal/compute/types.go`
3. Implement `Repository` interface for compute executions (Create, Update, Get, List operations)

**Phase 2: ComputeManager Integration**
1. Extend `ComputeManager` with tracking-aware provisioning methods
2. Update existing `ProvisionTenant()` to also create/update `ComputeExecution` records
3. Add status update logic after provider completion

**Phase 3: WorkflowClient Integration**
1. Add compute operation methods to `WorkflowClient` (or create `ComputeWorkflowClient`)
2. Implement request/response patterns for each workflow provider
3. Wire compute manager as a dependency

**Phase 4: Callback System**
1. Implement callback delivery abstraction in workflow layer
2. Add provider-specific callback implementations (Step Functions, Temporal)
3. Wire up compute manager to post callbacks on operation completion

**Phase 5: Testing & Validation**
1. Integration tests for workflow + compute operations
2. End-to-end tests with reconciliation controller
3. Error scenario testing (timeouts, provider failures, etc.)

**Rollback Strategy:**
- Initially enable via feature flag (e.g., `ENABLE_COMPUTE_WORKFLOWS`)
- If issues detected, disable flag and pause new compute workflows
- Keep existing `ProvisionTenant()` working independently for controllers that don't use workflows
- Database migrations are one-way; can write reverse migrations if needed

## Open Questions

1. **Execution ID format**: Should compute execution IDs follow same deterministic pattern as workflow execution IDs (tenant + operation type)? Or keep them separate?
2. **Timeout handling**: What timeout should compute operations have? Should it be configurable per provider? Should timeout lead to retry or abort?
3. **Partial failures**: If a compute operation partially succeeds (e.g., 2/3 containers launched), how should this be represented in the callback? As success? As partial-failure with details?
4. **Billing/metering integration**: Should compute execution tracking integrate with billing/metering, or is that out of scope?
5. **Provider-specific extensions**: If a provider needs to store extra metadata about compute executions, where should that go? (Separate table? JSONB column?)
