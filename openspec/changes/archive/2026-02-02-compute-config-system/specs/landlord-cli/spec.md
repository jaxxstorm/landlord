## MODIFIED Requirements

### Requirement: Verb-based command structure
The system SHALL expose verb-based commands for tenant operations, including create, list, get, set, and delete, accepting name or UUID identifiers.

#### Scenario: Create command
- **WHEN** a user runs `landlord-cli create --tenant-name <name> [--config <json>]`
- **THEN** the CLI SHALL issue a tenant create request to the Landlord API
- **AND** it SHALL include the compute_config payload when provided

#### Scenario: Set command
- **WHEN** a user runs `landlord-cli set --tenant-id <id-or-name> [--config <json>]`
- **THEN** the CLI SHALL issue a tenant update request with the compute_config payload
- **AND** it SHALL render a confirmation with the updated tenant details

### Requirement: CLI configuration
The system SHALL provide a CLI that loads configuration via Viper from flags, environment variables, and config files.

#### Scenario: API base URL configuration
- **WHEN** the CLI is started without an explicit API base URL flag
- **THEN** it SHALL use the configured default value from Viper
- **AND** it SHALL include the base URL in request targets

### Requirement: CLI compute config discovery
The system SHALL provide a CLI command to retrieve compute configuration schema from the API.

#### Scenario: Compute command
- **WHEN** a user runs `landlord-cli compute`
- **THEN** the CLI fetches the compute config discovery endpoint
- **AND** it renders the active provider and schema (and defaults if present)
