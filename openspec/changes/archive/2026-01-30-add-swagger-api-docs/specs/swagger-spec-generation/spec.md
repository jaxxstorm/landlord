## ADDED Requirements

### Requirement: Generate OpenAPI 3.0 specification from Go annotations
The system SHALL automatically generate a complete OpenAPI 3.0 specification from Swagger annotations embedded in Go source code using swaggo/swag.

#### Scenario: Spec generation during build
- **WHEN** the build process runs the swag init command
- **THEN** an OpenAPI 3.0 specification file is generated at `docs/swagger.json`

#### Scenario: Spec reflects all API endpoints
- **WHEN** all HTTP handlers have Swagger annotations
- **THEN** the generated spec includes all endpoints with methods, paths, parameters, and response schemas

#### Scenario: Spec is deterministic
- **WHEN** the same annotated source code is processed
- **THEN** the generated spec is identical across multiple runs

### Requirement: Maintain spec file in version control
The system SHALL keep the generated OpenAPI specification file in the repository for distribution and reference.

#### Scenario: Spec file committed to repo
- **WHEN** specs are generated
- **THEN** the specification file exists in the repository at a known location for serving and distribution

#### Scenario: Spec is accessible via HTTP
- **WHEN** the API server runs
- **THEN** the OpenAPI specification is available at a documented endpoint (e.g., `/api/swagger.json`)

### Requirement: Support multiple API versions in spec
The system SHALL allow documenting multiple API versions or revisions within the OpenAPI specification.

#### Scenario: Spec includes version information
- **WHEN** the spec is generated
- **THEN** it includes a version field matching the application version

#### Scenario: API servers configured in spec
- **WHEN** the spec is generated
- **THEN** it includes server information for different deployment environments (development, production)
