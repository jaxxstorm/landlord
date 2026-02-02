## MODIFIED Requirements

### Requirement: Update tenant properties
The system SHALL provide a PUT or PATCH endpoint at `/v1/tenants/{id}` that updates tenant properties identified by UUID or name.

#### Scenario: Valid update request
- **WHEN** a client sends a PUT request to `/v1/tenants/{id}` with valid update data
- **THEN** the system processes the update

#### Scenario: Tenant exists
- **WHEN** the update request specifies an existing tenant identifier
- **THEN** the system updates that tenant's properties

### Requirement: Document endpoint with Swagger
The system SHALL include Swagger/OpenAPI annotations for the update tenant endpoint.

#### Scenario: Swagger documentation present
- **WHEN** the OpenAPI spec is generated
- **THEN** it includes the PUT `/v1/tenants/{id}` endpoint

#### Scenario: Request schema documented
- **WHEN** viewing API documentation
- **THEN** the update request body schema shows updatable fields

#### Scenario: Response schemas documented
- **WHEN** viewing API documentation
- **THEN** 200 success, 404 not found, and 400 validation error responses are documented
