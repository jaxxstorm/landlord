## Why

Landlord’s AWS workflow and EC2 runner tests use mocks, which cannot verify SDK
requests, Step Functions execution behavior, or runner lifecycle interactions
against AWS-compatible services. Floci provides a free local emulator for Step
Functions, EC2, and EKS, enabling deterministic integration coverage in local
development and CI without AWS credentials.

## What Changes

- Add Floci-backed integration test infrastructure that starts an isolated local
  AWS emulator and configures AWS SDK clients through its endpoint and test
  credentials.
- Add end-to-end Step Functions integration coverage that creates a state
  machine, starts an execution, observes terminal status and history, and stops
  an active execution through the existing provider.
- Add EC2 runner integration coverage that launches, observes, replaces, and
  terminates a tagged runner through Floci’s EC2 emulation.
- Add a Docker-based EC2 lifecycle test that drives Landlord's API, PostgreSQL,
  Step Functions worker, and Floci as one end-to-end system.
- Add a Kubernetes compute runner backed by the Kubernetes API and Floci EKS,
  because no Kubernetes runner currently exists to exercise.
- Add Kubernetes runner integration coverage that provisions, observes, updates,
  and destroys tenant workloads against the Floci-provisioned cluster.
- Make all Floci tests opt-in through an integration build tag and document
  Docker, Floci image, EKS, and endpoint prerequisites. Unit tests remain the
  default test path.

## Capabilities

### New Capabilities
- `floci-integration-testing`: Isolated Floci test infrastructure and
  AWS-compatible integration coverage for Step Functions, EC2 runner, and
  Kubernetes runner paths.
- `kubernetes-compute-runner`: Provider-agnostic tenant workload lifecycle
  management through a Kubernetes cluster.

### Modified Capabilities
- `step-functions-provider`: Define AWS-compatible provider integration
  verification using Floci.

## Impact

- Adds testcontainers-based Floci fixture code and integration-only test
  dependencies.
- Adds a Kubernetes client dependency and `compute.kubernetes` configuration.
- Extends CI documentation and local test instructions; normal unit test runs
  do not start Docker or Floci.
- Requires Docker access for opt-in integration tests; EKS and EC2 emulation
  can start additional containers and may need increased test timeouts.
