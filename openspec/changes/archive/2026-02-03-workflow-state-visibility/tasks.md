## 1. Database schema updates

- [x] 1.1 Add migration to create workflow status columns on tenants table (workflow_sub_state, workflow_retry_count, workflow_error_message)
- [x] 1.2 Add down migration to drop workflow status columns
- [x] 1.3 Verify migrations are embedded and applied on startup

## 2. Tenant domain model updates

- [x] 2.1 Add WorkflowSubState, WorkflowRetryCount, WorkflowErrorMessage fields to tenant.Tenant struct with JSON tags
- [x] 2.2 Update tenant repository scan/insert/update mappings to include new fields
- [x] 2.3 Add repository query filters for workflow_sub_state, min_retry_count, and error presence

## 3. Workflow state mapping

- [x] 3.1 Define canonical workflow sub-state enum (running, waiting, backing-off, error, succeeded, failed)
- [x] 3.2 Implement state mapping helper for Step Functions execution states
- [x] 3.3 Implement state mapping helper for Restate invocation states
- [x] 3.4 Add retry/backoff detection logic for each provider where metadata is available
- [x] 3.5 Add unit tests for mapping helpers and unknown state fallback

## 4. Reconciler status enrichment

- [x] 4.1 Extend workflow ExecutionStatus handling to extract sub-state, retry count, and error message
- [x] 4.2 Add dirty-checking in reconciler to avoid updates when status fields unchanged
- [x] 4.3 Persist workflow status fields on tenant update during polling
- [x] 4.4 Clear or preserve workflow status fields on terminal success/failure per spec
- [x] 4.5 Add structured logs for sub-state transitions, retry increments, and error updates

## 5. API response updates

- [x] 5.1 Update GET /v1/tenants/{id} response to include workflow status fields
- [x] 5.2 Update LIST /v1/tenants response to include workflow status fields
- [x] 5.3 Add list query filters for workflow_sub_state, has_workflow_error, min_retry_count
- [x] 5.4 Update OpenAPI/Swagger schemas for tenant responses and list query parameters

## 6. CLI output updates

- [x] 6.1 Update CLI tenant list output to show workflow sub-state and retry count when present
- [x] 6.2 Update CLI tenant get output to show workflow sub-state, retry count, and error message

## 7. Tests

- [x] 7.1 Add unit tests for tenant repository workflow status persistence
- [x] 7.2 Add reconciler tests for sub-state extraction and dirty-checking behavior
- [x] 7.3 Add API tests for GET/LIST returning workflow status fields
- [x] 7.4 Add API tests for list filtering by workflow_sub_state and error presence

## 8. Documentation

- [x] 8.1 Update workflow provider interface docs with state mapping table
- [x] 8.2 Update API docs to describe workflow status fields in tenant responses
