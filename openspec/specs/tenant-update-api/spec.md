## ADDED Requirements

### Requirement: Update tenant properties
The system SHALL provide a PUT or PATCH endpoint at `/v1/tenants/{id}` that updates tenant properties identified by UUID or name.

#### Scenario: Valid update request
- **WHEN** a client sends a PUT request to `/v1/tenants/{id}` with valid update data
- **THEN** the system processes the update

#### Scenario: Tenant exists
- **WHEN** the update request specifies an existing tenant identifier
- **THEN** the system updates that tenant's properties

### Requirement: Support partial updates
The system SHALL allow updating individual tenant fields without requiring all fields.

#### Scenario: Update tenant name only
- **WHEN** an update request includes only the `name` field
- **THEN** the system updates the name and leaves other fields unchanged

#### Scenario: Update multiple fields
- **WHEN** an update request includes multiple fields
- **THEN** the system updates all specified fields atomically

### Requirement: Validate update data
The system SHALL validate update request data before applying changes.

#### Scenario: Invalid name in update
- **WHEN** an update request provides an empty or overly long name
- **THEN** the system returns HTTP 400 with validation error

#### Scenario: Valid update data
- **WHEN** all fields in the update request pass validation
- **THEN** the system applies the updates

### Requirement: Handle non-existent tenant
The system SHALL return error when attempting to update non-existent tenant.

#### Scenario: Update non-existent tenant
- **WHEN** an update request specifies a tenant identifier that does not exist
- **THEN** the system returns HTTP 404 Not Found

### Requirement: Return updated tenant
The system SHALL return the complete updated tenant resource upon successful update.

#### Scenario: Successful update
- **WHEN** a tenant is updated successfully
- **THEN** the system returns HTTP 200 with the updated tenant including all current fields

#### Scenario: Updated timestamp included
- **WHEN** a tenant is updated
- **THEN** the response includes an `updated_at` timestamp reflecting the update time

### Requirement: Prevent ID modification
The system SHALL not allow modifying the tenant ID through update operations.

#### Scenario: Attempt to change tenant ID
- **WHEN** an update request includes an `id` field
- **THEN** the system ignores it or returns HTTP 400

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
### Requirement: Support compute configuration updates
The system SHALL allow updating tenant compute configuration through update operations.

#### Scenario: Update compute configuration
- **WHEN** an update request includes a `compute_config` field
- **THEN** the system validates it against the active provider schema
- **AND** it updates Tenant.DesiredConfig only if validation succeeds

#### Scenario: Validate updated compute configuration
- **WHEN** compute configuration is provided in an update request
- **THEN** the active provider validates the new configuration before applying

#### Scenario: Invalid compute configuration in update
- **WHEN** an update request includes invalid compute_config
- **THEN** the system returns HTTP 400 with detailed validation errors from the provider

#### Scenario: Partial compute configuration update
- **WHEN** an update request provides partial compute_config (e.g., only env vars)
- **THEN** the system either replaces the entire compute_config or merges based on API semantics
