## Context

The landlord control plane uses a pluggable workflow provider architecture to abstract workflow orchestration engines. The workflow.Provider interface defines operations for workflow lifecycle management (create, execute, query, delete). A mock provider exists for testing, but a production implementation is needed.

AWS Step Functions is the initial target workflow engine for production deployments. Step Functions provides durable, serverless workflow orchestration with built-in retry logic, error handling, and visual workflow monitoring. It integrates naturally with AWS Lambda and ECS for executing tenant provisioning tasks.

**Current State:**
- workflow.Provider interface defined in `internal/workflow/types.go`
- Mock provider exists for unit tests
- Registry and Manager components ready to use providers
- No production workflow provider implemented

**Constraints:**
- Must implement all methods of workflow.Provider interface
- Must handle AWS Step Functions ASL (Amazon States Language) as workflow definition format
- Must support IAM role-based execution permissions
- Must provide idempotent operations (create, execute)
- Must translate Step Functions execution states to workflow.ExecutionState enum

## Goals / Non-Goals

**Goals:**
- Implement complete AWS Step Functions provider satisfying workflow.Provider interface
- Support idempotent workflow creation and execution
- Validate ASL definition syntax before creating state machines
- Map Step Functions execution states to workflow execution states
- Support execution history retrieval with event details
- Enable configuration of AWS region and IAM execution role
- Provide comprehensive test coverage with AWS SDK mocks

**Non-Goals:**
- Lambda function implementation (workflows reference existing functions)
- Step Functions state machine authoring tools or DSL
- Execution monitoring UI or dashboard
- Cross-region replication or high availability setup
- Cost optimization or execution metrics collection
- Support for Express workflows (Standard workflows only)

## Decisions

### Decision 1: In-tree provider vs separate module

**Choice:** Implement as in-tree provider in `internal/workflow/stepfunctions/`

**Rationale:**
- Step Functions is the primary production target, not an optional plugin
- Simpler dependency management and versioning
- Easier to maintain interface compatibility
- Can be extracted later if plugin architecture becomes beneficial

**Alternatives Considered:**
- External plugin loaded at runtime: Adds complexity, harder to version
- Separate Go module: Increases maintenance burden for core functionality

### Decision 2: AWS SDK v2 vs v1

**Choice:** Use AWS SDK for Go v2 (`github.com/aws/aws-sdk-go-v2`)

**Rationale:**
- SDK v2 is the current stable version with active support
- Better context integration for cancellation and timeouts
- More idiomatic Go patterns (uses standard library context)
- Modular package structure (only import SFN service)

**Alternatives Considered:**
- SDK v1: Legacy version, less idiomatic, larger dependency footprint

### Decision 3: Definition validation strategy

**Choice:** Basic JSON syntax validation + Step Functions API validation on create

**Rationale:**
- Step Functions API provides comprehensive ASL validation on CreateStateMachine
- Prevents duplicating complex ASL validation logic
- Faster failure feedback (create fails fast with clear errors)
- Ensures definition is actually executable by AWS

**Alternatives Considered:**
- Full ASL parser in Go: Complex to maintain, duplicates AWS logic
- No validation: Poor error messages, failures happen during execution

### Decision 4: Idempotency implementation

**Choice:** Use DescribeStateMachine to check existence before CreateStateMachine

**Rationale:**
- Step Functions CreateStateMachine fails if state machine already exists
- DescribeStateMachine is idempotent and returns full state machine details
- Allows returning success for existing state machines (true idempotency)
- Consistent with Provider interface contract

**Implementation:**
```go
// CreateWorkflow checks if state machine exists first
existingArn := buildStateMachineArn(spec.WorkflowID)
_, err := client.DescribeStateMachine(ctx, &sfn.DescribeStateMachineInput{
    StateMachineArn: aws.String(existingArn),
})
if err == nil {
    // State machine exists, return success
    return buildResultFromExisting(existingArn), nil
}
// State machine doesn't exist, create it
```

### Decision 5: Execution naming and uniqueness

**Choice:** Require caller to provide unique execution names via ExecutionInput.ExecutionName

**Rationale:**
- Matches Step Functions execution model (names must be unique per state machine)
- Gives caller control over idempotency (same name = idempotent)
- Prevents accidental duplicate executions
- Caller can implement retry logic with same name

