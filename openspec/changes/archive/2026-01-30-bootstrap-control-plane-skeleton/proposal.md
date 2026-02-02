## Why

We need a foundational control plane skeleton that provides the core HTTP API, database persistence, and configuration management for the tenant provisioning system. Without this foundation, we cannot implement tenant lifecycle operations or state management.

## What Changes

- Create HTTP API server using chi router with health and readiness endpoints
- Implement PostgreSQL database connection management with migrations support
- Add configuration management using alecthomas/kong for environment variables and CLI flags
- Set up structured logging with uber/zap (console-friendly for local dev, JSON for production)
- Establish project structure following ports-and-adapters architecture
- Add basic error handling patterns that bubble errors to CLI level

## Capabilities

### New Capabilities
- `http-api-server`: HTTP server with routing, middleware, and graceful shutdown
- `database-persistence`: PostgreSQL connection pooling, migrations, and health checks
- `configuration-management`: Environment-based configuration with CLI flag support
- `structured-logging`: Context-aware logging with development and production modes

### Modified Capabilities
<!-- No existing capabilities are being modified -->

## Impact

- New Go module with core dependencies (chi, kong, zap, pgx)
- Project directory structure: `cmd/`, `internal/`, `migrations/`
- Configuration via environment variables and CLI flags
- Database schema migration system
- Logging conventions for all future components
