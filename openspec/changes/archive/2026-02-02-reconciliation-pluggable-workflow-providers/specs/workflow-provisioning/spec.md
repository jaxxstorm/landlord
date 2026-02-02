## MODIFIED Requirements

### Requirement: Workflow execution is reconciler-driven and idempotent
The workflow execution system SHALL be triggered by the reconciliation controller, and providers SHALL treat repeated triggers as idempotent when the execution ID matches.

#### Scenario: Reconciler triggers workflow for requested tenant
- **WHEN** a tenant is in StatusRequested or StatusPlanning
- **THEN** the reconciliation controller SHALL invoke the workflow provider with a deterministic execution ID
- **AND** the execution input SHALL include tenant identity, action, and desired configuration

#### Scenario: Reconciler retries on transient failure
- **WHEN** the reconciliation controller fails to trigger a workflow (e.g., timeout, provider unavailable)
- **THEN** the tenant SHALL remain in a retriable state with null or unchanged workflow_execution_id
- **AND** the reconciliation controller SHALL retry on the next poll cycle

#### Scenario: Multiple reconciler attempts for same tenant
- **WHEN** the controller triggers a workflow multiple times for the same tenant within a short time window
- **THEN** the workflow provider MAY receive duplicate StartExecution calls with the same execution ID
- **AND** the provider SHALL handle this idempotently (second call is no-op or returns existing execution)
