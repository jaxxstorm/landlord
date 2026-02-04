# Controller Troubleshooting Guide

This guide helps diagnose and resolve common issues with the Landlord tenant reconciliation controller.

## Common Issues and Solutions

### Issue: Tenants Stuck in Planning/Provisioning State

**Symptoms:**
- Tenants remain in `planning`, `provisioning`, or `updating` status for extended time
- No progress in reconciliation despite waiting
- User sees tenant status unchanged after 10+ minutes

**Root Cause:**
- Workflow provider is unresponsive or slow
- Database connection issues
- Network connectivity problems
- Workflow provider configuration invalid

**Diagnosis:**

1. Check controller logs for errors:
```bash
# Look for error messages
grep -i "error\|failed" landlord.log | tail -20

# Check reconciliation logs
grep "reconciliation failed" landlord.log
```

2. Verify workflow provider health:
```bash
# For Docker provider:
docker ps
docker logs <container-id>

# For AWS Step Functions:
aws stepfunctions list-executions --state-machine-arn <arn> --status-filter FAILED
```

3. Check database connectivity:
```bash
# Verify database is accessible
psql -h <db-host> -U <user> -d <database> -c "SELECT 1"

# Check connection pool status if available
```

**Solutions:**

1. **Workflow Provider Unresponsive**
   - Restart workflow provider service
   - Check provider logs for errors
   - Verify provider has sufficient resources (CPU, memory)
   - Check network connectivity between controller and provider

2. **Workflow Timeout**
   - Increase `CONTROLLER_WORKFLOW_TRIGGER_TIMEOUT` if provider is slow but operational
   - Example: `CONTROLLER_WORKFLOW_TRIGGER_TIMEOUT=120s`
   - Check provider performance metrics

3. **Database Issues**
   - Verify database is running and accepting connections
   - Check database resource usage (disk, connections)
   - Look for slow queries in database logs
   - Verify index exists on `status` column: `SELECT * FROM pg_indexes WHERE tablename='tenants'`

4. **Network Issues**
   - Verify firewall rules allow controller to workflow provider
   - Check network latency: `ping <provider-host>`
   - Verify DNS resolution works

### Issue: Tenants Immediately Transition to Failed Status

**Symptoms:**
- Tenants created via API immediately transition to `failed` status
- No reconciliation attempts visible in logs
- `status_message` contains validation or configuration error

**Root Cause:**
- Invalid tenant configuration
- Required fields missing (e.g., no image specified)
- Workflow provider rejects configuration
- Invalid state transition

**Diagnosis:**

1. Check tenant status and message:
```bash
# Get tenant details
curl http://localhost:8080/tenants/{tenant_id}

# Check status_message field for error details
```

2. Check logs:
```bash
grep "fatal\|validation\|invalid" landlord.log
```

**Solutions:**

1. **Invalid Configuration**
   - Review API request, ensure all required fields present
   - Validate image URL is correct and accessible
   - Check configuration constraints (names, sizes, etc.)
   - Review workflow provider documentation for config requirements

2. **Workflow Provider Rejection**
   - Check workflow provider logs for error messages
   - Verify provider has necessary permissions/credentials
   - Test provider directly with sample configuration
   - Review provider error message in tenant `status_message`

### Issue: High Queue Depth / Slow Reconciliation

**Symptoms:**
- `queue_depth` metric continuously increasing
- Reconciliation taking longer than expected
- Many tenants in non-terminal states

**Root Cause:**
- Insufficient workers for load
- Workflow provider performance degraded
- Recurring transient errors causing retries
- Database query performance issues

**Diagnosis:**

1. Check metrics:
```bash
# Monitor queue depth trend
curl http://localhost:8080/metrics | grep queue_depth

# Monitor reconciliation duration
curl http://localhost:8080/metrics | grep reconciliation_duration
```

2. Check error rate:
```bash
# Look for reconciliation failures
grep "reconciliation failed" landlord.log | wc -l
grep "max retries exceeded" landlord.log | wc -l
```

3. Check worker status:
```bash
# Controller logs should show worker count at startup
grep "starting reconciler" landlord.log
```

**Solutions:**

1. **Increase Worker Count**
   ```yaml
   # In config.yaml or env var
   controller:
     workers: 8  # Increase from default 3
   ```
   - Each worker can handle one reconciliation simultaneously
   - Increase if CPU/memory available
   - Monitor resource usage after increase

2. **Investigate Workflow Provider Performance**
   ```bash
   # Check provider latency
   time curl <provider-health-endpoint>
   
   # Check provider resource usage
   docker stats <container> # for Docker provider
   ```
   - Profile provider performance
   - Scale provider horizontally/vertically
   - Check for provider bugs/deadlocks

3. **Fix Recurring Errors**
   - Identify error patterns in logs
   - Increase `CONTROLLER_WORKFLOW_TRIGGER_TIMEOUT` if timeouts
   - Fix underlying cause (network, permissions, etc.)

