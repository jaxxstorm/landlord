## ADDED Requirements

### Requirement: HTTP server starts and binds to configured port

The system SHALL start an HTTP server that binds to a configurable host and port.

#### Scenario: Server starts successfully
- **WHEN** the application starts with valid configuration
- **THEN** the HTTP server binds to the configured address
- **AND** the server accepts incoming connections

#### Scenario: Server fails to bind to port
- **WHEN** the configured port is already in use
- **THEN** the application fails to start with a clear error message
- **AND** the error indicates which port failed to bind

### Requirement: Server provides health check endpoint

The system SHALL expose a health check endpoint that returns server liveness status.

#### Scenario: Health check on running server
- **WHEN** a request is made to GET /health
- **THEN** the server responds with 200 OK
- **AND** the response body indicates server is alive

### Requirement: Server provides readiness check endpoint

The system SHALL expose a readiness check endpoint that verifies all dependencies are operational.

#### Scenario: All dependencies are ready
- **WHEN** a request is made to GET /ready
- **THEN** the server responds with 200 OK if database connection is healthy
- **AND** the response body indicates service is ready

#### Scenario: Database dependency is not ready
- **WHEN** a request is made to GET /ready and database is unavailable
- **THEN** the server responds with 503 Service Unavailable
- **AND** the response body indicates database is not ready

### Requirement: Server supports graceful shutdown

The system SHALL gracefully shut down when receiving termination signals.

#### Scenario: Graceful shutdown on SIGTERM
- **WHEN** the server receives SIGTERM signal
- **THEN** the server stops accepting new connections
- **AND** existing connections are allowed to complete
- **AND** the server shuts down within the configured timeout

#### Scenario: Force shutdown on timeout
- **WHEN** graceful shutdown exceeds the timeout period
- **THEN** the server forcefully terminates remaining connections
- **AND** exits with an appropriate status code

### Requirement: Server uses chi router for request routing

The system SHALL use chi router to handle HTTP request routing and middleware.

#### Scenario: Route registration
- **WHEN** routes are defined
- **THEN** chi router handles path matching and method filtering
- **AND** middleware is executed in registration order

### Requirement: Server logs HTTP requests

The system SHALL log all incoming HTTP requests with relevant metadata.

#### Scenario: Request logging
- **WHEN** an HTTP request is received
- **THEN** the server logs the request method, path, status code, and duration
- **AND** logs include correlation IDs for request tracing
