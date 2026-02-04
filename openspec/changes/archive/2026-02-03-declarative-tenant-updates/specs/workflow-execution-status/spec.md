## MODIFIED Requirements

### Requirement: Reconciler extracts enriched status from workflow providers
The reconciler SHALL poll workflow provider GetExecutionStatus() and extract sub-state, retry count, error details, and configuration hash.

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

#### Scenario: Extract config hash from execution metadata
- **WHEN** reconciler receives ExecutionStatus with Metadata
- **THEN** the reconciler SHALL extract config_hash value from metadata map
- **AND** use the hash for config change detection
- **AND** handle missing config_hash gracefully (for old executions without hash)
