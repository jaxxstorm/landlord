## 1. Configuration And AWS Seams

- [x] 1.1 Add the AWS EC2 SDK dependency and a narrow injectable EC2 client interface.
- [x] 1.2 Add `compute.ec2_runner` operator configuration for region, instance type, subnet, security groups, instance profile, optional key name, and optional assume role.
- [x] 1.3 Validate EC2 runner operator defaults and register `ec2-runner` as an enabled compute provider.
- [x] 1.4 Define and test the tenant `compute_config` JSON Schema requiring `ami` and `user_data` without exposing operator launch defaults.

## 2. EC2 Runner Provider

- [x] 2.1 Implement provider construction with standard AWS credential loading and test-injectable EC2 clients.
- [x] 2.2 Implement launch configuration parsing, validation, launch fingerprinting, and reserved managed-instance tags without logging user data.
- [x] 2.3 Implement idempotent provision using tag-based instance discovery and `RunInstances`.
- [x] 2.4 Implement status lookup with EC2 state mapping, instance identifiers, and available network addresses.
- [x] 2.5 Implement update as no-op for an unchanged launch fingerprint and replacement for changed AMI or user data.
- [x] 2.6 Implement idempotent destroy for managed non-terminated instances and missing runners.
- [x] 2.7 Map AWS API failures and context cancellation through existing compute error conventions.

## 3. Runtime Integration

- [x] 3.1 Register the EC2 runner in the Landlord control-plane bootstrap.
- [x] 3.2 Register the EC2 runner in the Step Functions Lambda bootstrap using the same configuration and credential behavior.
- [x] 3.3 Ensure shared lifecycle execution preserves EC2 runner resource identifiers and tracked workflow execution IDs.

## 4. Verification And Documentation

- [x] 4.1 Add provider unit tests for configuration validation, launch input, tagging, duplicate provision, status, update replacement, destroy, AWS errors, and context cancellation.
- [x] 4.2 Add configuration-discovery and bootstrap registration tests for `ec2-runner`.
- [x] 4.3 Add an end-to-end lifecycle test with mocked EC2 and workflow execution paths.
- [x] 4.4 Document EC2 runner configuration, IAM permissions, tagging ownership, user-data handling, update replacement behavior, and rollback.
- [x] 4.5 Run focused and full Go test suites, formatting, and OpenSpec validation; document AWS-compatible integration prerequisites.
