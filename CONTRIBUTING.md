# Contributing to Landlord

## API Documentation

The Landlord API is self-documenting through Swagger/OpenAPI annotations. When you add or modify HTTP endpoints, please add appropriate Swagger comments.

### Example: Annotated Handler

Here's an example of a properly documented handler:

```go
// handleGetTenant retrieves a tenant by ID
// @Summary Get a tenant
// @Description Retrieves details of a specific tenant
// @Tags tenants
// @Produce json
// @Param id path string true "Tenant ID" format(uuid)
// @Success 200 {object} domain.Tenant "Tenant details"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 404 {object} map[string]string "Tenant not found"
// @Router /api/tenants/{id} [get]
func (s *Server) handleGetTenant(w http.ResponseWriter, r *http.Request) {
    tenantID := chi.URLParam(r, "id")
    // Implementation...
}
```

### Another Example: POST with Request Body

```go
// handleCreateTenant creates a new tenant
// @Summary Create a tenant
// @Description Creates a new tenant with the provided configuration
// @Tags tenants
// @Accept json
// @Produce json
// @Param tenant body domain.CreateTenantRequest true "Tenant configuration"
// @Success 201 {object} domain.Tenant "Created tenant"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Server error"
// @Router /api/tenants [post]
func (s *Server) handleCreateTenant(w http.ResponseWriter, r *http.Request) {
    var req domain.CreateTenantRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        // Error handling...
    }
    // Implementation...
}
```

## Documenting Request/Response Schemas

Define your request and response types with clear field documentation:

```go
package domain

type CreateTenantRequest struct {
    // Name of the tenant
    Name string `json:"name" binding:"required"`
    // Optional description
    Description string `json:"description"`
}

type Tenant struct {
    // Unique tenant identifier
    ID string `json:"id" format:"uuid"`
    // Tenant name
    Name string `json:"name"`
    // Tenant description
    Description string `json:"description"`
    // Creation timestamp
    CreatedAt time.Time `json:"created_at" format:"date-time"`
}
```

## Regenerating Documentation

After adding or modifying handlers:

```bash
# Regenerate Swagger documentation
make swagger-docs

# View the updated documentation
# Open http://localhost:8080/api/docs in your browser
```

## Common Swagger Tags

| Tag | Purpose |
|-----|---------|
| `@Summary` | One-line summary of what the endpoint does |
| `@Description` | Longer description of functionality |
| `@Tags` | Comma-separated tags for grouping (e.g., "users,auth") |
| `@Param` | Document path/query/header parameters |
| `@Accept` | Content-Type the endpoint accepts (json, xml, etc.) |
| `@Produce` | Content-Type the endpoint produces |
| `@Success` | Success response with status code and schema |
| `@Failure` | Error response with status code and schema |
| `@Router` | HTTP method and URL path pattern |
| `@Deprecated` | Mark endpoint as deprecated |
| `@Security` | Security/auth requirements |

## Controller Development

### Architecture

The tenant reconciliation controller implements a continuous reconciliation pattern:

1. **Polling Loop**: Periodically queries database for tenants needing work
2. **Work Queue**: Rate-limited queue with exponential backoff for retries
3. **Worker Pool**: Concurrent workers process reconciliation tasks
4. **Workflow Integration**: Triggers provider-specific workflows for state transitions

### Key Files

- `internal/controller/reconciler.go` - Main reconciliation loop and orchestration
- `internal/controller/state_machine.go` - State transition logic
- `internal/controller/workflow_client.go` - Workflow provider integration
- `internal/controller/queue.go` - Work queue wrapper

### Development Guidelines

#### 1. State Machine Changes

When modifying state transitions:

1. Update `ValidTransitions` map in `internal/tenant/tenant.go`
2. Update state transition logic in `internal/controller/state_machine.go`
3. Add tests in `internal/controller/state_machine_test.go`
4. Update documentation in `docs/state-machine.md`

#### 2. Workflow Integration

When adding new workflow actions or providers:

1. Implement `workflow.Provider` interface in `internal/workflow/provider.go`
2. Update `TriggerWorkflow()` in `internal/controller/workflow_client.go`
3. Add determination logic in `DetermineAction()`
4. Create integration tests with mock provider

