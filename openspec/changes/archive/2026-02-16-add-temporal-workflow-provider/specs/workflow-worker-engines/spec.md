## MODIFIED Requirements

### Requirement: Worker job execution contract
The system SHALL define a stable job payload for workers that includes tenant identity, action, desired configuration, workflow execution ID, and cancellation context without direct database access.

#### Scenario: Dispatch create operation
- **WHEN** a tenant create operation is scheduled for execution
- **THEN** the worker engine SHALL receive a job payload including tenant ID, operation "create", and desired configuration

#### Scenario: Dispatch update operation
- **WHEN** a tenant update operation is scheduled for execution
- **THEN** the worker engine SHALL receive a job payload including tenant ID, operation "update", and desired configuration

#### Scenario: Dispatch delete operation
- **WHEN** a tenant delete operation is scheduled for execution
- **THEN** the worker engine SHALL receive a job payload including tenant ID, operation "delete", and desired configuration

#### Scenario: Dispatch cancellation signal
- **WHEN** a lifecycle execution is cancelled while work is in progress
- **THEN** the worker engine SHALL receive cancellation context associated with the workflow execution ID
- **AND** the worker engine SHALL stop work at safe boundaries and report terminal cancellation status

### Requirement: Compute engine resolution
The system SHALL allow worker engines to resolve compute engine selection from the request payload or provider defaults without requiring API lookups, and SHALL support all registered compute providers.

#### Scenario: Resolve compute engine from payload
- **WHEN** a worker job includes explicit compute engine information
- **THEN** the worker engine SHALL use the provided compute engine for execution

#### Scenario: Resolve compute engine from defaults
- **WHEN** a worker job omits explicit compute engine information
- **THEN** the worker engine SHALL fall back to default provider configuration or resolver logic

#### Scenario: Execute against registered compute providers
- **WHEN** a compute provider is registered and enabled in configuration
- **THEN** worker engines SHALL invoke compute operations through provider-agnostic compute interfaces
- **AND** workflow behavior SHALL remain consistent regardless of which registered compute provider is selected

#### Scenario: Unknown compute provider selection
- **WHEN** a worker job references a compute provider that is not registered
- **THEN** the worker engine SHALL fail with a validation error that identifies the unknown provider
- **AND** the failure SHALL be surfaced through workflow status and logs

