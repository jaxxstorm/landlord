# Specification: Step Functions Provider

## Purpose

Defines requirements for the AWS Step Functions workflow provider implementation that enables the landlord control plane to orchestrate tenant provisioning workflows using AWS Step Functions as the workflow engine.

## ADDED Requirements

### Requirement: Provider implements workflow.Provider interface

The system SHALL provide an AWS Step Functions provider that implements all methods of the workflow.Provider interface.

#### Scenario: Provider registration
- **WHEN** the Step Functions provider is created
- **THEN** it implements workflow.Provider interface
- **AND** it can be registered in the workflow registry
- **AND** it returns "step-functions" as its provider name

#### Scenario: Provider methods available
- **WHEN** the Step Functions provider is instantiated
- **THEN** all required interface methods are available: Name(), CreateWorkflow(), StartExecution(), GetExecutionStatus(), DeleteWorkflow()

### Requirement: Provider creates AWS Step Functions state machines

The system SHALL create AWS Step Functions state machines from workflow specifications.

#### Scenario: Create state machine with valid ASL definition
- **WHEN** CreateWorkflow is called with valid Amazon States Language definition
- **THEN** a Step Functions state machine is created in AWS
- **AND** the state machine ARN is returned in ResourceIDs
- **AND** the state machine uses the configured IAM execution role

#### Scenario: Create state machine with invalid ASL
- **WHEN** CreateWorkflow is called with invalid ASL syntax
- **THEN** the operation fails with workflow.ErrInvalidSpec
- **AND** the error message includes ASL validation details

#### Scenario: Idempotent state machine creation
- **WHEN** CreateWorkflow is called twice with the same WorkflowID
- **THEN** the first call creates the state machine
- **AND** the second call succeeds without error
- **AND** both calls return the same state machine ARN

### Requirement: Provider validates ASL definition syntax

The system SHALL validate Amazon States Language JSON syntax before creating state machines.

#### Scenario: ASL JSON syntax validation
- **WHEN** CreateWorkflow receives Definition field
- **THEN** the provider validates it is valid JSON
- **AND** the provider validates basic ASL structure (StartAt, States fields present)
- **AND** invalid JSON fails fast with clear error message

#### Scenario: AWS API ASL validation
- **WHEN** CreateWorkflow calls AWS CreateStateMachine API
- **THEN** AWS performs comprehensive ASL validation
- **AND** validation errors are captured and returned to caller
- **AND** no state machine is created if validation fails

### Requirement: Provider starts workflow executions

The system SHALL start Step Functions state machine executions with input parameters.

#### Scenario: Start execution with input
- **WHEN** StartExecution is called with WorkflowID and ExecutionInput
- **THEN** a Step Functions execution is started
- **AND** the execution ARN is returned
- **AND** the input JSON is passed to the execution
- **AND** execution state is returned as "running"

#### Scenario: Start execution with unique name
- **WHEN** StartExecution is called with ExecutionInput.ExecutionName
- **THEN** the execution uses the provided name
- **AND** the name must be unique per state machine
- **AND** duplicate names fail with appropriate error

#### Scenario: Idempotent execution start
- **WHEN** StartExecution is called twice with the same ExecutionName
- **THEN** the first call starts the execution
- **AND** the second call returns error indicating execution already exists
- **AND** caller can query execution status to verify state

### Requirement: Provider queries execution status

The system SHALL retrieve current execution status and history from Step Functions.

#### Scenario: Get execution status for running execution
- **WHEN** GetExecutionStatus is called for a running execution
- **THEN** current execution state is returned as "running"
- **AND** start time is included
- **AND** input JSON is included
- **AND** execution history events are included

#### Scenario: Get execution status for completed execution
- **WHEN** GetExecutionStatus is called for a completed execution
- **THEN** execution state is returned as "succeeded" or "failed"
- **AND** start and stop times are included
- **AND** output JSON is included (for succeeded)
- **AND** error details are included (for failed)
- **AND** execution history is included

