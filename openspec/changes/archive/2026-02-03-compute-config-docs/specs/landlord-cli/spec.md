## MODIFIED Requirements

### Requirement: Verb-based command structure
The system SHALL expose verb-based commands for tenant operations, including create, list, get, set, and delete, accepting name or UUID identifiers.

#### Scenario: Create command
- **WHEN** a user runs `landlord-cli create --tenant-name <name> --config <json|yaml|file://path>`
- **THEN** the CLI SHALL issue a tenant create request to the Landlord API
- **AND** it SHALL include the compute_config payload parsed from JSON or YAML input

#### Scenario: List command
- **WHEN** a user runs `landlord-cli list`
- **THEN** the CLI SHALL fetch the list of tenants from the Landlord API
- **AND** it SHALL render a formatted list of tenant identifiers and statuses

#### Scenario: Get command
- **WHEN** a user runs `landlord-cli get --tenant-id <id-or-name>`
- **THEN** the CLI SHALL fetch the tenant by UUID or name
- **AND** it SHALL render the tenant details in a readable, styled format

#### Scenario: Set command
- **WHEN** a user runs `landlord-cli set --tenant-id <id-or-name> --config <json|yaml|file://path>`
- **THEN** the CLI SHALL issue a tenant update request with the compute_config payload
- **AND** it SHALL include the compute_config payload parsed from JSON or YAML input

#### Scenario: Delete command
- **WHEN** a user runs `landlord-cli delete --tenant-id <id-or-name>`
- **THEN** the CLI SHALL issue a tenant delete request to the Landlord API
- **AND** it SHALL render a confirmation message
