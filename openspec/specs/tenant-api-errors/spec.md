## ADDED Requirements

### Requirement: Return HTTP 400 for unsupported API version
The system SHALL return HTTP 400 Bad Request when a request targets an unsupported API version.

#### Scenario: Unsupported version requested
- **WHEN** a client sends a request to `/v2/...` and v2 is not supported
- **THEN** the system returns HTTP 400 with error code `unsupported_version` and a list of supported versions

#### Scenario: Missing version prefix
- **WHEN** a client sends a request to an unversioned path (e.g., `/api/tenants`)
- **THEN** the system returns HTTP 400 with error code `version_required` and a list of supported versions

### Requirement: Return HTTP 400 for validation errors
The system SHALL return HTTP 400 Bad Request for client-side validation failures.

#### Scenario: Invalid request data
- **WHEN** a request fails validation (empty name, invalid UUID, malformed JSON)
- **THEN** the system returns HTTP 400

#### Scenario: Error includes validation details
- **WHEN** returning HTTP 400
- **THEN** the response body includes specific validation error messages

### Requirement: Return HTTP 404 for not found
The system SHALL return HTTP 404 Not Found when requested tenant does not exist.

#### Scenario: Get non-existent tenant
- **WHEN** a GET request specifies a tenant ID that does not exist
- **THEN** the system returns HTTP 404

#### Scenario: Update non-existent tenant
- **WHEN** an update request targets a non-existent tenant
- **THEN** the system returns HTTP 404

#### Scenario: Delete non-existent tenant
- **WHEN** a delete request targets a non-existent tenant
- **THEN** the system returns HTTP 404

### Requirement: Return HTTP 409 for conflicts
The system SHALL return HTTP 409 Conflict for operations that violate uniqueness constraints.

#### Scenario: Duplicate tenant name
- **WHEN** a creation request uses a name that already exists
- **THEN** the system returns HTTP 409 with conflict description

### Requirement: Return HTTP 500 for server errors
The system SHALL return HTTP 500 Internal Server Error for unexpected server-side failures.

#### Scenario: Database connection failure
- **WHEN** the system cannot connect to the database
- **THEN** it returns HTTP 500

#### Scenario: Unexpected error during processing
- **WHEN** an unhandled exception occurs
- **THEN** the system returns HTTP 500 with generic error message

### Requirement: Use consistent error response format
The system SHALL return errors in a standard JSON format across all tenant endpoints.

#### Scenario: Error response structure
- **WHEN** any error occurs
- **THEN** the response body includes `error` or `message` field with description

#### Scenario: Include error codes or types
- **WHEN** returning an error
- **THEN** the response optionally includes an error code or type for programmatic handling

#### Scenario: Include request ID
- **WHEN** an error occurs
- **THEN** the response includes a correlation or request ID for tracing

### Requirement: Avoid leaking sensitive information
The system SHALL not expose internal system details in error messages.

#### Scenario: Database errors sanitized
- **WHEN** a database error occurs
- **THEN** the error response does not include SQL statements, stack traces, or internal paths

#### Scenario: Validation errors are safe
- **WHEN** returning validation errors
- **THEN** error messages describe the issue without exposing system internals

### Requirement: Document error responses in Swagger
The system SHALL document all possible error responses in OpenAPI annotations.

#### Scenario: Common errors documented
- **WHEN** viewing API documentation for any tenant endpoint
- **THEN** it shows 400, 404, 500 responses with example schemas

#### Scenario: Endpoint-specific errors documented
- **WHEN** an endpoint can return specific errors (e.g., 409 for create)
- **THEN** those responses are documented in the OpenAPI spec
