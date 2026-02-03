## ADDED Requirements

### Requirement: Provider compute_config docs list all supported fields
The documentation SHALL list every supported compute_config field for each available provider.

#### Scenario: Docker compute_config reference completeness
- **WHEN** a reader views the Docker compute_config reference
- **THEN** it includes every Docker compute_config field supported by the provider
- **AND** each field includes a brief description and expected type

#### Scenario: ECS compute_config reference completeness
- **WHEN** a reader views the ECS compute_config reference
- **THEN** it includes every ECS compute_config field supported by the provider
- **AND** each field includes a brief description and expected type

#### Scenario: Mock compute_config reference completeness
- **WHEN** a reader views the mock compute_config reference
- **THEN** it includes every mock compute_config field supported by the provider
- **AND** each field includes a brief description and expected type

### Requirement: Provide full JSON and YAML examples
The documentation SHALL include complete, multi-line JSON and YAML compute_config examples for each provider.

#### Scenario: JSON example is complete and readable
- **WHEN** a reader views the JSON example for a provider
- **THEN** the example shows all supported compute_config fields
- **AND** the example is formatted across multiple lines for readability

#### Scenario: YAML example is complete and readable
- **WHEN** a reader views the YAML example for a provider
- **THEN** the example shows all supported compute_config fields
- **AND** the example is formatted across multiple lines for readability

### Requirement: Document compute_config file loading syntax
The documentation SHALL describe how to provide compute_config via `file://` paths in the CLI.

#### Scenario: File URI usage documented
- **WHEN** a reader views the compute_config reference docs
- **THEN** they can find an example that uses a `file://` URI with `landlord-cli --config`
