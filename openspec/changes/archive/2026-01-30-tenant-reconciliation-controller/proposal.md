## Why

Tenants created via the API remain in the `requested` state indefinitely because there is no controller to reconcile desired state with actual state. The control plane architecture requires a reconciliation loop that watches for tenant status changes and triggers workflow actions to provision resources. Without this, tenants exist only as database records with no actual infrastructure provisioned.

## What Changes

- Implement a tenant reconciliation controller that continuously monitors tenant status
- Use a lightweight workqueue-based reconciliation pattern inspired by Kubernetes controllers (using `k8s.io/client-go/util/workqueue`)
- Integrate with the pluggable workflow provider system to trigger provisioning workflows
- Support graceful startup, shutdown, and error handling with backoff/retry logic
- Add controller lifecycle management (start, stop, health checking)
- Implement status transition logic (requested → planning → provisioning → ready)
- Add observability through structured logging and metrics hooks

**Out of Scope**: Direct compute provider integration (will be handled by workflow execution in a future change)

## Capabilities

### New Capabilities

- `tenant-reconciliation`: Core reconciliation loop that watches tenant records in the database, detects state changes, and triggers appropriate workflow actions based on tenant status
- `workqueue-integration`: Workqueue-based event processing with rate limiting, exponential backoff, and graceful shutdown using `k8s.io/client-go/util/workqueue`
- `workflow-triggering`: Integration layer between the reconciler and workflow providers that translates tenant state changes into workflow execution requests

### Modified Capabilities

- `tenant-repository`: Add polling/watch methods to efficiently query tenants requiring reconciliation (e.g., tenants in non-terminal states, tenants with drift)

## Impact

**New Components:**
- `internal/controller/` - Controller package with reconciler, workqueue, and event handling
- `internal/controller/reconciler.go` - Main reconciliation logic
- `cmd/landlord/main.go` - Controller initialization and lifecycle management

**Dependencies:**
- Add `k8s.io/client-go` (workqueue only, ~minimal footprint)
- Add `k8s.io/apimachinery` (for backoff utilities)

**Modified Components:**
- `internal/tenant/repository.go` - Add ListTenantsForReconciliation() method
- `internal/workflow/manager.go` - May need triggering interface refinement
- `cmd/landlord/main.go` - Integrate controller startup/shutdown with HTTP server lifecycle

**Configuration:**
- Add controller configuration (reconciliation interval, worker count, queue settings)
- Add feature flags for enabling/disabling reconciliation

**Observability:**
- Structured logs for reconciliation events (start, success, failure, retry)
- Metrics for reconciliation latency, queue depth, error rates (hooks for future Prometheus integration)
