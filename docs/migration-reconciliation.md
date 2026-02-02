# Migration Guide: Reconciliation-Based Workflows

This guide covers the breaking change to a reconciliation-driven workflow architecture with stateless workers.

## What Changed

- Workflow invocation moved from API handlers to the reconciliation controller.
- Workflow workers are stateless and no longer connect to the database.
- Tenant lifecycle uses `requested → provisioning → ready` (with optional planning).
- Workflow status polling is now part of the controller loop.

## Required Updates

1. **Enable the controller**
   - Ensure `controller.enabled: true` in your configuration.
   - Set `controller.reconciliation_interval` and optionally `controller.status_poll_interval`.

2. **Pick a workflow provider**
   - Set `workflow.default_provider` (or `controller.workflow_provider` to override).
   - Configure provider credentials/endpoint as needed.

3. **Deploy stateless workers**
   - Remove DB credentials from worker environments.
   - Ensure workers can reach the compute provider and (optionally) Landlord API.

4. **Update any custom integrations**
   - If you had direct workflow triggers from the API, remove them and rely on reconciliation.
   - If you had worker code reading the database, migrate to use workflow payloads.

## Verification

- Create a tenant and confirm it moves from `requested` to `provisioning`.
- Confirm a workflow execution ID is stored on the tenant.
- Verify the reconciler updates the tenant to `ready` on success.

## Rollback Considerations

This change is not backward compatible with DB-connected workers. Rolling back requires:

- Reverting API workflow triggers and worker DB access.
- Disabling the controller or setting it to a no-op configuration.
