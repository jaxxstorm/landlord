## Context

Tenants are currently soft-deleted by setting status to `deleted` and keeping rows in the database. This blurs the line between a lifecycle terminal state and actual record removal. The create API also requires a `tenant-id`, which is confusing for users who primarily think in names. We need clearer lifecycle semantics and a more ergonomic identifier model, with names as the primary human reference and UUIDs as stable internal identifiers.

## Goals / Non-Goals

**Goals:**
- Introduce a lifecycle state `archived` to represent compute removal while retaining the tenant record.
- Treat `deleted` as a hard delete that removes the tenant record from storage.
- Replace create input `tenant-id` with `tenant-name`, enforcing uniqueness.
- Allow get/update/delete by either UUID or tenant name across API and CLI.
- Keep state transitions explicit and compatible with the existing reconciler/workflow pattern.

**Non-Goals:**
- Backward compatibility for existing API payloads or data.
- Introducing multi-tenant namespaces or name scoping beyond global uniqueness.
- Adding UI or access control changes.

## Decisions

- **Primary identifier choice**: Use a UUID as the immutable internal ID and a unique tenant name as the human identifier. The API will accept either for lookup. Alternative: name-only identifiers; rejected because UUIDs are useful for internal consistency and audit trails.
- **Lifecycle semantics**: Introduce `archived` as the post-delete compute state and reserve `deleted` for physical removal. Alternative: keep soft-delete only; rejected because it conflates lifecycle intent with persistence.
- **Delete flow**: Use the existing delete workflow to archive and then perform a hard delete after compute removal completes. Alternative: immediate hard delete; rejected because compute teardown needs a durable record and audit trail.
- **Schema change scope**: Change the persistence schema to store `name` with a unique constraint and remove reliance on user-supplied IDs. Alternative: add name alongside current ID without enforcement; rejected because uniqueness is required for name-based lookups.

## Risks / Trade-offs

- [Name collisions in existing data] → Since backward compatibility is not required, apply a destructive migration or reinitialize data in dev.
- [Hard deletes break audit expectations] → Keep archived records until explicit delete action and ensure API differentiates archive vs delete clearly.
- [Ambiguous identifier input] → Treat UUID format as ID; otherwise, resolve by name and return 404 if not found.

## Migration Plan

- Replace schema and data with new fields and constraints (no backward compatibility).
- Update API handlers and repository methods to accept name or UUID.
- Update workflow/reconciler to set archived status after compute removal and to hard delete on explicit delete completion.
- Update CLI commands and tests to use `tenant-name` on create and support name/UUID everywhere.

## Open Questions

- Should hard delete be a separate API action or follow after archive automatically when compute deletion completes?
- Should API responses always include both ID and name fields?
