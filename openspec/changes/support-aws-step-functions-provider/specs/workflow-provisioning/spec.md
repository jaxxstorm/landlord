## MODIFIED Requirements

### Requirement: Workflow execution is reconciler-driven and idempotent
The workflow execution system SHALL be triggered by the reconciliation controller, and providers SHALL treat repeated triggers for the same tenant lifecycle revision as idempotent. The Step Functions provider SHALL invoke one configured shared state machine and SHALL derive its AWS execution name from the tenant UUID, action, and desired-state revision.

#### Scenario: Reconciler triggers workflow for requested tenant
- **WHEN** a tenant is in StatusRequested or StatusPlanning
- **THEN** the reconciliation controller SHALL invoke the selected workflow provider with tenant identity, action, desired configuration, and a deterministic lifecycle revision
- **AND** the Step Functions provider SHALL start the configured shared state machine with a deterministic execution name

#### Scenario: Reconciler retries on transient failure
- **WHEN** the reconciliation controller fails to trigger a workflow due to a timeout or provider unavailability
- **THEN** the tenant SHALL remain in a retriable state with null or unchanged workflow_execution_id
- **AND** the reconciliation controller SHALL retry on the next poll cycle using the same lifecycle revision

#### Scenario: Multiple reconciler attempts for same tenant
- **WHEN** the controller triggers a workflow multiple times for the same tenant lifecycle revision
- **THEN** the workflow provider SHALL receive duplicate StartExecution calls with the same execution identity
- **AND** the provider SHALL return the existing execution instead of creating duplicate side effects

#### Scenario: Desired configuration changes after terminal failure
- **WHEN** a terminal workflow failure is followed by a changed desired configuration
- **THEN** the reconciliation controller SHALL derive a new lifecycle revision
- **AND** the workflow provider SHALL start a distinct execution for that revision
