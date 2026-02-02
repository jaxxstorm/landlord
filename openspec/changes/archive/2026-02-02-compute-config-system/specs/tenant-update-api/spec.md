## MODIFIED Requirements

### Requirement: Support compute configuration updates
The system SHALL allow updating tenant compute configuration through update operations.

#### Scenario: Update compute configuration
- **WHEN** an update request includes a `compute_config` field
- **THEN** the system validates it against the active provider schema
- **AND** it updates Tenant.DesiredConfig only if validation succeeds

#### Scenario: Validate updated compute configuration
- **WHEN** compute configuration is provided in an update request
- **THEN** the active provider validates the new configuration before applying

#### Scenario: Invalid compute configuration in update
- **WHEN** an update request includes invalid compute_config
- **THEN** the system returns HTTP 400 with detailed validation errors from the provider

#### Scenario: Partial compute configuration update
- **WHEN** an update request provides partial compute_config (e.g., only env vars)
- **THEN** the system either replaces the entire compute_config or merges based on API semantics
