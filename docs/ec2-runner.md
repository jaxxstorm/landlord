# EC2 Runner Compute Provider

The `ec2-runner` provider manages one EC2 instance per tenant. Workflow
providers continue to own lifecycle scheduling, retries, and status; EC2 Runner
only launches, observes, replaces, and terminates the instance.

## Configuration

Operator-owned launch topology is configured in `compute.ec2_runner`:

```yaml
compute:
  ec2_runner:
    region: us-west-2
    instance_type: t3.micro
    subnet_id: subnet-0123456789abcdef0
    security_group_ids:
      - sg-0123456789abcdef0
    instance_profile: LandlordTenantRunner
    key_name: landlord-debug
```

`region`, `instance_type`, `subnet_id`, and `security_group_ids` are required.
The instance profile and key name are optional. Keep subnets, security groups,
and IAM identities operator-owned; tenants select only their AMI and boot-time
user data:

```json
{
  "compute_provider": "ec2-runner",
  "ami": "ami-0123456789abcdef0",
  "user_data": "#!/bin/bash\necho tenant runner started"
}
```

Do not put long-lived secrets in `user_data`. Fetch them at boot from a managed
secret store using the instance profile.

## IAM

The worker or Step Functions Lambda role needs scoped permissions for:

- `ec2:DescribeInstances`
- `ec2:RunInstances`
- `ec2:TerminateInstances`
- `ec2:CreateTags`
- `iam:PassRole` for the configured instance profile

Restrict launch resources to approved AMIs, subnets, security groups, and
instance profiles. The control-plane Step Functions caller role does not need
these EC2 permissions.

## Lifecycle And Ownership

Landlord tags managed instances with `landlord.owner`, `landlord.tenant_id`,
and `landlord.provider`. Provision retries discover the existing non-terminated
instance through those tags instead of launching a duplicate.

Changing `ami` or `user_data` creates a replacement runner. Landlord launches
the replacement before terminating the prior runner, so updates can briefly
consume capacity for two instances. Updating unchanged launch input is a no-op.
Destroy terminates managed non-terminated instances and succeeds if none remain.

To roll back, change the tenant to another configured compute provider or
destroy the tenant. Existing EC2 runners can also be terminated with AWS tools;
do not remove Landlord ownership tags while Landlord is expected to manage them.

## Testing

Run unit and workflow-path coverage without AWS credentials:

```sh
go test ./internal/compute/providers/ec2runner ./internal/workflow/lifecycle
```

Before testing against AWS or an AWS-compatible endpoint, use a dedicated
account or VPC, an approved AMI, a disposable subnet and security group, and a
role scoped to the EC2 actions listed above. Confirm that user data has no
production secrets and terminate test runners after verification. The provider
does not require a separate workflow runner on the instance; use the normal
Landlord workflow path to exercise provision, update, status, and destroy.
