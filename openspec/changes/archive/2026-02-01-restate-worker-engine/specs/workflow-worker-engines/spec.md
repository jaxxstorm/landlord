## ADDED Requirements

### Requirement: Worker engine interface
The system SHALL define a workflow worker engine interface that supports registration, lifecycle control, and execution of workflow jobs.

#### Scenario: Worker engine registration
- **WHEN** the application boots a worker process
- **THEN** the worker engine implementation SHALL register itself with the worker registry
- **AND** the registry SHALL expose the worker by its engine name (for example, "restate")

#### Scenario: Worker engine lifecycle
- **WHEN** the worker engine Start method is invoked
- **THEN** the engine SHALL initialize dependencies and report readiness or error

### Requirement: Worker engine selection
The system SHALL select a worker engine based on the configured workflow engine name.

#### Scenario: Select restate worker engine
- **WHEN** the workflow engine is configured as "restate"
- **THEN** the worker runtime SHALL instantiate the restate worker engine

#### Scenario: Unknown engine
- **WHEN** the workflow engine name does not match any registered worker engine
- **THEN** worker startup SHALL fail with a clear configuration error

### Requirement: Worker job execution contract
The system SHALL define a stable job payload for workers that includes tenant identity and the requested lifecycle operation.

#### Scenario: Dispatch create operation
- **WHEN** a tenant create operation is scheduled for execution
- **THEN** the worker engine SHALL receive a job payload including tenant ID, operation "create", and requested configuration

#### Scenario: Dispatch update operation
- **WHEN** a tenant update operation is scheduled for execution
- **THEN** the worker engine SHALL receive a job payload including tenant ID, operation "update", and desired configuration

#### Scenario: Dispatch delete operation
- **WHEN** a tenant delete operation is scheduled for execution
- **THEN** the worker engine SHALL receive a job payload including tenant ID and operation "delete"

### Requirement: Compute engine resolution
The system SHALL allow worker engines to resolve compute engine selection from the Landlord server when not provided in the job payload.

#### Scenario: Resolve compute engine via API
- **WHEN** a worker job is missing explicit compute engine information
- **THEN** the worker engine SHALL query the Landlord API for the tenant's compute engine selection
- **AND** SHALL use the resolved compute engine for execution
