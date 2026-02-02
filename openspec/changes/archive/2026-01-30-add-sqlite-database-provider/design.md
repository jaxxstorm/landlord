## Context

Currently, the landlord system is tightly coupled to PostgreSQL through direct use of `pgx/v5/pgxpool` in `internal/database/database.go`. The `DatabaseConfig` struct assumes PostgreSQL connection parameters (host, port, user, password) and constructs PostgreSQL-specific connection strings. This creates barriers for:

- **Local development**: Developers must run PostgreSQL in Docker or locally
- **Testing**: Integration tests require PostgreSQL setup, slowing CI/CD
- **Lightweight deployments**: Demo/proof-of-concept deployments need external database infrastructure
- **Embedded use cases**: Single-binary deployments with embedded database not possible

The existing repository implementations in `internal/tenant/postgres` assume PostgreSQL-specific features like JSONB and pgx connection types.

## Goals / Non-Goals

**Goals:**
- Enable SQLite as an alternative database provider alongside PostgreSQL
- Support identical repository interface contract across both providers
- Allow provider selection via configuration without code changes
- Maintain backward compatibility with existing PostgreSQL deployments
- Support in-memory SQLite for fast testing
- Apply same migration files to both providers with compatible SQL syntax

**Non-Goals:**
- Remove PostgreSQL support or make it non-default
- Support runtime provider switching after application start
- Support database providers beyond PostgreSQL and SQLite in this change
- Optimize for SQLite-specific performance tuning beyond WAL mode
- Support SQLite encryption extensions (SQLCIPHER)
- Provide migration path from PostgreSQL to SQLite for existing data

## Decisions

### Decision 1: Use Provider Abstraction with Factory Pattern

**Choice**: Introduce a `Provider` interface that abstracts database operations, with factory function selecting implementation based on configuration.

**Rationale**: 
- Allows transparent swapping of database backends
- Keeps provider-specific code isolated
- Enables testing with different providers without changing application code
- Standard Go pattern for dependency injection

**Alternatives Considered**:
- **Build tags**: Would require separate binaries for each provider, complicating deployment
- **Conditional compilation**: Creates testing and maintenance burden
- **Interface wrapper at repository level**: Too high in the stack, duplicates connection management logic

**Implementation**:
```go
// internal/database/provider.go
type Provider interface {
    Pool() interface{} // Returns *pgxpool.Pool or *sql.DB
    Health(ctx context.Context) error
    Close()
}

func NewProvider(ctx context.Context, cfg *config.DatabaseConfig, logger *zap.Logger) (Provider, error)
```

### Decision 2: Use modernc.org/sqlite for Pure Go Implementation

**Choice**: Use `modernc.org/sqlite` as the SQLite driver instead of `github.com/mattn/go-sqlite3`.

**Rationale**:
- Pure Go implementation enables cross-compilation without CGO
- No CGO dependency simplifies builds and CI/CD
- Compatible with standard library `database/sql` interface
- Active maintenance and good performance characteristics

**Alternatives Considered**:
- **mattn/go-sqlite3**: Requires CGO, complicates cross-compilation. Widely used but build complexity outweighs benefits.
- **crawshaw/sqlite**: Lower-level API, excellent performance but different paradigm from pgx approach. Would require more adaptation work.

**Trade-offs**:
- Pure Go implementation slightly slower than CGO version (~10-20%)
- For landlord's use case (development, testing, small-scale deployments), simplicity outweighs performance difference

### Decision 3: Extend DatabaseConfig with Provider Field and Union Type

**Choice**: Add `Provider` field to `DatabaseConfig` with provider-specific configuration in nested structs.

**Configuration Structure**:
```go
type DatabaseConfig struct {
    Provider string `mapstructure:"provider" env:"DB_PROVIDER" default:"postgres"`
    
    // PostgreSQL-specific (existing fields)
    Host     string
    Port     int
    // ... existing fields
    
    // SQLite-specific
    SQLite SQLiteConfig `mapstructure:"sqlite"`
}

type SQLiteConfig struct {
    Path        string        `mapstructure:"path" env:"DB_SQLITE_PATH" default:"landlord.db"`
    BusyTimeout time.Duration `mapstructure:"busy_timeout" env:"DB_SQLITE_BUSY_TIMEOUT" default:"5s"`
    Pragmas     []string      `mapstructure:"pragmas" env:"DB_SQLITE_PRAGMAS"`
}
```

