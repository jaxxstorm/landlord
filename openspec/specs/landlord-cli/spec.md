# Specification: Landlord CLI

## Purpose

Define the CLI interface and behavior for interacting with the Landlord API. (TBD)

## ADDED Requirements

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

### Requirement: Verb-based command structure
The system SHALL expose verb-based commands for tenant operations, including create, list, get, set, and delete, accepting name or UUID identifiers.

#### Scenario: Create command
- **WHEN** a user runs `landlord-cli create --tenant-name <name> [--config <json>]`
- **THEN** the CLI SHALL issue a tenant create request to the Landlord API
- **AND** it SHALL include the compute_config payload when provided

#### Scenario: List command
- **WHEN** a user runs `landlord-cli list`
- **THEN** the CLI SHALL fetch the list of tenants from the Landlord API
- **AND** it SHALL render a formatted list of tenant identifiers and statuses

#### Scenario: Get command
- **WHEN** a user runs `landlord-cli get --tenant-id <id-or-name>`
- **THEN** the CLI SHALL fetch the tenant by UUID or name
- **AND** it SHALL render the tenant details in a readable, styled format

#### Scenario: Set command
- **WHEN** a user runs `landlord-cli set --tenant-id <id-or-name> [--config <json>]`
- **THEN** the CLI SHALL issue a tenant update request with the compute_config payload
- **AND** it SHALL render a confirmation with the updated tenant details

#### Scenario: Delete command
- **WHEN** a user runs `landlord-cli delete --tenant-id <id-or-name>`
- **THEN** the CLI SHALL issue a tenant delete request to the Landlord API
- **AND** it SHALL render a confirmation message

### Requirement: CLI compute config discovery
The system SHALL provide a CLI command to retrieve compute configuration schema from the API.

#### Scenario: Compute command
- **WHEN** a user runs `landlord-cli compute`
- **THEN** the CLI fetches the compute config discovery endpoint
- **AND** it renders the active provider and schema (and defaults if present)

### Requirement: Styled output
The CLI SHALL use Lip Gloss for output styling and Fang for CLI theming.

#### Scenario: Styled list output
- **WHEN** the CLI renders a list of tenants
- **THEN** the output SHALL include styled headers and rows
- **AND** the styling SHALL be applied consistently across commands

#### Scenario: Styled single-tenant output
- **WHEN** the CLI renders a single tenant (get or set)
- **THEN** the output SHALL include styled headers and fields
- **AND** the styling SHALL match the list output theme

### Requirement: Local dev execution
The CLI SHALL support local execution via `go run` without requiring a prebuilt binary.

#### Scenario: go run support
- **WHEN** a developer runs `go run ./cmd/cli`
- **THEN** the CLI SHALL load configuration and execute commands identically to a built binary
