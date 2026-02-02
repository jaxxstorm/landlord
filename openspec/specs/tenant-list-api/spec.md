## ADDED Requirements

### Requirement: List all tenants
The system SHALL provide a GET endpoint at `/v1/tenants` that returns a list of all tenants.

#### Scenario: Retrieve all tenants
- **WHEN** a client sends a GET request to `/v1/tenants`
- **THEN** the system returns an array of tenant resources

#### Scenario: No tenants exist
- **WHEN** there are no tenants in the system
- **THEN** the system returns HTTP 200 with an empty array

### Requirement: Support pagination
The system SHALL support pagination query parameters to limit result set size.

#### Scenario: Default pagination
- **WHEN** no pagination parameters are provided
- **THEN** the system returns all tenants or a default page size

#### Scenario: Limit parameter
- **WHEN** a `limit` query parameter is provided
- **THEN** the system returns at most that many tenants

#### Scenario: Offset parameter
- **WHEN** an `offset` query parameter is provided
- **THEN** the system skips that many tenants before returning results

#### Scenario: Combined limit and offset
- **WHEN** both `limit` and `offset` parameters are provided
- **THEN** the system applies offset first, then returns up to limit results

### Requirement: Return tenant metadata
The system SHALL include pagination metadata in list responses when pagination is used.

#### Scenario: Total count included
- **WHEN** pagination is active
- **THEN** the response includes total count of tenants

#### Scenario: Next page indicator
- **WHEN** more results exist beyond current page
- **THEN** the response indicates additional pages are available

### Requirement: Support filtering
The system SHALL allow filtering tenants by basic criteria.

#### Scenario: Filter by name prefix
- **WHEN** a `name` query parameter is provided
- **THEN** the system returns only tenants whose names contain that string

### Requirement: Return tenants in consistent order
The system SHALL return tenants in a deterministic sort order.

#### Scenario: Default sort order
- **WHEN** no sort parameter is specified
- **THEN** tenants are sorted by creation time (newest first) or by name

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
