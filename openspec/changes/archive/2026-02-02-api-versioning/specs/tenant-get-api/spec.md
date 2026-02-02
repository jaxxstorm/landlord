## MODIFIED Requirements

### Requirement: Retrieve tenant by ID
The system SHALL provide a GET endpoint at `/v1/tenants/{id}` that retrieves a specific tenant by its UUID or name.

#### Scenario: Valid tenant ID provided
- **WHEN** a client sends a GET request to `/v1/tenants/{id}` with a valid UUID
- **THEN** the system looks up the tenant by that ID

#### Scenario: Valid tenant name provided
- **WHEN** a client sends a GET request to `/v1/tenants/{id}` with a non-UUID value
- **THEN** the system looks up the tenant by name

#### Scenario: Tenant exists
- **WHEN** the requested tenant identifier exists in the system
- **THEN** the system returns HTTP 200 with the tenant resource

### Requirement: Document endpoint with Swagger
The system SHALL include Swagger/OpenAPI annotations for the get tenant endpoint.

#### Scenario: Swagger documentation present
- **WHEN** the OpenAPI spec is generated
- **THEN** it includes the GET `/v1/tenants/{id}` endpoint with path parameter

#### Scenario: Path parameter documented
- **WHEN** viewing API documentation
- **THEN** the `id` path parameter is documented as required UUID

#### Scenario: Response schemas documented
- **WHEN** viewing API documentation
- **THEN** both 200 success and 404 error responses are documented with schemas
