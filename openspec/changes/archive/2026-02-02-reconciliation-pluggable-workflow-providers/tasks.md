## 1. Align workflow provider interface with reconciliation model

- [x] 1.1 MODIFY `internal/workflow/provider.go`: add Invoke(workflowName, request) method to Provider interface
- [x] 1.2 MODIFY `internal/workflow/provider.go`: add GetWorkflowStatus(executionID) method returning simplified status
- [x] 1.3 Add ProvisionRequest struct (tenantID, computeProvider, desiredConfig) to provider.go
- [x] 1.4 Update workflow.Manager to support both old and new provider methods (compatibility layer)
- [x] 1.5 Update workflow.Registry to handle provider interface changes

## 2. Update Restate provider for reconciliation

- [x] 2.1 MODIFY `internal/workflow/providers/restate/provider.go`: implement new Invoke() method
- [x] 2.2 MODIFY `internal/workflow/providers/restate/provider.go`: implement GetWorkflowStatus() method (executionID)
- [x] 2.3 Map existing CreateWorkflow/StartExecution to new Invoke() internally
- [x] 2.4 Map existing GetExecutionStatus to new GetWorkflowStatus() format (executionID)

## 3. Extend existing reconciler for workflow invocation

- [x] 3.1 MODIFY `internal/controller/reconciler.go`: add workflow invocation logic for new tenants
- [x] 3.2 MODIFY `internal/controller/reconciler.go`: add status polling loop for in-flight workflows
- [x] 3.3 Update reconciler to use Provider.Invoke() when tenant.Status == StatusRequested
- [x] 3.4 Update reconciler to use Provider.GetWorkflowStatus() for StatusProvisioning tenants (by executionID)
- [x] 3.5 Add reconciler configuration for polling intervals

## 4. Align tenant model with workflow state

- [x] 4.1 Use existing tenant.Status (StatusRequested → StatusProvisioning → StatusReady) for workflow state
- [x] 4.2 MODIFY tenant.DesiredConfig to include compute provider configuration (no new DB field)
- [x] 4.3 MODIFY tenant.ObservedConfig to store compute results from workflows (no new DB field)
- [x] 4.4 Update tenant repository queries to filter by Status for reconciler loops
- [x] 4.5 Document mapping: tenant.Status is authoritative, workflow provider status updates it

## 5. Remove DB from Restate worker

- [x] 5.1 MODIFY `cmd/workers/restate/main.go`: remove tenant repository initialization (lines 88-95)
- [x] 5.2 MODIFY `cmd/workers/restate/main.go`: remove DB pool access from worker setup
- [x] 5.3 MODIFY Restate worker engine: remove tenantRepo parameter from NewWorkerEngine()
- [x] 5.4 Update worker to receive full tenant context from Restate workflow payload

## 6. Update worker handler to be stateless

- [x] 6.1 MODIFY `internal/workflow/providers/restate/tenant_service.go`: change Execute handler to accept ProvisionRequest
- [x] 6.2 Remove tenant repository queries from worker handler
- [x] 6.3 Use request payload (desiredConfig, computeProvider) instead of DB lookups
- [x] 6.4 Return compute results via Restate workflow context (not direct DB writes)
- [x] 6.5 Ensure Landlord API client is still available for optional status callbacks

## 7. Update Step Functions provider

- [x] 7.1 MODIFY `internal/workflow/providers/stepfunctions/provider.go`: implement new Invoke() method
- [x] 7.2 MODIFY `internal/workflow/providers/stepfunctions/provider.go`: implement GetWorkflowStatus() method
- [x] 7.3 Map existing Step Functions calls to new interface methods

## 8. Wire reconciler into Landlord API

- [x] 8.1 MODIFY `cmd/landlord/main.go`: ensure reconciler is started (already exists, verify config)
- [x] 8.2 Update reconciler config to specify workflow provider (restate vs step-functions)
- [x] 8.3 Verify reconciler uses workflow.Manager which wraps provider interface

## 9. Update all workflow provider call sites

- [x] 9.1 Update `internal/api/tenants.go`: set Status=StatusRequested and remove direct workflow trigger
- [x] 9.2 Update workflow.Manager call sites to use new provider methods
- [x] 9.3 Update any direct provider.CreateWorkflow() calls to use Manager.Invoke()
- [x] 9.4 Verify state machine transitions align with new reconciler logic

## 10. Tests and docs

- [x] 10.1 Add unit tests for new Provider.Invoke()/GetWorkflowStatus() methods
- [x] 10.2 Add unit tests for reconciler workflow invocation logic
- [x] 10.3 Add integration test for end-to-end provisioning via reconciliation
- [x] 10.4 Update README/docs: reconciler architecture, no worker DB access
- [x] 10.5 Add migration guide for existing deployments
