## ADDED Requirements

### Requirement: EC2 runner configuration accepts an AMI and user data
The system SHALL provide an `ec2-runner` compute provider whose tenant
`compute_config` requires non-empty `ami` and `user_data` string fields. The
provider SHALL validate this configuration before persistence or mutation and
SHALL not log user-data contents.

#### Scenario: Valid EC2 runner request
- **WHEN** a lifecycle request selects `ec2-runner` with a valid AMI and user data
- **THEN** the provider SHALL accept the configuration and use both values for the instance launch request

#### Scenario: Missing launch input
- **WHEN** an EC2 runner request omits or leaves empty the AMI or user data
- **THEN** the provider SHALL return a validation error identifying the missing field
- **AND** it SHALL not call AWS to launch an instance

### Requirement: EC2 runner provisioning is idempotent
The provider SHALL create at most one non-terminated EC2 runner per tenant and
SHALL tag each managed instance with deterministic provider and tenant
ownership data before returning success.

#### Scenario: First provision request
- **WHEN** a lifecycle request provisions a tenant with no managed EC2 runner
- **THEN** the provider SHALL launch one instance using the requested AMI and user data plus operator-configured launch defaults
- **AND** it SHALL return the instance ID and available network addresses in the provision result

#### Scenario: Repeated provision request
- **WHEN** a workflow retry provisions a tenant that already has a non-terminated managed EC2 runner
- **THEN** the provider SHALL return the existing runner result
- **AND** it SHALL not launch another instance

### Requirement: EC2 runner lifecycle operations preserve workflow safety
The provider SHALL expose EC2 instance status, perform a no-op update for an
unchanged launch input, replace a runner only when launch input changes, and
terminate managed runners idempotently on destroy.

#### Scenario: Status query
- **WHEN** the workflow runner queries an existing EC2 runner
- **THEN** the provider SHALL return the mapped EC2 lifecycle status, instance ID, and available addresses

#### Scenario: Update with changed AMI or user data
- **WHEN** a lifecycle update changes the requested AMI or user data
- **THEN** the provider SHALL launch and verify a replacement runner before terminating the prior managed runner
- **AND** it SHALL return the replacement instance identity

#### Scenario: Destroy request
- **WHEN** a lifecycle request destroys a tenant with managed EC2 runners
- **THEN** the provider SHALL terminate the managed non-terminated instances
- **AND** it SHALL succeed when no managed runner exists or it is already terminated

### Requirement: EC2 runner failures are actionable and auditable
The provider SHALL map AWS API errors to the existing compute error path,
preserve context cancellation, and record instance identifiers in compute
results and tracking without exposing user-data contents.

#### Scenario: AWS launch failure
- **WHEN** AWS rejects an instance launch
- **THEN** the provider SHALL return an actionable compute error
- **AND** it SHALL not report provisioning success

#### Scenario: Cancelled lifecycle request
- **WHEN** the workflow context is cancelled before an EC2 mutation completes
- **THEN** the provider SHALL return the context error
- **AND** the workflow runner SHALL retain its existing retry and status behavior
