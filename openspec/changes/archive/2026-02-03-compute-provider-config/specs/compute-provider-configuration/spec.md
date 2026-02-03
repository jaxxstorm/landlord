## ADDED Requirements

### Requirement: Provider-keyed compute configuration
The system SHALL represent compute configuration as provider-keyed blocks under `compute`, where each provider is enabled by presence of its configuration block.

#### Scenario: Single provider config
- **WHEN** the configuration file defines `compute.<provider>` for exactly one provider
- **THEN** the system enables that provider and starts successfully without requiring any other compute provider configuration

#### Scenario: Multiple provider configs
- **WHEN** the configuration file defines `compute.<provider>` blocks for multiple providers
- **THEN** the system enables only those providers and treats all other providers as disabled

### Requirement: Invalid or deprecated compute keys are rejected
The system SHALL reject legacy compute configuration keys and unknown provider names at startup validation.

#### Scenario: Legacy keys present
- **WHEN** configuration includes `compute.default_provider` or `compute.defaults`
- **THEN** the system fails validation with an error indicating these keys are no longer supported

#### Scenario: Unknown provider block
- **WHEN** configuration includes a `compute.<provider>` block for a provider that is not registered
- **THEN** the system fails validation with an error indicating the provider is unknown

### Requirement: Provider config validation on enablement
The system SHALL validate each enabled provider's configuration using its provider-specific schema before starting services.

#### Scenario: Missing required provider fields
- **WHEN** a provider block is present but required fields are missing
- **THEN** startup validation fails with a provider-specific error describing the missing fields
