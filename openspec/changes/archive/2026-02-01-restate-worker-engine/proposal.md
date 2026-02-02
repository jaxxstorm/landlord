## Why

The current Restate workflow integration requires manual worker registration and only drives tenant provisioning, which blocks automated end-to-end lifecycle execution. We need a pluggable worker engine that can register itself and execute create/update/delete workflows across workflow providers.

## What Changes

- Add a pluggable worker engine interface to the workflow subsystem, allowing multiple workflow engines (e.g., Restate, Step Functions) to register and run workers.
- Implement a full Restate worker engine that registers itself on startup and drives workflow execution.
- Extend the tenant workflow execution path to support create, update, and delete lifecycle operations.
- Add an end-to-end Restate integration test that provisions, updates, and deletes a tenant through the worker.
- Enable workers to resolve compute engine selection from the landlord server (e.g., via API lookup) to support dynamic compute targets.

## Capabilities

### New Capabilities
- `workflow-worker-engines`: Pluggable worker engine registration and execution for workflow providers.
- `restate-worker-engine`: Restate-specific worker implementation with automatic registration and lifecycle execution.
- `tenant-lifecycle-workflows`: Workflow-driven handling of tenant create, update, and delete operations.

### Modified Capabilities
- `restate-workflow-provider`: Extend provider requirements to include worker registration/execution semantics.
- `restate-workflow-registration`: Update registration requirements to allow automatic worker registration on startup.

## Impact

- Workflow engine interfaces and adapters, including Restate provider and worker code (`cmd/workers/restate`, `internal/workflow/providers/restate`).
- Landlord API/clients for compute engine discovery and worker configuration.
- Integration tests and local dev workflow for Restate worker startup and registration.
- Documentation/configuration related to workflow engine selection and worker registration.
