## Why

The current system requires PostgreSQL for all deployments, creating barriers for local development, testing, and lightweight deployments. A SQLite provider would enable developers to run the full system locally without external database dependencies and support simplified deployment scenarios for small-scale or embedded use cases.

## What Changes

- Add SQLite database provider implementation alongside existing PostgreSQL provider
- Support all existing database migrations (tenant schema, state history tables)
- Implement connection pooling semantics appropriate for SQLite (single writer, multiple readers)
- Provide configuration options for SQLite-specific settings (WAL mode, busy timeout, pragmas)
- Adapt optimistic locking implementation to work with SQLite's transaction model
- Maintain identical repository interface contract for transparent provider swapping

## Capabilities

### New Capabilities
- `sqlite-database-provider`: SQLite implementation of the database provider interface supporting migrations, connection management, and all repository operations defined in database-persistence spec

### Modified Capabilities
- `database-persistence`: Extend to support provider abstraction allowing PostgreSQL or SQLite backends with identical interface contract and behavior guarantees

## Impact

- **Code**: New provider implementation in `internal/database/providers/sqlite/`
- **Configuration**: Add database provider type selection (postgres/sqlite) and SQLite-specific config options
- **Dependencies**: Add SQLite driver (`modernc.org/sqlite` or `github.com/mattn/go-sqlite3`)
- **Migrations**: Migrations must be compatible with both PostgreSQL and SQLite SQL dialects
- **Testing**: Existing database tests should pass against both providers
- **Deployment**: Enables standalone binary deployments without external database for development/demo scenarios
