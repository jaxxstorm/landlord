# Tenant State Machine

This document describes the state machine that governs tenant lifecycle transitions in the Landlord controller.

## States

The tenant state machine defines the following states:

### Non-Terminal States

- **requested**: Initial state when a tenant is created via API. System is validating the request.
- **planning**: Optional state when a plan phase is enabled. Otherwise skipped by reconciliation.
- **provisioning**: Resources are being created. Workflow is executing, compute/networking being provisioned.
- **updating**: Tenant is being modified (image update, config change). Temporary state during reconciliation.
- **deleting**: Tenant deletion in progress. Resources are being torn down.

### Terminal States

- **ready**: Tenant is fully operational and serving traffic. Desired state matches observed state.
- **archived**: Tenant resources cleaned up; record retained for audit/history.
- **failed**: Operation failed, manual intervention may be required. StatusMessage contains error details.

## State Transitions

```
                           ┌─────────────┐
                           │  requested  │
                           └──────┬──────┘
                                  │
                        (optional planning)
                                  │
                                  v
                          ┌───────────────┐
                          │ provisioning  │
                          └──────┬────────┘
                                 │
                       ┌─────────┴──────────┐
                       │                    │
                   done             execution error
                       │                    │
                       v                    v
                  ┌────────┐          ┌──────────┐
                  │ ready  │          │  failed  │
                  └───┬────┘          └──────────┘
                      │
                      │ user triggers update/delete
                      │
                  ┌───┴──────────┐
                  │              │
            update request   delete request
                  │              │
                  v              v
             ┌────────┐     ┌─────────┐
             │updating│     │ deleting│
             └───┬────┘     └────┬────┘
                 │               │
                 │        ┌──────┴─────────┐
                 │        │                │
               done   done       execution error
                 │        │                │
                 v        v                v
             ┌────────┐ ┌─────────┐   ┌──────────┐
             │ ready  │ │ archived│   │  failed  │
             └────────┘ └─────────┘   └──────────┘
```

## Transition Rules

### From Requested

- → **provisioning**: When reconciliation invokes the provisioning workflow
- → **failed**: When validation fails or plan generation fails

### From Planning

- → **provisioning**: When plan succeeds and provisioning starts
- → **failed**: When plan fails

### From Provisioning

- → **ready**: When all resources are provisioned successfully
- → **failed**: When provisioning fails

### From Ready

- → **updating**: When configuration or image update is needed
- → **deleting**: When tenant deletion is requested
- No self-transitions (stays ready while healthy)

### From Updating

- → **ready**: When update completes successfully
- → **failed**: When update fails

### From Deleting

- → **archived**: When all resources are cleaned up
- → **failed**: When deletion fails

### From Failed

- → **deleting**: Only valid transition - clean up failed tenant

### From Archived

- Terminal state, no transitions possible

## Error Handling

When a workflow execution fails during any transition:

1. The error is classified as **retryable** or **fatal**
2. For retryable errors: Item is re-queued with exponential backoff (1s initial, 5min max)
3. For fatal errors: Tenant transitions to **failed** state immediately
4. After max retries (default 5): Tenant transitions to **failed** state

## Reconciliation Loop

The reconciler polls the database at configured intervals and processes tenants in non-terminal states:

1. Fetch all tenants in states: requested, planning, provisioning, updating, deleting
2. Add them to work queue for processing
3. Workers process queue items and trigger appropriate workflows
4. Successful workflows advance tenant to next state
5. Failed workflows either retry or transition to failed state

Workflow workers are stateless HTTP handlers invoked by the workflow provider; they do not read or write tenant state directly in the database.

Only tenants in **non-terminal** states (not ready, not archived, not failed) are included in reconciliation polling.

### Config Change Restart Behavior

When a tenant's configuration changes while a workflow is in-flight, the reconciler applies declarative semantics to ensure the workflow uses the latest config:

**Restart Conditions** (all must be true):
1. Tenant has an active workflow execution (`workflow_execution_id` != null)
2. Workflow is in **degraded state** (`workflow_sub_state` = "backing-off")
3. Configuration has changed (current config hash ≠ stored `workflow_config_hash`)

**Degraded States** (eligible for restart):
- `backing-off`: Workflow backing off due to failures, likely caused by config errors

**Healthy States** (NOT restarted):
- `running`: Actively provisioning, restart would lose progress
- `succeeded`: Completed successfully, no restart needed
- `failed`: Terminal state, handled separately
- `waiting`: Waiting for external event, not an error condition

**Restart Flow:**
```
Config Change Detected (hash comparison)
         ↓
Is workflow degraded? (backing-off)
         ↓ YES
Stop workflow execution (reason: "Configuration updated")
         ↓
Poll until terminal state (30s timeout)
         ↓
Clear execution_id, error_message, retry_count
         ↓
Trigger new workflow with updated config
         ↓
Store new config_hash in tenant record
         ↓
Continue normal reconciliation
```

**Example Scenario:**
1. Tenant created with invalid Docker image → workflow provisions → fails → backs off
2. User updates tenant config with correct image via API: `PUT /tenants/{id}`
3. Reconciler detects config change during next status poll (30s interval)
4. Reconciler stops backing-off workflow
5. Reconciler triggers new workflow with corrected image
6. New workflow succeeds → tenant transitions to `ready`

**Observability:**
- Log message: `"config changed while workflow degraded, restarting workflow"`
- Includes: `tenant_id`, `old_config_hash`, `new_config_hash`, `execution_id`
- Followed by: `"stopping workflow execution"`, `"new workflow triggered after config change"`

**Notes:**
- Config hash not retroactively added to old tenants (backward compatibility)
- Only JSON-serializable fields in `desired_config` are hashed
- Hash computation is deterministic (same config → same hash)
- Stop polling has 30-second timeout to prevent infinite wait

## Metrics

The state machine integrates with observability systems:

- **reconciliation_duration**: Time taken to complete one reconciliation cycle
- **state_transition_count**: Counter of state transitions by (from_state, to_state)
- **retry_count**: Histogram of retry counts before successful transition
- **error_rate**: Rate of failed reconciliations by error type