#### Scenario: Get execution status for non-existent execution
- **WHEN** GetExecutionStatus is called for execution that doesn't exist
- **THEN** the operation fails with workflow.ErrExecutionNotFound

### Requirement: Provider maps Step Functions states to workflow states

The system SHALL translate AWS Step Functions execution states to workflow.ExecutionState enum values.

#### Scenario: Map running state
- **WHEN** Step Functions execution state is RUNNING
- **THEN** workflow state is ExecutionStateRunning

#### Scenario: Map succeeded state
- **WHEN** Step Functions execution state is SUCCEEDED
- **THEN** workflow state is ExecutionStateSucceeded

#### Scenario: Map failed states
- **WHEN** Step Functions execution state is FAILED, TIMED_OUT, or ABORTED
- **THEN** workflow state is ExecutionStateFailed

### Requirement: Provider deletes state machines

The system SHALL delete AWS Step Functions state machines.

#### Scenario: Delete existing state machine
- **WHEN** DeleteWorkflow is called for existing state machine
- **THEN** the state machine is deleted from AWS
- **AND** the operation succeeds

#### Scenario: Delete non-existent state machine
- **WHEN** DeleteWorkflow is called for non-existent state machine
- **THEN** the operation fails with workflow.ErrWorkflowNotFound

#### Scenario: Delete state machine with active executions
- **WHEN** DeleteWorkflow is called for state machine with running executions
- **THEN** AWS returns error indicating active executions
- **AND** the error is returned to caller
- **AND** the state machine is not deleted

### Requirement: Provider uses AWS SDK v2

The system SHALL use AWS SDK for Go v2 for all Step Functions API calls.

#### Scenario: Initialize AWS SDK client
- **WHEN** the Step Functions provider is created
- **THEN** it initializes an AWS SDK v2 SFN client
- **AND** the client uses the configured AWS region
- **AND** the client respects context cancellation and timeouts

#### Scenario: SDK handles AWS credentials
- **WHEN** the provider makes AWS API calls
- **THEN** credentials are loaded from standard AWS credential chain
- **AND** IAM role, environment variables, or credential files are supported

### Requirement: Provider requires IAM execution role configuration

The system SHALL use a configured IAM role ARN for Step Functions state machine execution.

#### Scenario: Create state machine with role ARN
- **WHEN** CreateWorkflow is called
- **THEN** the configured IAM role ARN is used for state machine execution
- **AND** the role ARN is validated by AWS CreateStateMachine API

#### Scenario: Missing role ARN configuration
- **WHEN** provider is created without role ARN configuration
- **THEN** provider initialization fails with clear error message
- **AND** error indicates missing role ARN configuration

### Requirement: Provider respects context cancellation

The system SHALL respect context cancellation and timeouts for all AWS API operations.

#### Scenario: Context cancelled during API call
- **WHEN** a provider method is called with context
- **AND** the context is cancelled during AWS API call
- **THEN** the operation terminates promptly
- **AND** context.Canceled error is returned

#### Scenario: Context timeout during API call
- **WHEN** a provider method is called with context deadline
- **AND** the deadline is exceeded during AWS API call
- **THEN** the operation times out
- **AND** context.DeadlineExceeded error is returned

### Requirement: Provider wraps AWS errors with context

The system SHALL wrap AWS SDK errors with operation context and map to standard workflow errors where applicable.

#### Scenario: Map AWS not found error
- **WHEN** AWS API returns StateMachineDoesNotExist error
- **THEN** provider returns workflow.ErrWorkflowNotFound
- **AND** original AWS error is wrapped for debugging

#### Scenario: Map AWS validation error
- **WHEN** AWS API returns InvalidDefinition error
- **THEN** provider returns workflow.ErrInvalidSpec
- **AND** error message includes ASL validation details

#### Scenario: Preserve unknown AWS errors
- **WHEN** AWS API returns unexpected error
- **THEN** provider wraps error with operation context
- **AND** original error is preserved in error chain

### Requirement: Provider supports AWS region configuration

The system SHALL allow configuration of AWS region for Step Functions API calls.