4. **Optimize Database Queries**
   ```bash
   # Check index exists
   SELECT * FROM pg_indexes WHERE tablename='tenants' AND indexname LIKE '%status%';
   
   # Monitor query performance
   EXPLAIN ANALYZE SELECT * FROM tenants WHERE status IN ('requested', 'planning', ...);
   ```

### Issue: Memory Growth Over Time (Suspected Leak)

**Symptoms:**
- Controller memory usage continuously increases
- Eventually causes OOM killer to terminate container
- Issue occurs after many hours/days of operation

**Root Cause:**
- Goroutine leak
- Unbounded cache growth
- Database connection leak
- Retry state accumulation

**Diagnosis:**

1. Monitor memory usage:
```bash
# Get current memory usage
ps aux | grep landlord

# Monitor over time
watch -n 5 'ps aux | grep landlord'
```

2. Check logs for goroutine creation/destruction:
```bash
# Enable debug logging to see worker lifecycle
LANDLORD_LOG_LEVEL=debug ./landlord
```

3. Profile memory usage:
```bash
# If pprof enabled, fetch profile
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof heap.prof
```

**Solutions:**

1. **Restart Controller Periodically**
   - Most immediate solution: use process manager to restart daily
   - Example in systemd: `RestartForceExitStatus=137` (OOM code)

2. **Increase Available Memory**
   - Temporary solution until leak fixed
   - Container memory limit: `CONTAINER_MEMORY_LIMIT=4G`
   - Pod memory request/limit in Kubernetes

3. **Investigate Goroutine Leaks**
   - Check for goroutines that never terminate
   - Ensure context cancellation propagates
   - Verify retry loops have termination conditions

### Issue: Controller Doesn't Start

**Symptoms:**
- Application exits immediately on startup
- No error message or logs produced
- Process crashes silently

**Root Cause:**
- Configuration errors
- Database not accessible
- Port already in use
- Missing required dependencies

**Diagnosis:**

1. Check startup logs:
```bash
./landlord 2>&1 | head -50  # Capture both stdout and stderr
```

2. Verify configuration:
```bash
# Check for config file
ls -la config.yaml

# Validate YAML syntax
cat config.yaml | python3 -c "import sys, yaml; yaml.safe_load(sys.stdin)"
```

3. Check port availability:
```bash
netstat -tuln | grep 8080  # Check if port in use
```

**Solutions:**

1. **Configuration Error**
   - Review `config.yaml` for syntax errors
   - Ensure all required fields present
   - Check environment variables override correctly

2. **Database Connection**
   ```bash
   # Test database connectivity
   psql -h <db-host> -U <db-user> -d <db-name> -c "SELECT 1"
   
   # Check database URL in config
   ```
   - Verify database credentials
   - Ensure database network accessibility
   - Check firewall rules

3. **Port Already In Use**
   ```bash
   # Find process using port
   lsof -i :8080
   
   # Kill existing process or change port
   ```

### Issue: Workflow Not Restarting After Config Update

**Symptoms:**
- Updated tenant configuration via API (e.g., `PUT /tenants/{id}` or `cli set --tenant-name X --config ...`)
- Workflow remains in backing-off or retrying state
- Expected: workflow should restart with new config
- Actual: workflow continues with old config

**Root Cause:**
- Config change detection requires `workflow_config_hash` field in tenant record
- Old tenants (created before config hash feature) don't have stored hash
- Workflow not in degraded state (backing-off/retrying) eligible for restart
- Config change may not have actually changed the relevant fields

**How Config Change Restart Works:**

The controller automatically restarts degraded workflows when configuration changes:

1. **Detection**: When reconciling a tenant with an in-flight workflow:
   - Controller compares current config hash with stored `workflow_config_hash`
   - If hashes differ AND workflow is in degraded state → restart triggered

2. **Degraded States** (eligible for restart):
   - `SubStateBackingOff`: Workflow backing off due to failures
   - These workflows are stuck due to config errors, restart may fix them

3. **Healthy States** (NOT restarted on config change):
   - `SubStateRunning`: Actively provisioning, shouldn't interrupt
   - `SubStateSucceeded`: Completed successfully
   - `SubStateFailed`: Terminal failure, handled separately
   - `SubStateWaiting`: Waiting for external event

4. **Restart Flow**:
   ```
   Config Change Detected
   ↓
   Workflow in degraded state?
   ↓ (yes)
   Call provider.StopExecution("Configuration updated")
   ↓
   Poll until workflow reaches terminal state (30s timeout)
   ↓
   Clear execution_id, error_message, retry_count
   ↓
   Trigger new workflow with updated config
   ↓
   Store new config_hash
   ```

**Diagnosis:**

