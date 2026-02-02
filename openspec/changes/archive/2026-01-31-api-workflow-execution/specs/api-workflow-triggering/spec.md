## ADDED Requirements

### Requirement: API handlers trigger workflows immediately after database mutation
The API handlers SHALL trigger workflow execution immediately after successfully updating the database, without waiting for the reconciliation controller polling interval.

#### Scenario: POST /tenants triggers provisioning workflow
- **WHEN** a client creates a new tenant via POST /api/tenants
- **THEN** the API handler SHALL update the tenant status to "planning" and trigger the "plan" workflow action immediately after database commit

#### Scenario: PUT /tenants triggers update workflow
- **WHEN** a client updates an existing tenant via PUT /api/tenants/{id}
- **THEN** the API handler SHALL update the tenant status to "updating" and trigger the "update" workflow action immediately after database commit

#### Scenario: DELETE /tenants triggers deletion workflow
- **WHEN** a client deletes a tenant via DELETE /api/tenants/{id}
- **THEN** the API handler SHALL update the tenant status to "deleting" and trigger the "delete" workflow action immediately after database commit

### Requirement: API responses include workflow execution ID
The API response for tenant operations SHALL include the workflow execution ID to enable client-side tracking of provisioning progress.

#### Scenario: Successful workflow trigger returns execution ID
- **WHEN** the API successfully triggers a workflow
- **THEN** the response body MUST include the "workflow_execution_id" field with the unique execution identifier

#### Scenario: Workflow trigger failure omits execution ID
- **WHEN** the API fails to trigger a workflow
- **THEN** the response body MUST NOT include a "workflow_execution_id" field and SHALL return HTTP 500 status

### Requirement: API uses HTTP 202 Accepted for asynchronous operations
The API SHALL return HTTP 202 Accepted status code for tenant operations that trigger asynchronous workflows, indicating the request has been accepted but processing is not complete.

#### Scenario: Successful create returns 202
- **WHEN** POST /api/tenants successfully creates a tenant and triggers a workflow
- **THEN** the API SHALL return HTTP 202 Accepted status code

#### Scenario: Successful update returns 202
- **WHEN** PUT /api/tenants/{id} successfully updates a tenant and triggers a workflow
- **THEN** the API SHALL return HTTP 202 Accepted status code

#### Scenario: Successful delete returns 202
- **WHEN** DELETE /api/tenants/{id} successfully deletes a tenant and triggers a workflow
- **THEN** the API SHALL return HTTP 202 Accepted status code

### Requirement: Workflow client injection maintains pluggable provider abstraction
The API server SHALL use the existing WorkflowClient wrapper to trigger workflows, maintaining the pluggable provider abstraction without hardcoding specific workflow providers.

#### Scenario: API server receives workflow client at initialization
- **WHEN** the API server is initialized in cmd/landlord/main.go
- **THEN** the constructor MUST accept a WorkflowClient instance as a dependency

#### Scenario: API handlers trigger workflows via WorkflowClient
- **WHEN** an API handler needs to trigger a workflow
- **THEN** it SHALL call workflowClient.TriggerWorkflow() without direct dependency on workflow.Manager or specific providers

### Requirement: Workflow trigger failures return HTTP 500 with error details
The API SHALL return HTTP 500 Internal Server Error when workflow triggering fails, with error details in the response body for debugging.

#### Scenario: Workflow provider unavailable
- **WHEN** the workflow provider (Restate, Step Functions) is unavailable during workflow trigger
- **THEN** the API SHALL return HTTP 500 with error message "Failed to trigger workflow"

#### Scenario: Workflow trigger timeout
- **WHEN** the workflow trigger operation exceeds the timeout (default 30s)
- **THEN** the API SHALL return HTTP 500 with timeout error details

#### Scenario: Invalid workflow specification
- **WHEN** the workflow specification contains invalid parameters
- **THEN** the API SHALL return HTTP 400 Bad Request with validation error details

### Requirement: State machine validation prevents invalid transitions
The API handlers SHALL use the shared state machine logic to validate state transitions before triggering workflows, preventing invalid status changes.

#### Scenario: Valid transition is allowed
- **WHEN** a tenant in "ready" status receives a PUT request to update configuration
- **THEN** the API SHALL transition status to "updating" and trigger the update workflow

#### Scenario: Invalid transition is rejected
- **WHEN** a tenant in "failed" status receives a PUT request to update configuration
- **THEN** the API SHALL return HTTP 409 Conflict without triggering a workflow

#### Scenario: Terminal status prevents workflow triggering
- **WHEN** a tenant in "deleted" status receives any mutation request
- **THEN** the API SHALL return HTTP 410 Gone without triggering a workflow
