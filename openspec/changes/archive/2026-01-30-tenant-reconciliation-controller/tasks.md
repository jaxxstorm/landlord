## 1. Project Setup

- [x] 1.1 Add k8s.io/client-go and k8s.io/apimachinery dependencies to go.mod
- [x] 1.2 Create internal/controller package directory structure
- [x] 1.3 Add controller configuration to internal/config/config.go (reconciliation_interval, workers, timeouts)
- [x] 1.4 Document controller configuration in docs/configuration.md

## 2. Repository Layer

- [x] 2.1 Add ListTenantsForReconciliation() method to internal/tenant/repository.go interface
- [x] 2.2 Implement ListTenantsForReconciliation() in internal/tenant/postgres/repository.go to query non-terminal states
- [x] 2.3 Add unit tests for ListTenantsForReconciliation() with multiple status scenarios
- [x] 2.4 Verify query performance with database indexes on status column

## 3. Workqueue Integration

- [x] 3.1 Create internal/controller/queue.go with NewRateLimitingQueue() function
- [x] 3.2 Configure exponential backoff rate limiter (1s base, 5min max)
- [x] 3.3 Add workqueue wrapper methods (Add, Get, Done, AddRateLimited, ShutDown)
- [x] 3.4 Write unit tests for queue initialization and basic operations
- [x] 3.5 Add queue depth metrics hooks (prepare for future Prometheus integration)

## 4. Core Reconciler Logic

- [x] 4.1 Create internal/controller/reconciler.go with Reconciler struct
- [x] 4.2 Implement NewReconciler() constructor with dependencies (repo, workflow manager, queue, config)
- [x] 4.3 Implement Start() method to launch polling loop and workers
- [x] 4.4 Implement pollLoop() method to query database at configured interval
- [x] 4.5 Implement runWorker() method to process items from queue
- [x] 4.6 Implement reconcile() method with tenant fetch, status validation, and workflow triggering
- [x] 4.7 Add graceful shutdown with Stop() method and context cancellation
- [x] 4.8 Implement shutdown timeout (default 30s) with forced exit

## 5. State Machine Implementation

- [x] 5.1 Create internal/controller/state_machine.go with state transition logic
- [x] 5.2 Implement nextStatus() function for valid transitions (requested→planning, planning→provisioning, provisioning→ready)
- [x] 5.3 Implement shouldReconcile() function to check if tenant status requires action
- [x] 5.4 Add validation to reject invalid status transitions
- [x] 5.5 Write unit tests for all valid and invalid state transitions
- [x] 5.6 Document state machine flow diagram in docs/

## 6. Workflow Integration

- [x] 6.1 Create internal/controller/workflow_client.go wrapper around workflow manager
- [x] 6.2 Implement triggerWorkflow() method with action determination (plan, provision, update, delete)
- [x] 6.3 Add workflow trigger timeout (default 30s) with context
- [x] 6.4 Implement workflow response handling (success→next status, retryable error→requeue, fatal error→failed status)
- [x] 6.5 Add workflow provider readiness check before triggering
- [x] 6.6 Store workflow execution ID in tenant record
- [x] 6.7 Write integration tests with mock workflow provider

## 7. Error Handling and Retry

- [x] 7.1 Implement error classification (retryable vs fatal)
- [x] 7.2 Add retry counter tracking per tenant reconciliation
- [x] 7.3 Implement max retry limit (default 5) with failure state transition
- [x] 7.4 Add exponential backoff timing to requeued items
- [x] 7.5 Handle tenant not found scenarios gracefully
- [x] 7.6 Write unit tests for retry scenarios and backoff behavior

## 8. Observability

- [x] 8.1 Add structured logging to internal/controller/logging.go
- [x] 8.2 Emit logs for reconciliation start (tenant_id, status)
- [x] 8.3 Emit logs for reconciliation success (tenant_id, prev_status, new_status, duration)
- [x] 8.4 Emit logs for reconciliation errors (tenant_id, error, retry_count)
- [x] 8.5 Emit logs for shutdown events (active_workers, queued_items)
- [x] 8.6 Add metrics hooks for future Prometheus integration (reconciliation_duration, queue_depth, error_rate)

## 9. Controller Lifecycle

- [x] 9.1 Update cmd/landlord/main.go to initialize controller with dependencies
- [x] 9.2 Start controller after database migrations and before HTTP server
- [x] 9.3 Add controller health check endpoint integration
- [x] 9.4 Implement graceful shutdown coordination between HTTP server and controller
- [x] 9.5 Add controller readiness check (workqueue initialized, workers started)

## 10. Testing

- [x] 10.1 Write unit tests for reconciler with mock repository and workflow manager
- [x] 10.2 Write integration tests with test database and mock workflow provider
- [x] 10.3 Test concurrent worker behavior with multiple tenants
- [ ] 10.4 Test graceful shutdown with in-flight reconciliations
- [ ] 10.5 Test error scenarios (database unavailable, workflow provider down)
- [ ] 10.6 Test backoff and retry behavior with simulated failures
- [ ] 10.7 Load test with 100+ tenants to validate polling efficiency

## 11. Documentation

- [x] 11.1 Update README.md with controller architecture overview
- [x] 11.2 Document controller configuration options in docs/configuration.md
- [x] 11.3 Add state machine diagram to docs/
- [x] 11.4 Document reconciliation flow in docs/tenant-lifecycle.md
- [x] 11.5 Add troubleshooting guide for common controller issues
- [x] 11.6 Update CONTRIBUTING.md with controller development guidelines

## 12. Validation

- [x] 12.1 Create test tenant via API and verify automatic reconciliation
- [x] 12.2 Verify tenant transitions through states (requested→planning→provisioning→ready)
- [ ] 12.3 Verify workflow provider is triggered with correct actions
- [ ] 12.4 Test failure scenarios and verify retry behavior
- [ ] 12.5 Test graceful shutdown during active reconciliations
- [ ] 12.6 Verify no memory leaks with long-running controller
