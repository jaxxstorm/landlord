## MODIFIED Requirements

### Requirement: Expose compute config schema by provider
The system SHALL expose the compute provider configuration schema via an API endpoint for a requested provider, including the `ec2-runner` provider when it is configured and enabled.

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

#### Scenario: EC2 runner schema
- **WHEN** a client requests discovery for an enabled `ec2-runner` provider
- **THEN** the schema SHALL require non-empty `ami` and `user_data` string fields
- **AND** it SHALL not expose operator-owned launch defaults as tenant inputs
