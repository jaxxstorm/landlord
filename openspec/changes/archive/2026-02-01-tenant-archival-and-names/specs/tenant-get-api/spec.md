## MODIFIED Requirements

### Requirement: Retrieve tenant by ID
The system SHALL provide a GET endpoint at `/api/tenants/{id}` that retrieves a specific tenant by its UUID or name.

#### Scenario: Valid tenant ID provided
- **WHEN** a client sends a GET request to `/api/tenants/{id}` with a valid UUID
- **THEN** the system looks up the tenant by that ID

#### Scenario: Valid tenant name provided
- **WHEN** a client sends a GET request to `/api/tenants/{id}` with a non-UUID value
- **THEN** the system looks up the tenant by name

#### Scenario: Tenant exists
- **WHEN** the requested tenant identifier exists in the system
- **THEN** the system returns HTTP 200 with the tenant resource

### Requirement: Return complete tenant data
The system SHALL return all tenant fields in the response.

#### Scenario: Response includes tenant details
- **WHEN** a tenant is retrieved successfully
- **THEN** the response includes `id`, `name`, `created_at`, `updated_at`, and all other tenant properties

### Requirement: Handle non-existent tenant
The system SHALL return appropriate error when requested tenant does not exist.

#### Scenario: Tenant identifier not found
- **WHEN** a GET request specifies an identifier that does not exist
- **THEN** the system returns HTTP 404 Not Found with descriptive error

### Requirement: Validate UUID format
The system SHALL validate UUID format only when the identifier parses as a UUID.

#### Scenario: Invalid UUID format
- **WHEN** a GET request provides a malformed UUID
- **THEN** the system returns HTTP 400 Bad Request with validation error

#### Scenario: Valid UUID format
- **WHEN** a GET request provides a properly formatted UUID
- **THEN** the system processes the request
