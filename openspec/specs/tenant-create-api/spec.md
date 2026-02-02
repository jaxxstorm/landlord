## ADDED Requirements

### Requirement: Accept tenant creation requests
The system SHALL provide a POST endpoint at `/api/tenants` that accepts JSON payloads to create new tenants.

#### Scenario: Valid tenant creation request
- **WHEN** a client sends a POST request to `/api/tenants` with valid tenant data
- **THEN** the system accepts the request and processes it

#### Scenario: Request includes required fields
- **WHEN** the request body contains `name` field
- **THEN** the system validates and accepts the request

### Requirement: Validate tenant name
The system SHALL validate that tenant names are non-empty strings with reasonable length constraints.

#### Scenario: Tenant name is provided
- **WHEN** a creation request includes a `name` field with 1-255 characters
- **THEN** the system accepts the name as valid

#### Scenario: Tenant name is empty
- **WHEN** a creation request has an empty or whitespace-only `name`
- **THEN** the system returns HTTP 400 with validation error

#### Scenario: Tenant name is too long
- **WHEN** a creation request has a `name` exceeding 255 characters
- **THEN** the system returns HTTP 400 with validation error

### Requirement: Generate unique tenant ID
The system SHALL generate a unique UUID for each newly created tenant.

#### Scenario: New tenant gets unique ID
- **WHEN** a tenant is successfully created
- **THEN** the system assigns a UUID that does not conflict with existing tenants

### Requirement: Return created tenant
The system SHALL return the complete tenant resource upon successful creation.

#### Scenario: Successful tenant creation
- **WHEN** a valid tenant creation request is processed
- **THEN** the system returns HTTP 201 with the created tenant including ID, name, and timestamps

#### Scenario: Response includes all tenant fields
- **WHEN** a tenant is created
- **THEN** the response includes `id`, `name`, `created_at`, and any other tenant properties

### Requirement: Handle duplicate names
The system SHALL enforce tenant name uniqueness and return appropriate errors for duplicates.

#### Scenario: Tenant name already exists
- **WHEN** a creation request uses a name that already exists
- **THEN** the system returns HTTP 409 Conflict with descriptive error

### Requirement: Document endpoint with Swagger
The system SHALL include Swagger/OpenAPI annotations for the create tenant endpoint.

#### Scenario: Swagger documentation present
- **WHEN** the OpenAPI spec is generated
- **THEN** it includes the POST `/api/tenants` endpoint with request/response schemas

#### Scenario: Request schema documented
- **WHEN** viewing API documentation
- **THEN** the create tenant request body schema shows required and optional fields

#### Scenario: Response schemas documented
- **WHEN** viewing API documentation
- **THEN** both 201 success and error responses are documented with schemas
### Requirement: Accept provider-specific compute configuration
The system SHALL accept compute configuration specific to the active compute provider (Docker, ECS, Kubernetes).

#### Scenario: Compute configuration provided
- **WHEN** a creation request includes a `compute_config` field
- **THEN** the system validates it against the active provider's schema
- **AND** the request is rejected with HTTP 400 if the schema validation fails

#### Scenario: Valid Docker compute configuration
- **WHEN** a creation request includes Docker-specific configuration (e.g., env vars, volumes, network mode)
- **THEN** the Docker provider validates and accepts it

#### Scenario: Invalid compute configuration structure
- **WHEN** a creation request includes `compute_config` that doesn't match provider schema
- **THEN** the system returns HTTP 400 with detailed validation errors

#### Scenario: Missing required compute configuration fields
- **WHEN** a creation request omits required provider configuration fields
- **THEN** the system returns HTTP 400 listing missing fields

### Requirement: Validate compute configuration at API ingress
The system SHALL validate compute configuration immediately upon request receipt, before database operations.

#### Scenario: Early validation prevents wasted resources
- **WHEN** an invalid compute configuration is submitted
- **THEN** the system rejects it with HTTP 400 before creating database records or provisioning resources

#### Scenario: Provider performs validation
- **WHEN** compute configuration is provided
- **THEN** the active compute provider validates the configuration structure, types, and constraints

### Requirement: Store compute configuration with tenant
The system SHALL persist compute configuration as part of tenant desired state.

#### Scenario: Compute configuration stored
- **WHEN** a tenant is created with compute_config
- **THEN** the configuration is stored in Tenant.DesiredConfig for later provisioning

#### Scenario: Configuration retrievable
- **WHEN** a tenant is retrieved via GET /api/tenants/{id}
- **THEN** the response includes the compute_config that was provided at creation
