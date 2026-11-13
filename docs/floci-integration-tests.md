# Floci Integration Tests

Floci-backed tests verify the Step Functions provider and EC2 runner against
AWS-compatible local services. They are opt-in and do not run as part of the
default Go test suite.

## Prerequisites

- Docker Engine must be running and able to mount `/var/run/docker.sock`.
- The first run pulls `floci/floci:latest` and the Docker image used by the EC2
  runner test.

## Run Locally

```bash
make test-floci
```

The fixture starts an isolated Floci container, provides disposable AWS
credentials and endpoint configuration, and cleans up created resources. No
AWS account or credentials are required.

Run the standard Docker-free suite with:

```bash
make test
```

## EC2 End-To-End Stack

The EC2 end-to-end stack is separate from the provider integration tests. It
builds the Linux custom-runtime Lambda from `cmd/workers/stepfunctions`,
generates a stack-specific configuration, and creates the Lambda, IAM roles,
and Standard Step Functions state machine in an isolated Floci account.

```bash
make floci-e2e-up
```

The command prints a 12-digit Floci account, local API URL, and the generated
state file under `dist/floci-e2e/`. It uses random loopback ports and a unique
Docker network, so multiple stacks can run at the same time. Use the reported
suffix to stop a stack:

```bash
LANDLORD_FLOCI_E2E_ID=floci-123456789012 make floci-e2e-down
```

To build, bootstrap, verify API readiness, and clean up in one command, run:

```bash
make test-floci-e2e-stack
```

This verifies only the reusable stack bootstrap; the API tenant lifecycle test
is intentionally separate. Run the complete API lifecycle test with:

```bash
make test-floci-e2e
```

This opt-in test starts an isolated stack, creates an EC2 tenant through the
public API, waits for `ready` or `failed`, deletes the tenant, then verifies
the API no longer returns the tenant and Floci has no active EC2 runner with
the tenant's Landlord ownership tags. Docker must allow the Floci container to
mount `/var/run/docker.sock`, because Floci runs Lambda and EC2 runners as
Docker containers.
