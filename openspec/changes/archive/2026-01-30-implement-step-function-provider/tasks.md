# Tasks: Implement Step Functions Provider

## 1. Package Structure and Dependencies

- [x] 1.1 Create `internal/workflow/stepfunctions/` package directory
- [x] 1.2 Add AWS SDK v2 dependencies to go.mod: `github.com/aws/aws-sdk-go-v2`, `github.com/aws/aws-sdk-go-v2/service/sfn`
- [x] 1.3 Add AWS SDK config package: `github.com/aws/aws-sdk-go-v2/config`
- [x] 1.4 Run `go mod tidy` to download dependencies

## 2. Provider Configuration

- [x] 2.1 Add StepFunctionsConfig struct to `internal/config/config.go` with Region and RoleARN fields
- [x] 2.2 Update WorkflowConfig to include StepFunctionsConfig
- [x] 2.3 Add configuration validation for required RoleARN
- [x] 2.4 Add environment variable mapping for AWS region and role ARN

## 3. Provider Struct and Constructor

- [x] 3.1 Create `internal/workflow/stepfunctions/provider.go` with Provider struct
- [x] 3.2 Add fields: client (sfniface.ClientAPI), config, logger, accountID
- [x] 3.3 Implement New() constructor that initializes AWS SDK client
- [x] 3.4 Add accountID retrieval using STS GetCallerIdentity
- [x] 3.5 Implement Name() method returning "step-functions"

## 4. State Machine ARN Building

- [x] 4.1 Create `arn.go` file with ARN building utilities
- [x] 4.2 Implement buildStateMachineARN(workflowID, region, accountID) function
- [x] 4.3 Implement buildExecutionARN(workflowID, executionName, region, accountID) function
- [x] 4.4 Add ARN parsing helper for extracting components

## 5. CreateWorkflow Implementation

- [x] 5.1 Implement CreateWorkflow() method in provider.go
- [x] 5.2 Add validation for WorkflowSpec required fields (WorkflowID, Definition)
- [x] 5.3 Validate Definition is valid JSON
- [x] 5.4 Check if state machine exists using DescribeStateMachine (for idempotency)
- [x] 5.5 Call CreateStateMachine API if state machine doesn't exist
- [x] 5.6 Map AWS errors to workflow errors (InvalidDefinition → ErrInvalidSpec)
- [x] 5.7 Build and return CreateWorkflowResult with state machine ARN
- [x] 5.8 Add logging for operation start, success, and failure

## 6. StartExecution Implementation

- [x] 6.1 Implement StartExecution() method in provider.go
- [x] 6.2 Validate ExecutionInput required fields (ExecutionName)
- [x] 6.3 Build state machine ARN from WorkflowID
- [x] 6.4 Call StartExecution API with execution name and input JSON
- [x] 6.5 Map AWS errors (ExecutionAlreadyExists, StateMachineDoesNotExist)
- [x] 6.6 Build and return ExecutionResult with execution ARN and state
- [x] 6.7 Add logging for execution start

## 7. State Mapping

- [x] 7.1 Create `state.go` file with state mapping logic
- [x] 7.2 Implement mapExecutionState(sfnState) function
- [x] 7.3 Map RUNNING → ExecutionStateRunning
- [x] 7.4 Map SUCCEEDED → ExecutionStateSucceeded
- [x] 7.5 Map FAILED/TIMED_OUT/ABORTED → ExecutionStateFailed
- [x] 7.6 Add tests for all state mappings

## 8. GetExecutionStatus Implementation

- [x] 8.1 Implement GetExecutionStatus() method in provider.go
- [x] 8.2 Build execution ARN from WorkflowID and ExecutionID
- [x] 8.3 Call DescribeExecution API
- [x] 8.4 Call GetExecutionHistory API to retrieve events
- [x] 8.5 Parse execution details (state, timestamps, input, output)
- [x] 8.6 Map Step Functions state to workflow.ExecutionState
- [x] 8.7 Convert execution history events to workflow.HistoryEvent format
- [x] 8.8 Build and return ExecutionStatus with all details
- [x] 8.9 Map AWS ExecutionDoesNotExist → workflow.ErrExecutionNotFound

## 9. DeleteWorkflow Implementation

