## Context

The landlord application currently has a minimal HTTP API with only health check endpoints (`/health` and `/ready`). The system has an existing tenant domain layer (`internal/tenant/`) with database repository implementations for PostgreSQL and SQLite, but no HTTP API exposure for tenant management operations.

The existing codebase uses:
- **HTTP Framework**: go-chi/chi for routing
- **Database Layer**: Abstract `database.Provider` interface with PostgreSQL and SQLite implementations
- **Tenant Repository**: Already implements `Create`, `Get`, `List`, and `Delete` operations at the database level
- **Domain Models**: `internal/domain/` exists but tenant API models need to be defined
- **API Documentation**: Swagger/OpenAPI infrastructure is already in place with Redoc UI at `/api/docs`

Current constraints:
- Must maintain compatibility with existing database layer
- Follow established patterns from health check endpoints
- Integrate with existing middleware (logging, CORS, request ID)
- Build on existing Swagger documentation infrastructure

## Goals / Non-Goals

**Goals:**
- Expose complete tenant CRUD operations through RESTful HTTP endpoints
- Provide request/response models that are clean and API-friendly (may differ from database models)
- Implement comprehensive input validation before database operations
- Document all endpoints with Swagger annotations for automatic OpenAPI spec generation
- Return standardized error responses across all tenant operations
- Support pagination for list operations to handle large tenant counts
- Maintain consistency with existing API patterns

