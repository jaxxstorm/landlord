## Context

The landlord system currently has a functional reconciliation controller that polls the database every 10 seconds and triggers workflows for tenants in non-terminal states. The HTTP API handlers (POST /tenants, PUT /tenants/{id}, DELETE /tenants/{id}) only perform database mutations without triggering workflows, creating a 0-10 second delay before infrastructure changes begin.

**Current State:**
- API handlers in `internal/api/server.go` create/update/delete tenant records
- Reconciliation controller polls database and triggers workflows via `WorkflowClient`
- Workflow manager (`internal/workflow/manager.go`) provides provider abstraction (Restate, Step Functions)
- Compute providers (Docker, ECS) are invoked by workflows, not directly by API
- Tenant table has `workflow_execution_id` column for tracking

**Constraints:**
- Must maintain pluggable workflow provider abstraction (no hardcoding Restate/Step Functions)
- Must maintain pluggable compute provider abstraction (no hardcoding Docker/ECS)
- Must coordinate with existing reconciliation controller to avoid duplicate triggers
- Must handle transaction atomicity (database update + workflow trigger as a unit)
- Single-process operation (API server and controller run in same process)
- Must preserve RESTful API semantics

**Stakeholders:**
- API users (immediate feedback on provisioning actions)
- Workflow providers (receive triggers from both API and controller)
- Operations team (monitoring, debugging workflow execution)

## Goals / Non-Goals

**Goals:**
- API handlers immediately trigger workflows after database mutation (no waiting for controller poll)
- Maintain dual-trigger pattern: API provides immediate response, controller provides eventual consistency/retry
- Return workflow execution ID in API responses for tracking
- Prevent duplicate workflow triggers when API and controller both attempt to trigger
- Preserve pluggable provider architecture (workflow and compute providers)
- Update tenant status appropriately before triggering workflow (requested→planning, ready→updating, etc.)
- Handle workflow trigger failures gracefully with appropriate HTTP status codes

**Non-Goals:**
- Synchronous workflow execution (API still returns immediately, workflows run async)
- Removing the reconciliation controller (it provides retry/recovery capabilities)
- Direct compute provider integration from API (workflows remain the orchestration layer)
- Webhook/callback mechanisms for workflow completion (future enhancement)
- Multi-process coordination or distributed locking (single process for now)

## Decisions

### Decision 1: API Handler Integration Pattern

**Choice:** Inject workflow manager into API server, use existing `WorkflowClient` pattern

**Rationale:**
- **Consistency**: Controller already uses `WorkflowClient` wrapper around workflow manager
- **Reusability**: Same workflow triggering logic (action determination, input preparation) used by API and controller
- **Testability**: Existing `WorkflowClient` is well-tested, can mock workflow manager in API tests
- **Separation of concerns**: API handlers focus on HTTP concerns, `WorkflowClient` handles workflow triggering

**Implementation:**
- Add `workflowClient *controller.WorkflowClient` field to `Server` struct
- Inject workflow client in `New()` constructor alongside tenant repository
- Call `workflowClient.TriggerWorkflow()` in POST/PUT/DELETE handlers after database operations

**Alternatives considered:**
- Direct workflow manager integration in API: Duplicates logic from `WorkflowClient`
- New API-specific workflow wrapper: Creates code duplication
- Inline workflow calls in handlers: Mixes concerns, harder to test

**Trade-off:** Creates dependency from `internal/api` → `internal/controller`, but acceptable since both are internal packages

### Decision 2: Deduplication Strategy

**Choice:** Idempotency via database `workflow_execution_id` field, controller checks before re-triggering

**Rationale:**
- **Simple**: No distributed locks, no coordination primitives needed
- **Database-backed**: Tenant table already has `workflow_execution_id` column
- **Race-tolerant**: If API and controller both trigger, second trigger is harmless (workflow manager handles duplicate execution IDs)
- **Observable**: Execution ID in database provides audit trail of which trigger won the race

**Implementation:**
- API handler: Update tenant status + set `workflow_execution_id`, then trigger workflow
- Controller: In `reconcile()`, check if `workflow_execution_id` is already set and workflow is still running
- If execution ID exists and workflow is active, skip triggering (no-op)
- If execution ID exists but workflow failed/completed, re-trigger with new ID (retry case)

**Alternatives considered:**
- Distributed locks (Redis, etcd): Adds external dependency, overkill for single-process
- Database advisory locks: PostgreSQL-specific, breaks SQLite compatibility
- Optimistic locking with version field: Already used for concurrent updates, doesn't prevent double-triggers
- Skip controller triggering entirely: Loses retry/recovery capabilities

