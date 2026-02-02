## ADDED Requirements

### Requirement: CLI configuration
The system SHALL provide a CLI that loads configuration via Viper from flags, environment variables, and config files.

#### Scenario: Configuration precedence
- **WHEN** configuration is provided via multiple sources
- **THEN** CLI flags MUST override environment variables
- **AND** environment variables MUST override config files

#### Scenario: API base URL configuration
- **WHEN** the CLI is started without an explicit API base URL flag
- **THEN** it SHALL use the configured default value from Viper
- **AND** it SHALL include the base URL in request targets

### Requirement: Verb-based command structure
The system SHALL expose verb-based commands for tenant operations.

#### Scenario: Create command
- **WHEN** a user runs `landlord-cli create`
- **THEN** the CLI SHALL issue a tenant create request to the Landlord API
- **AND** it SHALL render a success message with the tenant identifier

#### Scenario: List command
- **WHEN** a user runs `landlord-cli list`
- **THEN** the CLI SHALL fetch the list of tenants from the Landlord API
- **AND** it SHALL render a formatted list of tenant identifiers and statuses

#### Scenario: Delete command
- **WHEN** a user runs `landlord-cli delete <tenant-id>`
- **THEN** the CLI SHALL issue a tenant delete request to the Landlord API
- **AND** it SHALL render a confirmation message

### Requirement: Styled output
The CLI SHALL use Lip Gloss for output styling and Fang for CLI theming.

#### Scenario: Styled list output
- **WHEN** the CLI renders a list of tenants
- **THEN** the output SHALL include styled headers and rows
- **AND** the styling SHALL be applied consistently across commands

### Requirement: Local dev execution
The CLI SHALL support local execution via `go run` without requiring a prebuilt binary.

#### Scenario: go run support
- **WHEN** a developer runs `go run ./cmd/cli`
- **THEN** the CLI SHALL load configuration and execute commands identically to a built binary
