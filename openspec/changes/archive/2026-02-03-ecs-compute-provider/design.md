## Context

Landlord currently has compute provider interfaces, registry, and workflow-to-compute integration, but no ECS provider implementation. The workflow worker (e.g., Restate) builds a `compute.TenantComputeSpec` from the tenant desired config and invokes compute providers via the registry. The API server also relies on the active compute provider to validate and expose `compute_config` schema, so the ECS provider must validate without requiring AWS access.

## Goals / Non-Goals

**Goals:**
- Implement an ECS compute provider that provisions one ECS service per tenant compute instance.
- Accept ECS-specific inputs via provider config (task definition ARN, cluster ID/ARN, service options).
- Integrate the provider into the workflow worker compute registry so workflows can invoke it.
- Use the AWS SDK default credential chain on the worker, with primitives for future assume-role support.
- Keep provider configuration and credential helpers reusable for future compute providers.

**Non-Goals:**
- Cross-account assume-role provisioning in this change (only primitives and configuration shape).
- Generating or managing ECS task definitions from `ContainerSpec`.
- Multi-region provisioning, autoscaling policies, or advanced networking beyond required ECS fields.

## Decisions

- **Provider location and interface:** Add `internal/compute/providers/ecs` implementing `compute.Provider`, mirroring patterns from the Docker provider (validation, schema, defaults, and idempotent operations).
- **ECS config surface:** Introduce an ECS-specific provider config payload (JSON) with required `cluster_arn` and `task_definition_arn`, plus optional `service_name`, `desired_count`, `launch_type`, `subnets`, `security_groups`, `assign_public_ip`, and tags/labels passthrough. This config lives in `TenantComputeSpec.ProviderConfig` and is validated via JSON Schema + `ValidateConfig`.
- **Mapping to ECS resources:** Provision = create (or ensure) a service named deterministically from `tenant_id` (e.g., `landlord-tenant-<id>`). Update = update service desired count and task definition if changed. Destroy = delete service and wait for stabilization (best-effort, idempotent).
- **Credential handling:** Add a small AWS config helper (e.g., `internal/cloud/awsconfig`) that builds an `aws.Config` using the SDK default chain (env, shared config/SSO, EC2/ECS metadata). Include optional assume-role fields in the provider config (role ARN, external ID, session name) but only use them when provided.
- **Worker vs server responsibility:** The API server uses the provider only for schema/config validation (no AWS calls). AWS SDK clients are initialized lazily during Provision/Update/Destroy on the worker to avoid requiring credentials on the server.
- **ContainerSpec usage:** ECS provider ignores `ContainerSpec` content for provisioning because the task definition already encapsulates containers. Validation will allow a single container (as produced by the workflow worker) but will not require it to match the task definition.

## Risks / Trade-offs

- [ECS API latency and eventual consistency] → Mitigation: treat operations as idempotent, return `in_progress` where appropriate, and rely on workflow callbacks for completion tracking.
- [Service name collisions across tenants/environments] → Mitigation: include tenant ID and an optional prefix in the provider config; document required uniqueness.
- [Task definition/cluster mismatches] → Mitigation: validate required fields and add targeted AWS error mapping to compute errors to aid troubleshooting.
- [Credential resolution surprises on worker hosts] → Mitigation: document credential chain behavior and allow explicit role configuration when needed.

## Migration Plan

- Register the ECS provider in the workflow worker compute registry and update worker configuration defaults/examples.
- Deploy worker changes first (so workflows can call the provider), then roll out server docs/config changes.
- No data migrations required; existing tenants unaffected until they opt into `ecs` as compute provider.

## Open Questions

- Should ECS support both Fargate and EC2 launch types, and what defaults should we choose?
- Do we need to surface capacity provider strategy or service discovery settings in provider config?
- Should we validate that the task definition is compatible with requested networking options (subnets/security groups)?
