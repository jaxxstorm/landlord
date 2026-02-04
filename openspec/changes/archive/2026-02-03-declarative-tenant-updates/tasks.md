## 1. Config Hash Utilities

- [x] 1.1 Create computeConfigHash function that computes SHA256 hash of tenant compute_config JSON
- [x] 1.2 Add unit tests for hash computation with identical configs producing same hash
- [x] 1.3 Add unit tests for different configs producing different hashes
- [x] 1.4 Add unit test for null/empty config producing consistent hash

## 2. Workflow Metadata Enhancement

- [x] 2.1 Update workflowClient.TriggerWorkflow to compute config hash before starting workflow
- [x] 2.2 Add config_hash to ExecutionInput.Metadata when calling StartExecution
- [x] 2.3 Add unit test verifying config_hash is included in workflow metadata
- [x] 2.4 Update integration tests to verify metadata contains config_hash

## 3. Degraded Workflow Detection

- [x] 3.1 Create isDegradedWorkflow helper function checking for SubStateBackingOff or SubStateRetrying
- [x] 3.2 Add unit tests for degraded state classification (backing-off, retrying = degraded)
- [x] 3.3 Add unit tests verifying running/succeeded/failed states are NOT degraded
- [x] 3.4 Document degraded states in code comments

## 4. Config Change Detection in Reconciler

- [x] 4.1 Extract config_hash from ExecutionStatus.Metadata in reconciler status polling
- [x] 4.2 Add helper function to compare current tenant config hash with workflow metadata hash
- [x] 4.3 Handle missing config_hash gracefully (old executions without hash)
- [x] 4.4 Add unit tests for config change detection logic
- [x] 4.5 Add logging when config change is detected (include old and new hashes)

## 5. Workflow Stop on Config Change

- [x] 5.1 Add logic in reconciler to call provider.StopExecution when config changed + workflow degraded
- [x] 5.2 Pass "Configuration updated" as stop reason
- [x] 5.3 Add polling loop to wait for workflow State == StateDone after stop
- [x] 5.4 Add timeout for stop polling with error handling
- [x] 5.5 Add unit tests for stop logic with mock workflow provider
- [ ] 5.6 Add integration test for stopping backing-off workflow on config change

## 6. Workflow Restart After Stop

- [x] 6.1 Clear tenant.WorkflowExecutionID after successful workflow stop
- [x] 6.2 Update tenant record in database with cleared execution ID
- [x] 6.3 Call workflowClient.TriggerWorkflow with tenant and "update" operation
- [x] 6.4 Update tenant.WorkflowExecutionID with new execution ID
- [x] 6.5 Reset tenant.WorkflowRetryCount to 0 for fresh workflow
- [x] 6.6 Clear tenant.WorkflowErrorMessage on restart
- [x] 6.7 Add unit tests for restart sequence
- [ ] 6.8 Add integration test for full stop → start → new execution flow

## 7. State Machine Updates

- [x] 7.1 Verify shouldReconcile includes checking for active workflows eligible for restart
- [x] 7.2 Add reconciler state handling for config-change restart scenario
- [x] 7.3 Update reconciler to skip config change check if workflow execution ID is empty
- [x] 7.4 Add unit tests for reconciler state transitions with config changes

## 8. Preserve Healthy Workflows

- [x] 8.1 Add guard condition: skip restart if workflow is in SubStateRunning
- [x] 8.2 Add guard condition: skip restart if workflow State == StateDone (terminal)
- [x] 8.3 Add unit tests verifying running workflows are NOT restarted on config change
- [x] 8.4 Add unit tests verifying completed workflows are NOT restarted on config change
- [ ] 8.5 Add integration test confirming healthy workflow continues with config change

## 9. Error Handling and Edge Cases

- [x] 9.1 Handle StopExecution failure with retry on next reconciliation pass
- [x] 9.2 Handle TriggerWorkflow failure after stop with proper error logging
- [x] 9.3 Add timeout for stop polling to prevent infinite wait
- [x] 9.4 Handle race condition where workflow completes during stop attempt
- [x] 9.5 Add unit tests for error scenarios

## 10. Logging and Observability

- [x] 10.1 Add structured logging when config change detected (include tenant ID, old hash, new hash)
- [x] 10.2 Add structured logging when stopping workflow due to config change
- [x] 10.3 Add structured logging when starting new workflow after config change
- [ ] 10.4 Add metrics/counters for config-based workflow restarts (optional)
- [x] 10.5 Update reconciler logging to distinguish config-change restarts from other restarts

## 11. Integration Tests

- [ ] 11.1 Test: create tenant with bad config → workflow backs off → update config → verify restart
- [ ] 11.2 Test: create tenant with bad config → workflow backs off → update unrelated field → verify NO restart
- [ ] 11.3 Test: create tenant → workflow running → update config → verify NO restart (workflow continues)
- [ ] 11.4 Test: create tenant → workflow succeeds → update config → verify NO restart
- [ ] 11.5 Test: verify new workflow includes updated config_hash in metadata
- [ ] 11.6 Test: multiple rapid config changes don't cause duplicate restarts

## 12. Documentation

- [x] 12.1 Update controller troubleshooting docs with config change restart behavior
- [x] 12.2 Add code comments explaining config hash computation and comparison
- [x] 12.3 Update API documentation noting that config updates trigger async workflow restart
- [x] 12.4 Document reconciler behavior for config-change scenarios
