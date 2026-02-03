## Context

Currently, Landlord's reconciler polls workflow providers via `GetExecutionStatus()` to check if workflows are running, but only extracts high-level state (`StateRunning`, `StateSucceeded`, `StateFailed`). Provider-specific details like backing-off, waiting for retries, or intermediate error states are lost in translation. The tenant persists only `Status` (lifecycle phase like "provisioning") and `StatusMessage` (human-readable string), making it impossible for API consumers to distinguish between normal progress and degraded states.

Existing workflow providers (Restate, Step Functions) already return richer execution details in their `ExecutionStatus` responses, but the reconciler discards this information. The `workflow.ExecutionStatus` type includes fields like `State`, `Error`, and provider-specific history, but lacks standardized sub-state classification.

## Goals / Non-Goals

**Goals:**
- Surface workflow execution sub-states (backing-off, retrying, error) in tenant API responses
- Define provider-agnostic canonical workflow states that map to Step Functions, Restate, and future providers
- Persist workflow execution details (sub-state, retry count, error message) in tenant database records
- Update reconciler to extract and store enriched status from workflow providers
- Maintain backward compatibility with existing `tenant.Status` lifecycle states

**Non-Goals:**
- Real-time streaming of workflow events (polling-based approach is sufficient)
- Workflow execution history/audit trail (existing state history table handles transitions)
- Provider-specific UI or debugging tools (focus on API-level observability)
- Changing existing workflow provider interfaces (use existing `GetExecutionStatus` method)

## Decisions

### Decision 1: Add sub-state fields to tenant model, not new lifecycle states

Add new fields (`WorkflowSubState`, `WorkflowRetryCount`, `WorkflowErrorMessage`) to the tenant model and database schema instead of expanding the `Status` enum with intermediate states.

**Rationale:**
- `tenant.Status` represents lifecycle phase (requested, provisioning, ready, failed) and drives state machine transitions
- Workflow execution details are orthogonal to lifecycle - a tenant can be "provisioning" with sub-states of "running", "backing-off", or "retrying"
- Separating concerns keeps state machine logic clean while providing rich observability
- API consumers can filter by `Status` for broad lifecycle queries and inspect sub-state for details

**Alternatives considered:**
- Add new `Status` values like `StatusProvisioningBackoff` or `StatusProvisioningRetrying`
  - Rejected: Pollutes state machine with transient states, makes transitions complex
- Store workflow details only in `StatusMessage` as JSON
  - Rejected: Makes API consumers parse strings, prevents structured filtering/querying

### Decision 2: Define canonical workflow sub-states with provider mapping

Introduce a `WorkflowSubState` enum with provider-agnostic values that workflow providers map their native states to:
- `running`: Workflow is actively executing (Step Functions: RUNNING, Restate: running/active)
- `waiting`: Workflow is paused waiting for external input (Step Functions: N/A, Restate: suspended)
- `backing-off`: Workflow is in exponential backoff between retries (provider-specific retry logic)
- `error`: Workflow encountered an error but may retry (transient failure state)
- `succeeded`: Workflow completed successfully (Step Functions: SUCCEEDED, Restate: completed)
- `failed`: Workflow failed terminally (Step Functions: FAILED/TIMED_OUT/ABORTED, Restate: failed)

**Rationale:**
- Provider-agnostic nomenclature allows swapping workflow engines without API contract changes
- Clear semantics: "backing-off" explicitly indicates retry behavior vs generic "error"
- Maps cleanly to existing Step Functions and Restate states based on current provider code
- Future providers (Temporal, etc.) can map their states to these canonical values

**Alternatives considered:**
- Use provider-specific states directly (e.g., expose Step Functions "RUNNING" vs Restate "active")
  - Rejected: Couples API to workflow provider, breaks abstraction
- Use only high-level states (running, succeeded, failed)
  - Rejected: Loses critical observability into retry/backoff behavior

### Decision 3: Reconciler extracts sub-state from ExecutionStatus.State and Error fields

