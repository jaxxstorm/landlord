## Why

Local development and testing require a simple, reproducible environment that demonstrates the full Landlord control plane in action. Currently, developers must manually orchestrate multiple services (PostgreSQL, Landlord, Restate) and manage configuration files. A Docker Compose configuration with unified example and test configurations will significantly reduce setup friction and enable developers to test the complete system locally, including Docker compute provisioning and workflow orchestration.

## What Changes

- **New**: `docker-compose.yml` - Development environment with PostgreSQL, Landlord (with Docker compute provider), and Restate workflow engine
- **New**: `config.example.yaml` - Comprehensive example configuration with all options documented and commented
- **New**: `test.config.yaml` - Minimal test configuration for Docker Compose mounted into Landlord container
- **Consolidation**: All existing config examples (`config.yaml`, `config.json`, `config.sqlite.yaml`, `config.docker.yaml`) consolidated into single `config.example.yaml` with complete option reference
- **Enhancement**: Configuration documentation clarifying each option's purpose, defaults, and use cases

## Capabilities

### New Capabilities

- `docker-compose-setup`: Complete Docker Compose configuration for local testing with all integrated services (PostgreSQL, Landlord, Restate)
- `unified-config-example`: Comprehensive example configuration consolidating all service options with documentation
- `test-configuration`: Pre-configured test settings for Docker Compose local development environment
- `local-development-guide`: Documentation for developers on running the complete system locally

### Modified Capabilities

- `database-persistence`: No spec-level changes; test configuration demonstrates SQLite and PostgreSQL options
- `configuration-management`: No spec-level changes; enhanced with consolidated example and test configurations

## Impact

- **Developers**: Simplified onboarding - `docker-compose up` now runs complete system
- **Testing**: Local integration testing of Landlord with Docker compute and Restate workflows
- **Documentation**: Single reference point for all configuration options
- **File Organization**: Cleaner root directory with consolidated config examples
- **CI/CD**: Can leverage docker-compose configuration for automated testing

## Configuration Strategy

**config.example.yaml**: Reference documentation
- All possible configuration options
- Detailed comments explaining each section
- Multiple provider examples (PostgreSQL, SQLite)
- Multiple workflow providers (Mock, Restate, Step Functions)
- Defaults clearly marked

**test.config.yaml**: Docker Compose testing (minimal, specific)
- PostgreSQL database configuration (matches docker-compose service)
- Docker compute provider enabled
- Restate workflow provider configured
- All other options use sensible defaults

**docker-compose.yml**: Service orchestration
- PostgreSQL service with initialization
- Landlord service (built from current codebase)
- Restate service
- Network configuration for inter-service communication
- Volume mounts for configuration and persistence
