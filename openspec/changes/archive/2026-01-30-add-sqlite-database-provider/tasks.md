## 1. Provider Abstraction Setup

- [x] 1.1 Create `internal/database/provider.go` with Provider interface definition (Pool, Health, Close methods)
- [x] 1.2 Create `internal/database/factory.go` with NewProvider factory function
- [x] 1.3 Create directory structure `internal/database/providers/postgres/`
- [x] 1.4 Create directory structure `internal/database/providers/sqlite/`

## 2. Configuration Updates

- [x] 2.1 Add `Provider` field to `DatabaseConfig` in `internal/config/database.go` with default "postgres"
- [x] 2.2 Create `SQLiteConfig` struct in `internal/config/database.go` with Path, BusyTimeout, Pragmas fields
- [x] 2.3 Add `SQLite SQLiteConfig` field to `DatabaseConfig`
- [x] 2.4 Update `DatabaseConfig.Validate()` to handle both provider types
- [x] 2.5 Add `SQLiteConfig.Validate()` method for SQLite-specific validation
- [x] 2.6 Add environment variable tags for SQLite config (DB_SQLITE_PATH, DB_SQLITE_BUSY_TIMEOUT, DB_SQLITE_PRAGMAS)

## 3. PostgreSQL Provider Refactoring

- [x] 3.1 Move existing `DB` struct to `internal/database/providers/postgres/provider.go`
- [x] 3.2 Implement Provider interface for PostgreSQL provider
- [x] 3.3 Move connection pool initialization logic to PostgreSQL provider
- [x] 3.4 Ensure PostgreSQL provider implements Pool() returning *pgxpool.Pool
- [x] 3.5 Update PostgreSQL provider Health() method to match interface
- [x] 3.6 Update PostgreSQL provider Close() method to match interface
- [x] 3.7 Add PostgreSQL-specific retry logic in provider initialization

## 4. SQLite Provider Implementation

- [x] 4.1 Add `modernc.org/sqlite` dependency to go.mod
- [x] 4.2 Add `github.com/jmoiron/sqlx` dependency to go.mod
- [x] 4.3 Create `internal/database/providers/sqlite/provider.go` with SQLiteProvider struct
- [x] 4.4 Implement SQLiteProvider.Pool() returning *sqlx.DB
- [x] 4.5 Implement SQLiteProvider initialization with connection string handling (file path and URI)
- [x] 4.6 Implement pragma application on initialization (journal_mode=WAL, busy_timeout, foreign_keys, synchronous)
- [x] 4.7 Implement SQLiteProvider.Health() with simple SELECT query
- [x] 4.8 Implement SQLiteProvider.Close() with connection cleanup
- [x] 4.9 Add support for in-memory SQLite (":memory:" and "file::memory:?cache=shared")
- [x] 4.10 Configure connection pool settings appropriate for SQLite (single writer consideration)
- [x] 4.11 Add logging for SQLite initialization and pragma settings

## 5. Factory Implementation

- [x] 5.1 Implement NewProvider factory in `internal/database/factory.go`
- [x] 5.2 Add provider type switch (postgres/sqlite) in factory
- [x] 5.3 Route to PostgreSQL provider for "postgres" type
- [x] 5.4 Route to SQLite provider for "sqlite" type
- [x] 5.5 Return error for unknown provider types with helpful message
- [x] 5.6 Add logging for provider selection

## 6. Migration Compatibility Updates

- [x] 6.1 Update migrations to replace `UUID` with `TEXT` for SQLite compatibility
- [x] 6.2 Update migrations to replace `JSONB` with `JSON` for SQLite compatibility
- [x] 6.3 Update migrations to replace `TIMESTAMP WITH TIME ZONE` with `TIMESTAMP`
- [x] 6.4 Update migrations to replace `gen_random_uuid()` with application-level UUID generation
- [x] 6.5 Update migrations to replace `NOW()` with `CURRENT_TIMESTAMP`
- [x] 6.6 Verify GIN indexes are SQLite-compatible or add conditional logic
- [x] 6.7 Update `RunMigrations()` in `internal/database/database.go` to handle both provider connection string formats
- [x] 6.8 Add SQLite migration driver import (`_ "github.com/golang-migrate/migrate/v4/database/sqlite3"`)
- [x] 6.9 Test migrations apply cleanly to both PostgreSQL and SQLite

