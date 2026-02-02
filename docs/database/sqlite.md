# SQLite Database Provider

Landlord supports SQLite as an alternative database provider alongside PostgreSQL. This provides a lightweight, zero-dependency option for local development, testing, and single-instance deployments.

## Overview

The SQLite provider uses [modernc.org/sqlite](https://gitlab.com/cznic/sqlite), a pure Go SQLite implementation that requires no CGO or external C dependencies. This makes it easy to build and deploy across platforms.

## When to Use SQLite

**Good Use Cases:**
- Local development and testing
- Single-instance deployments
- Embedded applications
- Demo environments
- CI/CD testing pipelines
- Development laptops without PostgreSQL

**Not Recommended For:**
- Multi-instance deployments (SQLite is file-based)
- High-concurrency workloads (single writer limitation)
- Distributed systems requiring replication
- Production deployments requiring HA/DR
- Applications with heavy write workloads

## Configuration

### Basic File-Based Configuration

```yaml
database:
  provider: sqlite
  sqlite:
    path: "landlord.db"
    busyTimeout: 5s
```

### In-Memory Configuration (Testing Only)

```yaml
database:
  provider: sqlite
  sqlite:
    path: ":memory:"
    busyTimeout: 5s
```

**Warning:** In-memory databases are destroyed when the application stops. Use only for testing.

### Advanced Configuration with Pragmas

```yaml
database:
  provider: sqlite
  sqlite:
    path: "landlord.db"
    busyTimeout: 10s
    pragmas:
      - "PRAGMA cache_size=-64000"  # 64MB cache
      - "PRAGMA mmap_size=268435456"  # 256MB memory-mapped I/O
```

## Default Settings

The SQLite provider automatically applies these pragmas for optimal performance:

| Pragma | Value | Purpose |
|--------|-------|---------|
| `journal_mode` | `WAL` | Write-Ahead Logging for better concurrency |
| `synchronous` | `NORMAL` | Balanced safety and performance |
| `foreign_keys` | `ON` | Enforce foreign key constraints |
| `temp_store` | `MEMORY` | Use memory for temporary tables |
| `busy_timeout` | User-configured | How long to wait when database is locked |

## File Paths

### Relative Paths
```yaml
sqlite:
  path: "data/landlord.db"
```
Relative to the working directory where landlord runs.

### Absolute Paths
```yaml
sqlite:
  path: "/var/lib/landlord/landlord.db"
```

### URI Format
```yaml
sqlite:
  path: "file:/var/lib/landlord/landlord.db?mode=rwc"
```

Supported query parameters:
- `mode=ro` - Read-only
- `mode=rw` - Read-write (default)
- `mode=rwc` - Read-write-create (default)
- `mode=memory` - In-memory database

## Migration Compatibility

All migrations have been made compatible with both PostgreSQL and SQLite:

| PostgreSQL Type | SQLite Type | Notes |
|-----------------|-------------|-------|
| `UUID` | `TEXT` | 36-character string format |
| `JSONB` | `JSON` | SQLite JSON1 extension |
| `TIMESTAMP WITH TIME ZONE` | `TIMESTAMP` | Stored as UTC |
| `gen_random_uuid()` | Application-generated | UUID v4 in app code |
| `NOW()` | `CURRENT_TIMESTAMP` | Server-side timestamp |

## Performance Considerations

### Write Concurrency
SQLite uses a write lock - only one writer at a time. For write-heavy workloads, use PostgreSQL.

### Busy Timeout
The `busyTimeout` setting controls how long to wait when the database is locked:
```yaml
sqlite:
  busyTimeout: 5s  # Wait up to 5 seconds for lock
```

### WAL Mode Benefits
Write-Ahead Logging (enabled by default):
- Readers don't block writers
- Writers don't block readers
- Better concurrency than default rollback journal
- Creates additional files: `landlord.db-wal`, `landlord.db-shm`

### Cache Size
Increase cache for better performance:
```yaml
sqlite:
  pragmas:
    - "PRAGMA cache_size=-64000"  # Negative = KB (64MB)
```

## Environment Variables

```bash
# Use SQLite provider
export DB_PROVIDER=sqlite
export SQLITE_PATH=landlord.db
export SQLITE_BUSY_TIMEOUT=5s
```

See [configuration.md](../configuration.md) for complete environment variable list.

## Examples

### Development Configuration
```yaml
database:
  provider: sqlite
  sqlite:
    path: "dev.db"
    busyTimeout: 5s

log:
  level: debug
  development: true
```

### Testing Configuration
```yaml
database:
  provider: sqlite
  sqlite:
    path: ":memory:"
    busyTimeout: 1s
```

### Production Single-Instance
```yaml
database:
  provider: sqlite
  sqlite:
    path: "/var/lib/landlord/data/landlord.db"
    busyTimeout: 10s
    pragmas:
      - "PRAGMA cache_size=-128000"  # 128MB cache
      - "PRAGMA mmap_size=536870912"  # 512MB mmap
```

## Switching Between Providers

### PostgreSQL to SQLite

1. Export data from PostgreSQL:
```bash
pg_dump -U postgres landlord > backup.sql
```

2. Update configuration:
```yaml
database:
  provider: sqlite
  sqlite:
    path: "landlord.db"
```

3. Run migrations (will create new SQLite database):
```bash
./landlord --config config.yaml
```

4. Import data (requires schema-compatible dump)

### SQLite to PostgreSQL

1. Backup SQLite:
```bash
sqlite3 landlord.db .dump > backup.sql
```

2. Update configuration:
```yaml
database:
  provider: postgres
  host: localhost
  port: 5432
  user: postgres
  password: secret
  database: landlord
```

3. Run migrations:
```bash
./landlord --config config.yaml
```

4. Import data (requires schema-compatible dump)

## Troubleshooting

### Database Locked Errors
```
database is locked
```

**Solutions:**
- Increase `busyTimeout`
- Ensure WAL mode is enabled (default)
- Check for long-running transactions
- Verify only one writer instance is running

### File Permission Errors
```
unable to open database file
```

**Solutions:**
- Check file permissions: `chmod 644 landlord.db`
- Verify directory is writable: `chmod 755 /path/to/directory`
- Ensure user has write access to database directory

### Performance Issues
```
slow query performance
```

**Solutions:**
- Increase cache size with `PRAGMA cache_size`
- Enable memory-mapped I/O with `PRAGMA mmap_size`
- Add indexes for frequently queried columns
- Consider PostgreSQL for high-concurrency workloads

### WAL Files Growing Large
```
landlord.db-wal is huge
```

**Solutions:**
- WAL checkpoint happening automatically
- Can force checkpoint: `PRAGMA wal_checkpoint(TRUNCATE)`
- Normal behavior under heavy write load
- Files cleaned up automatically when connections close

## Limitations

1. **Single Writer**: Only one process can write at a time
2. **File-Based**: Cannot be used across multiple servers
3. **No Network Access**: Must be on same filesystem
4. **Limited Replication**: No built-in streaming replication like PostgreSQL
5. **Type System**: Less strict than PostgreSQL (e.g., flexible typing)

## Best Practices

1. **Use WAL Mode**: Enabled by default, don't disable
2. **Set Busy Timeout**: At least 5 seconds for production
3. **Regular Backups**: Use `sqlite3 .dump` or copy file when app is stopped
4. **Monitor File Size**: Plan for growth, SQLite doesn't shrink automatically
5. **Graceful Shutdown**: Ensures WAL checkpoint and clean closure
6. **Development Only**: Unless you have a specific single-instance use case

## Comparison with PostgreSQL

| Feature | SQLite | PostgreSQL |
|---------|--------|------------|
| Installation | None (embedded) | Separate server |
| Concurrency | Single writer | Multi-writer |
| Network Access | No | Yes |
| Replication | No | Yes (streaming) |
| Full-Text Search | FTS5 extension | Built-in |
| JSON Support | JSON1 extension | Native JSONB |
| Type System | Flexible | Strict |
| Transactions | ACID | ACID |
| Performance | Excellent (single) | Excellent (multi) |

## See Also

- [Configuration Management](../configuration.md)
- [Database Migrations](../database-persistence/)
- [PostgreSQL Provider](./postgres.md)
