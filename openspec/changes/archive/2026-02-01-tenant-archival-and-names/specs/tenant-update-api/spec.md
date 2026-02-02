## MODIFIED Requirements

### Requirement: Update tenant properties
The system SHALL provide a PUT or PATCH endpoint at `/api/tenants/{id}` that updates tenant properties identified by UUID or name.

#### Scenario: Valid update request
- **WHEN** a client sends a PUT request to `/api/tenants/{id}` with valid update data
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
