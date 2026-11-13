## 1. Floci Test Infrastructure

- [x] 1.1 Add the Go dependencies and shared `integration`-build-tag fixture that starts an isolated Floci container with Docker access.
- [x] 1.2 Configure the fixture with static AWS credentials, a unique test account/resource prefix, endpoint overrides, bounded readiness polling, and cleanup.
- [x] 1.3 Add and document a local Floci integration-test target, including Docker prerequisites, without affecting `go test ./...`.
- [x] 1.4 Add a Docker-capable CI job that runs the Floci integration-test target.

## 2. Step Functions And EC2 Coverage

- [x] 2.1 Extend Step Functions integration coverage to create disposable Floci state-machine resources and verify start, status/history, duplicate, and stop behavior.
- [x] 2.2 Add EC2 runner integration coverage that verifies tagged provision, status, launch-input replacement, and idempotent destroy against Floci EC2.
- [x] 2.3 Ensure integration cleanup removes all created Step Functions and EC2 resources even after test assertion failures.

## 3. Kubernetes Compute Runner

- [x] 3.1 Add Kubernetes provider configuration, validation, defaults, and registration with the compute provider registry.
- [x] 3.2 Implement Kubernetes API client construction and tenant namespace/Deployment naming, labels, and ownership handling.
- [x] 3.3 Implement and unit-test Kubernetes runner provision, status, update, and idempotent destroy operations.
- [x] 3.4 Add Floci EKS integration coverage that creates an ephemeral cluster, obtains scoped Kubernetes access, and validates the Kubernetes runner lifecycle.

## 4. Verification

- [x] 4.1 Run focused unit tests for the new Kubernetes provider and existing Step Functions and EC2 runner packages.
- [x] 4.2 Run `go test ./...` without integration tags to verify Docker-free default test behavior.
- [x] 4.3 Run the Floci integration suite on a Docker-capable host and record any unsupported Floci behavior in follow-up work.

## 5. EC2 End-To-End Coverage

- [x] 5.1 Add a Docker-based Floci test stack containing the Landlord API, PostgreSQL, and lifecycle worker, with disposable Step Functions and EC2 bootstrap resources.
- [x] 5.2 Add an opt-in EC2 tenant lifecycle test that drives the API through the full stack and verifies the terminal result and runner cleanup.