## 7. Repository Updates

- [x] 7.1 Update `internal/tenant/postgres/repository.go` constructor to accept `interface{}` pool parameter
- [x] 7.2 Add type assertion in PostgreSQL repository constructor for *pgxpool.Pool
- [x] 7.3 Update repository initialization in main application code to use Provider.Pool()
- [x] 7.4 If needed, create SQLite-specific repository implementation in `internal/tenant/sqlite/`
- [x] 7.5 Ensure repository error handling works identically across providers
- [x] 7.6 Verify optimistic locking works with SQLite transaction model

## 8. Integration with Application

- [x] 8.1 Update `internal/database/database.go` New() function to use NewProvider factory
- [x] 8.2 Update application startup in `cmd/landlord/main.go` to pass provider to repositories
- [x] 8.3 Ensure graceful shutdown closes provider connections properly
- [x] 8.4 Update health check endpoints to use Provider.Health() interface method

## 9. Testing

- [x] 9.1 Create `internal/database/provider_test.go` with factory tests
- [x] 9.2 Create `internal/database/providers/postgres/provider_test.go` with PostgreSQL provider tests
- [x] 9.3 Create `internal/database/providers/sqlite/provider_test.go` with SQLite provider tests
- [x] 9.4 Add test for in-memory SQLite initialization
- [x] 9.5 Add test for SQLite pragma application verification
- [x] 9.6 Create migration compatibility test suite running against both providers
- [x] 9.7 Update existing repository tests to run against both PostgreSQL and SQLite
- [x] 9.8 Add integration test using SQLite in-memory for fast CI execution
- [x] 9.9 Verify all database-persistence spec scenarios pass with both providers
- [x] 9.10 Add test for concurrent operations with SQLite (WAL mode verification)

## 10. Configuration Examples

- [x] 10.1 Create example config file for PostgreSQL provider in `docs/examples/config-postgres.yaml`
- [x] 10.2 Create example config file for SQLite file-based provider in `docs/examples/config-sqlite.yaml`
- [x] 10.3 Create example config file for SQLite in-memory provider in `docs/examples/config-sqlite-memory.yaml`
- [x] 10.4 Update `config.yaml` in project root to show provider field with postgres default
- [x] 10.5 Update `config.json` if present to include provider configuration

## 11. Documentation

- [x] 11.1 Update `docs/configuration.md` with database provider selection documentation
- [x] 11.2 Document SQLite configuration options (path, busy timeout, pragmas)
- [x] 11.3 Document use cases for each provider (PostgreSQL: production, SQLite: dev/test)
- [x] 11.4 Add section on SQLite limitations (single writer, not for high-concurrency)
- [x] 11.5 Document in-memory SQLite setup for testing
- [x] 11.6 Add troubleshooting section for common SQLite issues (database locked, etc.)
- [x] 11.7 Update README.md with quickstart using SQLite for local development

## 12. CI/CD Updates

- [x] 12.1 Update CI pipeline to run tests against PostgreSQL provider
- [x] 12.2 Update CI pipeline to run tests against SQLite provider
- [x] 12.3 Add CI job for in-memory SQLite tests (fast feedback)
- [x] 12.4 Ensure go.mod and go.sum are committed with new dependencies
- [x] 12.5 Verify cross-compilation works with pure Go SQLite driver

## 13. Validation and Cleanup

- [x] 13.1 Run full test suite with PostgreSQL provider (ensure backward compatibility)
- [x] 13.2 Run full test suite with SQLite provider (ensure feature parity)
- [x] 13.3 Test application startup with both providers via environment variables
- [x] 13.4 Verify migrations work correctly on fresh SQLite database
- [x] 13.5 Verify migrations work correctly on existing PostgreSQL database
- [x] 13.6 Test graceful shutdown and cleanup with both providers
- [x] 13.7 Run linter and fix any issues introduced by changes
- [x] 13.8 Review code for PostgreSQL-specific assumptions that need abstraction
- [x] 13.9 Verify logging output is helpful for both providers
- [x] 13.10 Update CHANGELOG.md or release notes with SQLite provider feature
