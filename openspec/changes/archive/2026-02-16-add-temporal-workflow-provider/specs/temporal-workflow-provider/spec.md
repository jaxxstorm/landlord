## ADDED Requirements

### Requirement: Temporal provider implements lifecycle workflow contract
The Temporal workflow provider SHALL implement the workflow provider interface for tenant lifecycle orchestration, including validate, create, start, status, stop, and delete operations.

#### Scenario: Validate and create workflow definition
- **WHEN** the workflow manager calls `Validate` and `CreateWorkflow` for a Temporal workflow
- **THEN** the provider MUST validate provider configuration and workflow definition input
- **AND** `CreateWorkflow` MUST be idempotent for the same workflow ID

#### Scenario: Shared lifecycle workflow ID is available by default
- **WHEN** controller-triggered lifecycle actions use the shared workflow ID `tenant-provisioning`
- **THEN** the Temporal provider SHALL accept invocation without requiring a per-tenant pre-registration step
- **AND** repeated startup or registration calls MUST remain idempotent

#### Scenario: Start lifecycle execution
- **WHEN** the workflow manager calls `StartExecution` for a tenant lifecycle action (create, update, or delete)
- **THEN** the provider SHALL start a Temporal workflow execution asynchronously
- **AND** return a non-empty execution ID with an initial running or waiting state

#### Scenario: Query and delete workflow resources
- **WHEN** the workflow manager calls `GetExecutionStatus` or `DeleteWorkflow`
- **THEN** the provider SHALL return provider status mapped to Landlord canonical status semantics
- **AND** `DeleteWorkflow` MUST be idempotent for missing or previously deleted workflow resources

### Requirement: Temporal provider supports idempotent cancellation
The Temporal workflow provider SHALL implement stop and cancellation behavior that is safe under retries and duplicate requests.

#### Scenario: Cancel running execution
- **WHEN** the workflow manager calls `StopExecution` for a running Temporal execution
- **THEN** the provider SHALL issue Temporal cancellation or termination using the supplied reason
- **AND** the execution SHALL transition toward a terminal state without duplicate side effects

#### Scenario: Repeat cancellation request
- **WHEN** `StopExecution` is called multiple times for the same execution ID
- **THEN** the provider MUST handle repeated calls idempotently
- **AND** return success or an equivalent already-terminal result without failing the reconciliation loop

### Requirement: Temporal provider preserves lifecycle payload and failure semantics
The Temporal workflow provider SHALL execute tenant lifecycle workflows using payload-provided tenant state and SHALL surface actionable failure and retry context.

#### Scenario: Payload-driven execution without worker database reads
- **WHEN** a Temporal execution is started from API or reconciler triggers
- **THEN** execution input MUST include tenant identity, lifecycle action, desired configuration, and compute selection data
- **AND** worker behavior SHALL be driven by that payload without direct tenant database reads

#### Scenario: Transient Temporal backend failure
- **WHEN** provider operations fail due to transient backend or transport errors
- **THEN** the provider SHALL return retryable errors that preserve idempotent retry behavior in the caller
- **AND** provider logs MUST include tenant ID, workflow ID, and execution ID when available
