## ADDED Requirements

### Requirement: Workflows receive callbacks on compute operation state changes
The workflow system SHALL support receiving notifications when compute operations change state, enabling dynamic workflow decisions.

#### Scenario: Workflow continues after compute operation completes
- **WHEN** a compute provisioning operation reaches a terminal state
- **THEN** the workflow system SHALL receive a callback notification
- **AND** the workflow execution SHALL resume or branch based on the result

#### Scenario: Workflow handles successful compute provisioning
- **WHEN** a compute provisioning operation succeeds and provides resource identifiers
- **THEN** the workflow SHALL receive a "compute-provisioned" callback
- **AND** the workflow logic can proceed to the next stage (e.g., health checks, monitoring setup)

#### Scenario: Workflow handles compute provisioning failure
- **WHEN** a compute provisioning operation fails
- **THEN** the workflow SHALL receive a "compute-failed" callback with error details
- **AND** the workflow logic can execute error handling (e.g., retry, rollback, notify operator)

### Requirement: Compute callbacks include sufficient context for workflow decisions
Callback messages SHALL contain all information needed for workflows to make routing decisions without additional queries.

#### Scenario: Success callback includes compute resource identifiers
- **WHEN** a compute provisioning operation completes successfully and returns a callback
- **THEN** the callback message SHALL include the resource identifiers (ARN, container ID, pod name)
- **AND** the workflow can reference these identifiers in subsequent operations

#### Scenario: Failure callback includes structured error information
- **WHEN** a compute provisioning operation fails and returns a callback
- **THEN** the callback SHALL include error code, message, and retry eligibility flags
- **AND** the workflow can determine if the operation is retriable or requires manual intervention

#### Scenario: Callback includes operation metadata
- **WHEN** a compute operation completes
- **THEN** the callback SHALL include the operation type (provision, update, delete), tenant ID, and execution duration
- **AND** the workflow can use this for logging, metrics, and audit trails

### Requirement: Workflow callbacks are delivered reliably
Compute operation callbacks SHALL be delivered to workflows with at-least-once semantics to prevent loss of critical status information.

#### Scenario: Callback is retried if workflow is temporarily unavailable
- **WHEN** a compute operation completes and attempts to notify the workflow
- **THEN** if the workflow system is temporarily unavailable, the callback SHALL be retried
- **AND** retries SHALL continue until the workflow acknowledges receipt

#### Scenario: Duplicate callbacks are handled idempotently
- **WHEN** a callback is delivered multiple times due to retry logic
- **THEN** the workflow system SHALL deduplicate based on operation ID and state
- **AND** the workflow execution state SHALL not be corrupted by duplicate callbacks

### Requirement: Workflows can optionally poll for compute status
In addition to callbacks, workflows SHALL support explicit status polling for compute operations.

#### Scenario: Workflow polls compute execution status
- **WHEN** a workflow needs to check if a compute operation has completed
- **THEN** the workflow can call a status query endpoint without waiting for a callback
- **AND** the result SHALL reflect the current state of the compute operation

#### Scenario: Polling is used as fallback to callbacks
- **WHEN** a workflow has not received an expected callback within a timeout
- **THEN** the workflow logic can explicitly poll the compute status
- **AND** based on the result, the workflow can proceed or trigger error handling

### Requirement: Callback routing supports multiple workflow providers
Callback delivery SHALL work with any pluggable workflow provider (Step Functions, Temporal, Restate).

#### Scenario: Step Functions workflow receives compute callback
- **WHEN** a compute operation completes in a Step Functions workflow
- **THEN** the callback SHALL be delivered as a task success/failure input
- **AND** the state machine can transition based on the result

#### Scenario: Temporal workflow receives compute callback
- **WHEN** a compute operation completes in a Temporal workflow
- **THEN** the callback SHALL be delivered to the appropriate activity or workflow signal
- **AND** the workflow can process the notification natively
