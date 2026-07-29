## Context

Landlord already routes lifecycle requests through a workflow runner and
delegates resource mutation to registered compute providers. ECS, Docker, and
mock providers implement this boundary, but none launch a dedicated EC2
instance. The new provider must fit the existing synchronous compute interface
without moving workflow scheduling, retry, or tenant-state lookup into EC2
code.

## Goals / Non-Goals

**Goals:**

- Provision one EC2 runner instance per tenant from a tenant-supplied AMI and
  user-data payload.
- Keep tenant input limited to `ami` and `user_data`; keep launch topology and
  identity defaults under operator configuration.
- Make provision, update, status, and destroy safe under workflow retries.
- Surface instance identity and network information through existing compute
  results, tracking, and workflow status.
- Support both existing workers and the Step Functions Lambda target.

**Non-Goals:**

- Run a workflow runner or scheduler on the EC2 instance.
- Add Auto Scaling groups, spot interruption recovery, multi-instance tenants,
  SSH management, AMI construction, or EC2 Fleet support.
- Add a database schema for EC2 resources; AWS tags are the source for resource
  discovery.
- Implement asynchronous readiness signals beyond EC2 instance state and
  available addresses.

## Decisions

### Implement `ec2-runner` as a compute provider

Add `internal/compute/providers/ec2runner` implementing `compute.Provider` and
register it wherever ECS is registered. The workflow runner continues to invoke
the generic lifecycle executor, so Temporal and Step Functions receive the same
behavior. A separate workflow provider was rejected because EC2 only changes
the runtime resource, not lifecycle scheduling semantics.

### Split tenant input from operator launch defaults

The provider-specific tenant `compute_config` contains only required `ami` and
`user_data` fields. `compute.ec2_runner` stores operator-owned defaults such as
region, instance type, subnet IDs, security-group IDs, instance profile, and
optional key name. This prevents tenants from selecting network placement or
AWS identities while still allowing each tenant to select its boot image and
bootstrap script.

### Find instances by deterministic ownership tags

Every instance receives provider, tenant ID, and managed-by tags. Provision
uses `DescribeInstances` with those tags before `RunInstances`; it returns the
existing non-terminated runner on a retry. Destroy terminates matching active
instances and treats an absent or already terminated runner as success. Tags
avoid an additional persistence model and preserve idempotency across process
restarts.

### Update by replacement when the launch input changes

AMI and user data are launch-time settings. An update compares the requested
launch fingerprint with tags on the active instance. If unchanged, it returns a
no-op success. If changed, it launches and verifies a replacement before
terminating the old runner. The workflow runner handles retries around this
boundary; no attempt is made to migrate application state between instances.

### Use AWS SDK v2 seams and standard credential loading

Use `internal/cloud/awsconfig` and a narrow EC2 API interface for unit tests.
The provider loads credentials from the standard chain and optional
operator-configured assume role. This matches ECS and enables Lambda execution
roles without static credentials.

## Risks / Trade-offs

- [A retry occurs after AWS accepts RunInstances but before the response is
  persisted] -> deterministic tags and pre-launch discovery return the active
  runner rather than creating another one.
- [Replacement temporarily runs two instances] -> launch and verify the
  replacement before terminating the old runner; document transient duplicate
  capacity and cost.
- [User data contains secrets] -> treat it as sensitive launch input, do not log
  it, and recommend fetching secrets at boot from a managed secret store.
- [An instance becomes unhealthy after EC2 reports running] -> expose EC2 state
  only; workload-level health checks remain a future runner contract.
- [Tag-based discovery returns unexpected manually tagged instances] -> use a
  reserved managed-by tag and document that operators must not reuse it.

## Migration Plan

1. Deploy code with the provider disabled by default.
2. Grant the worker and Lambda roles scoped EC2, tag, and describe permissions.
3. Add `compute.ec2_runner` defaults and enable the provider in a staging
   environment.
4. Submit a tenant request with `compute_provider: ec2-runner`, an AMI, and user
   data; verify launch, status, update replacement, and destroy.
5. Roll back by selecting another compute provider for new revisions. Existing
   EC2 runners remain manageable through Landlord destroy operations or AWS
   tooling.

## Open Questions

- Should update replacement wait for a future workload-level readiness signal
  before terminating the previous instance?
- Which operator defaults must be mandatory for the first release: one subnet,
  one or more security groups, and an instance profile, or only instance type?
