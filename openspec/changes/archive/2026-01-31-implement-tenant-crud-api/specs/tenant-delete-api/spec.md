## ADDED Requirements

### Requirement: Delete tenant by ID
The system SHALL provide a DELETE endpoint at `/api/tenants/{id}` that removes a tenant from the system.

#### Scenario: Valid delete request
- **WHEN** a client sends a DELETE request to `/api/tenants/{id}` with a valid UUID
- **THEN** the system attempts to delete that tenant

#### Scenario: Tenant exists
- **WHEN** the delete request specifies an existing tenant ID
- **THEN** the system removes the tenant from the database

### Requirement: Confirm successful deletion
The system SHALL return appropriate response upon successful tenant deletion.

#### Scenario: Successful deletion
- **WHEN** a tenant is deleted successfully
- **THEN** the system returns HTTP 204 No Content or HTTP 200 with confirmation

### Requirement: Handle non-existent tenant
The system SHALL return appropriate error when attempting to delete non-existent tenant.

#### Scenario: Delete non-existent tenant
- **WHEN** a delete request specifies a tenant ID that does not exist
- **THEN** the system returns HTTP 404 Not Found

### Requirement: Validate UUID format
The system SHALL validate that the tenant ID parameter is a valid UUID format.

#### Scenario: Invalid UUID format
- **WHEN** a DELETE request provides a malformed ID
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
