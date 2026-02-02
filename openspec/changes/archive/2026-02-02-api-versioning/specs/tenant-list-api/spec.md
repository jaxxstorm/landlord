## MODIFIED Requirements

### Requirement: List all tenants
The system SHALL provide a GET endpoint at `/v1/tenants` that returns a list of all tenants.

#### Scenario: Retrieve all tenants
- **WHEN** a client sends a GET request to `/v1/tenants`
- **THEN** the system returns an array of tenant resources

#### Scenario: No tenants exist
- **WHEN** there are no tenants in the system
- **THEN** the system returns HTTP 200 with an empty array

### Requirement: Document endpoint with Swagger
The system SHALL include Swagger/OpenAPI annotations for the list tenants endpoint.

#### Scenario: Swagger documentation present
- **WHEN** the OpenAPI spec is generated
- **THEN** it includes the GET `/v1/tenants` endpoint

#### Scenario: Query parameters documented
- **WHEN** viewing API documentation
- **THEN** pagination parameters (`limit`, `offset`) are documented as optional integers

#### Scenario: Response schema documented
- **WHEN** viewing API documentation
- **THEN** the 200 response shows an array of tenant objects