**Rationale**:
- Maintains backward compatibility (defaults to PostgreSQL)
- Clear separation of provider-specific configuration
- Single config struct keeps initialization simple
- Environment variable naming conventions follow existing patterns

**Alternatives Considered**:
- **Separate config types**: Would require interface for config, complicating initialization and validation
- **Config interface with implementations**: Over-engineering for two providers

### Decision 4: Adapt Migrations to Use Common SQL Subset

**Choice**: Modify existing migrations to use SQL syntax compatible with both PostgreSQL and SQLite.

**Type Mappings**:
- `UUID` → `TEXT` (store UUID strings in SQLite)
- `JSONB` → `JSON` (SQLite has JSON functions)
- `TIMESTAMP WITH TIME ZONE` → `TIMESTAMP` (SQLite stores as TEXT/INTEGER)
- `gen_random_uuid()` → Application generates UUIDs before INSERT
- `NOW()` → `CURRENT_TIMESTAMP` (works in both)

**Rationale**:
- Single set of migration files reduces maintenance burden
- golang-migrate supports multiple database drivers
- Type mappings maintain semantic equivalence
- UUID generation in application code is more portable

**Alternatives Considered**:
- **Separate migration files per provider**: Doubles migration maintenance burden
- **Database-specific pragmas in migrations**: Complicates migration logic, rejected in favor of provider initialization

**Trade-offs**:
- Lose PostgreSQL-specific optimizations (e.g., native UUID type storage)
- Acceptable trade-off for deployment flexibility

### Decision 5: Use database/sql.DB with sqlx for SQLite Provider

**Choice**: SQLite provider uses `database/sql.DB` wrapped with `jmoiron/sqlx` for query helpers, maintaining similar patterns to pgx usage.

**Rationale**:
- `database/sql` is standard library interface for SQLite drivers
- `sqlx` provides similar ergonomics to pgx (named parameters, struct scanning)
- Minimizes repository implementation differences between providers
- Well-tested, stable libraries

**Alternatives Considered**:
- **Pure database/sql**: More verbose, requires more boilerplate in repository implementations
- **Create custom wrapper**: Reinventing sqlx functionality, unnecessary

### Decision 6: Repository Interface Accepts interface{} Pool Type

**Choice**: Repository constructors accept `interface{}` for pool and type-assert to provider-specific type.

**Implementation**:
```go
func NewPostgresRepository(pool interface{}, logger *zap.Logger) (*Repository, error) {
    pgPool, ok := pool.(*pgxpool.Pool)
    if !ok {
        return nil, fmt.Errorf("expected *pgxpool.Pool, got %T", pool)
    }
    // ...
}
```

**Rationale**:
- Avoids defining overly abstract interfaces for connection pools
- Type assertion makes provider expectations explicit
- Simple, straightforward approach for two providers

**Alternatives Considered**:
- **Generic connection interface**: Would need to abstract Query/Exec/Begin - complex and unnecessary for two providers
- **Separate factory for repositories per provider**: Repository knows provider anyway via type assertion

### Decision 7: Enable WAL Mode and Configure Pragmas on SQLite Initialization

**Choice**: SQLite provider applies pragmas during connection initialization:
- `journal_mode=WAL` (Write-Ahead Logging)
- `busy_timeout=5000` (5 second default)
- `foreign_keys=ON`
- `synchronous=NORMAL` (balance durability/performance)

**Rationale**:
- WAL mode allows concurrent readers with single writer
- Busy timeout prevents immediate "database locked" errors
- Foreign keys must be explicitly enabled in SQLite
- NORMAL synchronous mode is appropriate for development/testing use case

