## ADDED Requirements

### Requirement: Test Configuration File

The system SHALL provide a `test.config.yaml` file specifically tuned for Docker Compose local testing, configured with PostgreSQL database, Docker compute provider, and Restate workflow engine with sensible defaults.

#### Scenario: Configuration matches docker-compose services
- **WHEN** test.config.yaml is reviewed
- **THEN** database host is set to `postgres` (service name), compute provider is `docker`, workflow provider is `restate`

#### Scenario: Configuration inherits sensible defaults
- **WHEN** test.config.yaml is reviewed
- **THEN** it contains only minimal necessary overrides; other values come from code defaults or config.example.yaml

#### Scenario: All services are correctly configured
- **WHEN** Landlord starts with test.config.yaml in docker-compose
- **THEN** it successfully connects to PostgreSQL (port 5432), configures Docker provider, and configures Restate (port 8080)

#### Scenario: Docker provider is configured for tenant provisioning
- **WHEN** reviewing compute section
- **THEN** docker provider is explicitly enabled with appropriate host configuration for in-container socket access

#### Scenario: Restate workflow provider is configured
- **WHEN** reviewing workflow section
- **THEN** restate endpoint is configured to match docker-compose service endpoint

### Requirement: Configuration Consistency with Docker Compose

The system SHALL ensure test.config.yaml settings align exactly with docker-compose.yml service definitions, avoiding connectivity or configuration mismatches.

#### Scenario: Database configuration matches service definition
- **WHEN** comparing test.config.yaml database section with docker-compose.yml postgres service
- **THEN** hostname, port, user, password, and database name all match

#### Scenario: Workflow provider endpoint matches service
- **WHEN** comparing test.config.yaml workflow.restate section with docker-compose.yml restate service
- **THEN** endpoint URL uses correct service hostname and port

#### Scenario: HTTP port is configurable
- **WHEN** reviewing test.config.yaml http section
- **THEN** port is explicitly set to 8080, matching docker-compose port mapping

### Requirement: Minimal Configuration Philosophy

The system SHALL keep test.config.yaml minimal and focused, containing only settings that differ from sensible defaults or are specific to the Docker Compose testing environment.

#### Scenario: File size is reasonable
- **WHEN** reviewing test.config.yaml
- **THEN** it contains fewer than 50 lines (excluding comments)

#### Scenario: Comments explain Docker Compose context
- **WHEN** reading test.config.yaml
- **THEN** comments clarify that values are tuned for Docker Compose environment and reference docker-compose.yml for service details

#### Scenario: Unused sections are omitted
- **WHEN** scanning test.config.yaml
- **THEN** no placeholder sections for unconfigured providers or unused options
