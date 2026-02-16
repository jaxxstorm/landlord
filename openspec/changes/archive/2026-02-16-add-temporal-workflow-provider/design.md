## Context

Landlord currently has workflow providers for `mock`, `step-functions`, and `restate`, plus a worker engine implementation for Restate. The workflow/provider abstractions already include lifecycle operations (`CreateWorkflow`, `StartExecution`, `GetExecutionStatus`, `StopExecution`, `DeleteWorkflow`, `Validate`) and a worker engine abstraction (`Register`, `Start`), but Temporal is not yet wired into configuration, provider registration, worker startup, or local development infrastructure.

This change is cross-cutting across:
- `internal/config` (workflow provider configuration and validation)
- `cmd/landlord` (provider registration and default provider resolution)
- `cmd/worker` and provider-specific worker entrypoints (worker engine registration/startup)
- `internal/workflow/providers/*` (new Temporal provider and worker engine)
- `docker-compose.yml` and dev config assets (local Temporal stack)
- unit/integration test suites

Key constraints:
- No new external HTTP API contract is required; existing tenant lifecycle endpoints continue to trigger workflows.
- Workflow and worker execution must remain idempotent and retry-safe.
- Compute execution must remain provider-agnostic across all registered compute providers.

## Goals / Non-Goals

**Goals:**
- Add a Temporal workflow provider implementation that satisfies the existing workflow provider contract.
- Add a Temporal worker engine implementation for tenant create/update/delete execution.
- Support cancellation via `StopExecution` and deterministic terminal state reporting.
- Keep compute behavior provider-agnostic and compatible with all registered compute providers.
- Provide local Docker Compose support for Temporal and an integration test path.
- Add unit and integration tests that validate provider behavior, worker behavior, status mapping, idempotency, and cancellation.

**Non-Goals:**
- Redesigning tenant lifecycle state names or API response schema.
- Replacing existing Restate or Step Functions providers.
- Introducing a new worker control-plane HTTP API.
- Solving multi-cluster or multi-namespace Temporal tenancy in this change.

## Decisions

### Decision 1: Add first-class Temporal configuration in `workflow` config

Add `TemporalConfig` to `internal/config/workflow.go` and extend `WorkflowConfig.Validate()` to accept `temporal` as a valid provider.

Proposed config surface:
- `workflow.temporal.host_port` (Temporal frontend target)
- `workflow.temporal.namespace`
- `workflow.temporal.task_queue`
- `workflow.temporal.retry_attempts`
- `workflow.temporal.timeout`
- worker-specific options mirroring existing worker startup patterns as needed

Rationale:
- Keeps provider configuration explicit, validated, and discoverable like Restate/Step Functions.
- Prevents runtime failures from malformed provider config.

Alternatives considered:
- Reuse untyped provider config map only: rejected due weak validation and poor UX.
- Encode Temporal config under generic worker section: rejected because provider config belongs to workflow provider concerns.

### Decision 2: Implement Temporal provider under `internal/workflow/providers/temporal`

Create a provider package implementing `workflow.Provider` with Temporal SDK client integration.

Behavior model:
- `Name()` returns `temporal`.
- `Validate()` validates workflow identifiers and provider config.
- `CreateWorkflow()` is idempotent and validates workflow metadata used by Temporal start calls.
- Controller-triggered lifecycle invokes use shared workflow ID `tenant-provisioning` (operation remains payload-driven), avoiding per-tenant registration churn.
- `StartExecution()` starts Temporal workflows with deterministic execution names for idempotency.
- `GetExecutionStatus()` maps Temporal statuses to canonical states/sub-states.
- `StopExecution()` requests cancellation (and uses termination fallback where appropriate).
- `DeleteWorkflow()` is idempotent and treated as logical unregistration for code-defined workflows.

Rationale:
- Preserves existing manager/controller integration with minimal orchestration changes.
- Keeps provider-specific logic isolated in a single adapter package.

Alternatives considered:
- New Temporal-specific manager bypassing `workflow.Provider`: rejected because it fragments provider abstraction.
- Embedding Temporal logic in controller layer: rejected due layering violations.

### Decision 3: Add Temporal worker engine that reuses compute abstractions

Create `internal/workflow/providers/temporal/worker_engine.go` implementing `workflow.WorkerEngine`, and add `cmd/workers/temporal/main.go` to run Temporal workers.

Worker behavior:
- Startup validates config, initializes Temporal worker, and registers workflows/activities.
- Lifecycle handlers for create/update/delete use payload-driven execution and shared compute manager abstractions.
- Compute provider resolution follows existing payload -> resolver -> default fallback behavior.
- Cancellation is handled at workflow/activity boundaries and is idempotent.

