# ECS Compute Provider

The ECS compute provider provisions tenant workloads as ECS services. Each tenant maps to one ECS service using a task definition you supply.

## Required defaults

Landlord requires a default ECS configuration at startup. Set `compute.ecs` in your config file with at least `cluster_arn`, `task_definition_arn`, and either `service_name` or `service_name_prefix`. Tenant `compute_config` values are merged on top of these defaults.

## Tenant compute_config example

```json
{
  "cluster_arn": "arn:aws:ecs:us-west-2:123456789012:cluster/landlord",
  "task_definition_arn": "arn:aws:ecs:us-west-2:123456789012:task-definition/tenant-app:12",
  "desired_count": 1,
  "launch_type": "FARGATE",
  "subnets": ["subnet-123", "subnet-456"],
  "security_groups": ["sg-123"],
  "assign_public_ip": true,
  "service_name_prefix": "landlord-tenant-",
  "tags": {
    "environment": "dev"
  }
}
```

## Credential resolution

The worker uses the AWS SDK default credential chain:

- Environment variables (e.g., `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
- Shared config/credentials files and AWS SSO
- ECS/EC2 metadata credentials

To use an assumed role, include an optional `assume_role` block in `compute_config`:

```json
{
  "cluster_arn": "arn:aws:ecs:us-west-2:123456789012:cluster/landlord",
  "task_definition_arn": "arn:aws:ecs:us-west-2:123456789012:task-definition/tenant-app:12",
  "service_name_prefix": "landlord-tenant-",
  "assume_role": {
    "role_arn": "arn:aws:iam::123456789012:role/landlord-ecs",
    "external_id": "tenant-provisioning",
    "session_name": "landlord-worker"
  }
}
```

The ECS provider never requires AWS credentials on the API server; credentials are only used on the worker when it makes ECS calls.
