## ADDED Requirements

### Requirement: Validate tenant name format
The system SHALL validate that tenant names meet format requirements.

#### Scenario: Name length validation
- **WHEN** a tenant name is provided in any request
- **THEN** the system validates it is between 1 and 255 characters

#### Scenario: Name cannot be empty
- **WHEN** a tenant name consists only of whitespace
- **THEN** the system rejects it with HTTP 400

#### Scenario: Name trimming
- **WHEN** a tenant name has leading or trailing whitespace
- **THEN** the system either trims it automatically or rejects it

### Requirement: Validate UUID format
The system SHALL validate UUID parameters in all tenant endpoints.

#### Scenario: Valid UUID v4 format
- **WHEN** a tenant ID is provided as a path parameter
- **THEN** the system validates it matches UUID format

#### Scenario: Invalid UUID format
- **WHEN** an ID parameter is not a valid UUID
- **THEN** the system returns HTTP 400 with clear validation error message

### Requirement: Validate request content type
The system SHALL verify that request bodies have correct content type headers.

#### Scenario: JSON content type required
- **WHEN** a request includes a body (POST, PUT, PATCH)
- **THEN** the Content-Type header MUST be application/json

#### Scenario: Missing content type
- **WHEN** a request with a body omits the Content-Type header
- **THEN** the system returns HTTP 400 or assumes application/json based on configuration

### Requirement: Validate JSON structure
The system SHALL validate that request bodies contain valid JSON.

#### Scenario: Malformed JSON
- **WHEN** a request body cannot be parsed as JSON
- **THEN** the system returns HTTP 400 with parse error details

#### Scenario: Valid JSON structure
- **WHEN** a request body is well-formed JSON
- **THEN** the system proceeds to field-level validation

### Requirement: Validate required fields
The system SHALL enforce that required fields are present in requests.

#### Scenario: Missing required field
- **WHEN** a creation request omits the required `name` field
- **THEN** the system returns HTTP 400 listing missing fields

#### Scenario: All required fields present
- **WHEN** all required fields are included
- **THEN** field-level validation proceeds

### Requirement: Reject unknown fields
The system SHALL handle unknown fields in request bodies according to configuration.

#### Scenario: Unknown field in strict mode
- **WHEN** a request includes fields not in the tenant schema and strict validation is enabled
- **THEN** the system returns HTTP 400 listing unrecognized fields

#### Scenario: Unknown field in lenient mode
- **WHEN** strict validation is disabled
- **THEN** the system ignores unknown fields

### Requirement: Return clear validation errors
The system SHALL provide descriptive validation error messages.

#### Scenario: Validation error response format
- **WHEN** validation fails
- **THEN** the error response includes field names, validation rules violated, and helpful messages

#### Scenario: Multiple validation errors
- **WHEN** multiple fields fail validation
- **THEN** the error response lists all validation failures, not just the first one
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