**Trade-off:** Small window where both API and controller might trigger workflows (if status update and workflow trigger aren't atomic), but workflow providers handle this gracefully

### Decision 3: Transaction Boundary

**Choice:** Update database first, then trigger workflow (non-atomic)

**Rationale:**
- **Practical**: Workflow trigger is external HTTP call (to Restate/AWS), cannot be part of database transaction
- **Failure modes preferred**: Better to have database updated with no workflow trigger (controller retries) than workflow triggered with no database update (orphan workflow)
- **Consistency**: Controller already uses this pattern (update DB, then trigger), API should match
- **Eventual consistency**: Reconciliation controller provides safety net if workflow trigger fails

**Implementation:**
1. Begin database transaction
2. Update tenant record (status change, set workflow_execution_id)
3. Commit transaction
4. Trigger workflow via WorkflowClient (outside transaction)
5. If workflow trigger fails, return HTTP 500 but tenant is in consistent state for controller retry

**Alternatives considered:**
- Atomic transaction with workflow trigger: Impossible (workflow is external HTTP call)
- Trigger workflow first, then update DB: Orphan workflows if DB update fails
- Two-phase commit: Too complex for this use case
- Message queue (Kafka, RabbitMQ): Adds dependency, overkill for single-process

**Trade-off:** Workflow trigger failure leaves tenant in intermediate state (e.g., "planning" status with no active workflow), but controller will retry within 10 seconds

### Decision 4: API Response Structure

**Choice:** Return workflow execution ID in response body, use HTTP 202 Accepted for async operations

**Rationale:**
- **Standard RESTful pattern**: 202 Accepted indicates "request accepted but processing not complete"
- **Tracking**: Execution ID allows clients to poll workflow status (future GET /tenants/{id}/workflow endpoint)
- **Backwards compatibility**: Can add execution_id field to existing response schemas without breaking existing clients
- **Observability**: Clients can log execution ID for debugging

**Response format:**
```json
{
  "id": "uuid",
  "tenant_id": "my-app",
  "status": "planning",
  "workflow_execution_id": "tenant-my-app-plan",
  // ... other tenant fields
}
```

**Status codes:**
- POST /tenants: 202 Accepted (provisioning started)
- PUT /tenants/{id}: 202 Accepted (update started)
- DELETE /tenants/{id}: 202 Accepted (deletion started)
- Workflow trigger failure: 500 Internal Server Error (with error details)

**Alternatives considered:**
- HTTP 200 OK: Misleading (implies operation completed)
- HTTP 201 Created: Only for POST, doesn't work for PUT/DELETE
- Separate response object with execution details: Over-engineered for MVP

**Trade-off:** 202 changes existing API contract (currently returns 200), may break clients expecting synchronous completion

### Decision 5: Error Handling Strategy

**Choice:** Return HTTP 500 on workflow trigger failure, rely on controller for retry

**Rationale:**
- **Honest failure reporting**: If workflow trigger fails, API should indicate failure (don't return 202)
- **Client retry**: Client can retry the entire operation (POST/PUT/DELETE)
- **Controller safety net**: Even if API returns 500, tenant is in retriable state for controller
- **Observability**: Error logs with execution ID attempt help debugging

**Error scenarios:**
1. **Workflow provider unavailable**: Return 500, log error, controller will retry
2. **Invalid workflow specification**: Return 400 Bad Request (client error)
3. **Database update fails**: Return appropriate error (409 Conflict for version mismatch, 500 for DB error)
4. **Workflow trigger timeout**: Return 500, controller will retry

**Implementation:**
```go
// After DB update
executionID, err := s.workflowClient.TriggerWorkflow(ctx, tenant, "plan")
if err != nil {
    s.logger.Error("workflow trigger failed", 
        zap.String("tenant_id", tenant.TenantID),
        zap.Error(err))
    return s.respondError(w, http.StatusInternalServerError, 
        "Failed to trigger provisioning workflow")
}
// Return 202 with execution ID
```

**Alternatives considered:**
- Return 202 even on failure: Misleading, client doesn't know operation failed
- Rollback database changes on workflow failure: Loses idempotency, complicates logic
- Queue failed triggers for later retry: Adds complexity, controller already provides this

**Trade-off:** Transient workflow provider failures cause API errors, but controller provides automatic retry

### Decision 6: Status Transition Handling

**Choice:** API handlers use state machine logic to determine next status before workflow trigger

**Rationale:**
- **Consistency**: Same state transitions as controller (requested→planning, ready→updating)
- **Validation**: Prevents invalid transitions (e.g., ready→planning)
- **Reusability**: Controller's state machine logic (`internal/controller/state_machine.go`) can be extracted to shared package

**Implementation:**
- Move `nextStatus()` and `shouldReconcile()` functions from `internal/controller/state_machine.go` to `internal/tenant/state_machine.go`
- API handlers call `tenant.NextStatus(currentStatus, action)` before updating database
- Reject invalid transitions with HTTP 409 Conflict

**Alternatives considered:**
- Hardcode status transitions in API handlers: Duplicates logic, diverges from controller
- Let workflow determine next status: Too much logic in workflow layer
- Allow any status transition: Breaks state machine invariants

**Trade-off:** Requires refactoring state machine code to shared package, but improves consistency

## Risks / Trade-offs

### Risk: Duplicate workflow triggers (API + controller race)

**Mitigation:**
- Workflow execution IDs are deterministic based on tenant ID + action + timestamp
- Workflow providers (Restate, Step Functions) handle duplicate execution IDs gracefully (idempotency)
- Controller checks `workflow_execution_id` field before triggering
- Logs include execution ID for debugging duplicate triggers

### Risk: Workflow trigger fails but database updated (partial failure)

**Mitigation:**
- Return HTTP 500 to indicate failure to client
- Tenant remains in intermediate state (e.g., "planning" without active workflow)
- Reconciliation controller retries within 10 seconds
- `workflow_execution_id` is NULL, indicating no successful trigger yet

### Risk: Increased API latency (workflow trigger adds HTTP call)

**Mitigation:**
- Workflow trigger has 30s timeout (configurable)
- Most workflow triggers complete in <100ms
- Future optimization: Async workflow trigger with job queue

### Risk: Breaking change to API response (202 vs 200 status codes)

**Mitigation:**
- Document in CHANGELOG as breaking change
- Version API if needed (e.g., /v2/tenants)
- Add feature flag to enable/disable immediate workflow triggering
- Monitor client errors after deployment

### Risk: State machine logic drift between API and controller

**Mitigation:**
- Extract state machine to shared `internal/tenant/state_machine.go` package
- Comprehensive unit tests for state transitions
- Integration tests covering both API and controller paths
- Code review checklist for state transition changes

## Migration Plan

### Phase 1: Refactor State Machine (Preparation)
1. Extract `nextStatus()`, `shouldReconcile()`, and transition validation from `internal/controller/state_machine.go` to `internal/tenant/state_machine.go`
2. Update controller to use new shared package
3. Add unit tests for state machine package
4. Deploy and verify controller still works correctly

### Phase 2: Integrate Workflow Client into API
1. Add `workflowClient` field to `Server` struct
2. Update `New()` constructor to accept workflow client
3. Update `cmd/landlord/main.go` to pass workflow client to API server
4. No behavior change yet (handlers don't call workflow client)
5. Deploy and verify no regressions

### Phase 3: Enable Workflow Triggering in API Handlers
1. Update `handleCreateTenant` to trigger "plan" workflow after database insert
2. Update `handleUpdateTenant` to trigger "update" workflow after database update
3. Update `handleDeleteTenant` to trigger "delete" workflow after status change to "deleting"
4. Change response status codes to 202 Accepted
5. Add integration tests for API workflow triggering
6. Deploy with feature flag disabled

### Phase 4: Enable Feature and Monitor
1. Enable feature flag in production
2. Monitor metrics: workflow trigger latency, error rates, duplicate triggers
3. Monitor API response times
4. Validate controller still handles retries correctly

### Rollback Strategy
- Disable feature flag to revert to database-only mutations
- Controller continues to handle workflow triggering
- No data loss or corruption (state machine maintains consistency)
- Rollback can happen at any phase without breaking state

### Validation
- Unit tests: State machine transitions, workflow trigger logic
- Integration tests: API → workflow → database flow
- Load tests: API latency under workflow triggering load
- Chaos tests: Workflow provider failures, database failures

## Open Questions

1. **Should we add a /tenants/{id}/workflow endpoint to query workflow execution status?**
   - Decision deferred: Current implementation returns execution ID in tenant response, dedicated endpoint can be added later

2. **How should we handle workflow provider circuit breaking?**
   - Decision deferred: Workflow manager doesn't currently implement circuit breaking, this is a future enhancement

3. **Should DELETE operation wait for workflow completion before removing database record?**
   - Decision: No, DELETE updates status to "deleting" and triggers workflow, controller removes record after workflow completes (soft delete pattern)

4. **What metrics should we track for API workflow triggering?**
   - Proposed: workflow_trigger_duration, workflow_trigger_errors_total, workflow_duplicates_prevented_total
   - Final metrics defined in observability spec
