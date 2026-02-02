## 1. API Versioning Foundation

- [x] 1.1 Define supported API versions (v1) in a central config or constant
- [x] 1.2 Add versioned chi subrouter mounting at `/v1`
- [x] 1.3 Implement version-required handling for unversioned paths returning `version_required`
- [x] 1.4 Implement unsupported-version handling returning `unsupported_version` with supported list

## 2. Update HTTP API Routes

- [x] 2.1 Update tenant create route to `/v1/tenants` and adjust handler wiring
- [x] 2.2 Update tenant list route to `/v1/tenants` and adjust handler wiring
- [x] 2.3 Update tenant get route to `/v1/tenants/{id}` and adjust handler wiring
- [x] 2.4 Update tenant update route to `/v1/tenants/{id}` and adjust handler wiring
- [x] 2.5 Update tenant delete route to `/v1/tenants/{id}` and adjust handler wiring

## 3. Validation, Errors, and Swagger

- [x] 3.1 Update request validation to accept only versioned paths
- [x] 3.2 Add error response mapping for missing/unsupported API versions
- [x] 3.3 Update OpenAPI/Swagger annotations to use `/v1` paths

## 4. CLI Updates

- [x] 4.1 Update CLI base URL resolution to append `/v1` when no version is provided
- [x] 4.2 Ensure CLI commands use versioned paths in requests and output references

## 5. Documentation and Examples

- [x] 5.1 Update README and docs to show `/v1` endpoints and supported versions
- [x] 5.2 Update any API examples or curl snippets to use `/v1` paths

## 6. Tests

- [x] 6.1 Update HTTP router tests to expect versioned paths
- [x] 6.2 Add tests for unversioned request rejection and unsupported version errors
