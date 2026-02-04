## MODIFIED Requirements

### Requirement: Workflow status reporting
The system SHALL publish workflow status updates including sub-state, retry count, error details, and config change detection for create, update, archive, and delete operations via the reconciler and workflow provider status polling.

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

#### Scenario: Config hash checked during status poll
- **WHEN** reconciler polls workflow execution status
- **THEN** the reconciler SHALL extract config_hash from execution metadata
- **AND** compare with current tenant compute_config hash
- **AND** trigger workflow restart if hashes differ and workflow is degraded

#### Scenario: Workflow restarted when config changes during backing-off
- **WHEN** workflow is in backing-off sub-state
- **AND** tenant compute_config has changed since workflow started
- **THEN** the reconciler SHALL stop the current workflow execution
- **AND** start a new workflow execution with updated config
- **AND** reset retry count and error state

#### Scenario: Workflow NOT restarted when config changes during healthy execution
- **WHEN** workflow is in running sub-state (actively provisioning)
- **AND** tenant compute_config has changed since workflow started
- **THEN** the reconciler SHALL NOT interrupt the current execution
- **AND** allow workflow to complete with original config
