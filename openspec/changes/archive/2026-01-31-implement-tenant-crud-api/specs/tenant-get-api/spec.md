## ADDED Requirements

### Requirement: Retrieve tenant by ID
The system SHALL provide a GET endpoint at `/api/tenants/{id}` that retrieves a specific tenant by its UUID.

#### Scenario: Valid tenant ID provided
- **WHEN** a client sends a GET request to `/api/tenants/{id}` with a valid UUID
- **THEN** the system looks up the tenant by that ID

#### Scenario: Tenant exists
- **WHEN** the requested tenant ID exists in the system
- **THEN** the system returns HTTP 200 with the tenant resource

### Requirement: Return complete tenant data
The system SHALL return all tenant fields in the response.

#### Scenario: Response includes tenant details
- **WHEN** a tenant is retrieved successfully
- **THEN** the response includes `id`, `name`, `created_at`, `updated_at`, and all other tenant properties

### Requirement: Handle non-existent tenant
The system SHALL return appropriate error when requested tenant does not exist.

#### Scenario: Tenant ID not found
- **WHEN** a GET request specifies a UUID that does not exist
- **THEN** the system returns HTTP 404 Not Found with descriptive error

### Requirement: Validate UUID format
The system SHALL validate that the tenant ID parameter is a valid UUID format.

#### Scenario: Invalid UUID format
- **WHEN** a GET request provides a malformed ID (not a valid UUID)
- **THEN** the system returns HTTP 400 Bad Request with validation error

#### Scenario: Valid UUID format
- **WHEN** a GET request provides a properly formatted UUID
- **THEN** the system processes the request

### Requirement: Document endpoint with Swagger
The system SHALL include Swagger/OpenAPI annotations for the get tenant endpoint.

#### Scenario: Swagger documentation present
- **WHEN** the OpenAPI spec is generated
- **THEN** it includes the GET `/api/tenants/{id}` endpoint with path parameter

#### Scenario: Path parameter documented
- **WHEN** viewing API documentation
- **THEN** the `id` path parameter is documented as required UUID

#### Scenario: Response schemas documented
- **WHEN** viewing API documentation
- **THEN** both 200 success and 404 error responses are documented with schemas