**Alternatives Considered:**
- Generate random execution names: Loses idempotency, harder to track
- Use WorkflowID + timestamp: Not idempotent, may duplicate on retries

### Decision 6: State mapping strategy

**Choice:** Map Step Functions execution states directly to workflow.ExecutionState

**Mapping:**
- `RUNNING` → `workflow.ExecutionStateRunning`
- `SUCCEEDED` → `workflow.ExecutionStateSucceeded`
- `FAILED` → `workflow.ExecutionStateFailed`
- `TIMED_OUT` → `workflow.ExecutionStateFailed`
- `ABORTED` → `workflow.ExecutionStateFailed`

**Rationale:**
- Simple 1:1 mapping for most states
- Failed/TimedOut/Aborted all represent terminal failures
- Matches semantic meaning of workflow execution states

### Decision 7: IAM role configuration

**Choice:** Require IAM role ARN in configuration, use existing roles

**Rationale:**
- Step Functions requires IAM role for execution permissions
- Role management is outside control plane scope (infrastructure concern)
- Allows operators to control permissions via IAM policies
- Follows AWS best practices (least privilege via roles)

**Configuration:**
```yaml
workflow:
  provider: step-functions
  step_functions:
    region: us-west-2
    role_arn: arn:aws:iam::123456789012:role/LandlordStepFunctionsRole
```

### Decision 8: Testing strategy

**Choice:** Use AWS SDK mock interfaces (`sfniface.ClientAPI`) for unit tests

**Rationale:**
- No dependency on real AWS account for tests
- Fast test execution without network calls
- Full control over API responses for edge cases
- Standard AWS SDK testing pattern

**Alternatives Considered:**
- LocalStack: Adds container dependency, slower, less reliable
- Integration tests only: Slow, requires AWS credentials, harder to test edge cases

### Decision 9: Error handling approach

**Choice:** Wrap AWS SDK errors with context, map to standard workflow errors

**Implementation:**
```go
if err != nil {
    if isStateMachineNotFound(err) {
        return nil, workflow.ErrWorkflowNotFound
    }
    return nil, fmt.Errorf("step functions create failed: %w", err)
}
```

**Rationale:**
- Preserves original error for debugging
- Maps to standard workflow error types for consistency
- Provides context about which operation failed

## Risks / Trade-offs

### Risk: AWS API rate limits
**Mitigation:** Use exponential backoff (built into AWS SDK), implement caching for DescribeStateMachine calls

### Risk: IAM permission errors
**Mitigation:** Document required IAM permissions clearly, provide validation helper that checks permissions on startup

### Risk: ASL definition errors not caught until execution
**Mitigation:** CreateStateMachine validates ASL syntax, execution failures surface quickly with clear error messages

### Risk: Step Functions cost at scale
**Trade-off:** Accept higher cost for operational simplicity. Standard workflows cost $0.025 per 1000 state transitions. For <100k executions/month, cost is minimal (<$100/month).

### Risk: AWS SDK version compatibility
**Mitigation:** Pin to specific SDK v2 version, test upgrades in staging before production

### Risk: Execution history size limits
**Trade-off:** Step Functions limits execution history to 25,000 events. For long-running workflows, history may be truncated. Document this limitation, recommend workflow design with bounded event counts.

## Migration Plan

**Not applicable** - This is a new provider implementation, no migration from existing system.

**Deployment Steps:**
1. Merge provider implementation
2. Add AWS SDK v2 dependencies to go.mod
3. Configure AWS credentials and region in deployment environment
4. Create IAM role with required Step Functions permissions
5. Update main.go to register Step Functions provider
6. Deploy and verify with test workflow
7. Monitor executions in AWS Step Functions console

**Rollback Strategy:**
- Remove Step Functions provider registration from main.go
- Fall back to mock provider for testing (not suitable for production)
- No data loss risk (workflows are stateless definitions)

## Open Questions

**Q: Should we support Step Functions Express workflows?**
- Express workflows are cheaper but have different execution model (no history)
- Decision: Start with Standard workflows only, add Express support if needed
- Status: Deferred

**Q: How to handle Step Functions service quotas (1000 state machines per region)?**
- Decision: Document quota limit, recommend prefix-based naming conventions
- Status: Deferred (not expected to hit limit initially)

**Q: Should we cache DescribeStateMachine responses?**
- Decision: Start without caching, add if DescribeStateMachine becomes bottleneck
- Status: Deferred
