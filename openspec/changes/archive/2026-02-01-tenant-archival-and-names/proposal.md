## Why

Tenant deletion currently only marks records as deleted, which conflates soft deletion with a final lifecycle state and keeps removed tenants in the database. At the same time, the create API requires a tenant ID, which is confusing for users who primarily think in names. We need clearer lifecycle semantics and more ergonomic identifiers now, before the app stabilizes.

## What Changes

- **BREAKING** Introduce a new terminal lifecycle status `archived` when compute has been removed; `deleted` will indicate permanent removal from the database.
- **BREAKING** Replace `tenant-id` on create with `tenant-name`, auto-generating the tenant UUID server-side.
- **BREAKING** Enforce unique tenant names and allow get/update/delete by either UUID or name.
- Update CLI and API request/response formats to reflect the new naming and status semantics.
- Update persistence and workflow logic to align with archived vs deleted transitions.

## Capabilities

### New Capabilities
- (none)

### Modified Capabilities
- `database-persistence`: store unique tenant names, adjust lifecycle fields for archived vs deleted semantics.
- `tenant-create-api`: accept tenant name, generate UUID, and return both identifiers.
- `tenant-get-api`: allow retrieval by UUID or name.
- `tenant-update-api`: allow updates by UUID or name.
- `tenant-delete-api`: separate archiving from hard delete and allow delete by UUID or name.
- `tenant-request-validation`: validate tenant name uniqueness and identifier parsing.
- `tenant-lifecycle-workflows`: add archived status and update transition rules.
- `landlord-cli`: create/get/update/delete by name or UUID and reflect new statuses.

## Impact

- API payloads, URL parameters, and validation rules.
- Database schema (tenant identifiers, uniqueness constraints, deletion semantics).
- Workflow execution and reconciler logic for archived vs deleted tenants.
- CLI flags and output, plus tests and fixtures across API/CLI layers.
