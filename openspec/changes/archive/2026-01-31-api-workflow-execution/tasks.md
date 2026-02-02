## 1. State Machine Refactoring (Preparation Phase)

- [x] 1.1 Create internal/tenant/state_machine.go package file
- [x] 1.2 Extract nextStatus() function from internal/controller/state_machine.go to internal/tenant/state_machine.go
- [x] 1.3 Extract shouldReconcile() function from internal/controller/state_machine.go to internal/tenant/state_machine.go
- [x] 1.4 Extract ValidateTransition() function from internal/controller/state_machine.go to internal/tenant/state_machine.go
- [x] 1.5 Update internal/controller/reconciler.go to import and use tenant.NextStatus() and tenant.ShouldReconcile()
- [x] 1.6 Update internal/controller/state_machine.go to import and use shared state machine functions
- [x] 1.7 Add unit tests for tenant.NextStatus() covering all valid transitions (requested→planning, planning→provisioning, etc.)
- [x] 1.8 Add unit tests for tenant.ShouldReconcile() covering terminal and non-terminal states
- [x] 1.9 Add unit tests for tenant.ValidateTransition() covering invalid transition rejection
- [x] 1.10 Run existing controller tests to verify no regressions from refactoring

## 2. API Server Workflow Client Integration

- [x] 2.1 Add workflowClient field to Server struct in internal/api/server.go
- [x] 2.2 Update New() constructor in internal/api/server.go to accept *controller.WorkflowClient parameter
- [x] 2.3 Store workflow client in Server.workflowClient field during initialization
- [x] 2.4 Update cmd/landlord/main.go to create WorkflowClient instance
- [x] 2.5 Pass WorkflowClient to API server constructor in cmd/landlord/main.go (after controller, before HTTP server start)
- [x] 2.6 Add nil check for workflowClient in API handlers (graceful degradation if not provided)
- [x] 2.7 Update API server unit tests to pass mock workflow client
- [x] 2.8 Verify build succeeds with new dependency (go build ./...)

## 3. POST /api/tenants Workflow Integration

- [x] 3.1 Update handleCreateTenant to set tenant status to "planning" instead of "requested"
- [x] 3.2 Add workflow trigger call after database commit: s.workflowClient.TriggerWorkflow(ctx, tenant, "plan")
- [x] 3.3 Store returned execution ID in tenant.WorkflowExecutionID field
- [x] 3.4 Update tenant record with execution ID in same transaction as initial create
- [x] 3.5 Change response status code from 200 to 202 Accepted
- [x] 3.6 Add workflow_execution_id field to JSON response
- [x] 3.7 Handle workflow trigger error: return HTTP 500 with error message "Failed to trigger provisioning workflow"
- [x] 3.8 Log workflow trigger success/failure with tenant_id and execution_id
- [x] 3.9 Add unit test for successful tenant creation with workflow trigger
- [x] 3.10 Add unit test for workflow trigger failure (provider unavailable)
- [x] 3.11 Add unit test for invalid tenant specification (400 Bad Request, no workflow trigger)

## 4. PUT /api/tenants/{id} Workflow Integration

- [x] 4.1 Add state machine validation in handleUpdateTenant using tenant.ValidateTransition()
- [x] 4.2 Determine next status using tenant.NextStatus(currentStatus, "update")
- [x] 4.3 Update tenant status to "updating" (or appropriate next status) in database transaction
- [x] 4.4 Add workflow trigger call after database commit: s.workflowClient.TriggerWorkflow(ctx, tenant, "update")
- [x] 4.5 Store returned execution ID in tenant.WorkflowExecutionID field
- [x] 4.6 Update tenant record with execution ID
- [x] 4.7 Change response status code from 200 to 202 Accepted
- [x] 4.8 Add workflow_execution_id field to JSON response
- [x] 4.9 Return HTTP 409 Conflict for invalid state transitions (e.g., updating terminal status tenant)
- [x] 4.10 Return HTTP 410 Gone for deleted tenants
- [x] 4.11 Handle workflow trigger error: return HTTP 500 with error message
- [x] 4.12 Add unit test for successful tenant update with workflow trigger
- [x] 4.13 Add unit test for invalid state transition rejection (409 Conflict)
- [x] 4.14 Add unit test for updating deleted tenant (410 Gone)

