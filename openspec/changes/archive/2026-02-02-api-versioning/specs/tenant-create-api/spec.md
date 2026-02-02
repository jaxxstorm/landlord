## MODIFIED Requirements

### Requirement: Accept tenant creation requests
The system SHALL provide a POST endpoint at `/v1/tenants` that accepts JSON payloads to create new tenants.

#### Scenario: Valid tenant creation request
- **WHEN** a client sends a POST request to `/v1/tenants` with valid tenant data
- **THEN** the system accepts the request and processes it

#### Scenario: Request includes required fields
- **WHEN** the request body contains `name` field
- **THEN** the system validates and accepts the request

### Requirement: Document endpoint with Swagger
The system SHALL include Swagger/OpenAPI annotations for the create tenant endpoint.

#### Scenario: Swagger documentation present
- **WHEN** the OpenAPI spec is generated
- **THEN** it includes the POST `/v1/tenants` endpoint with request/response schemas

#### Scenario: Request schema documented
- **WHEN** viewing API documentation
- **THEN** the create tenant request body schema shows required and optional fields

#### Scenario: Response schemas documented
- **WHEN** viewing API documentation
- **THEN** both 201 success and error responses are documented with schemas

### Requirement: Store compute configuration with tenant
The system SHALL persist compute configuration as part of tenant desired state.

#### Scenario: Compute configuration stored
- **WHEN** a tenant is created with compute_config
- **THEN** the configuration is stored in Tenant.DesiredConfig for later provisioning

#### Scenario: Configuration retrievable
- **WHEN** a tenant is retrieved via GET /v1/tenants/{id}
- **THEN** the response includes the compute_config that was provided at creation