- [x] 9.1 Implement DeleteWorkflow() method in provider.go
- [x] 9.2 Build state machine ARN from WorkflowID
- [x] 9.3 Call DeleteStateMachine API
- [x] 9.4 Map AWS StateMachineDoesNotExist → workflow.ErrWorkflowNotFound
- [x] 9.5 Handle active executions error gracefully
- [x] 9.6 Add logging for deletion operation

## 10. Error Handling

- [x] 10.1 Create `errors.go` file with error mapping utilities
- [x] 10.2 Implement isStateMachineNotFound(err) helper
- [x] 10.3 Implement isExecutionNotFound(err) helper
- [x] 10.4 Implement isInvalidDefinition(err) helper
- [x] 10.5 Implement wrapAWSError(err, operation) helper for context
- [x] 10.6 Add error wrapping for all AWS API calls

## 11. Testing - Unit Tests

- [x] 11.1 Create `internal/workflow/stepfunctions/provider_test.go`
- [x] 11.2 Create mock SFN client using sfniface.ClientAPI
- [x] 11.3 Test Name() returns "step-functions"
- [x] 11.4 Test CreateWorkflow success case
- [x] 11.5 Test CreateWorkflow idempotency (already exists)
- [x] 11.6 Test CreateWorkflow with invalid ASL
- [x] 11.7 Test StartExecution success case
- [x] 11.8 Test StartExecution with non-existent workflow
- [x] 11.9 Test GetExecutionStatus for running execution
- [x] 11.10 Test GetExecutionStatus for completed execution
- [x] 11.11 Test GetExecutionStatus for failed execution
- [x] 11.12 Test GetExecutionStatus for non-existent execution
- [x] 11.13 Test DeleteWorkflow success case
- [x] 11.14 Test DeleteWorkflow for non-existent workflow
- [x] 11.15 Test context cancellation handling

## 12. Testing - State Mapping Tests

- [x] 12.1 Create `state_test.go`
- [x] 12.2 Test RUNNING → ExecutionStateRunning
- [x] 12.3 Test SUCCEEDED → ExecutionStateSucceeded
- [x] 12.4 Test FAILED → ExecutionStateFailed
- [x] 12.5 Test TIMED_OUT → ExecutionStateFailed
- [x] 12.6 Test ABORTED → ExecutionStateFailed

## 13. Testing - ARN Building Tests

- [x] 13.1 Create `arn_test.go`
- [x] 13.2 Test buildStateMachineARN format
- [x] 13.3 Test buildExecutionARN format
- [x] 13.4 Test ARN parsing utilities

## 14. Testing - Error Handling Tests

- [x] 14.1 Create `errors_test.go`
- [x] 14.2 Test error detection helpers (isStateMachineNotFound, etc.)
- [x] 14.3 Test error wrapping preserves original error
- [x] 14.4 Test AWS error mapping to workflow errors

## 15. Integration with Main

- [x] 15.1 Update `cmd/landlord/main.go` to import stepfunctions package
- [x] 15.2 Initialize Step Functions provider with configuration
- [x] 15.3 Register provider in workflow registry
- [x] 15.4 Add conditional registration based on config (only if configured)
- [x] 15.5 Add startup logging for Step Functions provider registration

## 16. Documentation

- [x] 16.1 Add package documentation to provider.go
- [x] 16.2 Document IAM role requirements in comments
- [x] 16.3 Add example ASL definition in comments
- [x] 16.4 Document AWS region configuration options
- [x] 16.5 Update `docs/workflow-providers.md` with Step Functions section

## 17. Validation and Testing

- [x] 17.1 Run `go test -short ./internal/workflow/stepfunctions/...` - all tests pass
- [x] 17.2 Run `go build ./cmd/landlord` - builds successfully
- [x] 17.3 Run linter: `golangci-lint run ./internal/workflow/stepfunctions/`
- [x] 17.4 Verify mock provider still works (backwards compatibility)
- [x] 17.5 Test provider registration in registry

## Acceptance Criteria

- [x] All workflow.Provider interface methods implemented
- [x] State machine creation is idempotent
- [x] Execution status queries work for all states
- [x] ASL validation errors are caught and reported
- [x] AWS SDK errors are mapped to workflow errors
- [x] Context cancellation is respected
- [x] All unit tests pass
- [x] Provider can be registered and used via Manager
- [x] Configuration supports AWS region and IAM role ARN
- [x] Logging covers all operations
