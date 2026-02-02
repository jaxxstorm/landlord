## ADDED Requirements

### Requirement: Docker Compose Orchestration

The system SHALL provide a `docker-compose.yml` file that orchestrates three services: PostgreSQL database, Landlord control plane, and Restate workflow engine, configured for local development and testing.

#### Scenario: Compose file provides complete service definition
- **WHEN** developers run `docker-compose up` from the repository root
- **THEN** three services start successfully: postgresql (port 5432), landlord (port 8080), restate (port 8080)

#### Scenario: Services communicate on shared network
- **WHEN** Landlord container is running
- **THEN** it can connect to PostgreSQL service at hostname `postgres:5432` and Restate service at hostname `restate:8080` using Docker's internal DNS

#### Scenario: PostgreSQL initializes with Landlord schema
- **WHEN** PostgreSQL service starts
- **THEN** database migrations are applied automatically and schema is ready for Landlord

#### Scenario: Configuration is mounted into Landlord
- **WHEN** Landlord container starts
- **THEN** `test.config.yaml` is mounted at `/app/config.yaml` and used for configuration

#### Scenario: Docker socket is available to Landlord
- **WHEN** Landlord service is running
- **THEN** Docker socket from host (`/var/run/docker.sock`) is mounted into container, allowing container provisioning via Docker API

### Requirement: Persistent Data Storage

The system SHALL use named Docker volumes to persist PostgreSQL data between compose sessions, enabling developers to maintain database state across `docker-compose down/up` cycles.

#### Scenario: Database data persists across restarts
- **WHEN** docker-compose is stopped and restarted
- **THEN** PostgreSQL data remains intact from previous session

#### Scenario: Volume can be cleaned up
- **WHEN** developer runs `docker-compose down -v`
- **THEN** named volume is removed and next startup creates fresh database

### Requirement: Service Dependencies

The system SHALL ensure services start in correct order with proper dependency resolution: PostgreSQL must be ready before Landlord starts, Restate can start independently.

#### Scenario: Landlord waits for PostgreSQL
- **WHEN** all services are starting
- **THEN** Landlord service does not attempt connections until PostgreSQL reports healthy

#### Scenario: Startup completes successfully
- **WHEN** docker-compose up is executed
- **THEN** all services reach running state within 60 seconds with no errors

### Requirement: Environment Configuration

The system SHALL provide environment variables and file mounts to configure all three services without requiring changes to docker-compose.yml.

#### Scenario: Database credentials are configurable
- **WHEN** docker-compose.yml is reviewed
- **THEN** PostgreSQL user, password, and database name are visible as environment variables or file references

#### Scenario: Port mappings are explicit
- **WHEN** docker-compose.yml is reviewed
- **THEN** all service port mappings to host are documented and modifiable without code changes

#### Scenario: Service versions are pinned
- **WHEN** docker-compose.yml is reviewed
- **THEN** PostgreSQL image version and Restate image version are explicitly specified (not "latest")
