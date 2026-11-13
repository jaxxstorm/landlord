## ADDED Requirements

### Requirement: Floci integration tests are opt-in and isolated
The system SHALL provide Floci-backed integration tests behind an `integration`
build tag. Tests SHALL start an isolated Floci instance with Docker access,
static test credentials, and a unique account or resource prefix, and SHALL not
run in the default Go test suite.

#### Scenario: Default test suite
- **WHEN** a developer runs `go test ./...` without integration tags
- **THEN** the suite SHALL not require Docker, Floci, AWS credentials, or an EKS cluster

#### Scenario: Integration test invocation
- **WHEN** a developer runs the Floci integration suite on a Docker-capable host
- **THEN** the fixture SHALL start Floci and configure AWS-compatible clients with its endpoint and test credentials
- **AND** the fixture SHALL remove test resources and stop Floci after the suite

### Requirement: Floci validates Step Functions and EC2 runner lifecycle behavior
The integration suite SHALL exercise the Step Functions provider for execution
start, status/history retrieval, duplicate handling, and stop behavior, and
shall exercise EC2 runner provision, status, replacement, and destroy behavior.

#### Scenario: Stop Step Functions execution
- **WHEN** the suite stops an active Floci Step Functions execution
- **THEN** subsequent status retrieval SHALL report a terminal state without creating new lifecycle work

#### Scenario: EC2 runner replacement
- **WHEN** an EC2 runner update changes AMI or user data in the Floci environment
- **THEN** the provider SHALL launch a replacement before terminating the prior managed runner
- **AND** the suite SHALL observe the replacement instance identity

### Requirement: Floci validates the complete EC2 tenant lifecycle
The integration suite SHALL provide an opt-in Docker-based environment with
Landlord's API, PostgreSQL, lifecycle worker, and Floci to exercise an EC2
tenant lifecycle through the public API.

#### Scenario: End-to-end tenant lifecycle
- **WHEN** the suite submits an EC2 tenant lifecycle request to the Landlord API
- **THEN** the lifecycle worker SHALL complete the request through Floci-backed Step Functions and EC2
- **AND** the suite SHALL observe the terminal API result and cleanup of the managed runner
