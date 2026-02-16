## 1. Configuration and dependency setup

- [x] 1.1 Add Temporal SDK dependencies to `go.mod` and verify module resolution.
- [x] 1.2 Extend `internal/config/workflow.go` with `TemporalConfig` fields and validation logic.
- [x] 1.3 Update workflow default-provider validation to accept `temporal`.
- [x] 1.4 Add Temporal configuration examples to `config.example.yaml` and `docker.config.yaml` (or Temporal-specific local config).
- [x] 1.5 Add/adjust unit tests in `internal/config/workflow_test.go` for valid/invalid Temporal config.

## 2. Temporal workflow provider implementation

- [x] 2.1 Create `internal/workflow/providers/temporal` package scaffold (`provider.go`, `config.go`, `state.go`, `errors.go`).
- [x] 2.2 Implement `Name`, `Validate`, and `CreateWorkflow` with idempotent behavior.
- [x] 2.3 Implement `StartExecution` with deterministic execution naming for retry safety.
- [x] 2.4 Implement `GetExecutionStatus` and map Temporal-native states to canonical workflow states.
- [x] 2.5 Implement `StopExecution` cancellation/termination flow with idempotent repeated calls.
- [x] 2.6 Implement `DeleteWorkflow` behavior consistent with Temporal workflow-definition semantics.
- [x] 2.7 Add provider unit tests for success, transient failures, idempotency, and cancellation.

## 3. Temporal worker engine implementation

- [x] 3.1 Implement `internal/workflow/providers/temporal/worker_engine.go` to satisfy `workflow.WorkerEngine`.
- [x] 3.2 Implement worker registration/start behavior for Temporal namespace/task queue.
- [x] 3.3 Implement create/update/delete workflow handlers using payload-driven tenant state.
- [x] 3.4 Wire compute-provider resolution via existing resolver + compute registry abstractions.
- [x] 3.5 Add cancellation-aware activity/workflow handling at safe boundaries.
- [x] 3.6 Add worker engine unit tests for startup, compute resolution, and cancellation behavior.

## 4. Application and worker runtime wiring

- [x] 4.1 Register Temporal provider in `cmd/landlord/main.go` when Temporal config is present.
- [x] 4.2 Ensure controller workflow-provider selection supports `temporal` default/override paths.
- [x] 4.3 Add `cmd/workers/temporal/main.go` entrypoint with provider-specific worker bootstrap.
- [x] 4.4 Register Temporal worker in worker registry flow and select by configured provider.
- [x] 4.5 Update worker build assets (Dockerfile/build target) to produce Temporal worker binary.

## 5. Workflow state mapping and provider parity updates

- [x] 5.1 Update state-mapping logic to include Temporal statuses and retry/backoff metadata.
- [x] 5.2 Add tests for Temporal-to-canonical state conversion and unknown-state fallback behavior.
- [x] 5.3 Ensure cancellation terminal states are surfaced consistently through tenant workflow status fields.
- [x] 5.4 Validate parity with existing provider contract methods and error semantics.

## 6. Local Docker Compose support for Temporal

- [x] 6.1 Add Temporal services (and optional UI) to compose assets via profile or overlay file.
- [x] 6.2 Add Temporal worker service to compose assets with required env/config mounts.
- [x] 6.3 Add compose health checks and startup dependencies so integration tests wait for readiness.
- [x] 6.4 Add developer run instructions for Temporal local stack startup and teardown.

## 7. Integration and end-to-end testing

- [x] 7.1 Add Temporal provider integration tests for workflow registration/start/status/stop paths.
- [x] 7.2 Add end-to-end API lifecycle integration tests for create/update/delete via Temporal.
- [x] 7.3 Add integration test for cancelling an in-flight Temporal lifecycle workflow.
- [x] 7.4 Add compute-provider compatibility tests that exercise Temporal workflows across enabled compute providers.
- [x] 7.5 Ensure test suites are stable under retries and repeated registration/start operations.

## 8. Final verification

- [x] 8.1 Run targeted unit test packages for config, workflow manager, and Temporal provider/worker packages.
- [x] 8.2 Run Temporal integration suite against local Docker Compose stack.
- [x] 8.3 Verify no regressions in existing Restate and Step Functions tests.
- [x] 8.4 Confirm `openspec status --change add-temporal-workflow-provider` reports all artifacts complete.