## 5. DELETE /api/tenants/{id} Workflow Integration

- [x] 5.1 Update handleDeleteTenant to set tenant status to "deleting" instead of immediate deletion
- [x] 5.2 Add workflow trigger call after status update: s.workflowClient.TriggerWorkflow(ctx, tenant, "delete")
- [x] 5.3 Store returned execution ID in tenant.WorkflowExecutionID field
- [x] 5.4 Update tenant record with execution ID and "deleting" status
- [x] 5.5 Change response status code from 204 No Content to 202 Accepted
- [x] 5.6 Return tenant record in response body (instead of empty response) with workflow_execution_id
- [x] 5.7 Return HTTP 410 Gone for already deleted tenants without triggering workflow
- [x] 5.8 Handle workflow trigger error: return HTTP 500 with error message
- [x] 5.9 Add unit test for successful tenant deletion with workflow trigger
- [x] 5.10 Add unit test for deleting non-existent tenant (404 Not Found)
- [x] 5.11 Add unit test for deleting already deleted tenant (410 Gone)

## 6. GET Endpoints Workflow Status Support

- [x] 6.1 Verify GET /api/tenants returns workflow_execution_id field in response (already in tenant model)
- [x] 6.2 Verify GET /api/tenants/{id} returns workflow_execution_id field in response
- [x] 6.3 Add unit test for GET returning tenant with active workflow (non-null execution ID)
- [x] 6.4 Add unit test for GET returning tenant without active workflow (null execution ID)

## 7. Controller Deduplication Logic

- [x] 7.1 Update controller reconcile() to check if workflow_execution_id is non-null before triggering
- [x] 7.2 Add workflow status check: call workflowClient.GetExecutionStatus(executionID) if execution ID exists
- [x] 7.3 Skip workflow trigger if execution status is "running" or "pending"
- [x] 7.4 Re-trigger workflow with new execution ID if status is "completed" or "failed" and tenant still needs reconciliation
- [x] 7.5 Log deduplication events: "skipping trigger, workflow already active" with execution_id
- [x] 7.6 Log re-trigger events: "re-triggering after workflow completion/failure" with old and new execution IDs
- [x] 7.7 Add unit test for controller skipping tenant with active workflow
- [x] 7.8 Add unit test for controller re-triggering after workflow completion
- [x] 7.9 Add unit test for controller re-triggering after workflow failure

## 8. Workflow Execution ID Format

- [x] 8.1 Verify WorkflowClient generates deterministic execution IDs: "tenant-{tenant_id}-{action}"
- [x] 8.2 Update WorkflowClient.TriggerWorkflow() to use tenant_id in execution ID format if not already
- [x] 8.3 Add unit test verifying execution ID format for plan action: "tenant-my-app-plan"
- [x] 8.4 Add unit test verifying execution ID format for provision action: "tenant-my-app-provision"
- [x] 8.5 Add unit test verifying execution ID format for update action: "tenant-my-app-update"
- [x] 8.6 Add unit test verifying execution ID format for delete action: "tenant-my-app-delete"

## 9. Workflow Provider Idempotency

- [x] 9.1 Verify mock workflow provider StartExecution handles duplicate execution IDs idempotently
- [x] 9.2 Add test for workflow provider receiving duplicate StartExecution calls (returns existing execution)
- [x] 9.3 Document idempotency requirements in workflow provider interface comments
- [x] 9.4 Update Restate provider (if applicable) to handle duplicate execution IDs idempotently
- [x] 9.5 Update Step Functions provider (if applicable) to handle duplicate execution IDs idempotently

## 10. Trigger Source Tracking

- [x] 10.1 Add trigger_source field to ExecutionInput in internal/workflow/types.go
- [x] 10.2 Update WorkflowClient.TriggerWorkflow() to accept trigger source parameter (default "controller")
- [x] 10.3 Update API handlers to pass trigger_source="api" when calling TriggerWorkflow
- [x] 10.4 Update controller reconciler to pass trigger_source="controller" when calling TriggerWorkflow
- [x] 10.5 Log trigger source alongside execution ID in workflow manager
- [x] 10.6 Add unit test verifying API triggers include trigger_source="api"
- [x] 10.7 Add unit test verifying controller triggers include trigger_source="controller"

