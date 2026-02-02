## ADDED Requirements

### Requirement: Docker Compose Setup Documentation

The system SHALL provide comprehensive documentation explaining how to use docker-compose.yml, including quickstart instructions, architecture overview, and troubleshooting for common issues.

#### Scenario: Quickstart guide is provided
- **WHEN** developer wants to start using docker-compose
- **THEN** documentation includes a "Quick Start" section with: clone repo, run `docker-compose up`, verify services are running

#### Scenario: Prerequisites are clearly listed
- **WHEN** reading setup documentation
- **THEN** it explicitly lists: Docker Desktop (macOS/Windows) or Docker Engine (Linux), docker-compose version requirement, minimum system resources

#### Scenario: Service architecture is explained
- **WHEN** reviewing documentation
- **THEN** it includes diagram or description of three services, their roles, and how they communicate

#### Scenario: Port mappings are documented
- **WHEN** looking for network information
- **THEN** documentation lists all exposed ports: Landlord (8080), Restate (8080), PostgreSQL (5432)

#### Scenario: Docker socket mounting is explained
- **WHEN** seeking to understand Docker compute provisioning
- **THEN** documentation explains how socket is mounted, why it's needed, and security implications

#### Scenario: Data persistence is documented
- **WHEN** wondering how PostgreSQL data is managed
- **THEN** documentation explains named volume usage and how to reset database (`docker-compose down -v`)

### Requirement: Troubleshooting Guide

The system SHALL provide troubleshooting documentation for common issues developers encounter when using docker-compose setup.

#### Scenario: Port conflict issues are addressed
- **WHEN** developer encounters port already in use error
- **THEN** documentation explains how to identify conflicting process and either stop it or remap ports in docker-compose.yml

#### Scenario: Docker socket permission issues are addressed
- **WHEN** Landlord cannot access Docker socket
- **THEN** documentation explains group membership, provides fix for macOS vs Linux, includes verification steps

#### Scenario: Service connectivity issues are addressed
- **WHEN** Landlord cannot connect to PostgreSQL or Restate
- **THEN** documentation provides `docker-compose logs` command and explains how to diagnose network issues

#### Scenario: Database initialization issues are addressed
- **WHEN** migrations fail to apply on startup
- **THEN** documentation explains migration setup and provides manual migration steps if needed

#### Scenario: Image pulling issues are addressed
- **WHEN** `docker-compose up` fails with image not found errors
- **THEN** documentation explains Docker registry access and provides solutions for offline/air-gapped environments

### Requirement: Platform-Specific Guidance

The system SHALL provide specific guidance for running docker-compose on different platforms (macOS, Linux, Windows) where behavior or setup differs.

#### Scenario: macOS instructions are provided
- **WHEN** developer on macOS reads setup guide
- **THEN** documentation covers Docker Desktop installation, socket path differences, and any macOS-specific gotchas

#### Scenario: Linux instructions are provided
- **WHEN** developer on Linux reads setup guide
- **THEN** documentation covers Docker Engine installation, docker-compose installation separate from desktop, and user group setup

#### Scenario: Windows instructions are provided
- **WHEN** developer on Windows reads setup guide
- **THEN** documentation covers WSL 2 / Hyper-V requirements, Docker Desktop for Windows, and socket path under Windows

### Requirement: Next Steps and Further Development

The system SHALL provide guidance on what developers can do next after the compose setup is running, including testing capabilities and extending the setup.

#### Scenario: API testing is documented
- **WHEN** developer wants to test Landlord API
- **THEN** documentation provides example API calls (curl or similar) to health endpoints and basic provisioning

#### Scenario: Tenant provisioning example is provided
- **WHEN** developer wants to create a test tenant
- **THEN** documentation includes example JSON payload and endpoint for tenant provisioning via Docker compute provider

#### Scenario: Workflow testing is documented
- **WHEN** developer wants to test workflow execution
- **THEN** documentation explains how to trigger workflows and check Restate status

#### Scenario: Log inspection is documented
- **WHEN** developer needs to debug issues
- **THEN** documentation explains `docker-compose logs` options to view logs from specific services
