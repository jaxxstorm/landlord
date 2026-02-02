## Context

Landlord currently has a Restate workflow provider and a standalone worker entrypoint, but worker registration is manual and only tenant provisioning workflows are supported. The change introduces a pluggable worker engine API so different workflow providers can register and run workers, and extends workflows to cover create/update/delete lifecycle operations. Workers should be able to resolve compute engine selection from the Landlord server to avoid hard-coding compute targets.

## Goals / Non-Goals

**Goals:**
- Define a workflow worker engine interface and lifecycle contract that supports multiple providers.
- Implement a Restate worker engine that self-registers on startup and executes tenant create/update/delete workflows.
- Enable workers to fetch compute engine type (and related config) from the Landlord API.
- Provide an end-to-end Restate test that covers create, update, and delete flows.

**Non-Goals:**
- Implement non-Restate worker engines beyond defining interfaces (e.g., Step Functions worker).
- Redesign existing tenant API contracts or persistence models.
- Add UI or non-essential operational tooling.

## Decisions

- Introduce a `WorkflowWorkerEngine` interface parallel to the existing workflow provider abstraction, with explicit hooks for registration and workflow execution. This keeps worker lifecycle concerns separate from API-side workflow orchestration and allows providers to implement only the worker side.
- Restate worker engine will register itself on startup using Restate admin APIs, using configurable endpoint and namespace settings. This avoids manual registration and enables automated local and CI workflows.
- Tenant lifecycle workflows will be modeled as explicit operations (create, update, delete) with shared execution path and common plan/transition handling. This makes it easier to extend to future operations while keeping idempotency guarantees.
- Workers will resolve compute engine type and configuration by querying the Landlord API (or internal service) at startup and/or per job. This allows compute targets to be changed without redeploying workers and keeps the worker stateless.
- End-to-end tests will run the Restate worker alongside a test server and exercise create, update, and delete flows to ensure registration and workflow dispatch work together.

## Risks / Trade-offs

- [Risk] Worker registration relies on Restate admin availability and auth. -> Mitigation: retries with backoff and clear failure logging; allow startup to fail fast in CI.
- [Risk] Fetching compute configuration remotely introduces latency/failure modes. -> Mitigation: cache responses with TTL and allow override via config for local tests.
- [Risk] Adding lifecycle operations can widen surface area for inconsistencies. -> Mitigation: reuse shared validation and reconciliation paths and add integration tests for each operation.
- [Trade-off] Keeping worker engines separate from workflow providers adds an additional abstraction layer. -> Mitigation: keep the interface small and mirror provider naming to reduce cognitive load.
