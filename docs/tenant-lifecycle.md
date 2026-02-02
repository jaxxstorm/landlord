# Tenant Lifecycle and Reconciliation

This document describes the complete lifecycle of a tenant in Landlord and how the controller manages state transitions.

## Tenant Lifecycle Overview

A tenant's lifecycle begins when it's created via the API and ends when it's deleted. Throughout its lifecycle, the reconciliation controller ensures the actual deployed state matches the desired configuration.

## State Transitions and Workflow

### 1. Creation Phase

**Step 1: API Request**
- User sends tenant creation request via REST API
- API validates request (required fields, naming constraints, etc.)
- Tenant is stored in database with `requested` status

**Step 2: Planning Phase**
- Controller detects tenant in `requested` status
- Controller triggers "plan" workflow with desired configuration
- Workflow provider (Docker, ECS, Step Functions, etc.) generates deployment plan
- Plan specifies which resources need to be created, their configuration, etc.

**Step 3: Provisioning Phase**
- If plan succeeds, tenant transitions to `provisioning` status
- Controller triggers "provision" workflow
- Workflow provider allocates and configures resources:
  - Creates compute instances or containers
  - Sets up networking, load balancers, DNS
  - Configures application container with desired image and config
  - Performs initial health checks

**Step 4: Ready Phase**
- If provisioning succeeds, tenant transitions to `ready` status
- Tenant is now serving traffic
- Controller continues monitoring for changes

### 2. Operational Phase

**Steady State**
- Tenant remains in `ready` status while healthy
- Controller periodically checks status (every `CONTROLLER_RECONCILIATION_INTERVAL`)
- If observed state matches desired state, no action is taken

**Updates**
- When image or configuration needs to change:
  - User updates tenant via API
  - Tenant transitions to `updating` status
  - Controller triggers "update" workflow
  - Workflow provider performs rolling update or blue-green deployment
  - Once update completes, tenant returns to `ready` status

### 3. Deletion Phase

**Step 1: Deletion Request**
- User requests tenant deletion via API
- Tenant transitions to `deleting` status
- Controller triggers "delete" workflow

**Step 2: Resource Cleanup**
- Workflow provider tears down all resources:
  - Stops and removes compute instances
  - Deletes networking resources
  - Removes DNS entries
  - Cleans up persistent storage

**Step 3: Deleted Phase**
- If cleanup succeeds, tenant transitions to `deleted` status
- Row remains in database (soft delete) for audit trail
- Tenant no longer consumes resources

## Error Handling and Retry Logic

### Transient Errors (Retryable)

When a workflow fails with transient errors (network timeouts, temporary provider unavailability):

1. Reconciliation is **not immediately re-attempted**
2. Tenant is re-queued with **exponential backoff**:
   - First retry: 1 second
   - Second retry: ~2 seconds
   - Third retry: ~4 seconds
   - Continues doubling until reaching 5 minute maximum
3. After each retry, controller checks if workflow succeeds
4. Tenant progresses to next state on success

### Fatal Errors (Non-Retryable)

When a workflow fails with fatal errors (invalid configuration, provider rejecting request):

1. Reconciliation immediately fails
2. Tenant transitions to `failed` status
3. StatusMessage contains error details
4. Manual investigation and intervention required
5. Operator can:
   - Fix configuration and update tenant (triggers new reconciliation)
   - Delete tenant to clean up
   - Monitor logs and workflow provider for root cause

### Max Retries Exceeded

If a transient error persists across all retry attempts:

1. Retry counter reaches `CONTROLLER_MAX_RETRIES` limit (default: 5)
2. Tenant automatically transitions to `failed` status
3. Same as fatal error handling - requires manual intervention

## Reconciliation Loop Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                  Reconciliation Controller                       │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │         Polling Loop (every reconciliation_interval)      │  │
│  │                                                           │  │
│  │  1. Query database for tenants in non-terminal states    │  │
│  │  2. For each tenant:                                      │  │
│  │     - Add tenant_id to work queue                         │  │
│  │     - Queue handles deduplication automatically          │  │
│  └──────────────────────────────────────────────────────────┘  │
│                           ↓                                      │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │     Work Queue (Rate-Limited with Exponential Backoff)   │  │
│  │                                                           │  │
│  │  - Tracks processing state for each tenant               │  │
│  │  - Automatically throttles retries                        │  │
│  │  - Prevents duplicate processing                         │  │
│  └──────────────────────────────────────────────────────────┘  │
│                           ↓                                      │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Worker Pool (CONTROLLER_WORKERS concurrent workers)      │  │
│  │                                                           │  │
│  │  Each worker:                                             │  │
│  │  1. Get item from queue (blocking)                        │  │
│  │  2. Fetch tenant from database                            │  │
│  │  3. Check if status still needs reconciliation            │  │
│  │  4. Determine action (plan, provision, update, delete)   │  │
│  │  5. Trigger workflow with timeout protection             │  │
│  │  6. On success: transition to next state                 │  │
│  │  7. On retryable error: re-queue with backoff            │  │
│  │  8. On fatal error: transition to failed state           │  │
│  │  9. Mark item as done in queue                           │  │
│  └──────────────────────────────────────────────────────────┘  │
│                           ↓                                      │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │       Database Updates (With Optimistic Locking)         │  │
│  │                                                           │  │
│  │  - Status transitions use version checking               │  │
│  │  - Prevents lost updates from concurrent modifications  │  │
│  │  - Logs status changes for audit trail                   │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Monitoring and Observability

