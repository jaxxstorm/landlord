# Proposal: Bootstrap Database Connections and Schema

## Problem Statement

The tenant provisioning system requires durable persistence to:
- Track tenant lifecycle state (requested → provisioning → ready → deleting → deleted)
- Store desired and observed state for each tenant
- Maintain audit history of all state transitions
- Enable querying tenant status via API
- Support reconciliation and drift detection

Currently, no database layer exists beyond basic connection setup. We need a complete schema and abstraction layer to support tenant lifecycle management.

## Why This Change Now

This is foundational infrastructure required before implementing tenant CRUD operations. Without it:
- Cannot persist tenant state across control plane restarts
- Cannot track lifecycle transitions or audit history
- Cannot implement reconciliation logic
- Cannot provide tenant status APIs

This change unblocks all tenant management features.

## Goals

1. **Define pluggable database abstraction** following ports-and-adapters architecture
2. **Design tenant lifecycle schema** tracking current state, desired state, and history
3. **Implement PostgreSQL adapter** as initial concrete implementation
4. **Support audit trail** for all tenant state transitions
5. **Enable idempotent operations** via schema design

## Non-Goals

- Complex query optimization (initial implementation favors simplicity)
- Multi-tenant database isolation (single control plane database)
- Database migrations framework (will use existing golang-migrate)
- Real-time streaming/CDC (polling-based queries sufficient initially)

## Success Criteria

- Tenant CRUD operations can persist and retrieve state
- State transitions are auditable with timestamp and reason
- Database implementation is swappable via interface
- Schema supports querying by tenant ID, status, and timestamps
- Idempotent updates work correctly (repeated writes don't corrupt state)

## Alternatives Considered

### In-Memory Storage
**Rejected**: No durability, loses state on restart, unsuitable for production control plane.

### Document Database (MongoDB/DynamoDB)
**Deferred**: PostgreSQL provides stronger consistency guarantees and simpler audit trail queries. May revisit for specific use cases.

### Event Sourcing
**Deferred**: Adds complexity for initial implementation. Current state + audit log is simpler and sufficient for MVP.

## Dependencies

- Existing database connection setup (already implemented)
- Existing migration infrastructure (golang-migrate already configured)

## Risks

- **Schema evolution**: Future tenant features may require schema changes. Mitigated by: using migrations, designing extensible schema.
- **Performance at scale**: Initial design optimizes for correctness over performance. Mitigated by: adding indexes as needed, profiling before optimization.

## Timeline Estimate

- Schema design and interface definition: Small (covered in this change)
- PostgreSQL implementation: Small (straightforward CRUD)
- Testing and validation: Small

Total: Single focused implementation session
