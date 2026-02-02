## MODIFIED Requirements

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
