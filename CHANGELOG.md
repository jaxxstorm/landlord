# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added

- **SQLite Database Provider**: Added SQLite as an alternative database provider alongside PostgreSQL
  - Support for both file-based and in-memory SQLite databases
  - Provider abstraction with factory pattern for transparent backend selection
  - Automatic schema migration compatibility between PostgreSQL and SQLite
  - Configuration-driven provider selection with backward compatibility (defaults to PostgreSQL)
  - SQLite uses Write-Ahead Logging (WAL) mode for improved concurrent read performance
  - In-memory database support for fast testing and CI/CD pipelines
  - Comprehensive documentation and example configurations for SQLite setup

### Changed

- Refactored database initialization to use provider abstraction pattern
- Database configuration now includes `provider` field to select between "postgres" and "sqlite"
- PostgreSQL-specific code moved to dedicated provider package (`internal/database/providers/postgres/`)
- Repository constructors now accept interface{} pool parameter for provider flexibility
- Health check endpoint now uses Provider interface for both database backends

### Migration Notes

- All existing database migrations remain compatible and now work with both PostgreSQL and SQLite
- Type mappings for cross-database compatibility:
  - UUID → TEXT
  - JSONB → JSON
  - TIMESTAMP WITH TIME ZONE → TIMESTAMP
- No data migration required for existing PostgreSQL deployments (backward compatible)
- New SQLite deployments use identical schema through shared migration files

## Features

### Database Provider Selection

```yaml
# PostgreSQL (default)
database:
  provider: postgres
  host: localhost
  port: 5432
  user: landlord
  password: secret
  database: landlord_db

# SQLite (file-based)
database:
  provider: sqlite
  sqlite:
    path: landlord.db
    busy_timeout: 5s

# SQLite (in-memory for testing)
database:
  provider: sqlite
  sqlite:
    path: ":memory:"
```

### Use Cases

- **PostgreSQL**: Production deployments, high-concurrency environments, distributed systems
- **SQLite**: Local development, testing, single-instance deployments, CI/CD pipelines, embedded applications

### Documentation

- New [SQLite Provider Documentation](docs/database/sqlite.md) with configuration, usage, and troubleshooting
- Example configurations for PostgreSQL and SQLite deployments
- Updated [Configuration Guide](docs/configuration.md) with provider selection details

---

For migration instructions and detailed documentation, see:
- [Configuration Guide](docs/configuration.md)
- [SQLite Provider Documentation](docs/database/sqlite.md)
- Example Configurations in `docs/examples/`
