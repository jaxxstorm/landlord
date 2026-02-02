## Why

Workers currently connect directly to the database, coupling workflow execution to Postgres and making providers difficult to swap. We need Landlord to own tenant state (Postgres), workflow providers to orchestrate execution (Restate/Temporal), and workers to be stateless handlers invoked by providers.

## What Changes

- **BREAKING** Remove database access from workers; workers become stateless HTTP handlers.
- **EXTEND** `workflow.Provider` interface with Invoke()/GetWorkflowStatus() methods (keep existing methods for compatibility).
- **MODIFY** Restate and Step Functions providers to implement new interface methods.
- **ENHANCE** existing `internal/controller/reconciler.go` to invoke workflows and poll status.
- **REUSE** tenant.Status, tenant.DesiredConfig, tenant.ObservedConfig for workflow state (no new DB fields).
- **UPDATE** worker handlers to receive full context from workflow payload (no DB queries).

## Capabilities

### Modified Capabilities
- `workflow-provider-interface`: Extend Provider with Invoke()/GetWorkflowStatus() methods.
- `tenant-reconciler`: Enhance existing reconciler with workflow invocation and status polling.
- `workflow-worker-engines`: Workers become stateless, receive full context in payload.
- `workflow-provisioning`: Reconciler invokes providers via workflow.Manager.
- `workflow-status-tracking`: Reconciler polls provider status, updates tenant.Status.
- `restate-worker-engine`: Remove DB access; receive context from Restate payload.
- `restate-workflow-provider`: Implement new Provider interface methods.
- `step-functions-provider`: Implement new Provider interface methods.
- `tenant-lifecycle-workflows`: Use existing tenant.Status for workflow state.
- `api-workflow-triggering`: Create tenant sets Status=StatusRequested; reconciler handles rest.

## Impact

- Worker binaries (no DB credentials required).
- Landlord API (reconciler runs in main process or sidecar).
- Workflow provider interfaces and implementations.
- Configuration (reconciler interval, provider endpoints).
