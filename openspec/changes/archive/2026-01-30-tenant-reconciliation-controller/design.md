## Context

The landlord control plane currently has no mechanism to act on tenants after they're created via the API. Tenants remain in the `requested` state because there's no reconciliation loop to detect state changes and trigger workflow execution. 

**Current State:**
- API creates tenant records with status="requested"
- Workflow providers (Restate, Step Functions) are registered but never invoked
- Compute providers (Docker, ECS) exist but have no trigger mechanism
- No continuous reconciliation or drift detection

**Constraints:**
- Must be lightweight - avoid full K8s controller-runtime dependency
- Should integrate with existing workflow provider abstraction
- Must support graceful shutdown for zero-downtime deployments
- Needs to handle database polling efficiently (avoid constant queries)
- Must coexist with HTTP API server in single process

**Stakeholders:**
- Tenant lifecycle management (primary use case)
- Workflow orchestration system integration
- Future: Multi-tenant controller patterns, horizontal scaling

## Goals / Non-Goals

**Goals:**
- Implement a reconciliation controller that processes tenants through their lifecycle (requested → planning → provisioning → ready)
- Use `k8s.io/client-go/util/workqueue` for reliable event processing with backoff/retry
- Integrate with workflow provider system to trigger provisioning workflows
- Support graceful startup/shutdown and error handling
- Add observability hooks (structured logging, future metrics)
- Ensure single-process operation with HTTP server

**Non-Goals:**
- Horizontal scaling / distributed controller pattern (future)
- Direct compute provider integration (workflows handle this)
- Watch-based notifications (stick with polling for simplicity)
- Complex scheduling / priority handling
- Multi-namespace or tenant isolation (single control plane for now)

## Decisions

### Decision 1: Polling vs Watch Pattern

**Choice:** Database polling with configurable interval

**Rationale:**
- **Simplicity**: No need for database triggers, event streams, or change data capture
- **Portability**: Works across PostgreSQL and SQLite without vendor-specific features
- **Sufficient**: For MVP with <1000 tenants, polling every 5-10s is acceptable
- **Future-proof**: Can optimize to watch/notify pattern later without changing reconciler interface

**Alternatives considered:**
- PostgreSQL LISTEN/NOTIFY: Requires PG-specific code, adds complexity
- CDC with Debezium: Too heavy for current scale
- Event-driven with message queue: Adds another dependency

**Trade-off:** Slightly higher latency (up to polling interval) vs simplicity

### Decision 2: Workqueue Implementation

**Choice:** Use `k8s.io/client-go/util/workqueue.RateLimitingInterface`

**Rationale:**
- **Battle-tested**: Used by all K8s controllers, proven at massive scale
- **Lightweight**: Just the workqueue package, not full controller-runtime (~200KB)
- **Features we need**: Rate limiting, exponential backoff, graceful shutdown, metrics hooks
- **Familiar**: Many Go developers know this pattern

**Alternatives considered:**
- Custom queue: Reinventing wheel, missing edge cases
- Go channels: No rate limiting, backoff, or retry logic
- Full controller-runtime: Too heavy (30+ MB), K8s dependencies we don't need

**Trade-off:** K8s dependency (acceptable for well-isolated workqueue package)

### Decision 3: Reconciliation Trigger Strategy

**Choice:** Periodic list-watch with workqueue deduplication

**Rationale:**
- List all tenants in non-terminal states (requested, planning, provisioning, updating, deleting)
- Enqueue each tenant ID into workqueue (automatically deduplicates)
- Workers pull from queue and reconcile one tenant at a time
- Failed reconciliations get re-queued with exponential backoff

**Alternatives considered:**
- Single-pass batch reconciliation: No retry logic, loses failed items
- Event-driven per-tenant: Requires database triggers or API hooks
- Continuous streaming: Complexity without clear benefit at current scale

**Trade-off:** Potential duplicate work if tenant changes during reconciliation cycle

### Decision 4: Worker Concurrency Model

**Choice:** Configurable worker pool (default: 3 workers)

**Rationale:**
- Multiple workers for throughput (parallel tenant reconciliation)
- Workqueue handles synchronization and deduplication
- Workflow provider calls may be slow (HTTP to Restate/AWS) - parallelism helps
- Low enough default to avoid overwhelming downstream systems

**Alternatives considered:**
- Single worker: Too slow for bulk operations
- Dynamic scaling: Adds complexity without clear need
- Per-tenant goroutine: No bounds, potential resource exhaustion

**Trade-off:** More workers = more pressure on workflow providers and database

### Decision 5: Status Transition Logic

**Choice:** State machine in reconciler with validation

**Rationale:**
- Enforce valid transitions (e.g., requested → planning → provisioning → ready)
- Fail fast on invalid transitions with clear errors
- Reconciler owns state transitions, not API or workflows
- Status transitions are audit-logged to `tenant_state_history` table

**Alternatives considered:**
- Workflows manage transitions: Couples workflow implementation to status model
- No validation: Allows invalid states, hard to debug
- Database constraints: Too rigid, harder to evolve

