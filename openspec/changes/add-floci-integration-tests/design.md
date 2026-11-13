## Context

The Step Functions provider and EC2 runner rely primarily on mock SDK clients.
Floci exposes AWS-compatible Step Functions and EC2 APIs locally and runs EKS
as a Docker-backed k3s cluster. Landlord has no Kubernetes compute provider, so
Kubernetes coverage requires a small provider implementation as well as test
infrastructure.

## Goals / Non-Goals

**Goals:**

- Run opt-in integration tests against an isolated Floci container using Docker.
- Validate the Step Functions provider at the AWS API boundary.
- Validate EC2 runner lifecycle operations against a real AWS-compatible EC2 service.
- Add a Kubernetes compute provider that manages tenant workloads through the Kubernetes API exposed by Floci EKS.
- Keep normal unit tests independent of Docker, Floci, and AWS credentials.

**Non-Goals:**

- Replace mock-client tests or require Floci in default CI.
- Add production EKS provisioning, cluster lifecycle management, Helm support, ingress, autoscaling, or multi-cluster scheduling.
- Treat Floci behavior as proof of identical behavior in every AWS region.

## Decisions

### Use opt-in Go integration tests with a shared Floci fixture

Add an `integration` build tag and a fixture that starts `floci/floci` through
Testcontainers with the Docker socket mounted. Tests set static credentials,
`AWS_ENDPOINT_URL`, and a unique account/region. This keeps default `go test
./...` fast and lets CI opt in only where Docker is available. A standalone
Docker Compose service was rejected because it introduces shared state and port
coordination across tests.

### Create resources per test and clean them up

Each test creates uniquely named Step Functions resources, EC2 runners, and EKS
clusters or namespaces and registers cleanup with `t.Cleanup`. Tests use bounded
polling for asynchronous state. Cleanup failures are reported but do not hide
assertion failures. The fixture uses in-memory Floci storage so a failed run
cannot affect later runs.

### Implement Kubernetes runner against the Kubernetes API

Add a `kubernetes` compute provider that uses a kubeconfig or API endpoint from
operator configuration. Provision creates a tenant namespace and a Deployment
from tenant container configuration; status reads Deployment and Pod readiness;
update applies the desired Deployment; destroy removes the namespace. Floci EKS
supplies the test cluster. Directly using EKS AWS APIs for workload lifecycle
was rejected because workloads are managed through Kubernetes, not EKS.

### Keep integration tests focused on lifecycle contracts

Step Functions tests cover start, describe/history, duplicate behavior, and
stop. EC2 tests cover tagged launch, status, replacement, and termination.
Kubernetes tests cover namespace/workload create, status, update, and destroy.
These scenarios prove Landlord's provider boundaries while unit tests retain
fine-grained validation and AWS error coverage.

### Run the EC2 end-to-end path in a disposable Docker stack

The end-to-end suite will start PostgreSQL, the Landlord API, and Floci on an
isolated Docker network. Before the API becomes ready, a bootstrap step will
create the Floci IAM roles, upload a Linux custom-runtime Lambda ZIP containing
the root-level `bootstrap` binary built from `cmd/workers/stepfunctions`, and
create a Step Functions state machine that invokes that Lambda. The Lambda ZIP
will include a generated configuration that points at the stack PostgreSQL and
Floci endpoints and enables the EC2 runner with Floci's default subnet and
security group.

The test will submit an EC2 tenant request to the public API, poll the public
tenant and workflow status endpoints until a terminal result, then delete the
tenant and verify that the tagged Floci EC2 runner is terminated. The worker is
run as a Lambda inside Floci rather than as the existing Restate worker
container because Step Functions invokes Lambda directly. Reusing the Restate
Compose service was rejected because it does not exercise the Step Functions
production path.

The stack will use a unique Docker Compose project name and Floci account per
test run. Cleanup will delete the tenant, stop active executions, remove
bootstrap resources, and run Compose teardown with volumes. The end-to-end
suite will have its own Make target and CI job because Lambda packaging and
image builds make it slower than the provider-level Floci suite.

## Risks / Trade-offs

- [Floci EKS or Step Functions behavior differs from AWS] -> retain mock tests
  and document Floci as AWS-compatible integration coverage, not a production substitute.
- [Docker is unavailable in CI] -> integration tests are build-tagged and skip
  with an actionable prerequisite message.
- [EKS startup is slow] -> use one fixture per package and bounded readiness
  polling with explicit integration timeouts.
- [Test resources leak after a failure] -> use unique names, `t.Cleanup`, and
  fixture teardown; in-memory Floci state disappears at container shutdown.
- [Lambda cannot reach PostgreSQL or Floci] -> place services on one isolated
  Docker network and configure container-reachable endpoint hostnames.
- [E2E image builds make feedback slow] -> keep the suite opt-in and run it in
  a dedicated CI job with an explicit timeout.

## Migration Plan

1. Add the Kubernetes runner disabled by default and unit-test its API client seam.
2. Add the Floci fixture and opt-in Step Functions and EC2 integration tests.
3. Add Floci EKS integration coverage for the Kubernetes runner.
4. Add the EC2 end-to-end Docker stack and its dedicated CI job.
5. Enable the provider-level `integration` test job only on Docker-capable CI runners.
6. Roll back by omitting the integration build tag and leaving the Kubernetes provider unconfigured.

## Open Questions

- Does Floci expose a stable, testable kubeconfig retrieval API for EKS, or should tests use its documented endpoint and generated credentials directly?
- Should the Kubernetes provider support a shared namespace with tenant labels instead of one namespace per tenant in a later change?