**Alternatives Considered**:
- **Let users configure all pragmas**: Too much configuration burden for common case
- **Memory-only mode by default**: Users can explicitly configure `:memory:` if needed

## Risks / Trade-offs

### [Risk] Migration syntax compatibility issues
**Mitigation**: Create comprehensive test suite that runs migrations against both PostgreSQL and SQLite in CI. Add migration validation script that checks for provider-specific syntax.

### [Risk] Performance degradation with SQLite under load
**Mitigation**: Document SQLite as suitable for development/testing/small deployments. PostgreSQL remains default and recommended for production. Add metrics/logging to detect SQLite bottlenecks.

### [Risk] Repository behavior differences between providers
**Mitigation**: Shared repository test suite runs against both providers. Any behavioral difference fails tests. Document known limitations if unavoidable.

### [Risk] SQLite single-writer bottleneck
**Mitigation**: WAL mode + busy timeout handles moderate concurrency. Document that SQLite is not suitable for high-concurrency production workloads. Connection pool configured appropriately (max 1 writer).

### [Trade-off] UUID stored as TEXT in SQLite vs native UUID in PostgreSQL
**Impact**: Minor storage overhead (~25 bytes vs 16 bytes), no functional difference. Query performance equivalent for indexed lookups.

### [Trade-off] Loss of PostgreSQL JSONB operators for complex queries
**Impact**: Current tenant queries don't use advanced JSONB operators. If needed in future, can add provider-specific query implementations.

### [Trade-off] Migration complexity increased
**Impact**: Migrations must work with both providers. Adds testing burden. Benefit of deployment flexibility outweighs cost.

## Migration Plan

### Phase 1: Provider Abstraction (Non-Breaking)
1. Introduce `Provider` interface in `internal/database/provider.go`
2. Refactor existing PostgreSQL code into `internal/database/providers/postgres/`
3. Update `database.New()` to use factory pattern (defaults to PostgreSQL)
4. No configuration changes required - existing deployments continue working

### Phase 2: SQLite Provider Implementation
1. Add `modernc.org/sqlite` dependency
2. Implement SQLite provider in `internal/database/providers/sqlite/`
3. Update `DatabaseConfig` to include `Provider` field and `SQLiteConfig`
4. Update factory to handle SQLite provider type

### Phase 3: Migration Compatibility
1. Update migrations to use common SQL subset
2. Modify `RunMigrations()` to use provider-appropriate connection string format
3. Add migration tests against both providers

### Phase 4: Repository Updates
1. Update repository constructors to accept `interface{}` pool
2. Add type assertions for provider-specific types
3. If needed, add provider-specific repository implementations in `internal/tenant/sqlite/`

### Phase 5: Testing & Documentation
1. Run full test suite against both providers
2. Add integration tests with SQLite in-memory mode
3. Update documentation with provider selection examples
4. Add configuration examples for SQLite development setups

### Rollback Strategy
- All changes are additive and backward compatible
- If issues arise, default provider remains PostgreSQL
- SQLite provider can be disabled by setting `DB_PROVIDER=postgres` (default)
- No data migration needed since SQLite is for new/test deployments

### Deployment Checklist
- [ ] Add `modernc.org/sqlite` to go.mod
- [ ] Run `go mod tidy && go mod vendor` if using vendoring
- [ ] Update CI to test against both database providers
- [ ] Update developer documentation with SQLite setup instructions
- [ ] Add example configurations for both providers

## Open Questions

1. **Should we support runtime detection of provider from connection string?**
   - Currently requires explicit `provider` config field
   - Could auto-detect from connection string format (postgres:// vs file:)
   - Decision: Explicit is better than implicit, keep provider field required

2. **Should we provide migration tool to convert PostgreSQL data to SQLite?**
   - Not in scope for this change
   - SQLite is for new/dev deployments, not migrating production data
   - If needed, can be separate utility

3. **Connection pool configuration for SQLite - should we restrict max connections?**
   - SQLite has single-writer limitation
   - Should we enforce MaxConnections=1 programmatically or trust configuration?
   - Decision: Set sensible defaults but allow override for read-heavy workloads