**Non-Goals:**
- Authentication and authorization (existing system doesn't have auth yet; separate concern)
- Tenant isolation or multi-tenancy security model (separate feature)
- Bulk operations (create/update/delete multiple tenants in one request)
- Advanced filtering beyond basic name search
- Versioning of tenant resources or audit trails
- Webhooks or event notifications for tenant changes
- Rate limiting or throttling (API-wide concern, not tenant-specific)

## Decisions

### Decision: Map database models to API-specific DTOs
**Choice**: Create separate request/response structs in a new package (e.g., `internal/api/models/`) rather than directly exposing database domain models.

**Rationale**:
- Decouples API contracts from database schema
- Allows API to evolve independently of storage layer
- Enables different field names, validation rules, and JSON serialization between API and DB
- Provides cleaner separation of concerns

**Alternatives considered**:
- Use domain models directly: Simpler but couples API to database schema changes
- Use inline anonymous structs: More concise but harder to document and reuse

### Decision: Use chi's URL parameter extraction for tenant IDs
**Choice**: Use `chi.URLParam(r, "id")` to extract tenant IDs from URL paths.

**Rationale**:
- Consistent with go-chi patterns
- Simple and performant
- Already used in the codebase for routing

**Alternative considered**:
- gorilla/mux style: Would require framework change

### Decision: Validate UUIDs at handler level before database lookup
**Choice**: Parse and validate UUID format in HTTP handlers using `uuid.Parse()` before calling repository methods.

**Rationale**:
- Fail fast with clear HTTP 400 errors for malformed IDs
- Prevents unnecessary database queries for invalid UUIDs
- Provides better error messages to clients

### Decision: Return 404 for non-existent resources, 204 for successful deletion
**Choice**: Use HTTP 404 Not Found when tenant doesn't exist, HTTP 204 No Content for successful DELETE operations.

**Rationale**:
- Follows REST conventions
- 404 clearly indicates resource state
- 204 for DELETE is idiomatic and indicates no response body

**Alternative considered**:
- 200 with empty body: Valid but 204 is more semantically precise

### Decision: Implement pagination with limit/offset query parameters
**Choice**: Use `?limit=10&offset=20` query parameters for list endpoint pagination.

**Rationale**:
- Simple and widely understood pagination pattern
- Easy to implement with SQL LIMIT/OFFSET
- Predictable for API clients

**Alternatives considered**:
- Cursor-based pagination: More complex, better for large datasets but overkill for tenant lists
- Page number based: Less flexible than offset

### Decision: Use go-chi middleware for common concerns
**Choice**: Leverage existing middleware for logging, request ID, timeouts, and panic recovery.

**Rationale**:
- Consistency with existing API patterns
- Avoid duplicating cross-cutting concerns
- Middleware already configured in server setup

### Decision: Return structured error responses with standard format
**Choice**: Define error response struct with `error` or `message` field, use consistently across all endpoints.

**Rationale**:
- Predictable error format for API clients
- Easier to parse programmatically
- Can include additional context (error codes, validation details) in consistent structure

### Decision: Add Swagger annotations directly in handler code
**Choice**: Place Swagger comments immediately above each handler function.

**Rationale**:
- Co-locates documentation with implementation
- Easy to keep docs in sync with code changes
- Consistent with existing Swagger setup in the project

### Decision: Single compute provider per landlord instance
**Choice**: Landlord instances are configured with a single compute provider at startup (e.g., Docker, ECS, or Kubernetes). The provider type is not specified per-tenant at runtime.

**Rationale**:
- Simplifies deployment model - one landlord instance manages one provider type
- Eliminates need for multi-provider routing logic and configuration complexity
- Each environment (dev/staging/prod) typically uses one provider type consistently
- Allows provider-specific optimizations and assumptions in the codebase

**Implications**:
- Provider type is configured in `config.yaml` at startup
- Tenants inherit the provider type from the landlord instance they're managed by
- Multiple providers requires multiple landlord instances (e.g., one for Docker dev, one for ECS prod)

**Alternative considered**:
- Multi-provider per instance: More flexible but significantly more complex, unnecessary for typical use cases

### Decision: Provider-specific compute configuration at tenant creation
**Choice**: Tenant create/update API accepts provider-specific compute configuration that is validated against the active provider's schema at API ingress.

**Rationale**:
- Validates configuration early, failing fast before wasting compute resources
- Provides clear feedback to API users about what configuration is valid
- Provider validates its own configuration format (Docker validates Docker config, ECS validates ECS config)
- Configuration is stored in `Tenant.DesiredConfig` as the source of truth

**Flow**:
1. API receives tenant create request with `compute_config` field
2. API delegates to active compute provider for schema validation
3. Provider validates configuration structure, types, and constraints
4. If valid, configuration is stored in `Tenant.DesiredConfig`
5. Later, during provisioning, provider reads the same configuration back

**Example configurations**:
- **Docker**: `{"env": {"PORT": "8080"}, "volumes": ["/data:/data"], "network_mode": "bridge"}`
- **ECS**: `{"task_role_arn": "arn:aws:iam::...", "environment": [{"name": "PORT", "value": "8080"}]}`
- **Kubernetes**: `{"env": [{"name": "PORT", "value": "8080"}], "volumes": [{"name": "data", "mountPath": "/data"}]}`

**Alternatives considered**:
- Opaque map[string]string: Simpler storage but no validation, poor UX
- Validate at provisioning time: Wastes resources, slow feedback loop
- Typed union of all provider configs: Couples API to all providers, harder to extend

## Risks / Trade-offs

### Risk: UUID validation adds extra parsing step
**Mitigation**: UUID parsing is very fast; performance impact negligible. Can benchmark if needed.

### Risk: Lack of authentication allows unrestricted tenant operations
**Mitigation**: Out of scope for this change. Will be addressed in separate authentication feature. Document clearly in API docs that endpoints are currently unprotected.

### Risk: No tenant name uniqueness enforcement at database level
**Mitigation**: Check existing tenant repository implementation. If uniqueness not enforced by database constraints, add explicit checks in handlers and return 409 Conflict. Consider adding unique constraint in future migration.

### Risk: Pagination with offset can have consistency issues with concurrent modifications
**Trade-off**: Offset-based pagination is simpler but can skip/duplicate items if data changes during pagination. Acceptable for tenant lists (low write frequency). Document limitation. Can revisit with cursor-based pagination if needed.

### Risk: Large tenant lists could cause memory pressure
**Mitigation**: Default limit on list endpoint (e.g., 100 items max). Require clients to paginate for larger datasets. Log warning if fetching very large result sets.

### Risk: Delete operation doesn't check for dependent resources
**Mitigation**: If tenants have associated resources (schemas, workflows), deletion should cascade or be blocked. Check existing repository implementation. Document behavior clearly.

### Trade-off: DTOs add mapping code between API and domain layers
**Acceptance**: Mapping overhead is worth the decoupling. Keep mapping functions simple and testable. Consider using a mapping library if complexity grows.

## Migration Plan

1. **Create API models package** (`internal/api/models/`)
   - Define `CreateTenantRequest`, `UpdateTenantRequest`, `TenantResponse` structs
   - Add JSON tags and validation tags
   - Include `compute_config` field as `json.RawMessage` or `map[string]interface{}`
   - Create mapping functions to/from domain models

2. **Add provider configuration validation**
   - Extend compute provider interface with `ValidateConfig(config json.RawMessage) error` method
   - Implement validation in Docker provider using Docker-specific schema
   - Return detailed validation errors (field names, constraint violations)

3. **Implement HTTP handlers** in `internal/api/tenants.go` (new file)
   - `handleCreateTenant` - POST /api/tenants (validates compute_config via provider)
   - `handleGetTenant` - GET /api/tenants/{id}
   - `handleListTenants` - GET /api/tenants
   - `handleUpdateTenant` - PUT /api/tenants/{id} (validates compute_config if present)
   - `handleDeleteTenant` - DELETE /api/tenants/{id}

4. **Add Swagger annotations** to each handler
   - Document request/response schemas including compute_config structure
   - Specify HTTP status codes (400 for invalid compute_config)
   - Add parameter descriptions and examples

5. **Register routes** in `internal/api/server.go`
   - Add tenant routes to router in `registerRoutes` method
   - Ensure middleware chain is applied

6. **Add validation helpers**
   - UUID parsing/validation function
   - Request body validation (possibly using validator library)
   - Compute config validation via provider integration
   - Error response formatting helpers

7. **Test implementation**
   - Unit tests for handlers with mock repository and mock provider
   - Test compute_config validation with valid/invalid Docker configs
   - Integration tests with real database and Docker provider
   - Manual testing with curl/Postman

8. **Regenerate Swagger docs**
   - Run `make swagger-docs` to update OpenAPI spec
   - Verify all endpoints appear in `/api/docs`
   - Verify compute_config schema is documented

9. **Update documentation**
   - Add tenant API examples to README with provider-specific compute_config examples
   - Document pagination behavior
   - Note authentication absence
   - Document provider-specific configuration requirements

## Open Questions

- **Tenant name uniqueness**: Is it enforced at database level or should API handle it? Check existing repository code.
- **Update semantics**: Full replacement (PUT) or partial update (PATCH)? Start with PUT, add PATCH if needed.
- **Soft delete**: Should tenants be soft-deleted or hard-deleted? Check repository implementation.
- **List default limit**: What's a reasonable default? Suggest 50 or 100.
- **Filter capabilities**: Besides name prefix, what other filters are useful? Priority, status? Leave for future iteration.
- **Error code standards**: Should we use custom error codes or stick with HTTP status codes only? HTTP status codes sufficient for now.
