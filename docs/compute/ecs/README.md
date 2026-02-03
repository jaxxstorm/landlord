# ECS Compute Provider

The ECS compute provider provisions tenant workloads as ECS services. Each tenant maps to one ECS service using a task definition you supply.

## Required defaults

Landlord requires a default ECS configuration at startup. Set `compute.ecs` in your config file with at least `cluster_arn`, `task_definition_arn`, and either `service_name` or `service_name_prefix`. Tenant `compute_config` values are merged on top of these defaults.

## Tenant compute_config reference

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `cluster_arn` | string | yes | ECS cluster ARN |
| `task_definition_arn` | string | yes | ECS task definition ARN |
| `service_name` | string | conditional | Service name (required if `service_name_prefix` is not set) |
| `service_name_prefix` | string | conditional | Service name prefix (required if `service_name` is not set) |
| `region` | string | no | AWS region override for ECS calls |
| `desired_count` | integer | no | Desired task count (>= 0) |
| `launch_type` | string | no | Launch type (`EC2`, `FARGATE`, `EXTERNAL`) |
| `subnets` | array<string> | no | Subnet IDs for awsvpc networking |
| `security_groups` | array<string> | no | Security group IDs for awsvpc networking |
| `assign_public_ip` | boolean | no | Assign public IP for awsvpc networking |
| `tags` | object<string,string> | no | Tags applied to ECS service |
| `assume_role` | object | no | Assume-role configuration (see below) |

### `assume_role` fields

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `role_arn` | string | yes | Role ARN to assume |
| `external_id` | string | no | External ID for assume role |
| `session_name` | string | no | Session name for assume role |

## Full JSON example

```json
{
  "cluster_arn": "arn:aws:ecs:us-west-2:123456789012:cluster/landlord",
  "task_definition_arn": "arn:aws:ecs:us-west-2:123456789012:task-definition/tenant-app:12",
  "service_name_prefix": "landlord-tenant-",
  "region": "us-west-2",
  "desired_count": 1,
  "launch_type": "FARGATE",
  "subnets": ["subnet-123", "subnet-456"],
  "security_groups": ["sg-123"],
  "assign_public_ip": true,
  "tags": {
    "environment": "dev"
  },
  "assume_role": {
    "role_arn": "arn:aws:iam::123456789012:role/landlord-ecs",
    "external_id": "tenant-provisioning",
    "session_name": "landlord-worker"
  }
}
```

## Full YAML example

```yaml
cluster_arn: "arn:aws:ecs:us-west-2:123456789012:cluster/landlord"
task_definition_arn: "arn:aws:ecs:us-west-2:123456789012:task-definition/tenant-app:12"
service_name_prefix: "landlord-tenant-"
region: "us-west-2"
desired_count: 1
launch_type: "FARGATE"
subnets:
  - "subnet-123"
  - "subnet-456"
security_groups:
  - "sg-123"
assign_public_ip: true
tags:
  environment: "dev"
assume_role:
  role_arn: "arn:aws:iam::123456789012:role/landlord-ecs"
  external_id: "tenant-provisioning"
  session_name: "landlord-worker"
```

### Using file:// with the CLI

```bash
go run ./cmd/cli create --tenant-name demo \
  --config file:///path/to/ecs-compute-config.yaml
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
