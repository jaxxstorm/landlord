# Database Types

Landlord persists tenant state, transitions, and audit history in a database. The database provider is pluggable and configured in `config.yaml`.

## Supported providers

| Provider | Use case | Notes |
| --- | --- | --- |
| postgres | Production | Durable, multi-tenant capable, supports migrations |
| sqlite | Local development and tests | File-based, single-writer constraints |

## PostgreSQL

PostgreSQL is the recommended production database.

```yaml
database:
  provider: postgres
  host: localhost
  port: 5432
  user: landlord
  password: landlord_password
  database: landlord_db
  ssl_mode: prefer
```

## SQLite

SQLite is useful for local development and tests.

```yaml
database:
  provider: sqlite
  sqlite:
    path: landlord.db
```

## Migrations

Migrations live under `migrations/` and are applied at startup by the database provider.