#### 3. Reconciliation Loop Changes

When modifying core reconciliation logic:

1. Ensure context cancellation is handled properly
2. Update error classification in `IsRetryableError()`
3. Add structured logging at key points
4. Write integration tests in `internal/controller/reconciler_test.go`

#### 4. Error Handling

Error handling patterns:

- **Retryable Errors**: Network timeouts, provider temporarily unavailable
  - Action: Re-queue with exponential backoff
  - Log: Include `retry_count`
  
- **Fatal Errors**: Invalid configuration, permanent provider failure
  - Action: Transition to `failed` status
  - Log: Include error message in `status_message`

- **Unknown Errors**: Unexpected error types
  - Default to retryable to be safe
  - Log with full error details for investigation

### Testing Guidelines

#### Integration Tests

Write integration tests for:
- Happy path: successful state transitions
- Error scenarios: retries, failures
- Concurrency: multiple tenants processing
- Database: state persistence, optimistic locking
- Workflows: provider interactions

Use `internal/controller/reconciler_test.go` as reference.

#### Mocking

For testing without real providers:
- Use `MockWorkflowProvider` in test files
- Simulate errors with `simulateError` flag
- Test timeout scenarios with `simulateSlowExecution`

#### Test Database

Integration tests use PostgreSQL in Docker:
```bash
# Tests automatically start/stop containers
go test ./internal/controller -v
```

### Performance Considerations

#### Database Queries

- `ListTenantsForReconciliation()` is called frequently
- Ensure `status` column is indexed
- Monitor query duration in reconciliation logs

#### Goroutine Management

- Workers are spawned by `Start()` and cleaned up by `Stop()`
- Context cancellation must propagate to all workers
- Test graceful shutdown scenarios

#### Queue Management

- Work queue deduplicates entries automatically
- Exponential backoff prevents overwhelming providers
- Monitor queue depth in metrics

### Logging Best Practices

Use structured logging throughout:

```go
// Good: Structured fields
r.logger.Info("tenant reconciled successfully",
    zap.String("tenant_id", tenantID),
    zap.String("previous_status", string(previousStatus)),
    zap.String("new_status", string(t.Status)),
    zap.Duration("duration", duration))

// Less useful: String formatting
r.logger.Info(fmt.Sprintf("tenant %s reconciled in %v", tenantID, duration))
```

Log at appropriate levels:
- `Debug`: Detailed information (tenant fetch, status check)
- `Info`: Important events (reconciliation start/success, workflow trigger)
- `Error`: Error conditions (reconciliation failure, max retries)
- `Warn`: Unusual but recoverable situations (timeout but will retry)

### Making a PR

When submitting controller changes:

1. **Update Documentation**
   - Update relevant docs if behavior changes
   - Add comments explaining complex logic
   - Update examples if configuration changes

2. **Write Tests**
   - Add integration tests for new functionality
   - Test error paths and edge cases
   - Run full test suite locally

3. **Performance Impact**
   - Consider impact on database query volume
   - Test with realistic tenant counts (100+)
   - Monitor resource usage in tests

4. **Backward Compatibility**
   - Ensure existing tenants continue working
   - Test migration scenarios
   - Support graceful degradation for new features

5. **Documentation Updates**
   - Update `docs/state-machine.md` if transitions change
   - Update `docs/configuration.md` if new config options
   - Update `docs/controller-troubleshooting.md` if new failure modes

## Testing

Write tests for new endpoints alongside Swagger documentation:

```bash
make test
```

## Pull Requests

1. Add Swagger annotations to all new endpoints
2. Regenerate documentation: `make swagger-docs`
3. Commit the generated `docs/` changes
4. Ensure tests pass: `make test`
5. Build locally: `make build`

## Useful Resources

- [Swaggo Documentation](https://github.com/swaggo/swag)
- [OpenAPI 3.0 Specification](https://spec.openapis.org/oas/v3.0.3)
- [Go-Chi Router Documentation](https://github.com/go-chi/chi)
