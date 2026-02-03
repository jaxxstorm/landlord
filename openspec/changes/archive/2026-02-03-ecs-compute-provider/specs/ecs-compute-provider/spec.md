## ADDED Requirements

### Requirement: ECS compute provider provisioning
The system SHALL provide an ECS compute provider that provisions tenant compute as a single ECS service per tenant using provider-specific configuration.

#### Scenario: Provision tenant service
- **WHEN** a workflow provisions compute with provider type "ecs" and a valid ECS provider config
- **THEN** the system creates or ensures an ECS service for the tenant and returns a successful provision result with the service identifier

### Requirement: ECS provider configuration validation
The system SHALL validate ECS provider configuration for required fields and surface validation errors before provisioning or persisting compute_config.

#### Scenario: Reject missing required ECS fields
- **WHEN** the ECS provider config omits required fields such as task definition ARN or cluster identifier
- **THEN** the system rejects the request with a validation error describing the missing fields

### Requirement: AWS credential resolution on worker
The system SHALL use the standard AWS SDK credential chain on the workflow worker to authenticate ECS operations, with optional assume-role parameters when configured.

#### Scenario: Use default credential chain
- **WHEN** the worker provisions ECS compute without explicit credentials configured
- **THEN** the system resolves credentials from the AWS SDK default chain and proceeds with ECS API calls

### Requirement: ECS service lifecycle management
The system SHALL support idempotent update and destroy operations for ECS services created for tenants.

#### Scenario: Update service configuration
- **WHEN** a workflow updates tenant compute with a changed ECS config or task definition
- **THEN** the system updates the existing ECS service and reports the update status

#### Scenario: Destroy service
- **WHEN** a workflow deletes tenant compute for provider type "ecs"
- **THEN** the system deletes the ECS service and completes without error if it already does not exist
