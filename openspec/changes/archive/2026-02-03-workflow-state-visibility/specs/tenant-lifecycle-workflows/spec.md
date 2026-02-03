## MODIFIED Requirements

### Requirement: Workflow status reporting
The system SHALL publish workflow status updates including sub-state, retry count, and error details for create, update, archive, and delete operations via the reconciler and workflow provider status polling.

#### Scenario: Status updates during execution
- **WHEN** a lifecycle workflow changes state
- **THEN** the reconciler SHALL update tenant status and status message based on provider status
- **AND** the reconciler SHALL update workflow sub-state from provider execution state
- **AND** the reconciler SHALL update workflow retry count from provider metadata
- **AND** the reconciler SHALL update workflow error message when errors occur
- **AND** status history MAY be recorded when supported
- **AND** the status SHALL be retrievable via the tenant status API

#### Scenario: Sub-state reflects workflow progress
- **WHEN** workflow provider status is polled
- **THEN** tenant.WorkflowSubState SHALL be updated to canonical sub-state
- **AND** valid sub-states include "running", "waiting", "backing-off", "error", "succeeded", "failed"

#### Scenario: Retry count tracks workflow retry attempts
- **WHEN** workflow encounters errors and retries
- **THEN** tenant.WorkflowRetryCount SHALL increment with each retry attempt
- **AND** retry count is extracted from provider-specific retry metadata

#### Scenario: Error message provides failure context
- **WHEN** workflow execution fails or errors
- **THEN** tenant.WorkflowErrorMessage SHALL be populated with provider error message
- **AND** error message is cleared when workflow recovers from error state
