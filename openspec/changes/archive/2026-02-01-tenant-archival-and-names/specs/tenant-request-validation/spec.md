## MODIFIED Requirements

### Requirement: Validate tenant name format
The system SHALL validate that tenant names meet format requirements and are unique.

#### Scenario: Name length validation
- **WHEN** a tenant name is provided in any request
- **THEN** the system validates it is between 1 and 255 characters

#### Scenario: Name cannot be empty
- **WHEN** a tenant name consists only of whitespace
- **THEN** the system rejects it with HTTP 400

#### Scenario: Name uniqueness
- **WHEN** a tenant name is provided on create or rename
- **THEN** the system checks for existing tenants with that name
- **AND** returns HTTP 409 if the name is already in use

### Requirement: Validate UUID format
The system SHALL validate UUID parameters only when identifiers parse as UUIDs.

#### Scenario: Valid UUID v4 format
- **WHEN** a tenant ID is provided as a path parameter and parses as UUID
- **THEN** the system validates it matches UUID format

#### Scenario: Invalid UUID format
- **WHEN** an ID parameter parses as UUID but is malformed
- **THEN** the system returns HTTP 400 with clear validation error message
