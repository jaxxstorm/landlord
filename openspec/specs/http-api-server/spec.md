## ADDED Requirements

### Requirement: POST /api/tenants creates tenant and triggers provisioning workflow
The API SHALL create a new tenant record and immediately trigger a provisioning workflow, returning HTTP 202 Accepted with the workflow execution ID.

#### Scenario: Successful tenant creation with workflow trigger
- **WHEN** a client sends POST /api/tenants with valid tenant specification
- **THEN** the API SHALL create the tenant record with status "planning"
- **AND** SHALL trigger the "plan" workflow action
- **AND** SHALL return HTTP 202 Accepted with the tenant record including workflow_execution_id

#### Scenario: Workflow trigger fails after tenant creation
- **WHEN** the tenant is created successfully but workflow trigger fails
- **THEN** the API SHALL return HTTP 500 Internal Server Error
- **AND** the tenant record SHALL remain in the database with status "planning" and null workflow_execution_id for controller retry

#### Scenario: Invalid tenant specification
- **WHEN** a client sends POST /api/tenants with invalid or missing required fields
- **THEN** the API SHALL return HTTP 400 Bad Request without creating a tenant or triggering a workflow

### Requirement: PUT /api/tenants/{id} updates tenant and triggers update workflow
The API SHALL update an existing tenant record and immediately trigger an update workflow, returning HTTP 202 Accepted with the workflow execution ID.

#### Scenario: Successful tenant update with workflow trigger
- **WHEN** a client sends PUT /api/tenants/{id} with updated tenant specification
- **THEN** the API SHALL update the tenant record with status "updating"
- **AND** SHALL trigger the "update" workflow action
- **AND** SHALL return HTTP 202 Accepted with the updated tenant record including workflow_execution_id

#### Scenario: Update on non-existent tenant
- **WHEN** a client sends PUT /api/tenants/{id} for a non-existent tenant
- **THEN** the API SHALL return HTTP 404 Not Found without triggering a workflow

#### Scenario: Update on terminal status tenant
- **WHEN** a client attempts to update a tenant in "deleted" or "failed" status
- **THEN** the API SHALL return HTTP 409 Conflict without triggering a workflow

### Requirement: DELETE /api/tenants/{id} marks tenant for deletion and triggers deletion workflow
The API SHALL update the tenant status to "deleting" and immediately trigger a deletion workflow, returning HTTP 202 Accepted with the workflow execution ID.

#### Scenario: Successful tenant deletion with workflow trigger
- **WHEN** a client sends DELETE /api/tenants/{id}
- **THEN** the API SHALL update the tenant status to "deleting"
- **AND** SHALL trigger the "delete" workflow action
- **AND** SHALL return HTTP 202 Accepted with the tenant record including workflow_execution_id

#### Scenario: Delete on non-existent tenant
- **WHEN** a client sends DELETE /api/tenants/{id} for a non-existent tenant
- **THEN** the API SHALL return HTTP 404 Not Found without triggering a workflow

#### Scenario: Delete on already deleted tenant
- **WHEN** a client sends DELETE /api/tenants/{id} for a tenant already in "deleted" status
- **THEN** the API SHALL return HTTP 410 Gone without triggering a workflow

### Requirement: GET /api/tenants returns list of tenants with workflow status
The API SHALL return all tenants including their workflow_execution_id field for status tracking.

#### Scenario: List includes workflow execution IDs
- **WHEN** a client sends GET /api/tenants
- **THEN** the response SHALL include an array of tenant records
- **AND** each tenant record MUST include the workflow_execution_id field (null if no active workflow)

### Requirement: GET /api/tenants/{id} returns single tenant with workflow status
The API SHALL return a single tenant including its workflow_execution_id field for status tracking.

#### Scenario: Get tenant with active workflow
- **WHEN** a client sends GET /api/tenants/{id} for a tenant with an active workflow
- **THEN** the response SHALL include the tenant record with non-null workflow_execution_id

#### Scenario: Get tenant without active workflow
- **WHEN** a client sends GET /api/tenants/{id} for a tenant in "ready" status
- **THEN** the response SHALL include the tenant record with null workflow_execution_id

#### Scenario: Get non-existent tenant
- **WHEN** a client sends GET /api/tenants/{id} for a non-existent tenant
- **THEN** the API SHALL return HTTP 404 Not Found

### Requirement: Document endpoint with Swagger
The system SHALL include Swagger/OpenAPI annotations for all HTTP endpoints.

#### Scenario: Swagger documentation present
- **WHEN** the OpenAPI spec is generated
- **THEN** it includes the compute config discovery endpoint with request/response schemas

#### Scenario: Response schema documented
- **WHEN** viewing API documentation
- **THEN** the compute config discovery response schema shows provider identifier, schema, and defaults