1. Check if config hash is stored:
```bash
# Query tenant record
curl http://localhost:8080/tenants/{tenant_id} | jq '.workflow_config_hash'

# If null, tenant created before feature was added
```

2. Check workflow sub-state:
```bash
# Look at workflow_sub_state field
curl http://localhost:8080/tenants/{tenant_id} | jq '.workflow_sub_state'

# Should be "backing-off" or "retrying" for restart to trigger
```

3. Check controller logs for config change detection:
```bash
grep "config changed while workflow degraded" landlord.log
grep "restarting workflow" landlord.log
```

4. Verify config actually changed:
```bash
# Compare current config hash
echo '{"your":"config"}' | sha256sum

# vs stored hash in tenant.workflow_config_hash
```

**Solutions:**

1. **Old Tenant Without Config Hash**
   - Manually stop and restart workflow:
     ```bash
     # Stop workflow via provider (e.g., docker stop, aws stepfunctions stop-execution)
     # Controller will detect stopped state and trigger new workflow on next reconciliation
     ```
   - Or update to a newer status to trigger new workflow:
     ```bash
     # Archive then recreate tenant
     cli archive --tenant-name X
     cli create --tenant-name X --config new-config.yaml
     ```

2. **Workflow Not in Degraded State**
   - If workflow is healthy (running), wait for completion before updating config
   - If workflow is failed (terminal), update will trigger new workflow on next status change
   - If backing off, verify config change was saved:
     ```bash
     curl http://localhost:8080/tenants/{tenant_id} | jq '.desired_config'
     ```

3. **Config Change Not Detected**
   - Verify controller is reconciling tenant:
     ```bash
     grep "reconciling tenant.*{tenant_id}" landlord.log | tail -5
     ```
   - Check reconciliation interval: `CONTROLLER_RECONCILIATION_INTERVAL` (default 30s)
   - Verify workflow execution ID exists:
     ```bash
     curl http://localhost:8080/tenants/{tenant_id} | jq '.workflow_execution_id'
     ```

4. **Stop Execution Timeout**
   - Check logs for timeout errors:
     ```bash
     grep "timeout waiting for workflow to stop" landlord.log
     ```
   - Workflow provider may be slow to stop execution
   - Increase stop polling timeout in code (default 30s)
   - Verify provider's StopExecution API is working:
     ```bash
     # For Docker:
     docker stop <container-id>  # Should complete in < 10s
     
     # For AWS Step Functions:
     aws stepfunctions stop-execution --execution-arn <arn>
     ```

**Observability:**

Monitor these log messages for config change restart behavior:

```
# Config change detected
INFO config changed while workflow degraded, restarting workflow
  tenant_id=<uuid>
  execution_id=<id>
  old_config_hash=<hash>
  new_config_hash=<hash>

# Workflow stop initiated
INFO stopping workflow execution
  tenant_id=<uuid>
  execution_id=<id>

# Workflow reached terminal state
INFO workflow reached terminal state
  execution_id=<id>
  state=cancelled

# New workflow triggered
INFO new workflow triggered after config change
  tenant_id=<uuid>
  new_execution_id=<id>
  action=provision
```

**Known Limitations:**

- Config hash not retroactively added to old tenants (backward compatibility)
- Only degraded workflows are restarted (by design, to preserve in-flight work)
- Stop polling has 30-second timeout (tunable in code)
- Hash computation includes entire DesiredConfig (no selective field hashing)

## Monitoring and Prevention

### Recommended Alerts

Configure monitoring/alerting for these metrics:

1. **Queue Depth Growing** (alert if > 100 tenants for 5+ minutes)
   - Indicates reconciliation falling behind
   
2. **High Error Rate** (alert if > 10% failures)
   - Indicates systemic issue with provider or configuration

3. **Reconciliation Duration High** (alert if > 60 seconds)
   - Indicates performance degradation

4. **Controller Not Running**
   - Process health check, restart if failed

### Preventive Actions

1. **Regular Health Checks**
   - Monitor controller logs for patterns
   - Test workflow provider connectivity periodically
   - Verify database performance

2. **Capacity Planning**
   - Load test with realistic tenant counts
   - Monitor resource usage during peak times
   - Plan scaling before hitting limits

3. **Version Updates**
   - Keep Landlord updated with latest patches
   - Review release notes for bug fixes
   - Test updates in staging before production

## Getting Help

If issue persists after following this guide:

1. **Collect Diagnostics**
   ```bash
   # Collect logs, metrics, configuration
   # DO NOT include credentials
   ```

2. **File Issue** with:
   - Symptoms and when they started
   - Landlord version
   - Configuration (without credentials)
   - Relevant logs (with timestamps)
   - Metrics snapshots

3. **Check Documentation**
   - [State Machine](state-machine.md)
   - [Tenant Lifecycle](tenant-lifecycle.md)
   - [Configuration Guide](configuration.md)