**Trade-off:** Reconciler must understand full state machine (documented in tenant domain)

### Decision 6: Workflow Integration Pattern

**Choice:** Reconciler calls workflow manager, which delegates to provider

**Rationale:**
- Workflow manager is already abstraction over Restate/Step Functions/Mock
- Reconciler stays decoupled from specific workflow implementations
- Workflow provider handles execution tracking, not reconciler
- Clean separation: Controller (orchestration) vs Workflow (execution)

**Pattern:**
```
Reconciler → Workflow Manager → Provider (Restate/SFN)
```

**Alternatives considered:**
- Direct provider calls: Tight coupling, hard to test
- Reconciler manages workflow state: Duplicates workflow provider responsibilities
- Event-driven: Adds message bus complexity

**Trade-off:** Two layers of abstraction (acceptable for clean design)

### Decision 7: Configuration Structure

**Choice:** Add `controller` section to config with tuning knobs

**Rationale:**
- Reconciliation interval, worker count, retry settings need tuning per environment
- Feature flag for disabling controller (testing, migrations)
- Queue settings (max retries, rate limit) exposed for operators

**Example:**
```yaml
controller:
  enabled: true
  reconciliation_interval: 10s
  worker_count: 3
  max_retries: 5
  rate_limit_per_second: 10
```

**Alternatives considered:**
- Hard-coded values: No production flexibility
- Environment variables only: Config file preferred for structured settings
- Per-tenant configuration: Over-engineering for current needs

**Trade-off:** More configuration surface area to document and validate

## Risks / Trade-offs

**Risk:** Polling inefficiency at scale (thousands of tenants)  
**Mitigation:** Add filtering to repository query (only non-terminal states, only changed since last poll). Future: Migrate to watch pattern if needed. Monitor query performance.

**Risk:** Reconciler crash loses in-flight work  
**Mitigation:** Workqueue is in-memory (acceptable for MVP). Failed items re-appear on next polling cycle. Future: Add persistent queue if needed.

**Risk:** Multiple reconcilers running (accidental multi-instance)  
**Mitigation:** Document single-instance requirement. Future: Add leader election if horizontal scaling needed.

**Risk:** Slow workflow provider calls block reconciliation  
**Mitigation:** Worker parallelism + HTTP client timeouts. Workflow calls should be async (fire-and-forget). Monitor workflow provider latency.

**Risk:** Database connection pool exhaustion from polling  
**Mitigation:** Polling is infrequent (10s default). Repository uses connection pooling. Monitor active connections.

**Risk:** Tight coupling to K8s workqueue package  
**Mitigation:** Workqueue is well-isolated and stable. Interface can be abstracted if needed. Low risk given adoption.

**Trade-off:** Polling adds latency (up to interval duration) vs simplicity  
**Acceptable:** For provisioning workflows that take minutes, 10s latency is negligible.

**Trade-off:** In-memory queue loses work on crash vs persistence complexity  
**Acceptable:** Tenants re-appear in next polling cycle. Workflows should be idempotent anyway.

## Migration Plan

**Phase 1: Development (this change)**
1. Add controller package with reconciler and workqueue setup
2. Add repository method for listing tenants requiring reconciliation
3. Wire controller into main.go with configuration
4. Test with existing mock workflow provider
5. Validate graceful shutdown

**Phase 2: Integration Testing**
1. Deploy with Restate workflow provider
2. Create test tenants via API, verify reconciliation
3. Monitor logs for reconciliation events
4. Test failure scenarios (workflow errors, database unavailable)
5. Validate retry/backoff behavior

**Phase 3: Production Rollout**
1. Deploy with controller disabled (`controller.enabled: false`)
2. Enable controller in staging environment
3. Monitor metrics (reconciliation latency, queue depth, error rate)
4. Gradually enable in production with low worker count
5. Tune configuration based on observed behavior

**Rollback Strategy:**
- Set `controller.enabled: false` in configuration
- Restart service (controller stops, API continues)
- No data loss (tenants remain in database)
- Re-enable after issue resolution

**Backward Compatibility:**
- API endpoints unchanged (tenants still created with status="requested")
- Database schema unchanged (no new columns)
- Workflow provider interface unchanged
- If controller disabled, system behaves as before (no reconciliation)

## Open Questions

1. **Reconciliation interval tuning**: Is 10s the right default? Should it be adaptive based on tenant count?
   - Resolution: Start with 10s, make configurable, monitor in production

2. **Workflow idempotency**: Do all workflow providers handle duplicate execution requests gracefully?
   - Resolution: Document idempotency requirement, test with each provider, add request deduplication if needed

3. **Error handling granularity**: Should different error types have different retry strategies?
   - Resolution: Start simple (all errors use exponential backoff), refine based on error patterns in production

4. **Observability**: What metrics are most important for controller health?
   - Resolution: Start with: reconciliation latency, queue depth, error rate, tenant status distribution. Add more based on operational needs.

5. **Multi-instance safety**: What happens if operator accidentally runs multiple instances?
   - Resolution: Document single-instance requirement. Add leader election in future change if horizontal scaling needed.