## 11. Error Handling and Status Codes

- [x] 11.1 Add error response helper for HTTP 500: "Failed to trigger workflow"
- [x] 11.2 Add error response helper for HTTP 400: "Invalid workflow specification"
- [x] 11.3 Add error response helper for HTTP 409: "Invalid state transition"
- [x] 11.4 Add error response helper for HTTP 410: "Tenant deleted"
- [x] 11.5 Update API handlers to use appropriate error responses for workflow trigger failures
- [x] 11.6 Add structured logging for all error scenarios (tenant_id, error, status_code)
- [x] 11.7 Add unit test for workflow provider unavailable (500 Internal Server Error)
- [x] 11.8 Add unit test for workflow trigger timeout (500 Internal Server Error)
- [x] 11.9 Add unit test for invalid workflow specification (400 Bad Request)

## 12. Integration Tests

- [x] 12.1 Create integration test: POST /api/tenants triggers plan workflow and returns 202 with execution ID
- [x] 12.2 Create integration test: PUT /api/tenants/{id} triggers update workflow and returns 202 with execution ID
- [x] 12.3 Create integration test: DELETE /api/tenants/{id} triggers delete workflow and returns 202 with execution ID
- [x] 12.4 Create integration test: API trigger followed by controller poll shows controller skips (deduplication)
- [x] 12.5 Create integration test: API trigger fails, controller retries successfully
- [x] 12.6 Create integration test: Workflow completes, controller re-triggers if status still requires reconciliation
- [x] 12.7 Create integration test: Concurrent API and controller triggers result in single workflow execution
- [x] 12.8 Run all integration tests with testcontainers and mock workflow provider

## 13. Documentation

- [x] 13.1 Update API documentation for POST /api/tenants: mention 202 status code, workflow_execution_id field
- [x] 13.2 Update API documentation for PUT /api/tenants/{id}: mention 202 status code, workflow_execution_id field
- [x] 13.3 Update API documentation for DELETE /api/tenants/{id}: mention 202 status code, changed response format
- [x] 13.4 Document workflow trigger error scenarios in API documentation
- [x] 13.5 Update CHANGELOG.md with breaking changes (202 status codes, DELETE response format)
- [x] 13.6 Document deduplication strategy in docs/tenant-lifecycle.md
- [x] 13.7 Update README.md with API workflow triggering behavior
- [x] 13.8 Update Swagger/OpenAPI spec with new response codes and workflow_execution_id field

## 14. Observability

- [x] 14.1 Add log event for API workflow trigger success (tenant_id, execution_id, trigger_source)
- [x] 14.2 Add log event for API workflow trigger failure (tenant_id, error, trigger_source)
- [x] 14.3 Add log event for controller deduplication (tenant_id, execution_id, skipped reason)
- [x] 14.4 Add metrics hook for workflow trigger duration (prepare for Prometheus)
- [x] 14.5 Add metrics hook for workflow trigger errors count (prepare for Prometheus)
- [x] 14.6 Add metrics hook for duplicate triggers prevented count (prepare for Prometheus)
- [x] 14.7 Document metrics in docs/configuration.md

## 15. Validation and Testing

- [x] 15.1 Run all unit tests: go test ./internal/api/...
- [x] 15.2 Run all controller tests: go test ./internal/controller/...
- [x] 15.3 Run all integration tests: go test ./... with integration tag
- [x] 15.4 Verify build succeeds: go build ./...
- [x] 15.5 Manual test: Create tenant via API, verify workflow triggered immediately
- [x] 15.6 Manual test: Update tenant via API, verify workflow triggered immediately
- [x] 15.7 Manual test: Delete tenant via API, verify workflow triggered immediately
- [x] 15.8 Manual test: Controller poll after API trigger shows skip behavior
- [x] 15.9 Load test: 100+ concurrent API requests, verify workflow triggering performance
- [x] 15.10 Verify no memory leaks with long-running API server and workflow triggering