### Key Metrics to Monitor

- **reconciliation_duration**: How long each reconciliation takes
- **queue_depth**: Number of tenants pending reconciliation
- **retry_count**: Distribution of retry counts before success
- **error_rate**: Percentage of failed reconciliation attempts
- **state_transition_count**: Tenants transitioning between states

### Logs to Check

Watch for these log patterns:

```
# Normal operation
"reconciling tenant" tenant_id=... status=...
"workflow triggered" tenant_id=... action=... execution_id=...
"tenant reconciled successfully" tenant_id=... previous_status=... new_status=...

# Errors to investigate
"reconciliation failed" tenant_id=... error=... retry_count=...
"max retries exceeded, marking tenant as failed" tenant_id=...
"reconciler shutdown timeout exceeded"
```

## Performance Considerations

### Tuning Reconciliation Interval

- **Too Frequent** (1-2s): High database load, more CPU usage, faster response to changes
- **Too Infrequent** (60s+): Lower database load, slow response to state changes
- **Recommended**: 5-15s for most deployments

### Tuning Worker Count

- **Too Few** (1-2): Slow processing, queue backlog, high latency for reconciliations
- **Too Many** (20+): High resource usage, potential database connection exhaustion
- **Recommended**: 3-8 workers, tune based on load testing

### Database Optimization

- Ensure `status` column is indexed (already configured)
- Monitor query performance on `ListTenantsForReconciliation`
- Consider connection pooling settings if hitting limits

## Graceful Shutdown

During application shutdown:

1. API server stops accepting new requests
2. Controller stops polling for new tenants
3. Controller waits up to `CONTROLLER_SHUTDOWN_TIMEOUT` for in-flight reconciliations
4. Any reconciliations not completed within timeout are terminated
5. Workers and polling loop exit gracefully
6. Application exits

To minimize data loss during shutdown:
- Ensure `CONTROLLER_SHUTDOWN_TIMEOUT` is sufficient for typical reconciliation duration
- Monitor logs for "shutdown timeout exceeded" messages
- Consider increasing timeout if tenants frequently exceed it

## Troubleshooting

### Tenants Stuck in Non-Terminal States

**Symptoms**: Tenants remain in `planning`, `provisioning`, or `updating` for extended period

**Diagnosis**:
1. Check controller logs for reconciliation errors
2. Verify workflow provider is healthy and responding
3. Check database connectivity
4. Look for "max retries exceeded" or error messages

**Resolution**:
- Restart workflow provider if unresponsive
- Check network connectivity between controller and provider
- Increase `CONTROLLER_WORKFLOW_TRIGGER_TIMEOUT` if provider is slow but operational
- Check tenant logs and provider-specific error messages

### High Queue Depth

**Symptoms**: `queue_depth` metric is high and not decreasing

**Diagnosis**:
1. Check reconciliation duration - reconciliations taking longer than expected
2. Check number of healthy workers - verify all workers are running
3. Check error rate - repeated failures causing retries

**Resolution**:
- Increase `CONTROLLER_WORKERS` to process queue faster
- Optimize workflow provider performance
- Investigate and fix recurring errors
- Increase machine resources if CPU/memory bottlenecked

### Reconciliation Timeout Errors

**Symptoms**: Many "workflow trigger timeout" errors in logs

**Diagnosis**:
1. Workflow provider is slow but not completely unavailable
2. Network latency to provider is high
3. Provider is overwhelmed with requests

**Resolution**:
- Increase `CONTROLLER_WORKFLOW_TRIGGER_TIMEOUT`
- Reduce `CONTROLLER_WORKERS` to lower load on provider
- Check provider metrics and scale if needed
- Verify network connectivity and latency