#### Scenario: Configure AWS region
- **WHEN** provider is initialized with region configuration
- **THEN** all AWS API calls use the configured region
- **AND** state machines are created in the configured region

#### Scenario: Default region fallback
- **WHEN** no region is explicitly configured
- **THEN** provider uses AWS SDK default region resolution
- **AND** region from environment variables or AWS config is used

### Requirement: Provider builds state machine ARNs correctly

The system SHALL construct AWS Step Functions state machine ARNs following AWS ARN format.

#### Scenario: Build state machine ARN
- **WHEN** provider needs to reference a state machine
- **THEN** ARN is constructed as: arn:aws:states:{region}:{accountID}:stateMachine:{workflowID}
- **AND** region is from provider configuration
- **AND** account ID is determined from AWS credentials

#### Scenario: Use ARN in AWS API calls
- **WHEN** provider calls DescribeStateMachine or DeleteStateMachine
- **THEN** the constructed ARN is used in the API call
- **AND** AWS validates the ARN format

### Requirement: Provider includes execution history in status

The system SHALL retrieve and include execution history events when querying execution status.

#### Scenario: Retrieve execution history
- **WHEN** GetExecutionStatus is called
- **THEN** provider calls GetExecutionHistory AWS API
- **AND** execution events are included in ExecutionStatus response
- **AND** events are ordered chronologically

#### Scenario: Parse execution events
- **WHEN** execution history events are retrieved
- **THEN** events are converted to workflow.HistoryEvent format
- **AND** event timestamp, type, and details are preserved

### Requirement: Provider is thread-safe

The system SHALL provide thread-safe provider implementation for concurrent use.

#### Scenario: Concurrent API calls
- **WHEN** multiple goroutines call provider methods simultaneously
- **THEN** all operations complete without data races
- **AND** AWS SDK client handles concurrent requests safely

### Requirement: Provider validates workflow specification

The system SHALL validate workflow specification before attempting AWS API calls.

#### Scenario: Validate required fields
- **WHEN** CreateWorkflow receives a WorkflowSpec
- **THEN** provider validates WorkflowID is non-empty
- **AND** provider validates Definition is non-empty
- **AND** provider validates ProviderType matches "step-functions"

#### Scenario: Validation failure handling
- **WHEN** specification validation fails
- **THEN** provider returns workflow.ErrInvalidSpec
- **AND** no AWS API calls are made
- **AND** error message indicates which field failed validation

### Requirement: Provider supports workflow tagging

The system SHALL apply tags to Step Functions state machines from workflow specification.

#### Scenario: Create state machine with tags
- **WHEN** CreateWorkflow receives WorkflowSpec with Tags field
- **THEN** tags are applied to created state machine
- **AND** tags are visible in AWS Step Functions console

#### Scenario: Tags included in ResourceIDs
- **WHEN** CreateWorkflow succeeds
- **THEN** CreateWorkflowResult includes tags in metadata
- **AND** tags can be retrieved via DescribeStateMachine

### Requirement: Provider handles AWS throttling

The system SHALL handle AWS API rate limiting with exponential backoff retry.

#### Scenario: Throttling with retry
- **WHEN** AWS API returns ThrottlingException
- **THEN** provider retries with exponential backoff
- **AND** retry logic is handled by AWS SDK
- **AND** operation succeeds after retry

#### Scenario: Max retries exceeded
- **WHEN** AWS API continues throttling after max retries
- **THEN** provider returns error with throttling details
- **AND** caller can retry with backoff

### Requirement: Provider logs operations

The system SHALL log all workflow operations for observability.

#### Scenario: Log workflow creation
- **WHEN** CreateWorkflow is called
- **THEN** provider logs workflow ID and operation start
- **AND** provider logs success or failure outcome
- **AND** logs include state machine ARN on success

#### Scenario: Log execution start
- **WHEN** StartExecution is called
- **THEN** provider logs workflow ID and execution name
- **AND** provider logs execution ARN on success

#### Scenario: Log errors
- **WHEN** any operation fails
- **THEN** provider logs error details
- **AND** logs include operation context (workflow ID, execution ID)
