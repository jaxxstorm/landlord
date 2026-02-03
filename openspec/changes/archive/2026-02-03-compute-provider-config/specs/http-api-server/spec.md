## MODIFIED Requirements

### Requirement: Document endpoint with Swagger
The system SHALL include Swagger/OpenAPI annotations for all HTTP endpoints.

#### Scenario: Swagger documentation present
- **WHEN** the OpenAPI spec is generated
- **THEN** it includes the compute config discovery endpoint with request/response schemas and a provider identifier parameter

#### Scenario: Response schema documented
- **WHEN** viewing API documentation
- **THEN** the compute config discovery response schema shows provider identifier, schema, and defaults
