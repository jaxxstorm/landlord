## 1. Worker Engine Abstractions

- [x] 1.1 Add workflow worker engine interface and registry in the workflow subsystem
- [x] 1.2 Wire worker engine selection to configured workflow engine name
- [x] 1.3 Define worker job payload model for tenant lifecycle operations

## 2. Restate Worker Engine Implementation

- [x] 2.1 Implement restate worker engine registration/startup with Restate admin API integration
- [x] 2.2 Add worker configuration loading/validation for admin endpoint, namespace, and auth
- [x] 2.3 Implement restate worker handlers for tenant create/update/delete workflows

## 3. Tenant Lifecycle Workflow Support

- [x] 3.1 Extend workflow execution path to include update and delete operations
- [x] 3.2 Add status reporting for create/update/delete lifecycle transitions
- [x] 3.3 Ensure lifecycle workflows are idempotent and share reconciliation logic

## 4. Compute Engine Resolution

- [x] 4.1 Add worker-side Landlord API client for compute engine lookup
- [x] 4.2 Cache compute engine resolution with TTL and allow config override for tests

## 5. Restate Registration Updates

- [x] 5.1 Update restate workflow provider configuration struct to include worker registration fields
- [x] 5.2 Ensure workflows and worker services register on startup when configured

## 6. Testing and Validation

- [x] 6.1 Add Restate end-to-end test covering tenant create, update, and delete via worker
- [x] 6.2 Add unit tests for worker engine selection and configuration validation
- [x] 6.3 Document local dev workflow for running Restate worker with auto-registration