The reconciler's status polling loop already calls `GetExecutionStatus()`. Enhance it to:
1. Map `ExecutionStatus.State` to canonical `WorkflowSubState`
2. Extract retry count from provider-specific logic (Step Functions event history count, Restate retry metadata)
3. Store `ExecutionStatus.Error.Message` in `WorkflowErrorMessage` when present
4. Update tenant record with enriched fields alongside `StatusMessage` update

**Rationale:**
- Minimal changes to reconciler flow - reuses existing polling logic
- No changes to workflow provider interface - `GetExecutionStatus` already returns needed data
- Centralizes mapping logic in reconciler rather than duplicating in each provider
- Preserves existing `StatusMessage` for backward compatibility

**Alternatives considered:**
- Add new `GetDetailedStatus()` method to workflow provider interface
  - Rejected: Unnecessary - existing method returns sufficient data
- Store full `ExecutionStatus` JSON in tenant record
  - Rejected: Over-normalization, difficult to query, includes provider-specific noise

### Decision 4: Database schema adds three nullable fields

Add to `tenants` table:
```sql
workflow_sub_state VARCHAR(50) NULL
workflow_retry_count INTEGER NULL DEFAULT 0
workflow_error_message TEXT NULL
```

Fields are nullable and only populated when `workflow_execution_id` is set and status is in-flight (provisioning, updating, deleting).

**Rationale:**
- Nullable design: fields only meaningful when workflow is active
- `workflow_retry_count` defaults to 0 for initial execution
- `workflow_error_message` stores last known error, cleared on retry success
- No foreign keys or joins needed - denormalized for query performance

**Alternatives considered:**
- Create separate `workflow_executions` table with 1:N relationship
  - Rejected: Overkill for current use case, adds join complexity
- Store in existing `status_message` as structured JSON
  - Rejected: Already decided against for query ergonomics

## Risks / Trade-offs

**[Risk]** Provider-specific retry semantics may not map cleanly to canonical states → **Mitigation:** Document mapping expectations in workflow provider interface spec, use best-effort mapping with fallback to "running"

**[Risk]** Retry count extraction requires parsing provider-specific history (Step Functions) or metadata (Restate) → **Mitigation:** Make retry count optional; reconciler returns 0 if unavailable; spec this as "SHOULD" not "MUST"

**[Risk]** Increased database writes on every reconciler poll cycle → **Mitigation:** Only update tenant record if sub-state/retry/error fields actually changed (dirty checking in reconciler)

**[Risk]** Migration adds new fields to high-traffic `tenants` table → **Mitigation:** Fields are nullable with no constraints, migration is non-blocking; old code ignores new fields

**[Trade-off]** Denormalized storage of workflow details duplicates data from provider → Accepted for query performance and API simplicity

**[Trade-off]** Polling-based approach has latency vs real-time event streaming → Accepted; reconciler interval (10-30s) is sufficient for observability use case

## Migration Plan

1. Add database migration with three new nullable columns to `tenants` table
2. Update `tenant.Tenant` Go struct with new fields and JSON tags
3. Extend `workflow.ExecutionStatus` type with sub-state mapping helper methods
4. Update reconciler's status polling logic to extract and persist enriched fields
5. Extend tenant GET/LIST API response serialization to include new fields
6. Update CLI output formatters to display sub-state information
7. Add integration tests verifying sub-state transitions during workflow execution

**Rollback strategy:** 
- Migration has corresponding `.down.sql` to drop columns
- Old code ignores new fields (backward compatible)
- If bugs discovered, can NULL out fields and redeploy without migration rollback

**Deployment:**
1. Deploy database migration (non-breaking, adds nullable columns)
2. Deploy API/reconciler code with new fields (reads/writes new columns)
3. Verify in staging with Restate/Step Functions providers
4. Roll out to production with monitoring on reconciler update rate

## Open Questions

- Should we expose workflow execution history/timeline in API responses? (Deferred: current scope is current state only)
- What retry count threshold should trigger alerting? (Deferred: observability/alerting is separate concern)
- Should we add GraphQL subscriptions for real-time sub-state updates? (Deferred: polling is sufficient for v1)
- How should we handle workflow providers that don't support retry count extraction? (Answer: return 0, document as optional)
