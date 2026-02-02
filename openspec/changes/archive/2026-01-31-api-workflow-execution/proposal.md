## Why

The landlord API currently creates tenant records in the database but doesn't trigger actual provisioning workflows. Users can POST to create tenants, but those tenants remain in "requested" state until the reconciliation controller picks them up (up to 10s delay). Similarly, PATCH and DELETE operations update database records without triggering corresponding infrastructure changes. This creates a disconnect between API semantics (RESTful operations should complete the action) and actual behavior (database mutation only).

The reconciliation controller (already implemented) handles eventual consistency and retries, but the API should immediately trigger workflows for responsive user experience and clear API contracts.

## What Changes

- **API handlers trigger workflows**: POST /tenants triggers provisioning workflow, PATCH triggers update workflow, DELETE triggers deletion workflow
- **Immediate workflow execution**: API responses include workflow execution ID, don't wait for controller polling
- **Status tracking integration**: API updates tenant status appropriately (requested→planning, ready→updating, etc.) before triggering workflow
- **Dual-trigger pattern**: Both API (immediate) and controller (eventual consistency/retry) can trigger workflows, with deduplication
- **Pluggable provider abstraction**: Maintain workflow provider abstraction (Restate, Step Functions, etc.) and compute provider abstraction (Docker, ECS, etc.) - no hardcoding
- **Health checking**: CREATE operations provision compute until running and healthy, API tracks provisioning progress
- **Error handling**: API returns appropriate errors when workflow triggering fails, tenant state reflects failure

## Capabilities

### New Capabilities
- `api-workflow-triggering`: HTTP API handlers trigger workflow execution for tenant lifecycle operations (create, update, delete)
- `workflow-status-tracking`: Track workflow execution state and propagate status updates back to tenant records in database

### Modified Capabilities
- `http-api-server`: Existing API handlers (CreateTenant, UpdateTenant, DeleteTenant) modified to trigger workflows instead of just database updates
- `workflow-provisioning`: Existing workflow integration enhanced to support both API-triggered and controller-triggered execution paths with deduplication

## Impact

**Affected Code:**
- `internal/api/server.go` - Tenant CRUD handlers (CreateTenant, UpdateTenant, DeleteTenant)
- `internal/api/handlers/*.go` - Individual handler implementations need workflow client integration
- `internal/workflow/manager.go` - May need deduplication logic to prevent double-triggering (API + controller)
- `internal/controller/reconciler.go` - Controller should detect already-triggered workflows and not re-trigger

**API Behavior Changes:**
- POST /tenants now triggers provisioning workflow immediately, returns execution ID
- PATCH /tenants/{id} triggers update workflow immediately
- DELETE /tenants/{id} triggers deletion workflow immediately
- GET /tenants returns workflow execution status alongside tenant data

**Database Schema:**
- Tenant table already has `workflow_execution_id` column (added in controller work)
- May need additional fields for tracking API-triggered vs controller-triggered workflows

**Dependencies:**
- Existing workflow provider abstraction (internal/workflow/)
- Existing compute provider abstraction (internal/compute/)
- Reconciliation controller coordination (avoid duplicate workflow triggers)
- Database transaction handling (atomic status update + workflow trigger)

**User-Visible Changes:**
- Faster response time (no waiting for controller poll)
- API responses include workflow execution ID for tracking
- Clearer API semantics (POST creates AND provisions, DELETE removes infrastructure)
