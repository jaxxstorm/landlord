## MODIFIED Requirements

### Requirement: Validate compute configuration structure
The system SHALL validate compute configuration against provider-specific schemas.

#### Scenario: Provider validates configuration
- **WHEN** a request includes compute_config
- **THEN** the active compute provider validates the configuration structure and constraints

#### Scenario: Docker-specific validation
- **WHEN** landlord is running with Docker provider and compute_config is provided
- **THEN** Docker provider validates fields like env, volumes, network_mode, ports

#### Scenario: Invalid compute configuration type
- **WHEN** a compute_config field has wrong type (e.g., string instead of object)
- **THEN** the system returns HTTP 400 with type error details

#### Scenario: Missing required compute fields
- **WHEN** compute_config omits required provider fields
- **THEN** the system returns HTTP 400 listing missing required fields

#### Scenario: Unknown compute configuration fields
- **WHEN** compute_config includes fields not recognized by the provider
- **THEN** the system either rejects with HTTP 400 or ignores them based on provider policy

### Requirement: Validate compute configuration constraints
The system SHALL enforce provider-specific constraints on configuration values.

#### Scenario: Docker volume path validation
- **WHEN** Docker compute_config includes invalid volume mount paths
- **THEN** Docker provider returns validation error

#### Scenario: Port number validation
- **WHEN** compute_config includes port numbers outside valid range (1-65535)
- **THEN** the provider returns HTTP 400 with constraint violation

#### Scenario: Environment variable name validation
- **WHEN** compute_config includes invalid env var names
- **THEN** the provider validates and returns HTTP 400 if invalid
