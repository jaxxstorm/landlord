# Specification: Landlord CLI

## MODIFIED Requirements

### Requirement: Verb-based command structure
The system SHALL expose verb-based commands for tenant operations, including create, list, get, set, and delete.

#### Scenario: Create command
- **WHEN** a user runs `landlord-cli create`
- **THEN** the CLI SHALL issue a tenant create request to the Landlord API
- **AND** it SHALL render a success message with the tenant identifier

#### Scenario: List command
- **WHEN** a user runs `landlord-cli list`
- **THEN** the CLI SHALL fetch the list of tenants from the Landlord API
- **AND** it SHALL render a formatted list of tenant identifiers and statuses

#### Scenario: Get command
- **WHEN** a user runs `landlord-cli get --tenant-id <id-or-name>`
- **THEN** the CLI SHALL fetch the tenant from the Landlord API
- **AND** it SHALL render the tenant details in a readable, styled format

#### Scenario: Set command
- **WHEN** a user runs `landlord-cli set --tenant-id <id-or-name> [--image <image>] [--config <config>]`
- **THEN** the CLI SHALL issue a tenant update request with only the provided fields
- **AND** it SHALL render a confirmation with the updated tenant details

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

#### Scenario: Styled single-tenant output
- **WHEN** the CLI renders a single tenant (get or set)
- **THEN** the output SHALL include styled headers and fields
- **AND** the styling SHALL match the list output theme
