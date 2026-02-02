## ADDED Requirements

### Requirement: Trigger workflow based on tenant status
The reconciler SHALL trigger appropriate workflows through the workflow manager based on the tenant's current status.

#### Scenario: Planning workflow is triggered for requested tenants
- **WHEN** a tenant with status "requested" is reconciled
- **THEN** the reconciler MUST call workflow manager with action "plan" and tenant details

#### Scenario: Provisioning workflow is triggered for planning tenants
- **WHEN** a tenant with status "planning" is reconciled AND planning is complete
- **THEN** the reconciler MUST call workflow manager with action "provision" and tenant details

#### Scenario: Update workflow is triggered for configuration changes
- **WHEN** a tenant with status "ready" has detected configuration drift
- **THEN** the reconciler MUST call workflow manager with action "update" and tenant details

#### Scenario: Deletion workflow is triggered for deleting tenants
- **WHEN** a tenant with status "deleting" is reconciled
- **THEN** the reconciler MUST call workflow manager with action "delete" and tenant details

### Requirement: Pass tenant context to workflow execution
The reconciler SHALL provide complete tenant information to the workflow manager for execution.

#### Scenario: Tenant metadata is included in workflow input
- **WHEN** a workflow is triggered
- **THEN** the input MUST include tenant ID, name, organization ID, and configuration

#### Scenario: Current status is included in workflow input
- **WHEN** a workflow is triggered
- **THEN** the input MUST include the tenant's current status

#### Scenario: Resource specifications are included
- **WHEN** a provisioning workflow is triggered
- **THEN** the input MUST include compute specifications (CPU, memory, storage)

### Requirement: Handle workflow execution responses
The reconciler SHALL process workflow execution results and update tenant status accordingly.

#### Scenario: Successful workflow updates tenant status
- **WHEN** a workflow completes successfully
- **THEN** the reconciler MUST update the tenant to the next status in the state machine

#### Scenario: Failed workflow marks tenant as failed
- **WHEN** a workflow execution fails with a non-retryable error
- **THEN** the reconciler MUST update the tenant status to "failed" and record the error

#### Scenario: Retryable workflow errors requeue tenant
- **WHEN** a workflow execution fails with a retryable error
- **THEN** the reconciler MUST NOT update tenant status and SHALL requeue for retry

#### Scenario: Workflow execution ID is stored
- **WHEN** a workflow is triggered successfully
- **THEN** the reconciler MUST store the execution ID in the tenant record

### Requirement: Validate workflow provider availability
The reconciler SHALL verify the workflow manager is ready before attempting to trigger workflows.

#### Scenario: Workflow manager readiness is checked
- **WHEN** the reconciler attempts to trigger a workflow
- **THEN** it MUST verify the workflow manager is initialized and not nil

#### Scenario: Missing workflow provider is handled gracefully
- **WHEN** no workflow provider is configured
- **THEN** the reconciler MUST log an error and requeue the tenant for retry

#### Scenario: Workflow provider errors are retried
- **WHEN** the workflow manager returns a transient error (connection timeout, rate limit)
- **THEN** the reconciler MUST requeue the tenant with exponential backoff

### Requirement: Support workflow execution timeouts
The reconciler SHALL enforce timeouts when triggering workflows to prevent indefinite blocking.

#### Scenario: Workflow trigger has default timeout
- **WHEN** a workflow is triggered without explicit timeout
- **THEN** the operation MUST complete or fail within 30 seconds

#### Scenario: Workflow trigger timeout is configurable
- **WHEN** controller.workflow_trigger_timeout is configured
- **THEN** the timeout MUST match the configured value

#### Scenario: Timeout is treated as retryable error
- **WHEN** a workflow trigger times out
- **THEN** the reconciler MUST treat it as a retryable error and requeue the tenant

### Requirement: Log workflow interactions
The reconciler SHALL emit structured logs for all workflow trigger attempts.

#### Scenario: Workflow trigger attempt is logged
- **WHEN** the reconciler triggers a workflow
- **THEN** a log entry with level INFO MUST include tenant_id, workflow_action, workflow_provider

#### Scenario: Workflow success is logged
- **WHEN** a workflow is triggered successfully
- **THEN** a log entry with level INFO MUST include tenant_id, execution_id, duration

#### Scenario: Workflow error is logged
- **WHEN** a workflow trigger fails
- **THEN** a log entry with level ERROR MUST include tenant_id, error_message, is_retryable
