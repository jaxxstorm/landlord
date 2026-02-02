## ADDED Requirements

### Requirement: Workflow execution supports dual-trigger pattern
The workflow execution system SHALL support being triggered by both the API handlers and the reconciliation controller without creating duplicate executions.

#### Scenario: API triggers workflow first
- **WHEN** an API handler triggers a workflow for a tenant
- **THEN** the workflow SHALL execute with a deterministic execution ID
- **AND** subsequent controller polls SHALL detect the existing execution ID and skip re-triggering

#### Scenario: Controller triggers workflow when API fails
- **WHEN** an API handler fails to trigger a workflow (e.g., timeout, provider unavailable)
- **THEN** the tenant SHALL remain in a retriable state with null workflow_execution_id
- **AND** the reconciliation controller SHALL trigger the workflow on the next poll cycle

#### Scenario: Both API and controller attempt to trigger simultaneously
- **WHEN** API and controller both attempt to trigger a workflow for the same tenant within a short time window
- **THEN** the workflow provider MAY receive duplicate StartExecution calls with the same execution ID
- **AND** the provider SHALL handle this idempotently (second call is no-op or returns existing execution)

### Requirement: Execution ID format enables deduplication
The workflow execution ID SHALL follow a deterministic format based on tenant ID and action to enable deduplication across API and controller triggers.

#### Scenario: Execution ID includes tenant identifier
- **WHEN** a workflow is triggered for tenant "my-app" with action "plan"
- **THEN** the execution ID MUST include "my-app" to enable tenant-specific deduplication

#### Scenario: Execution ID includes action type
- **WHEN** a workflow is triggered with action "provision"
- **THEN** the execution ID MUST include "provision" to distinguish from other actions (plan, update, delete)

#### Scenario: Execution ID is deterministic
- **WHEN** the same tenant and action are triggered multiple times
- **THEN** the execution ID MUST be identical to enable provider-level deduplication

### Requirement: StartExecution is idempotent for duplicate execution IDs
The workflow provider's StartExecution method SHALL handle duplicate calls with the same execution ID idempotently.

#### Scenario: Second StartExecution call with same ID returns existing execution
- **WHEN** StartExecution is called twice with the same execution ID
- **THEN** the second call SHALL NOT create a new execution
- **AND** SHALL return the execution ID of the existing execution without error

#### Scenario: Concurrent StartExecution calls with same ID
- **WHEN** StartExecution is called concurrently from multiple callers with the same execution ID
- **THEN** only one execution SHALL be created
- **AND** all callers SHALL receive the same execution ID in the result

### Requirement: Controller checks workflow execution status before re-triggering
The reconciliation controller SHALL check if a workflow is already active before triggering a new workflow execution.

#### Scenario: Controller skips tenant with active workflow
- **WHEN** the controller polls and finds a tenant with non-null workflow_execution_id
- **THEN** the controller SHALL call GetExecutionStatus to verify the workflow is still running
- **AND** SHALL skip triggering if the execution status is "running" or "pending"

#### Scenario: Controller re-triggers after workflow completion
- **WHEN** the controller finds a tenant with workflow_execution_id pointing to a completed execution
- **THEN** the controller SHALL trigger a new workflow with a new execution ID if the tenant status still requires reconciliation

#### Scenario: Controller re-triggers after workflow failure
- **WHEN** the controller finds a tenant with workflow_execution_id pointing to a failed execution
- **THEN** the controller SHALL trigger a new workflow with a new execution ID following retry/backoff logic

### Requirement: Workflow manager provides trigger source tracking
The workflow execution input SHALL include metadata indicating the trigger source (API or controller) for observability.

#### Scenario: API-triggered workflow includes source metadata
- **WHEN** an API handler triggers a workflow via WorkflowClient
- **THEN** the execution input MUST include "trigger_source": "api"

#### Scenario: Controller-triggered workflow includes source metadata
- **WHEN** the reconciliation controller triggers a workflow
- **THEN** the execution input MUST include "trigger_source": "controller"

#### Scenario: Trigger source is logged for debugging
- **WHEN** a workflow execution is started
- **THEN** the workflow manager SHALL log the trigger source alongside the execution ID for debugging duplicate triggers
