## MODIFIED Requirements

### Requirement: Provisioning request payloads for Step Functions
The Step Functions provider SHALL start lifecycle executions from a payload that includes tenant identity, operation, desired configuration, compute-provider selection, and request metadata without requiring the state machine or executor to read tenant state from the Landlord database.

#### Scenario: Provisioning input for provision execution
- **WHEN** the reconciliation controller starts a provision execution
- **THEN** the execution input SHALL include tenant ID, tenant UUID, operation, desired configuration, selected compute provider when available, and metadata
- **AND** the lifecycle executor SHALL use that payload as its tenant-state source

#### Scenario: Provisioning input for delete execution
- **WHEN** the reconciliation controller starts a delete execution
- **THEN** the execution input SHALL include tenant ID, action, and desired configuration
- **AND** the lifecycle executor SHALL execute deletion without querying tenant state from the database

## ADDED Requirements

### Requirement: Step Functions provider uses a configured Standard state machine
The Step Functions provider SHALL use `workflow.step_functions.state_machine_arn` as the target for all lifecycle executions and SHALL require a region and state-machine ARN when it is selected as the default provider.

#### Scenario: Valid provider configuration
- **WHEN** `workflow.default_provider` is `step-functions` and the region and state-machine ARN are configured
- **THEN** Landlord SHALL initialize an AWS Step Functions client using its configured caller credentials
- **AND** it SHALL register the `step-functions` provider

#### Scenario: Missing state-machine ARN
- **WHEN** `workflow.default_provider` is `step-functions` and `state_machine_arn` is empty
- **THEN** configuration validation SHALL fail before reconciliation starts
- **AND** the error SHALL identify the missing state-machine ARN

### Requirement: Step Functions starts retry-safe executions
The provider SHALL derive a deterministic valid execution name from the tenant UUID, lifecycle operation, and desired-state revision and SHALL use the configured state-machine ARN when calling StartExecution.

#### Scenario: Duplicate trigger for unchanged desired state
- **WHEN** the reconciler repeats a trigger for the same tenant UUID, operation, and desired-state revision
- **THEN** the provider SHALL use the same execution name and serialized input
- **AND** it SHALL return the existing execution ARN instead of creating duplicate lifecycle work

#### Scenario: Desired state changes after a terminal execution
- **WHEN** a new desired-state revision is reconciled for a tenant after a terminal execution
- **THEN** the provider SHALL derive a different execution name for the new revision
- **AND** it SHALL start a new execution on the configured state machine

### Requirement: Step Functions reports execution state and diagnostics
The provider SHALL query AWS execution state and map it to Landlord execution state, input, output, timestamps, error details, and diagnostic history.

#### Scenario: Running execution status
- **WHEN** the reconciler queries a running Step Functions execution
- **THEN** the provider SHALL return `running` with the AWS execution ARN and start time
- **AND** it SHALL preserve the execution input for diagnostic use

#### Scenario: Failed execution status
- **WHEN** AWS reports an execution as failed, timed out, or aborted
- **THEN** the provider SHALL return a terminal Landlord execution state with AWS error and cause details
- **AND** it SHALL include available execution history events

#### Scenario: Execution does not exist
- **WHEN** AWS reports that an execution ARN does not exist
- **THEN** the provider SHALL return `workflow.ErrExecutionNotFound`
- **AND** it SHALL not report the execution as running

### Requirement: Step Functions stop requests are safe to repeat
The provider SHALL request AWS execution stop for active executions and SHALL treat a repeated stop of an already terminal or previously stopped execution as successful.

#### Scenario: Stop active execution
- **WHEN** tenant deletion or reconciliation requests a stop for an active execution
- **THEN** the provider SHALL call StopExecution with the supplied reason
- **AND** subsequent status polling SHALL observe the AWS terminal state

#### Scenario: Repeat stop request
- **WHEN** StopExecution is called more than once for the same execution ARN
- **THEN** the provider SHALL not create additional work or return a false success state
- **AND** it SHALL return success when the execution is already terminal or absent due to the prior stop
