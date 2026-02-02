## MODIFIED Requirements

### Requirement: Worker job execution contract
The system SHALL define a stable job payload for workers that includes tenant identity, action, desired configuration, and workflow execution ID without direct database access.

#### Scenario: Dispatch create operation
- **WHEN** a tenant create operation is scheduled for execution
- **THEN** the worker engine SHALL receive a job payload including tenant ID, operation "create", and desired configuration

#### Scenario: Dispatch update operation
- **WHEN** a tenant update operation is scheduled for execution
- **THEN** the worker engine SHALL receive a job payload including tenant ID, operation "update", and desired configuration

#### Scenario: Dispatch delete operation
- **WHEN** a tenant delete operation is scheduled for execution
- **THEN** the worker engine SHALL receive a job payload including tenant ID, operation "delete", and desired configuration

### Requirement: Compute engine resolution
The system SHALL allow worker engines to resolve compute engine selection from the request payload or provider defaults without requiring API lookups.

#### Scenario: Resolve compute engine from payload
- **WHEN** a worker job includes explicit compute engine information
- **THEN** the worker engine SHALL use the provided compute engine for execution

#### Scenario: Resolve compute engine from defaults
- **WHEN** a worker job omits explicit compute engine information
- **THEN** the worker engine SHALL fall back to default provider configuration or resolver logic
