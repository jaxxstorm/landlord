## Why

Landlord needs a concrete compute provider to provision tenant runtime on AWS; ECS is the first supported target but the codebase lacks a completed ECS provider. Shipping an ECS compute provider now unblocks workflow-driven tenant provisioning and creates a foundation that can later expand to cross-account and non-ECS compute targets.

## What Changes

- Add a first-class ECS compute provider that provisions one ECS service per compute instance.
- Wire the workflow agent to call the ECS compute provider when executing tenant plans.
- Define worker inputs for ECS provisioning (task definition ARN, cluster ID, service-level settings).
- Implement AWS credential resolution on the worker using the standard AWS SDK chain, with primitives to support future assume-role flows.
- Update documentation with an ECS compute provider example and configuration guidance.

## Capabilities

### New Capabilities
- `ecs-compute-provider`: Provision and manage tenant compute on AWS ECS as services, including worker inputs and credential resolution.

### Modified Capabilities

## Impact

- New compute provider implementation and interfaces in worker-side compute code.
- Workflow agent integration changes to invoke compute provisioning.
- AWS SDK dependencies and credential-loading behavior in worker runtime.
- Documentation updates covering ECS compute configuration and example usage.
