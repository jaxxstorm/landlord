## ADDED Requirements

### Requirement: Tenant records store workflow execution ID
The tenant database record SHALL store the workflow execution ID in the workflow_execution_id column to track which workflow is processing the tenant.

#### Scenario: Workflow execution ID set on API trigger
- **WHEN** an API handler triggers a workflow and receives an execution ID
- **THEN** the tenant record MUST be updated with the workflow_execution_id value in the same transaction as the status update

#### Scenario: Workflow execution ID preserved across status transitions
- **WHEN** a tenant transitions through multiple statuses (planning → provisioning → ready)
- **THEN** the workflow_execution_id MUST remain set until the workflow completes or fails

#### Scenario: Workflow execution ID cleared on completion
- **WHEN** a workflow completes successfully and tenant reaches "ready" status
- **THEN** the workflow_execution_id MAY be cleared or retained for audit purposes

### Requirement: GET requests return workflow execution status
The API SHALL return the workflow_execution_id in the response body for GET /api/tenants and GET /api/tenants/{id} requests to enable client-side status tracking.

#### Scenario: Tenant with active workflow returns execution ID
- **WHEN** a client retrieves a tenant that has an active workflow
- **THEN** the response SHALL include the workflow_execution_id field with the current execution identifier

#### Scenario: Tenant without workflow returns null execution ID
- **WHEN** a client retrieves a tenant in "ready" status with no active workflow
- **THEN** the response SHALL include workflow_execution_id field with null value

### Requirement: Duplicate workflow triggers are prevented via execution ID check
The system SHALL check the workflow_execution_id field to prevent duplicate workflow triggers when both API and controller attempt to process the same tenant.

#### Scenario: Controller skips tenant with existing execution ID
- **WHEN** the reconciliation controller polls and finds a tenant with a non-null workflow_execution_id
- **THEN** the controller SHALL verify the workflow is still active before skipping the trigger

#### Scenario: Controller re-triggers after workflow failure
- **WHEN** the reconciliation controller polls and finds a tenant with workflow_execution_id pointing to a failed workflow
- **THEN** the controller SHALL trigger a new workflow with a new execution ID

#### Scenario: API overwrites execution ID on retry
- **WHEN** a client retries a failed operation (POST/PUT/DELETE)
- **THEN** the API SHALL overwrite the previous workflow_execution_id with the new execution ID

### Requirement: Workflow execution ID format is deterministic
The workflow execution ID SHALL follow a deterministic format to enable idempotency and debugging: "tenant-{tenant_id}-{action}".

#### Scenario: Plan workflow execution ID format
- **WHEN** the API triggers a "plan" workflow for tenant "my-app"
- **THEN** the execution ID MUST be "tenant-my-app-plan"

#### Scenario: Provision workflow execution ID format
- **WHEN** the controller triggers a "provision" workflow for tenant "my-app"
- **THEN** the execution ID MUST be "tenant-my-app-provision"

#### Scenario: Update workflow execution ID format
- **WHEN** the API triggers an "update" workflow for tenant "my-app"
- **THEN** the execution ID MUST be "tenant-my-app-update"

### Requirement: Database transaction ensures atomicity of status and execution ID updates
The system SHALL update the tenant status and workflow_execution_id in a single database transaction to prevent inconsistent state.

#### Scenario: Transaction commits both fields
- **WHEN** an API handler updates status to "planning" and sets workflow_execution_id
- **THEN** both fields MUST be committed in a single database transaction

#### Scenario: Transaction rollback on database error
- **WHEN** the database update fails after status change but before execution ID is set
- **THEN** the entire transaction MUST be rolled back, leaving tenant in previous state

#### Scenario: Workflow trigger outside transaction
- **WHEN** the database transaction commits successfully
- **THEN** the workflow trigger SHALL occur after the commit, outside the transaction boundary
