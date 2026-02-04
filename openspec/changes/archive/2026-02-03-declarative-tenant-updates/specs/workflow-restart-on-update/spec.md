## ADDED Requirements

### Requirement: Reconciler detects config changes during workflow execution
The reconciler SHALL detect when a tenant's compute configuration has changed while a workflow execution is in progress.

#### Scenario: Config hash stored in workflow metadata
- **WHEN** reconciler starts a workflow execution
- **THEN** the system SHALL compute a SHA256 hash of the tenant's compute_config
- **AND** include the hash in the workflow execution metadata under key "config_hash"

#### Scenario: Config hash comparison on status poll
- **WHEN** reconciler polls workflow execution status
- **THEN** the system SHALL retrieve the config_hash from execution metadata
- **AND** compute the current tenant's compute_config hash
- **AND** compare the two hashes to detect config changes

#### Scenario: Config unchanged
- **WHEN** workflow execution metadata config_hash matches current tenant config hash
- **THEN** the system SHALL NOT trigger workflow restart
- **AND** continue normal workflow status polling

### Requirement: Reconciler stops degraded workflows on config change
The reconciler SHALL stop workflow executions in degraded states when configuration changes are detected.

#### Scenario: Stop backing-off workflow on config change
- **WHEN** workflow is in backing-off sub-state
- **AND** tenant compute_config hash differs from workflow execution metadata config_hash
- **THEN** the reconciler SHALL call provider.StopExecution with reason "Configuration updated"
- **AND** poll execution status until State == StateDone

#### Scenario: Stop retrying workflow on config change
- **WHEN** workflow is in retrying sub-state
- **AND** tenant compute_config hash differs from workflow execution metadata config_hash
- **THEN** the reconciler SHALL call provider.StopExecution with reason "Configuration updated"
- **AND** poll execution status until State == StateDone

#### Scenario: Preserve running workflows on config change
- **WHEN** workflow is in running sub-state (actively provisioning)
- **AND** tenant compute_config changes
- **THEN** the reconciler SHALL NOT stop the workflow
- **AND** allow the current execution to complete

#### Scenario: Preserve completed workflows on config change
- **WHEN** workflow is in succeeded or failed terminal state
- **AND** tenant compute_config changes
- **THEN** the reconciler SHALL NOT attempt to stop the workflow
- **AND** handle terminal state through normal reconciliation flow

### Requirement: Reconciler starts new workflow after stopping old execution
The reconciler SHALL start a fresh workflow execution with updated configuration after successfully stopping a degraded workflow.

#### Scenario: Clear execution ID before starting new workflow
- **WHEN** reconciler successfully stops workflow due to config change
- **THEN** the system SHALL set tenant.WorkflowExecutionID to empty string
- **AND** update the tenant record in database

#### Scenario: Start new workflow with updated config
- **WHEN** tenant.WorkflowExecutionID is cleared after config change stop
- **THEN** the reconciler SHALL call workflowClient.TriggerWorkflow with tenant and operation "update"
- **AND** include new config_hash in execution metadata
- **AND** update tenant.WorkflowExecutionID with new execution ID

#### Scenario: Idempotent workflow start
- **WHEN** starting new workflow after config change
- **THEN** the workflow provider SHALL enforce idempotency based on ExecutionName
- **AND** return existing execution if already started with same name

### Requirement: Config hash computed consistently
The system SHALL compute configuration hashes in a deterministic, consistent manner.

#### Scenario: Hash deterministic for same config
- **WHEN** computing hash for identical compute_config twice
- **THEN** the system SHALL produce identical hash values
- **AND** use SHA256 algorithm

#### Scenario: Hash different for different configs
- **WHEN** compute_config values differ (even minor changes)
- **THEN** the system SHALL produce different hash values
- **AND** detect the configuration change

#### Scenario: Hash null/empty config consistently
- **WHEN** compute_config is null or empty
- **THEN** the system SHALL compute hash of empty/null value consistently
- **AND** treat all null/empty configs as equivalent

### Requirement: Workflow restart preserves audit trail
The system SHALL maintain clear audit trail when workflows are restarted due to config changes.

#### Scenario: Log workflow stop reason
- **WHEN** reconciler stops workflow due to config change
- **THEN** the system SHALL log the stop action with reason "Configuration updated"
- **AND** include old and new config hashes in log context

#### Scenario: New execution ID recorded
- **WHEN** new workflow starts after config-change restart
- **THEN** the system SHALL record the new execution ID in tenant record
- **AND** new execution ID differs from stopped execution ID

#### Scenario: Workflow retry count resets
- **WHEN** new workflow starts after config-change restart
- **THEN** the system SHALL reset tenant.WorkflowRetryCount to 0
- **AND** new workflow starts with clean retry state
