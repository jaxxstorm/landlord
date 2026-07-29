## Why

Landlord can provision container-oriented compute targets but cannot run a
tenant workload on a dedicated EC2 instance. An EC2 runner lets tenants supply
the machine image and boot-time setup required by their workload while retaining
the existing workflow runner for lifecycle orchestration, retries, and status.

## What Changes

- Add an `ec2-runner` compute provider that creates, observes, updates, and
  destroys one EC2 instance for a tenant lifecycle request.
- Define tenant-facing EC2 runner configuration with a required AMI and user
  data. The workflow runner remains responsible for scheduling and executing
  lifecycle operations; the EC2 provider only manages the instance resource.
- Require deterministic tenant ownership tags and idempotent lookup before
  launching an instance so retries do not create duplicate runners.
- Return the EC2 instance ID and available network addresses as compute result
  data for workflow status, audit, and troubleshooting.
- Surface AWS launch, describe, termination, and invalid-configuration failures
  through the existing compute error path without attempting compensation for a
  successfully launched instance.
- Add AWS SDK seams, configuration validation, documentation, and unit-first
  tests. **BREAKING** only if an operator selects `ec2-runner`; existing compute
  providers and workflow behavior are unchanged.

## Capabilities

### New Capabilities
- `ec2-runner`: Payload-driven, idempotent EC2 instance lifecycle management
  using an AMI and user-data configuration.

### Modified Capabilities
- `compute-config-discovery`: Expose `ec2-runner` configuration validation,
  schema, and defaults alongside existing compute providers.

## Impact

- Adds an AWS EC2 compute-provider adapter, its configuration model, and AWS
  SDK v2 client seams.
- Extends compute provider registration in the control plane and Step Functions
  Lambda bootstrap so both execution targets can select `ec2-runner`.
- Requires scoped EC2 IAM permissions for the runtime identity and network
  defaults controlled by the operator, not the tenant payload.
- Adds unit tests for AMI and user-data validation, idempotent launch, status,
  update, deletion, AWS failures, and context cancellation; AWS-compatible
  integration coverage remains opt-in.
