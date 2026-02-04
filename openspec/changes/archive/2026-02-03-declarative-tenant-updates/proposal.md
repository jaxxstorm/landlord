## Why

When a tenant's workflow enters a backing-off or retrying state due to invalid configuration, fixing the tenant's config via `set` does not stop the failing workflow. The old workflow continues to retry with the bad configuration indefinitely. This differs from delete/archive operations which properly stop old workflows before starting new ones. Users must manually intervene to stop backing-off workflows, making the system less declarative and requiring operational overhead to recover from config errors.

## What Changes

- Add workflow lifecycle management to the tenant update flow
- Detect and stop in-flight workflows that are in degraded states (backing-off, retrying) when config is updated
- Start a fresh workflow execution with the corrected configuration after stopping the old one
- Preserve existing behavior for workflows in healthy states (running, completed)
- Add reconciliation logic to detect when a tenant config change should trigger workflow restart
- Update tenant update API to handle workflow restart coordination

## Capabilities

### New Capabilities
- `workflow-restart-on-update`: Coordinate stopping degraded workflows and starting fresh ones when tenant config changes
- `degraded-workflow-detection`: Identify workflows in backing-off/retrying states that should be stopped on config update

### Modified Capabilities
- `tenant-update-api`: Add workflow restart orchestration when config changes while workflow is degraded
- `tenant-lifecycle-workflows`: Define when workflows should be restarted vs continued during updates
- `workflow-execution-status`: Ensure sub-state visibility supports degraded state detection for restart decisions

## Impact

**Affected Code:**
- `internal/controller/reconciler.go` - Add logic to detect config changes and degraded workflow states
- `internal/api/tenants.go` - Update endpoint to coordinate workflow restart on config update
- `internal/workflow/` - Add workflow stop/restart coordination methods
- `internal/tenant/` - Track previous config state to detect meaningful changes

**API Changes:**
- No breaking changes to tenant update API contract
- Internal workflow coordination added transparently

**Behavior Changes:**
- Tenant updates will now automatically restart backing-off/retrying workflows
- Users get declarative recovery from bad configs without manual intervention
- Workflows in healthy states (running, completed) continue unaffected
