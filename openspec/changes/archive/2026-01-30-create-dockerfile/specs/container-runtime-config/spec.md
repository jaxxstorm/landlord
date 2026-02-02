## ADDED Requirements

### Requirement: Container runtime configuration for environment variables
The system SHALL accept Landlord configuration via environment variables at runtime, enabling cloud-native deployment patterns.

#### Scenario: Environment variables override defaults
- **WHEN** a container starts with environment variables like `LANDLORD_LOG_LEVEL=debug`
- **THEN** the running Landlord application uses the provided values to configure its behavior

#### Scenario: Configuration persists across container restarts
- **WHEN** a container is stopped and restarted with the same environment variables
- **THEN** the configuration remains consistent without requiring image rebuilds

### Requirement: Health check endpoint is accessible
The system SHALL expose an HTTP health check endpoint that container orchestrators can use to verify Landlord is running.

#### Scenario: Health check responds to orchestrator
- **WHEN** a container orchestrator sends a request to the health check endpoint (e.g., `/health`)
- **THEN** the endpoint responds with HTTP 200 status code and a healthy status indicator

#### Scenario: Health check detects unhealthy state
- **WHEN** Landlord encounters a critical error (e.g., database connection failure)
- **THEN** the health check endpoint responds with a non-200 status code (e.g., 503) to signal unhealthy state

### Requirement: Container runs as nonroot user
The system SHALL configure the container to run with a nonroot user for security compliance.

#### Scenario: Process runs with nonroot privileges
- **WHEN** the container starts the Landlord binary
- **THEN** the process runs as a nonroot user (e.g., `nonroot` uid 65532) to minimize privilege escalation risk

### Requirement: Port binding configuration
The system SHALL expose the Landlord HTTP API port (default 8080) from the container.

#### Scenario: API port is accessible from host
- **WHEN** a container maps port 8080 to the host (e.g., `-p 8080:8080`)
- **THEN** the Landlord API is accessible at `http://localhost:8080` and `/api/docs` returns the Swagger documentation

### Requirement: Container can be stopped gracefully
The system SHALL properly handle SIGTERM signals to allow graceful shutdown.

#### Scenario: Container shutdown completes cleanly
- **WHEN** a container orchestrator sends SIGTERM (e.g., `docker stop`)
- **THEN** the Landlord process shuts down cleanly, closes database connections, and exits within 30 seconds
