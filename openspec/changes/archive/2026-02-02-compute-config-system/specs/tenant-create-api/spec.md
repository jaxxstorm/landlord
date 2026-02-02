## MODIFIED Requirements

### Requirement: Accept provider-specific compute configuration
The system SHALL accept compute configuration specific to the active compute provider (Docker, ECS, Kubernetes).

#### Scenario: Compute configuration provided
- **WHEN** a creation request includes a `compute_config` field
- **THEN** the system validates it against the active provider's schema
- **AND** the request is rejected with HTTP 400 if the schema validation fails

#### Scenario: Valid Docker compute configuration
- **WHEN** a creation request includes Docker-specific configuration (e.g., env vars, volumes, network mode)
- **THEN** the Docker provider validates and accepts it

#### Scenario: Invalid compute configuration structure
- **WHEN** a creation request includes `compute_config` that doesn't match provider schema
- **THEN** the system returns HTTP 400 with detailed validation errors

#### Scenario: Missing required compute configuration fields
- **WHEN** a creation request omits required provider configuration fields
- **THEN** the system returns HTTP 400 listing missing fields

### Requirement: Validate compute configuration at API ingress
The system SHALL validate compute configuration immediately upon request receipt, before database operations.

#### Scenario: Early validation prevents wasted resources
- **WHEN** an invalid compute configuration is submitted
- **THEN** the system rejects it with HTTP 400 before creating database records or provisioning resources

#### Scenario: Provider performs validation
- **WHEN** compute configuration is provided
- **THEN** the active compute provider validates the configuration structure, types, and constraints

### Requirement: Store compute configuration with tenant
The system SHALL persist compute configuration as part of tenant desired state.

#### Scenario: Compute configuration stored
- **WHEN** a tenant is created with compute_config
- **THEN** the configuration is stored in Tenant.DesiredConfig for later provisioning

#### Scenario: Configuration retrievable
- **WHEN** a tenant is retrieved via GET /api/tenants/{id}
- **THEN** the response includes the compute_config that was provided at creation
