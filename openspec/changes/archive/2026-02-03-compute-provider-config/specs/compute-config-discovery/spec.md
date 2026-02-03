## MODIFIED Requirements

### Requirement: Expose compute config schema by provider
The system SHALL expose the compute provider configuration schema via an API endpoint for a requested provider.

#### Scenario: Compute config discovery response
- **WHEN** a client requests the compute config discovery endpoint with a provider identifier
- **THEN** the response includes the requested compute provider identifier
- **AND** the response includes a JSON Schema describing the provider-specific compute_config payload

#### Scenario: Provider not enabled
- **WHEN** a client requests a provider that is not configured/enabled
- **THEN** the API responds with a clear error indicating the provider is unavailable

#### Scenario: Schema format
- **WHEN** the schema is returned
- **THEN** it is a valid JSON Schema (draft 2020-12)
- **AND** it reflects the provider’s required and optional fields

### Requirement: Provide optional default compute config
The system SHALL return provider defaults when they exist for the requested provider.

#### Scenario: Defaults available
- **WHEN** the requested compute provider defines default configuration
- **THEN** the discovery response includes a defaults object

#### Scenario: Defaults not available
- **WHEN** the requested compute provider has no defaults
- **THEN** the discovery response omits defaults or sets it to null explicitly

### Requirement: CLI exposes compute config discovery
The system SHALL provide a CLI command to retrieve and display compute config discovery data for a specified provider.

#### Scenario: CLI compute command
- **WHEN** a user runs `landlord-cli compute --provider <name>`
- **THEN** the CLI fetches the discovery endpoint for that provider
- **AND** it renders the provider name and schema (and defaults if present)