Rationale:
- Mirrors successful Restate worker pattern while keeping Temporal-specific code isolated.
- Ensures compute-provider compatibility without hardcoding provider-specific branches.

Alternatives considered:
- Extend existing Restate worker binary with Temporal codepaths only: rejected to avoid tight coupling and accidental provider leakage.
- Build a new worker contract: rejected because existing `WorkerEngine` contract already fits.

### Decision 4: Register Temporal provider/worker through existing registries

Update startup wiring to register Temporal provider when configured:
- `cmd/landlord/main.go` registers temporal provider into `workflow.Registry`.
- worker startup registers temporal worker engine into `workflow.WorkerRegistry` and selects by configured workflow provider.

Rationale:
- Keeps provider selection centralized and consistent with existing default/override behavior.
- Avoids conditional logic spread across controller and API layers.

Alternatives considered:
- Hard-select Temporal based on env flags: rejected because it bypasses existing configuration model.

### Decision 5: Add Temporal local stack via Docker Compose overlay

Add Temporal services and worker to local development through compose assets (either extending `docker-compose.yml` with profiles or adding a dedicated overlay file).

Required local components:
- Temporal server/frontend dependencies
- Temporal UI (optional but useful for debugging)
- Landlord API container
- Temporal worker container
- Existing database and compute dependencies

Rationale:
- Provides repeatable local integration testing and debugging.
- Avoids requiring cloud services for baseline Temporal validation.

Alternatives considered:
- Testcontainers-only integration path: rejected as sole path because local operator workflow still needs compose support.
- Replacing Restate compose setup entirely: rejected to preserve existing workflows.

### Decision 6: Expand tests with provider parity and compute compatibility focus

Test plan structure:
- Unit tests:
  - `internal/config/workflow_test.go` for temporal config validation/default provider support.
  - `internal/workflow/providers/temporal/*_test.go` for provider methods, state mapping, idempotency, cancellation.
  - Temporal worker engine tests for compute resolution, cancellation handling, and startup registration behavior.
- Integration tests:
  - End-to-end create/update/delete/cancel flows with Temporal in local stack.
  - Workflow registration/startup readiness checks.
  - Compute compatibility coverage across enabled compute providers (always docker/mock locally; provider-specific environments can add ECS coverage when configured).

Rationale:
- Aligns test coverage with failure-prone areas: status mapping, retries, cancellation, and provider selection.

Alternatives considered:
- Unit-only coverage: rejected because provider wiring and lifecycle behavior require integration verification.

## Risks / Trade-offs

- [Temporal semantics differ from Restate/Step Functions] -> Mitigation: explicit state mapping tests and canonical mapping requirements in spec deltas.
- [Cancellation can be nondeterministic if activities are not cancellation-aware] -> Mitigation: enforce cancellation checks at activity boundaries and verify in integration tests.
- [Compose complexity increases local startup time and maintenance overhead] -> Mitigation: isolate Temporal services via profiles/overlay and keep defaults stable.
- [Compute-provider parity regressions in worker path] -> Mitigation: route all compute actions through existing `compute.Manager` interfaces and add provider-matrix tests for enabled providers.
- [Workflow config validation drift between providers] -> Mitigation: add explicit config validation tests and keep provider-specific validation in `internal/config/workflow.go`.

## Migration Plan

1. Add Temporal config types and validation in `internal/config/workflow.go` and update example configs.
2. Add Temporal provider package and register it in `cmd/landlord/main.go` when configured.
3. Add Temporal worker engine package and worker entrypoint (`cmd/workers/temporal/main.go`), then register/select via `workflow.WorkerRegistry`.
4. Add Docker Compose Temporal services and worker wiring, plus Temporal-specific local config values.
5. Add/adjust unit tests for config, provider behavior, worker behavior, and state mapping.
6. Add integration tests for end-to-end lifecycle and cancellation with Temporal.
7. Validate `go test` for affected packages and run Temporal local integration suite.
8. Rollout by setting `workflow.default_provider=temporal` in target environments only after verification.

Rollback strategy:
- Revert `workflow.default_provider` to `restate` or `step-functions`.
- Stop Temporal worker containers/processes.
- Keep Temporal provider code inactive when not configured.

## Open Questions

- Should `CreateWorkflow`/`DeleteWorkflow` for Temporal be modeled as logical no-op operations with validation only, or should we maintain explicit workflow registration metadata in persistence?
- Should worker startup fail hard if Temporal namespace/task queue do not exist, or auto-create namespace in local/dev only?
- Do we want ECS-enabled Temporal integration tests in default CI, or only in a credentialed optional job?
- Do we need a shared generic worker binary that can host multiple worker engines, or keep provider-specific binaries long-term?
