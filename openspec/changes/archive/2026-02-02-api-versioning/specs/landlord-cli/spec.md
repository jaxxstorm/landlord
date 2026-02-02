## MODIFIED Requirements

### Requirement: CLI configuration
The system SHALL provide a CLI that loads configuration via Viper from flags, environment variables, and config files.

#### Scenario: Configuration precedence
- **WHEN** configuration is provided via multiple sources
- **THEN** CLI flags MUST override environment variables
- **AND** environment variables MUST override config files

#### Scenario: API base URL configuration
- **WHEN** the CLI is started without an explicit API base URL or version override
- **THEN** it SHALL use the configured default base URL and append the current stable API version prefix (`/v1`)
- **AND** it SHALL include the versioned base URL in request targets

#### Scenario: Explicit versioned base URL
- **WHEN** a user provides an API base URL that already includes a version prefix
- **THEN** the CLI SHALL use it as-is without adding another version segment
