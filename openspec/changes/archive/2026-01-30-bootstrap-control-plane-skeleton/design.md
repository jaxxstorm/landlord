## Context

The tenant provisioning system needs a foundational control plane to handle HTTP API requests, persist data, and manage configuration. Currently, no infrastructure exists - this is the initial bootstrap. The system follows a ports-and-adapters architecture with pluggable components, so the skeleton must establish patterns that support future extension points.

## Goals / Non-Goals

**Goals:**
- Create a runnable HTTP server with basic routing and middleware
- Establish database connection patterns with health checks and migrations
- Define configuration loading and validation approach
- Set up logging conventions that support both development and production environments
- Structure the codebase to support ports-and-adapters architecture

**Non-Goals:**
- Implementing tenant domain logic (comes after skeleton)
- Workflow or compute engine integrations (future phases)
- Authentication or authorization (basic stubs acceptable)
- Production deployment configuration (local development focus)

## Decisions

### Decision 1: Use chi for HTTP routing

**Choice:** chi router over alternatives (gin, echo, stdlib)

**Rationale:** 
- chi is idiomatic Go using stdlib http.Handler
- Lightweight with strong middleware composition support
- Good fit for ports-and-adapters (handlers are just functions)
- Project convention established in project.md

**Alternatives considered:**
- stdlib net/http: Too low-level for API routing patterns
- gin/echo: More opinionated, heavier frameworks

### Decision 2: Use pgx for PostgreSQL

**Choice:** pgx/v5 over database/sql with pq

**Rationale:**
- Better performance and PostgreSQL-specific features
- Native connection pooling (pgxpool)
- Supports both high-level and low-level APIs
- Industry standard for modern Go + Postgres

**Alternatives considered:**
- database/sql + pq: Less efficient, generic interface
- ORM (GORM, ent): Too much abstraction for control plane needs

### Decision 3: Database migrations using golang-migrate

**Choice:** golang-migrate for schema versioning

**Rationale:**
- CLI and library support for migration execution
- Version-based migrations with up/down support
- Widely adopted, integrates with pgx
- Supports embedding migrations in binary

**Alternatives considered:**
- Goose: Similar but less active maintenance
- SQL files only: Need programmatic execution support

### Decision 4: Configuration with kong

**Choice:** alecthomas/kong for config loading

**Rationale:**
- Unified env vars, CLI flags, and config files
- Type-safe configuration with struct tags
- Help text generation from structs
- Project convention per project.md

**Alternatives considered:**
- viper: More complex, too much flexibility
- envconfig: CLI flags require separate solution

### Decision 5: Structured logging with zap

**Choice:** uber/zap for structured logging

**Rationale:**
- High performance structured logging
- Development (console) and production (JSON) encoders
- Context propagation support for request tracing
- Project convention per project.md

**Alternatives considered:**
- logrus: Slower performance
- slog (stdlib): Newer, less mature ecosystem

### Decision 6: Project structure

**Structure:**
```
cmd/
  landlord/          # Main binary
internal/
  api/               # HTTP handlers (adapter)
  config/            # Configuration loading
  database/          # Database adapter
  domain/            # Domain logic (future)
migrations/          # SQL migration files
```

**Rationale:**
- Follows Go project layout conventions
- internal/ prevents external import
- Separates adapters (api, database) from domain
- Supports ports-and-adapters growth

## Risks / Trade-offs

**[Risk] Database connection failures during startup**
→ Mitigation: Implement retry logic with exponential backoff, separate health check from readiness check

**[Risk] Configuration validation failures are silent**
→ Mitigation: Validate all required config at startup, fail fast with clear error messages

**[Trade-off] Using pgx instead of database/sql**
→ Less portable to other databases, but not a concern given PostgreSQL commitment in project.md

**[Risk] Logging overhead in high-throughput scenarios**
→ Mitigation: zap is designed for performance, can adjust log levels per environment

**[Trade-off] No authentication in initial skeleton**
→ API is fully open, acceptable for initial development but must be addressed before multi-tenant deployment
