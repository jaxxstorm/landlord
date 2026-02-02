## ADDED Requirements

### Requirement: Compute operation failures are surfaced to workflows with structured error information
The system SHALL provide detailed, structured error information when compute operations fail to enable intelligent workflow error handling.

#### Scenario: Compute operation returns detailed error
- **WHEN** a compute provisioning operation fails (e.g., insufficient resources, invalid configuration)
- **THEN** the error response SHALL include an error code, message, and root cause
- **AND** the workflow system SHALL propagate this to the workflow execution context

#### Scenario: Workflow receives error classification
- **WHEN** a compute operation fails
- **THEN** the error response SHALL classify the failure as retriable, non-retriable, or requires-manual-intervention
- **AND** the workflow can decide next steps based on this classification

#### Scenario: Failed compute operation doesn't corrupt tenant state
- **WHEN** a compute operation fails partway through
- **THEN** the system SHALL rollback or mark partial resources for cleanup
- **AND** the tenant SHALL remain in a valid state for retry or recovery

### Requirement: Compute errors enable workflow-level retry logic
Workflow systems SHALL have the information needed to implement retry strategies for compute failures.

#### Scenario: Retriable error allows workflow retry
- **WHEN** a compute operation fails with a retriable error (e.g., timeout, temporary service unavailable)
- **THEN** the workflow can invoke the compute operation again with backoff
- **AND** the retry SHALL use the same execution ID for idempotency

#### Scenario: Non-retriable error triggers error handling
- **WHEN** a compute operation fails with a non-retriable error (e.g., invalid configuration, quota exceeded)
- **THEN** the workflow SHALL exit with an error state
- **AND** the tenant SHALL be marked as requiring manual investigation

#### Scenario: Partial compute failures are recoverable
- **WHEN** a compute operation partially succeeds (e.g., some containers launched, some failed)
- **THEN** the error response SHALL include details of what succeeded and what failed
- **AND** the workflow can decide whether to rollback everything or retry just the failed parts

### Requirement: Compute provider errors are standardized across providers
All compute providers SHALL return errors in a consistent format to simplify workflow error handling logic.

#### Scenario: Error format is consistent across ECS and Kubernetes providers
- **WHEN** compute provisioning fails in either ECS or Kubernetes
- **THEN** the error response format, error codes, and field names SHALL be identical
- **AND** workflow error handling logic works with any provider

#### Scenario: Provider-specific errors are mapped to standard codes
- **WHEN** an ECS operation fails with "ServiceUnavailable"
- **THEN** the system SHALL map this to a standard "PROVIDER_TEMPORARILY_UNAVAILABLE" code
- **AND** workflows treat similar failures identically regardless of underlying provider

### Requirement: Compute operation failures are logged for observability
Compute failures SHALL be logged with sufficient context for debugging without exposing sensitive details.

#### Scenario: Compute failure is logged with tenant and operation context
- **WHEN** a compute operation fails
- **THEN** the log entry SHALL include tenant ID, operation type, execution ID, and error details
- **AND** the log level SHALL be appropriate (error for permanent failures, warn for retriable failures)

#### Scenario: Provider-level debug information is captured
- **WHEN** a compute operation fails due to provider error
- **THEN** the provider's raw response (if not containing secrets) SHALL be captured in debug logs
- **AND** this information SHALL be available for ops debugging

### Requirement: Compute error recovery strategies are workflow-declarable
Workflows SHALL be able to declare how compute errors should be handled without hardcoding retry logic.

#### Scenario: Workflow declares max retry attempts
- **WHEN** a workflow defines a compute operation with retry configuration
- **THEN** the workflow system SHALL automatically retry up to the specified limit
- **AND** use exponential backoff between retries

#### Scenario: Workflow declares error-specific handling
- **WHEN** a compute operation fails with a specific error code
- **THEN** the workflow can declare different handling paths (e.g., retry for timeout, abort for invalid config)
- **AND** the workflow engine SHALL route to the appropriate handler

### Requirement: Compute operation timeouts are configurable and recoverable
Long-running compute operations SHALL support configurable timeouts with recovery options.

#### Scenario: Compute operation times out and is retriable
- **WHEN** a compute operation exceeds the configured timeout threshold
- **THEN** the system SHALL mark it as timed-out but retriable
- **AND** the workflow can retry with increased timeout or different parameters

#### Scenario: Timeout doesn't leave resources orphaned
- **WHEN** a compute operation times out
- **THEN** the system SHALL not immediately terminate provider-side operations
- **AND** the provider SHALL eventually complete or fail, allowing status updates
