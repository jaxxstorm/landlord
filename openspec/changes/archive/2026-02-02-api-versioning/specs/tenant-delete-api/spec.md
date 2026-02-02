## MODIFIED Requirements

### Requirement: Delete tenant by ID
The system SHALL provide a DELETE endpoint at `/v1/tenants/{id}` that archives a tenant by UUID or name and removes it from the database when deletion is finalized.

#### Scenario: Valid delete request
- **WHEN** a client sends a DELETE request to `/v1/tenants/{id}` with a valid UUID or name
- **THEN** the system attempts to delete that tenant

#### Scenario: Tenant exists
- **WHEN** the delete request specifies an existing tenant identifier
- **THEN** the system archives the tenant and triggers compute deletion

### Requirement: Prevent accidental deletion
The system SHALL require explicit tenant ID in the path to prevent bulk deletions.

#### Scenario: No bulk delete endpoint
- **WHEN** a DELETE request is sent to `/v1/tenants` without an ID
- **THEN** the system returns HTTP 405 Method Not Allowed or HTTP 400

### Requirement: Document endpoint with Swagger
The system SHALL include Swagger/OpenAPI annotations for the delete tenant endpoint.

#### Scenario: Swagger documentation present
- **WHEN** the OpenAPI spec is generated
- **THEN** it includes the DELETE `/v1/tenants/{id}` endpoint

#### Scenario: Path parameter documented
- **WHEN** viewing API documentation
- **THEN** the `id` path parameter is documented as required UUID

#### Scenario: Response codes documented
- **WHEN** viewing API documentation
- **THEN** 204/200 success, 404 not found, and 400 validation error responses are documented
