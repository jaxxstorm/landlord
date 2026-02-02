## Why

The workflow engine currently triggers in response to API requests and controller reconciliation, but the workflows themselves have no way to actually provision compute infrastructure. Workflows need to invoke a pluggable compute engine to create tenant runtime environments (ECS tasks, containers, etc.). This capability bridges the gap between workflow orchestration and actual infrastructure provisioning, enabling end-to-end tenant lifecycle management from API request to running compute.

## What Changes

- Workflows can invoke compute provisioning operations (create, update, delete tenant compute)
- Compute engine executions are tracked and their status is queryable by workflows
- Workflows receive compute-related callbacks (success, failure, status changes) to drive state transitions
- Compute provisioning failures are surfaced back to workflows for retry/rollback handling
- The compute engine provider interface is extended to support workflow-driven provisioning

## Capabilities

### New Capabilities

- `compute-provisioning-api`: Workflow-accessible API for triggering compute operations (create, update, delete)
- `compute-execution-tracking`: Track execution status of compute provisioning operations (pending, running, succeeded, failed)
- `compute-workflow-callbacks`: Workflows can subscribe to and react to compute operation state changes
- `compute-error-handling`: Structured error handling when compute provisioning fails, enabling workflow-level retry logic

### Modified Capabilities

- `workflow-provisioning`: Extended to include compute execution and callback handling as part of the workflow execution lifecycle

## Impact

- **Compute Provider Interface**: New methods for provisioning operations (e.g., `ProvisionTenant`, `UpdateTenant`, `DeleteTenant`)
- **Workflow Engine**: Must support invoking compute operations and receiving status callbacks
- **Database Schema**: New tables to track compute execution state and history
- **Controller Logic**: May need to handle compute-related state transitions and error conditions
- **Tenant State Machine**: New states may be needed for compute-specific phases (e.g., "compute-provisioning")
