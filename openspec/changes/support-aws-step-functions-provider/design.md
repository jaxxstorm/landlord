## Context

The existing `step-functions` provider is registered by the control plane but does not create AWS clients or make AWS calls. It manufactures ARNs with an empty account ID, ignores caller-supplied execution names, returns running for every status query, and cannot execute lifecycle work. The controller currently uses a tenant-specific workflow ID, but it never calls `CreateWorkflow`; a production implementation must therefore use an existing state machine rather than create one per tenant.

The repository already has payload-driven lifecycle logic in the Temporal worker and AWS credential loading in `internal/cloud/awsconfig`. Compute operations are synchronous from the workflow worker's perspective and the compute manager emits provider callbacks after tracked operations complete.

## Goals / Non-Goals

**Goals:**

- Execute tenant provision, update, and delete operations through AWS Step Functions Standard workflows.
- Preserve reconciliation idempotency, status polling, audit data, and provider-agnostic compute selection.
- Run lifecycle operations in a deployable Lambda target without requiring the Lambda to query tenant state.
- Expose actionable AWS failures, execution output, terminal state, and history through the existing workflow provider interface.
- Document roles, required permissions, deployment inputs, and local/AWS-compatible test paths.

**Non-Goals:**

- Create one state machine per tenant or dynamically author workflow definitions from the controller.
- Support Express state machines, cross-region failover, a workflow authoring DSL, or an execution UI.
- Make a stopped execution compensate for a compute change that already completed.
- Replace Restate or Temporal execution paths.

## Decisions

### Use one pre-provisioned Standard state machine

`workflow.step_functions.state_machine_arn` SHALL identify the environment's lifecycle state machine. The provider starts that machine for every tenant operation and treats the controller's tenant-specific workflow ID as an internal logical identifier, not an AWS resource name.

This avoids unbounded state-machine creation, matches the controller's absence of `CreateWorkflow` calls, and makes deployment ownership explicit. Standard workflows are selected because the reconciler needs stop, describe, output, and execution-history APIs. Dynamic state-machine management was considered but rejected because it would need a definition lifecycle, name ownership, quotas, and an additional deployment API that the control plane does not use.

### Invoke a Lambda lifecycle executor from ASL

The deployed ASL definition SHALL call a single Lambda handler synchronously and pass the full `workflow.ProvisionRequest` plus `$$.Execution.Id`. The handler SHALL dispatch provision, update, and delete through a new shared lifecycle executor extracted from the Temporal worker. The executor SHALL retain payload-first compute-provider resolution and tracked compute calls.

Lambda keeps the state machine serverless and isolates AWS compute credentials from the API process. Direct ASL ECS integrations were rejected because they would duplicate Landlord's compute validation, provider selection, tracking, and future provider support. Step Functions Activities were rejected because they require a permanently polling worker fleet and add task-token lifecycle infrastructure.

### Use explicit, deterministic execution names

The provider SHALL derive a valid, at-most-80-character execution name from the tenant UUID, normalized operation, and a stable desired-state revision hash. Repeated controller triggers for the same logical revision use the same name and input. On `ExecutionAlreadyExists`, the provider SHALL resolve and return the existing execution when it belongs to the same logical request; it SHALL not start a second execution.

Standard workflows retain execution names for 90 days. A changed desired-state hash creates a new logical revision. Failed executions with unchanged desired state remain terminal and are surfaced to reconciliation rather than silently being re-run under a different name. The controller's existing config-change retry path creates a new revision when desired state changes.

### Separate AWS caller identity from execution roles

The provider SHALL load API caller credentials through `internal/cloud/awsconfig`, including optional assume-role support. This identity needs Step Functions control-plane permissions. The configured state-machine ARN replaces the current overloaded `role_arn` setting. The state-machine execution role and Lambda execution role are deployment inputs, documented separately, and are not assumed by the provider unless an explicit caller-assume-role option is configured.

This follows the existing ECS credential pattern and prevents granting the control plane the workload's compute permissions.

### Use AWS response data as the execution record

`StartExecution` returns AWS's execution ARN. `GetExecutionStatus` calls `DescribeExecution`, maps `RUNNING`, `SUCCEEDED`, `FAILED`, `TIMED_OUT`, and `ABORTED`, and includes input, output, stop time, and terminal error details. It also retrieves paginated history for diagnostics. `StopExecution` calls the AWS stop API and is idempotent when the execution is already terminal or absent after a prior stop.

The current `ExecutionStatus` model is sufficient; no persistence migration is required because the tenant repository already persists the returned execution ARN and reconciler state fields.

### Cancellation is cooperative at lifecycle boundaries

Stopping an AWS execution prevents subsequent ASL states but cannot forcibly undo an already-issued ECS or other cloud-provider request. The ASL input includes the execution ARN. Before starting a mutable lifecycle operation, and between operation boundaries, the Lambda executor checks that the execution remains running. It returns a cancelled result without issuing further work when it has been stopped. Providers continue to make individual compute operations idempotent.

This gives safe cancellation boundaries without claiming transactional rollback of external cloud side effects.

## Risks / Trade-offs

- [Lambda invocation exceeds its runtime limit] → Keep the initial handler to bounded lifecycle API calls; return tracked operation results and use a callback-token design in a later change if compute becomes long-running.
- [A stop races with a cloud mutation] → Check execution status before each mutation, make mutations idempotent, and record the execution ARN in logs and compute tracking.
- [Duplicate execution cannot be identified from `ExecutionAlreadyExists`] → Use deterministic names and exact input; resolve the matching execution with the configured state-machine ARN and reject mismatched inputs.
- [AWS permissions or networking are misconfigured] → Validate configuration at startup, surface wrapped AWS errors, document roles, and cover the API boundary with mock-client tests.
- [Execution history is large] → Page history and bound retained events returned to callers; use CloudWatch as the complete operational log.
- [Lambda and worker logic drift] → Extract one lifecycle executor with tests and have Temporal and Lambda adapt it rather than duplicate operation branches.

## Migration Plan

1. Add state-machine ARN and caller credential configuration while retaining the provider name `step-functions`.
2. Deploy the ASL definition, Lambda image/function, state-machine role, Lambda role, logs, and controller role before enabling the provider.
3. Deploy the control plane and Lambda handler with `step-functions` configured in a staging environment; verify provision, update, delete, duplicate start, status, and stop behavior.
4. Enable `workflow.default_provider: step-functions` only after the configured ARN is reachable and validation passes.
5. Roll back by selecting Restate, Temporal, or mock as the default provider. Existing running Step Functions executions remain observable and can be stopped with AWS tooling; no database schema rollback is required.

## Open Questions

- Should deployment artifacts be an AWS SAM template, Terraform module, or documentation-only reference for the first implementation? This change will provide a versioned reference template; a repository-wide IaC standard can replace it later.
- What bounded number of history events should the API retain in `ExecutionStatus` before directing operators to CloudWatch?
- Should a later asynchronous compute provider use Step Functions callback tokens instead of the synchronous Lambda result path?
