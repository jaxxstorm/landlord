## Why

Landlord currently supports Restate and Step Functions workflows, but not Temporal, which blocks teams that standardize on Temporal from using Landlord's full tenant lifecycle orchestration. We need Temporal support now to keep workflow providers pluggable while preserving idempotent lifecycle semantics, cancellation behavior, and compute-provider independence.

## What Changes

- Add a Temporal workflow provider implementation that supports create, start, status, stop/cancel, delete, and validation through the existing workflow provider contract.
- Add a Temporal worker engine implementation for tenant create, update, and delete lifecycle operations using payload-provided tenant state (no direct database reads in workers).
- Add Temporal registration/initialization behavior so required workflows are available and invocable at startup.
- Add Docker Compose local development support for Temporal-backed lifecycle testing.
- Add unit and integration test coverage for Temporal provider/worker behavior, including cancellation, retries, status mapping, and idempotency.
- Ensure Temporal workflows execute tenant lifecycle operations against all supported compute providers via existing compute resolution abstractions.
- Define failure handling expectations: transient Temporal connectivity/availability issues are retried with backoff, terminal workflow failures surface actionable status and errors, and cancellation/stop operations are idempotent.
- Define observability expectations: Temporal workflow and worker operations emit structured logs and expose status transitions compatible with existing tenant status reporting.

## Capabilities

### New Capabilities
- `temporal-workflow-provider`: Temporal-backed implementation of the pluggable workflow provider interface, including lifecycle execution and cancellation.
- `temporal-worker-engine`: Temporal worker runtime for tenant create/update/delete workflows using provider-agnostic compute execution paths.
- `temporal-local-integration-test`: Local Docker Compose and integration test coverage for end-to-end Temporal workflow execution.

### Modified Capabilities
- `workflow-state-mapping`: Extend canonical workflow sub-state mapping requirements to include Temporal-native execution states and retry metadata.
- `workflow-worker-engines`: Clarify worker engine requirements for provider parity, including cancellation handling and compute-provider compatibility for Temporal.

## Impact

- **Code**: `internal/workflow/providers/temporal`, worker binaries/entrypoints, provider registration/bootstrap wiring, configuration plumbing, and Docker Compose assets.
- **Tests**: New unit tests for Temporal provider/worker behavior; new/updated integration tests for end-to-end tenant lifecycle with Temporal.
- **Dependencies**: Temporal Go SDK and local Temporal service containers for development/integration testing.
- **Operations**: New configuration surface for Temporal connectivity/namespace/task queue and worker startup behavior.
- **APIs**: No new external HTTP endpoints required; existing tenant lifecycle endpoints continue to drive workflow operations through the provider abstraction.
