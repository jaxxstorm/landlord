## ADDED Requirements

### Requirement: Expose active compute config schema
The system SHALL expose the active compute provider and its configuration schema via an API endpoint.

#### Scenario: Compute config discovery response
- **WHEN** a client requests the compute config discovery endpoint
- **THEN** the response includes the active compute provider identifier
- **AND** the response includes a JSON Schema describing the provider-specific compute_config payload

#### Scenario: Schema format
- **WHEN** the schema is returned
- **THEN** it is a valid JSON Schema (draft 2020-12)
- **AND** it reflects the provider’s required and optional fields

### Requirement: Provide optional default compute config
The system SHALL return provider defaults when they exist.

#### Scenario: Defaults available
- **WHEN** the active compute provider defines default configuration
- **THEN** the discovery response includes a defaults object

#### Scenario: Defaults not available
- **WHEN** the active compute provider has no defaults
- **THEN** the discovery response omits defaults or sets it to null explicitly

### Requirement: CLI exposes compute config discovery
The system SHALL provide a CLI command to retrieve and display compute config discovery data.

#### Scenario: CLI compute command
- **WHEN** a user runs `landlord-cli compute`
- **THEN** the CLI fetches the discovery endpoint
- **AND** it renders the provider name and schema (and defaults if present)
