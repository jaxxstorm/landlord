## ADDED Requirements

### Requirement: Compute execution status is tracked and queryable
The system SHALL track the status of compute provisioning operations throughout their lifecycle.

#### Scenario: Compute operation transitions through states
- **WHEN** a compute provisioning operation is initiated
- **THEN** it SHALL start in "pending" state
- **AND** transition to "running" when the provider begins execution
- **AND** eventually reach a terminal state ("succeeded", "failed", or "cancelled")

#### Scenario: Workflow can query compute execution status
- **WHEN** a workflow requests the status of a compute operation by execution ID
- **THEN** the system SHALL return the current state, progress indicators, and any error details
- **AND** the response SHALL be immediately available (not delayed by polling)

#### Scenario: Compute execution history is retained
- **WHEN** a compute operation completes
- **THEN** its status history SHALL be stored for audit and debugging
- **AND** the history SHALL include timestamps for each state transition

### Requirement: Compute executions have persistent identifiers
Each compute provisioning operation SHALL be assigned a unique, persistent identifier that enables tracking across system components.

#### Scenario: Compute execution ID is returned immediately
- **WHEN** a workflow initiates a compute provisioning operation
- **THEN** the operation SHALL return a unique execution ID immediately
- **AND** this ID SHALL be usable for subsequent status queries and callbacks

#### Scenario: Compute execution ID is deterministic for idempotent operations
- **WHEN** a compute operation is retried with the same parameters
- **THEN** the execution ID MAY be reused if the operation is idempotent
- **AND** duplicate execution ID calls SHALL not create redundant operations

### Requirement: Compute execution status includes resource identifiers
Status information for compute executions SHALL include identifiers for created resources.

#### Scenario: Successful compute execution returns resource IDs
- **WHEN** a compute provisioning operation succeeds
- **THEN** the status response SHALL include the compute resource identifiers (ECS task ARN, container ID, pod name, etc.)
- **AND** these identifiers SHALL be queryable by tenant ID for validation

#### Scenario: Failed execution includes error details
- **WHEN** a compute provisioning operation fails
- **THEN** the status SHALL include error messages and failure reasons
- **AND** sufficient context SHALL be available to diagnose the failure

### Requirement: Compute execution status updates are propagated to dependent systems
Status changes for compute operations SHALL trigger updates in dependent systems (tenant state, workflow state).

#### Scenario: Workflow receives notification of compute completion
- **WHEN** a compute operation reaches a terminal state
- **THEN** dependent systems (workflows, controllers) SHALL be notified
- **AND** these systems can act on the completion status

#### Scenario: Long-running compute operations are queryable
- **WHEN** a compute operation is still running
- **THEN** intermediate status (progress, logs, resource allocation details) SHALL be queryable
- **AND** workflows can check status without waiting for completion

### Requirement: Multiple compute executions can be tracked for a single tenant
A tenant lifecycle MAY involve multiple compute operations (initial provision, updates, rollbacks).

#### Scenario: Tenant can have concurrent compute operations
- **WHEN** a tenant is being updated while a previous operation is still running
- **THEN** the system SHALL track both operations independently
- **AND** each SHALL have its own execution ID and status history

#### Scenario: Execution history shows all operations for a tenant
- **WHEN** querying execution history for a tenant
- **THEN** the response SHALL include all historical compute operations in chronological order
- **AND** each entry SHALL include its unique execution ID and final status
