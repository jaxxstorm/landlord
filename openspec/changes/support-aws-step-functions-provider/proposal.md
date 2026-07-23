## Why

Landlord registers `step-functions` as a workflow provider, but its implementation only simulates workflow operations and leaves reconciled tenants permanently running. A production AWS-native execution path is needed so tenant lifecycle work can use durable Step Functions executions while preserving Landlord's existing idempotent reconciliation and compute-provider abstractions.

## What Changes

- Replace the simulated Step Functions provider with an AWS SDK v2 implementation for starting, querying, stopping, and reporting Standard workflow executions.
- Route all tenant lifecycle operations through one configured, pre-provisioned state machine instead of constructing a state machine identifier per tenant.
- Add a deployable Lambda lifecycle executor and state-machine definition that run provision, update, and delete operations from the execution payload using shared compute-provider logic.
- Define deterministic execution naming, duplicate-start recovery, cancellation behavior, terminal error/output reporting, and execution-history retrieval.
- **BREAKING** Require a configured Step Functions state-machine ARN when `workflow.default_provider` is `step-functions`; distinguish Landlord's caller credentials from the state machine's execution role.
- Document least-privilege IAM roles and add unit-first coverage plus AWS-compatible integration coverage.

## Capabilities

### New Capabilities
- `step-functions-lifecycle-executor`: Lambda-backed lifecycle execution and Amazon States Language orchestration for payload-driven provision, update, and delete operations.

### Modified Capabilities
- `step-functions-provider`: Real AWS Step Functions provider lifecycle, configuration, idempotency, status, error, and history behavior.
- `workflow-provisioning`: Deterministic, retry-safe Step Functions invocation through a configured shared state machine.
- `workflow-execution-status`: Step Functions terminal status, error, output, and history mapping for reconciler consumption.
- `workflow-worker-engines`: Shared payload-driven lifecycle execution usable by both existing workers and the Lambda executor.

## Impact

- Affects `internal/workflow/providers/stepfunctions`, workflow configuration and startup registration, controller workflow selection, and compute lifecycle execution code.
- Adds a Lambda handler, ASL definition, deployment/IAM documentation, and AWS SDK test seams.
- Requires AWS credentials for the control plane, a Standard Step Functions state machine, and separate state-machine and Lambda execution permissions in AWS deployments.
