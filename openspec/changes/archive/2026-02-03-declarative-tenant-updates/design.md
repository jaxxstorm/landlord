## Context

Currently, when a tenant's workflow enters a backing-off or retrying state due to bad configuration, updating the tenant config via the API does not stop the failing workflow. The workflow continues to retry with the original bad configuration indefinitely, even though the tenant record now has correct config. Users must manually intervene by deleting/archiving and recreating the tenant, which restarts the workflow.

This behavior differs from delete/archive operations, which properly stop in-flight workflows before completing. The lack of workflow restart on config updates makes the system less declarative - users can't simply "fix the config and move on."

**Current Architecture:**
- Reconciler polls tenants and checks workflow status
- When workflow completes (success/failure), reconciler updates tenant status
- Tenant update API (`PUT /tenants/{id}`) writes new config to DB but doesn't touch workflows
- Workflows run independently based on execution started at tenant creation/transition time

**Constraints:**
- Must preserve idempotency - multiple updates shouldn't cause duplicate restarts
- Must not disrupt healthy workflows (running provisioning, already completed)
- Must work with existing workflow provider interface (`StopExecution` already exists)
- Must integrate with existing reconciler polling architecture

## Goals / Non-Goals

**Goals:**
- Automatically restart workflows when tenant config changes while workflow is in degraded state
- Provide declarative recovery from configuration errors (set → fix)
- Maintain existing behavior for healthy workflows (no interruption)
- Preserve idempotency and audit trail of state transitions

**Non-Goals:**
- Hot-reload config into running workflows (workflows are immutable once started)
- Automatic rollback of bad config changes (users must explicitly revert)
- Workflow restart on every config change regardless of state (only restart degraded workflows)
- Changes to workflow provider implementations (use existing `StopExecution` interface)

## Decisions

### Decision 1: Detect Config Changes in Reconciler vs API Layer

**Chosen:** Reconciler detects and handles restart

**Rationale:**
- Reconciler already has workflow state visibility and orchestration logic
- API layer is thin - just validates and writes to DB
- Keeps workflow lifecycle decisions centralized in controller
- Reconciler polls regularly, so restart happens within reconciliation loop naturally

**Alternatives Considered:**
- API layer triggers restart synchronously: Would require exposing workflow operations to API layer, breaking separation of concerns. API responses would be slower.
- Event-driven webhook on config change: Adds complexity; reconciler already polls and handles workflow coordination.

### Decision 2: Config Change Detection Strategy

**Chosen:** Compare tenant config hash/fingerprint on each reconciliation pass

**Implementation:**
- Add `ConfigHash` field to Tenant model (compute SHA256 of compute_config JSON)
- Store hash when workflow starts
- On each reconcile pass, compare current tenant.ComputeConfig hash with stored hash in workflow metadata
- If hash differs + workflow in degraded state → trigger restart

**Rationale:**
- Simple to implement - just hash comparison
- Already have workflow metadata for storing context
- No new DB tables or migration complexity
- Works with existing reconciler polling architecture

**Alternatives Considered:**
- Track previous config in separate table: More storage overhead, additional queries
- Use tenant UpdatedAt timestamp: Not reliable (updates could be unrelated to compute config)
- Compare entire config objects: More expensive than hash comparison

### Decision 3: Which Workflow States Trigger Restart

**Chosen:** Only restart workflows in SubStateBackingOff or SubStateRetrying

**Rationale:**
- These states indicate workflow is failing due to provisioning issues (often config errors)
- Workflows in other states either:
  - Running actively (SubStateRunning) - shouldn't interrupt
  - Completed successfully (StateDone, SubStateSuccess) - no workflow to restart
  - Failed permanently (StateDone, SubStateError) - already terminal, handled separately

**Implementation Check:**
```go
if execStatus.State == workflow.StateActive && 
   (execStatus.SubState == workflow.SubStateBackingOff || 
    execStatus.SubState == workflow.SubStateRetrying) {
    // Check for config change and restart
}
```

**Alternatives Considered:**
- Restart all active workflows: Too aggressive, would interrupt healthy provisioning
- Include SubStateError: These are terminal failures, should be handled by existing error flow

### Decision 4: Restart Sequencing

**Chosen:** Stop old workflow → Wait for confirmation → Start new workflow

**Rationale:**
- Prevents duplicate workflows (two competing for same tenant)
- Clean audit trail (one execution ends, new one begins)
- Idempotent - if stop fails, reconciler retries; if start fails, existing error handling applies

**Implementation Flow:**
1. Call `workflowProvider.StopExecution(executionID, "Config updated")`
2. Poll workflow status until State == StateDone
3. Update tenant.WorkflowExecutionID = "" to clear old execution
4. Call `workflowClient.TriggerWorkflow(tenant, "update")` to start fresh workflow
5. Update tenant with new execution ID

**Alternatives Considered:**
- Start new workflow immediately: Risk of duplicate executions competing for same resources
- Stop and wait for user to manually trigger: Less declarative, more operational overhead

### Decision 5: Store Config Hash in Workflow Metadata

**Chosen:** Pass config hash in workflow execution metadata at start time

**Implementation:**
- When calling `StartExecution`, include `ConfigHash` in ExecutionInput.Metadata
- Reconciler reads metadata from `GetExecutionStatus` response
- Compare metadata ConfigHash with current tenant compute_config hash

**Rationale:**
- Workflow providers already support metadata (we recently fixed metadata visibility)
- No new DB fields or schema changes needed
- Hash travels with execution, making comparison simple
- Already have GetProvider() to access execution status with metadata

**Alternatives Considered:**
- Store hash in tenant DB record: Would need separate field, potential race conditions
- Track in separate reconciler state: Loses hash on controller restart

## Risks / Trade-offs

**Risk: Race condition during stop/start window**
→ **Mitigation:** Clear tenant.WorkflowExecutionID during stop phase. Reconciler won't poll status if execution ID is empty. Only set new execution ID after successful start.

**Risk: Thrashing - repeated config changes cause repeated restarts**
→ **Mitigation:** Workflow state must be degraded (backing-off/retrying) to trigger restart. If new workflow starts successfully and reaches running state, won't restart again unless it also degrades.

**Risk: Loss of workflow retry context**
→ **Trade-off:** Accepted. When workflow restarts, retry counters reset. This is desired behavior - the fix changes the input, so we want a fresh attempt, not continuation of old retry logic.

**Risk: StopExecution might not be immediate**
→ **Mitigation:** Reconciler polls until State == StateDone before starting new workflow. If provider takes time to stop, we wait. If stop fails, reconciler retries on next poll.

**Risk: Config hash collision (different configs produce same hash)**
→ **Mitigation:** Using SHA256 makes collision astronomically unlikely. For tenant config sizes (KB), collision probability is negligible.

## Migration Plan

**Deployment:**
1. Add ConfigHash field to ExecutionInput metadata (backward compatible - old executions won't have it)
2. Update workflowClient.TriggerWorkflow to compute and include config hash
3. Add config change detection logic to reconciler
4. Deploy controller with new code
5. Existing in-flight workflows without ConfigHash continue as before (no restarts)
6. New workflows started after deployment include hash and support restart behavior

**Rollback Strategy:**
- If issues arise, revert controller deployment
- Existing workflows continue unaffected (no DB schema changes)
- No data migration needed (hash is computed at runtime)

**Testing:**
- Unit test: Config hash computation and comparison logic
- Integration test: Create tenant with bad config → workflow backs off → update config → verify restart
- Integration test: Verify healthy workflows (running, completed) are NOT restarted on config change

## Open Questions

None at this time. Design is ready for implementation.
