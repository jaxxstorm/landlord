## 1. Database Schema & Migrations

- [x] 1.1 Create `migrations/XXXXXX_compute_executions.sql` migration for `compute_executions` table with columns: id, execution_id (unique), tenant_id (FK), workflow_execution_id, operation_type, status, resource_ids (JSONB), error_code, error_message, created_at, updated_at
- [x] 1.2 Create `migrations/XXXXXX_compute_execution_history.sql` migration for `compute_execution_history` table with columns: id, compute_execution_id (FK), status, details (JSONB), timestamp
- [x] 1.3 Test migrations run successfully on both PostgreSQL and SQLite
- [x] 1.4 Verify foreign keys are properly enforced and indexes are created for frequently queried columns (execution_id, tenant_id, status)

## 2. Data Models & Types

- [x] 2.1 Add `ComputeExecution` struct to `internal/compute/types.go` with fields matching database schema
- [x] 2.2 Add `ComputeExecutionHistory` struct to `internal/compute/types.go`
- [x] 2.3 Add `ComputeError` struct to `internal/compute/types.go` with fields: Code, Message, IsRetriable, ProviderError
- [x] 2.4 Define status constants for compute executions (pending, running, succeeded, failed)
- [x] 2.5 Define operation type constants (provision, update, delete)

## 3. Compute Execution Repository

- [x] 3.1 Create `internal/compute/execution_repository.go` with Repository interface
- [x] 3.2 Implement `CreateComputeExecution()` method to insert a new execution record
- [x] 3.3 Implement `UpdateComputeExecution()` method to update status, resource_ids, and error fields
- [x] 3.4 Implement `GetComputeExecution()` method to retrieve execution by ID
- [x] 3.5 Implement `ListComputeExecutions()` method to list by tenant_id with filtering
- [x] 3.6 Implement `AddExecutionHistory()` method to append history records
- [x] 3.7 Wire repository into database factory (`internal/database/factory.go`)
- [x] 3.8 Write unit tests for repository methods (success, not found, database errors)

## 4. ComputeManager Enhancements

- [x] 4.1 Add compute execution repository as dependency to ComputeManager
- [x] 4.2 Add deterministic execution ID generation function `GenerateComputeExecutionID(tenantID, operationType)` to ComputeManager
- [x] 4.3 Create `ProvisionTenantWithTracking()` method that: creates ComputeExecution with 'pending', calls provider, updates to 'running', handles completion (success → 'succeeded', failure → 'failed')
- [x] 4.4 Create `UpdateTenantWithTracking()` method following same pattern
- [x] 4.5 Create `DeleteTenantWithTracking()` method following same pattern
- [x] 4.6 Add `GetComputeExecution()` passthrough method for workflows to query status
- [x] 4.7 Extract and standardize provider errors into `ComputeError` format with IsRetriable classification
- [x] 4.8 Write unit tests for tracking methods (success paths, error paths, state transitions)

## 5. Workflow Integration - ComputeWorkflowClient

- [x] 5.1 Create `internal/workflow/compute_client.go` with ComputeWorkflowClient struct
- [x] 5.2 Add dependency injection of ComputeManager to ComputeWorkflowClient
- [x] 5.3 Add `ProvisionTenant()` method that calls ComputeManager and returns execution ID
- [x] 5.4 Add `UpdateTenant()` method that calls ComputeManager and returns execution ID
- [x] 5.5 Add `DeleteTenant()` method that calls ComputeManager and returns execution ID
- [x] 5.6 Add `GetComputeExecutionStatus()` method for polling execution status
- [x] 5.7 Add error handling to map provider errors to standardized ComputeError responses
- [x] 5.8 Write unit tests for ComputeWorkflowClient methods

## 6. Callback Delivery System

- [x] 6.1 Define `CallbackPayload` struct for compute operation results
- [x] 6.2 Create `WorkflowProvider.PostComputeCallback()` interface method in workflow provider layer
- [x] 6.3 Implement `PostComputeCallback()` for Step Functions provider: construct task success/failure input payload
- [x] 6.4 Implement `PostComputeCallback()` for Restate provider (if applicable)
- [x] 6.5 Add callback retry logic with exponential backoff (up to 3 retries)
- [x] 6.6 Store failed callbacks durably for manual retry/investigation
- [x] 6.7 Write tests for callback delivery (success, retry scenarios)

## 7. ComputeManager Integration with Callbacks

- [x] 7.1 Update ComputeManager to post callbacks after compute operations complete
- [x] 7.2 On successful provision/update/delete: post success callback with resource IDs
- [x] 7.3 On failed provision/update/delete: post failure callback with error details and retry eligibility
- [x] 7.4 Handle callback delivery failures gracefully (log, store for retry, don't crash ComputeManager)
- [x] 7.5 Integration tests for compute operation → callback flow

## 8. WorkflowClient Integration

- [x] 8.1 Update `internal/workflow/workflow_client.go` to expose compute operations
- [x] 8.2 Add `ProvisionTenant()` wrapper to WorkflowClient that delegates to ComputeWorkflowClient
- [x] 8.3 Add `UpdateTenant()` wrapper to WorkflowClient
- [x] 8.4 Add `DeleteTenant()` wrapper to WorkflowClient
- [x] 8.5 Add `GetComputeExecutionStatus()` wrapper to WorkflowClient
- [x] 8.6 Ensure WorkflowClient is injected into all necessary places (API handlers, controller)

## 9. Main Application Wiring

- [x] 9.1 Update `cmd/landlord/main.go` to instantiate compute execution repository
- [x] 9.2 Pass repository to ComputeManager constructor
- [x] 9.3 Instantiate ComputeWorkflowClient with ComputeManager
- [x] 9.4 Wire ComputeWorkflowClient into WorkflowClient
- [x] 9.5 Wire compute callbacks into workflow providers
- [x] 9.6 Add feature flag `ENABLE_COMPUTE_WORKFLOWS` for gradual rollout (default: false)

## 10. Integration Testing

- [x] 10.1 Verify workflow-triggered compute operations create execution records
- [x] 10.2 Verify compute execution transitions from pending→running→succeeded/failed
- [x] 10.3 Verify compute errors are properly classified (retriable vs non-retriable)
- [x] 10.4 Verify callback payload construction with execution details
- [x] 10.5 Verify workflow execution status queries work mid-operation
- [x] 10.6 Verify deterministic execution IDs enable idempotency
- [x] 10.7 Core integration flow: API creates tenant → workflow triggers → compute provisions

## 11. Error Scenario Validation

- [x] 11.1 Provider timeout errors classified as retriable
- [x] 11.2 Invalid config errors classified as non-retriable
- [x] 11.3 Callback errors logged gracefully without crashing manager
- [x] 11.4 Execution status queries handle not-found cases
- [x] 11.5 Concurrent execution tracking works independently per tenant

## 12. Documentation & Cleanup

- [x] 12.1 Code comments on execution flow and tracking methods
- [x] 12.2 ComputeExecution data model documented in types.go
- [x] 12.3 Callback payload structure documented in callback.go
- [x] 12.4 Unit tests provide usage examples for common scenarios
- [x] 12.5 Repository pattern consistent with tenant repository
- [x] 12.6 Compute provisioning workflow integration complete
- [x] 12.7 All new code follows project conventions and patterns
- [x] 12.8 Core compute provisioning workflow ready for production testing

