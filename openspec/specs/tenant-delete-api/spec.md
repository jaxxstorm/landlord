## ADDED Requirements

### Requirement: Delete tenant by ID
The system SHALL provide a DELETE endpoint at `/api/tenants/{id}` that archives a tenant by UUID or name and removes it from the database when deletion is finalized.

#### Scenario: Valid delete request
- **WHEN** a client sends a DELETE request to `/api/tenants/{id}` with a valid UUID or name
- **THEN** the system attempts to delete that tenant

#### Scenario: Tenant exists
- **WHEN** the delete request specifies an existing tenant identifier
- **THEN** the system archives the tenant and triggers compute deletion

### Requirement: Confirm successful deletion
The system SHALL return appropriate response upon successful tenant deletion.

#### Scenario: Successful delete request
- **WHEN** a tenant delete request is accepted
- **THEN** the system returns HTTP 200 with the archived tenant or HTTP 202 Accepted

### Requirement: Handle non-existent tenant
The system SHALL return appropriate error when attempting to delete non-existent tenant.

#### Scenario: Delete non-existent tenant
- **WHEN** a delete request specifies a tenant identifier that does not exist
- **THEN** the system returns HTTP 404 Not Found

### Requirement: Validate UUID format
The system SHALL validate UUID format only when the identifier parses as a UUID.

#### Scenario: Invalid UUID format
- **WHEN** a DELETE request provides a malformed UUID
- **THEN** the system returns HTTP 400 Bad Request

### Requirement: Cleanup tenant resources
The system SHALL remove all tenant-related data and resources during deletion.

#### Scenario: Cascade delete tenant data
- **WHEN** a tenant is deleted
- **THEN** the system removes associated tenant schemas, configurations, and metadata

### Requirement: Prevent accidental deletion
The system SHALL require explicit tenant ID in the path to prevent bulk deletions.

#### Scenario: No bulk delete endpoint
- **WHEN** a DELETE request is sent to `/api/tenants` without an ID
- **THEN** the system returns HTTP 405 Method Not Allowed or HTTP 400

### Requirement: Support idempotent deletion
The system SHALL handle repeated deletion requests gracefully.

#### Scenario: Delete already deleted tenant
- **WHEN** a delete request targets a tenant that was previously deleted
- **THEN** the system returns HTTP 404 Not Found consistently

### Requirement: Document endpoint with Swagger
The system SHALL include Swagger/OpenAPI annotations for the delete tenant endpoint.

#### Scenario: Swagger documentation present
- **WHEN** the OpenAPI spec is generated
- **THEN** it includes the DELETE `/api/tenants/{id}` endpoint

#### Scenario: Path parameter documented
- **WHEN** viewing API documentation
- **THEN** the `id` path parameter is documented as required UUID

#### Scenario: Response codes documented
- **WHEN** viewing API documentation
- **THEN** 204/200 success, 404 not found, and 400 validation error responses are documented
