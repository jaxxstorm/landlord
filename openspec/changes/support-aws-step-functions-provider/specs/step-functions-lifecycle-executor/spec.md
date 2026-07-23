## ADDED Requirements

### Requirement: State machine dispatches lifecycle requests to Lambda
Landlord SHALL provide a versioned Standard Step Functions state-machine definition that invokes a configured Lambda lifecycle executor synchronously with the provisioning payload and AWS execution ARN.

#### Scenario: Start provision lifecycle execution
- **WHEN** the Step Functions provider starts a provision request
- **THEN** the state machine SHALL invoke the Lambda executor with the request payload and execution ARN
- **AND** it SHALL return the executor result as execution output

#### Scenario: Lifecycle task failure
- **WHEN** the Lambda executor returns a retryable lifecycle failure
- **THEN** the state machine SHALL retry according to its bounded retry policy
- **AND** it SHALL fail the execution with actionable error details after retries are exhausted

### Requirement: Lambda executor performs payload-driven lifecycle work
The Lambda lifecycle executor SHALL dispatch provision, update, and delete operations through the shared lifecycle executor and SHALL not query the tenant repository for desired state.

#### Scenario: Provision from Lambda payload
- **WHEN** Lambda receives a provision payload with a registered compute provider
- **THEN** it SHALL provision through the provider-agnostic compute interfaces
- **AND** return the tracked execution result to the state machine

#### Scenario: Invalid Lambda payload
- **WHEN** Lambda receives a payload without a tenant identifier or with an unknown operation or compute provider
- **THEN** it SHALL return a non-retryable validation failure
- **AND** the state machine SHALL expose that failure in execution status

### Requirement: Lambda cancellation is cooperative and auditable
The Lambda lifecycle executor SHALL inspect the supplied Step Functions execution context before mutable operation boundaries and SHALL stop issuing new mutations when the execution has been stopped or is terminal.

#### Scenario: Stop before compute mutation
- **WHEN** the Step Functions execution is stopped before Lambda issues a compute mutation
- **THEN** Lambda SHALL return a cancelled result without invoking the compute provider
- **AND** logs and returned status SHALL include the workflow execution ARN

#### Scenario: Stop races with completed mutation
- **WHEN** a stop request races with a compute mutation that has already completed
- **THEN** Lambda SHALL not attempt an unsafe compensating mutation automatically
- **AND** the compute result and execution ARN SHALL remain available for audit and reconciliation

### Requirement: AWS deployment roles are least-privilege and distinct
The deployment reference SHALL document separate permissions for the Landlord control plane, Step Functions state machine, and Lambda lifecycle executor.

#### Scenario: Control-plane invocation permissions
- **WHEN** an operator configures the Step Functions provider
- **THEN** documentation SHALL identify the minimum StartExecution, DescribeExecution, StopExecution, and history-read permissions required by the caller role
- **AND** it SHALL not require the caller role to hold compute-provider mutation permissions

#### Scenario: State machine and Lambda permissions
- **WHEN** an operator deploys the state machine and Lambda executor
- **THEN** documentation SHALL identify the state-machine permission to invoke Lambda
- **AND** it SHALL identify the Lambda permissions required by configured compute providers and execution-status cancellation checks
