# Specification: Workflow Execution Status

## Purpose

Define enriched workflow execution status retrieval and persistence including sub-state, retry count, and error details.

## ADDED Requirements

### Requirement: Reconciler extracts enriched status from workflow providers
The reconciler SHALL poll workflow provider GetExecutionStatus() and extract sub-state, retry count, and error details.

#### Scenario: Extract sub-state from execution status
- **WHEN** reconciler receives ExecutionStatus from workflow provider
- **THEN** the reconciler SHALL map ExecutionStatus.State to canonical sub-state
- **AND** store the sub-state in tenant.WorkflowSubState field

#### Scenario: Extract retry count from execution metadata
- **WHEN** workflow provider execution includes retry metadata
- **THEN** the reconciler SHALL extract retry attempt count
- **AND** store the count in tenant.WorkflowRetryCount field

#### Scenario: Extract error message from failed executions
- **WHEN** ExecutionStatus includes an Error field with message
- **THEN** the reconciler SHALL extract the error message
- **AND** store it in tenant.WorkflowErrorMessage field

#### Scenario: Clear error message on successful retry
- **WHEN** workflow transitions from error to running after retry
- **THEN** the reconciler SHALL clear tenant.WorkflowErrorMessage
- **AND** increment tenant.WorkflowRetryCount

### Requirement: Tenant model includes workflow execution status fields
The tenant domain model SHALL include fields for workflow sub-state, retry count, and error message.

#### Scenario: Tenant struct includes WorkflowSubState
- **WHEN** defining the Tenant struct
- **THEN** it SHALL include WorkflowSubState string field
- **AND** the field is nullable (pointer or empty string when not set)

#### Scenario: Tenant struct includes WorkflowRetryCount
- **WHEN** defining the Tenant struct
- **THEN** it SHALL include WorkflowRetryCount integer field
- **AND** the field defaults to 0 for initial execution

#### Scenario: Tenant struct includes WorkflowErrorMessage
- **WHEN** defining the Tenant struct
- **THEN** it SHALL include WorkflowErrorMessage string field
- **AND** the field is nullable (pointer or empty string when no error)

### Requirement: Reconciler only updates tenant when workflow status changes
The reconciler SHALL implement dirty checking to avoid unnecessary database writes when workflow status is unchanged.

#### Scenario: Skip update when sub-state unchanged
- **WHEN** reconciler polls workflow status and sub-state matches current tenant.WorkflowSubState
- **THEN** the reconciler SHALL NOT update the tenant record
- **AND** no database write occurs

#### Scenario: Update when sub-state changes
- **WHEN** workflow sub-state differs from tenant.WorkflowSubState
- **THEN** the reconciler SHALL update the tenant record
- **AND** persist the new sub-state value

#### Scenario: Update when retry count changes
- **WHEN** workflow retry count differs from tenant.WorkflowRetryCount
- **THEN** the reconciler SHALL update the tenant record
- **AND** persist the new retry count

#### Scenario: Update when error message changes
- **WHEN** workflow error message differs from tenant.WorkflowErrorMessage
- **THEN** the reconciler SHALL update the tenant record
- **AND** persist the new error message

### Requirement: Workflow execution status fields are cleared when execution completes
The system SHALL clear workflow execution status fields when workflow reaches terminal state.

#### Scenario: Clear fields on successful completion
- **WHEN** workflow transitions to succeeded state
- **THEN** the system SHALL set WorkflowSubState to "succeeded"
- **AND** clear WorkflowExecutionID
- **AND** clear WorkflowRetryCount
- **AND** clear WorkflowErrorMessage

#### Scenario: Preserve fields on terminal failure
- **WHEN** workflow transitions to failed state
- **THEN** the system SHALL set WorkflowSubState to "failed"
- **AND** preserve WorkflowExecutionID for audit
- **AND** preserve WorkflowRetryCount for analysis
- **AND** preserve WorkflowErrorMessage for debugging

### Requirement: Workflow execution status is queryable via repository
The tenant repository SHALL support querying tenants by workflow sub-state.

#### Scenario: Query tenants by sub-state
- **WHEN** querying repository for tenants with specific workflow sub-state
- **THEN** the repository SHALL return tenants matching that sub-state
- **AND** results are filterable (e.g., all "backing-off" tenants)

#### Scenario: Query tenants with errors
- **WHEN** querying repository for tenants with non-null WorkflowErrorMessage
- **THEN** the repository SHALL return tenants in error condition
- **AND** results include error message content

#### Scenario: Query tenants with high retry counts
- **WHEN** querying repository for tenants with WorkflowRetryCount above threshold
- **THEN** the repository SHALL return tenants experiencing retry loops
- **AND** results are sortable by retry count

### Requirement: Workflow status updates are logged
The reconciler SHALL log workflow status changes with contextual information.

#### Scenario: Log sub-state transitions
- **WHEN** reconciler updates tenant workflow sub-state
- **THEN** a structured log entry SHALL be written
- **AND** log includes tenant ID, old sub-state, new sub-state, and execution ID

#### Scenario: Log retry count changes
- **WHEN** workflow retry count increases
- **THEN** a warning-level log entry SHALL be written
- **AND** log includes tenant ID, retry count, and error message if present

#### Scenario: Log error message changes
- **WHEN** workflow error message is set or updated
- **THEN** an error-level log entry SHALL be written
- **AND** log includes tenant ID, execution ID, and full error message
