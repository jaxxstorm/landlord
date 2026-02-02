## 1. API Models and DTOs

- [x] 1.1 Create internal/api/models package directory
- [x] 1.2 Define CreateTenantRequest struct with JSON tags and validation
- [x] 1.3 Add compute_config field to CreateTenantRequest (json.RawMessage or map[string]interface{})
- [x] 1.4 Define UpdateTenantRequest struct with JSON tags and validation
- [x] 1.5 Add compute_config field to UpdateTenantRequest (optional, for updates)
- [x] 1.6 Define TenantResponse struct with JSON tags
- [x] 1.7 Include compute_config in TenantResponse for retrieval
- [x] 1.8 Define ListTenantsResponse struct with pagination metadata
- [x] 1.9 Define ErrorResponse struct for standardized error messages
- [x] 1.10 Create mapping functions from domain.Tenant to TenantResponse
- [x] 1.11 Create mapping functions from API requests to domain models
- [x] 1.12 Add validation tags (e.g., using validator library)

## 1a. Compute Provider Validation Interface

- [x] 1a.1 Add ValidateConfig method to compute.Provider interface
- [x] 1a.2 Method signature: ValidateConfig(config json.RawMessage) error
- [x] 1a.3 Implement ValidateConfig in Docker provider
- [x] 1a.4 Docker validation: check env, volumes, network_mode, ports fields
- [x] 1a.5 Return detailed errors with field names and constraint violations
- [x] 1a.6 Add unit tests for Docker provider ValidateConfig
- [x] 1a.7 Test valid Docker configs pass validation
- [x] 1a.8 Test invalid Docker configs return appropriate errors

## 2. Handler Implementation - Create Tenant

- [x] 2.1 Create internal/api/tenants.go file
- [x] 2.2 Inject compute provider into API server/handlers
- [x] 2.3 Implement handleCreateTenant handler function
- [x] 2.4 Parse and validate JSON request body
- [x] 2.5 Validate tenant name (non-empty, length constraints)
- [x] 2.6 Extract compute_config from request
- [x] 2.7 Call provider.ValidateConfig(compute_config) for validation
- [x] 2.8 Return HTTP 400 with provider validation errors if invalid
- [x] 2.9 Check for duplicate tenant names
- [x] 2.10 Map compute_config to Tenant.DesiredConfig
- [x] 2.11 Call tenant repository Create method
- [x] 2.12 Handle repository errors (duplicate, database failures)
- [x] 2.13 Return HTTP 201 with created tenant response
- [x] 2.14 Return HTTP 400 for validation errors
- [x] 2.15 Return HTTP 409 for name conflicts
- [x] 2.16 Return HTTP 500 for server errors
- [x] 2.17 Add Swagger annotations for POST /api/tenants
- [x] 2.18 Document compute_config field in Swagger with provider-specific examples

## 3. Handler Implementation - Get Tenant

- [x] 3.1 Implement handleGetTenant handler function
- [x] 3.2 Extract tenant ID from URL path parameter
- [x] 3.3 Validate UUID format using uuid.Parse
- [x] 3.4 Return HTTP 400 for invalid UUID format
- [x] 3.5 Call tenant repository Get method
- [x] 3.6 Return HTTP 404 if tenant not found
- [x] 3.7 Map domain tenant to API response including compute_config
- [x] 3.8 Return HTTP 200 with tenant response
- [x] 3.9 Add Swagger annotations for GET /api/tenants/{id}
- [x] 3.10 Document compute_config in response schema

## 4. Handler Implementation - List Tenants

- [x] 4.1 Implement handleListTenants handler function
- [x] 4.2 Parse limit query parameter (default: 50 or 100)
- [x] 4.3 Parse offset query parameter (default: 0)
- [x] 4.4 Validate pagination parameters (non-negative integers)
- [x] 4.5 Parse optional name filter query parameter
- [x] 4.6 Call tenant repository List method with pagination
- [x] 4.7 Get total tenant count for metadata
- [x] 4.8 Map domain tenants to API response array
- [x] 4.9 Include pagination metadata in response (total, limit, offset)
- [x] 4.10 Return HTTP 200 with list response
- [x] 4.11 Handle empty list case (return empty array)
- [x] 4.12 Add Swagger annotations for GET /api/tenants

## 5. Handler Implementation - Update Tenant

- [x] 5.1 Implement handleUpdateTenant handler function
- [x] 5.2 Extract tenant ID from URL path parameter
- [x] 5.3 Validate UUID format
- [x] 5.4 Parse and validate JSON request body
- [x] 5.5 Validate update fields (name constraints)
- [x] 5.6 If compute_config present, call provider.ValidateConfig()
- [x] 5.7 Return HTTP 400 if compute_config validation fails
- [x] 5.8 Check if tenant exists before update
- [x] 5.9 Return HTTP 404 if tenant not found
- [x] 5.10 Map compute_config to Tenant.DesiredConfig if provided
- [x] 5.11 Call tenant repository Update method
- [x] 5.12 Handle update validation errors
- [x] 5.13 Return HTTP 200 with updated tenant response
- [x] 5.14 Return HTTP 400 for validation errors
- [x] 5.15 Add Swagger annotations for PUT /api/tenants/{id}
- [x] 5.16 Document compute_config field as optional in update

## 6. Handler Implementation - Delete Tenant

- [x] 6.1 Implement handleDeleteTenant handler function
- [x] 6.2 Extract tenant ID from URL path parameter
- [x] 6.3 Validate UUID format
- [x] 6.4 Check if tenant exists before deletion
- [x] 6.5 Return HTTP 404 if tenant not found
- [x] 6.6 Call tenant repository Delete method
- [x] 6.7 Handle cascade deletion if needed
- [x] 6.8 Return HTTP 204 No Content on success
- [x] 6.9 Return HTTP 500 for deletion failures
- [x] 6.10 Add Swagger annotations for DELETE /api/tenants/{id}

## 7. Validation and Error Handling

- [x] 7.1 Create validation helper function for UUIDs
- [x] 7.2 Create validation helper function for tenant names
- [x] 7.3 Create compute config validation wrapper calling provider.ValidateConfig
- [x] 7.4 Create error response formatting function
- [x] 7.5 Add request body parsing helper with error handling
- [x] 7.6 Implement consistent error response structure across handlers
- [x] 7.7 Include provider validation errors in API error responses
- [x] 7.8 Add correlation ID to error responses (from middleware)
- [x] 7.9 Log errors appropriately (with context and severity)

## 8. Route Registration

- [x] 8.1 Add tenant routes to registerRoutes in server.go
- [x] 8.2 Register POST /api/tenants route
- [x] 8.3 Register GET /api/tenants/{id} route
- [x] 8.4 Register GET /api/tenants route
- [x] 8.5 Register PUT /api/tenants/{id} route
- [x] 8.6 Register DELETE /api/tenants/{id} route
- [x] 8.7 Verify middleware chain applies to tenant routes
- [x] 8.8 Test route patterns don't conflict

## 9. Swagger Documentation

- [x] 9.1 Regenerate Swagger spec using make swagger-docs
- [x] 9.2 Verify all tenant endpoints appear in swagger.json
- [x] 9.3 Verify request schemas include compute_config field
- [x] 9.4 Add provider-specific examples to compute_config documentation
- [x] 9.5 Verify response schemas are documented correctly
- [x] 9.6 Verify error responses are documented (400, 404, 409, 500)
- [x] 9.7 Document that 400 errors can include provider validation failures
- [x] 9.8 Verify path parameters are documented
- [x] 9.9 Verify query parameters are documented
- [x] 9.10 Test /api/docs UI displays tenant endpoints
- [x] 9.11 Verify tenant operation tags group endpoints correctly

## 10. Unit Tests - Handlers

- [x] 10.1 Create internal/api/tenants_test.go file
- [x] 10.2 Create mock compute provider for testing
- [x] 10.3 Write test for handleCreateTenant success case with valid compute_config
- [x] 10.4 Write test for handleCreateTenant with invalid compute_config
- [x] 10.5 Write test for handleCreateTenant validation errors
- [x] 10.6 Write test for handleCreateTenant duplicate name
- [x] 10.7 Write test for handleGetTenant success case
- [x] 10.8 Write test for handleGetTenant not found
- [x] 10.9 Write test for handleGetTenant invalid UUID
- [x] 10.10 Write test for handleListTenants empty list
- [x] 10.11 Write test for handleListTenants with pagination
- [x] 10.12 Write test for handleUpdateTenant success case
- [x] 10.13 Write test for handleUpdateTenant with compute_config update
- [x] 10.14 Write test for handleUpdateTenant with invalid compute_config
- [x] 10.15 Write test for handleUpdateTenant not found
- [x] 10.16 Write test for handleDeleteTenant success case
- [x] 10.17 Write test for handleDeleteTenant not found
- [x] 10.18 Use mock tenant repository for all handler tests

## 11. Integration Tests

- [x] 11.1 Create integration test file with real database and Docker provider
- [x] 11.2 Test complete create tenant flow with valid Docker compute_config
- [x] 11.3 Test create tenant with invalid Docker compute_config (should fail)
- [x] 11.4 Test get tenant by ID flow (verify compute_config returned)
- [x] 11.5 Test list tenants with pagination
- [x] 11.6 Test update tenant flow with compute_config change
- [x] 11.7 Test delete tenant flow
- [x] 11.8 Test tenant name uniqueness enforcement
- [x] 11.9 Test cascade deletion if applicable
- [x] 11.10 Test error cases end-to-end
- [x] 11.11 Clean up test data after tests

## 12. Manual Testing

- [x] 12.1 Start server locally with Docker provider configured
- [x] 12.2 Test POST /api/tenants with Docker compute_config (curl or Postman)
- [x] 12.3 Test POST with invalid compute_config (verify 400 error)
- [x] 12.4 Test GET /api/tenants/{id} with valid ID (verify compute_config in response)
- [x] 12.5 Test GET /api/tenants with pagination parameters
- [x] 12.6 Test PUT /api/tenants/{id} to update tenant with new compute_config
- [x] 12.7 Test DELETE /api/tenants/{id}
- [x] 12.8 Test validation error cases (empty name, invalid UUID)
- [x] 12.9 Test 404 cases (non-existent tenant)
- [x] 12.10 Test duplicate name conflict
- [x] 12.11 Verify Swagger UI at /api/docs shows tenant operations with compute_config examples

## 13. Documentation

- [x] 13.1 Add tenant API examples to README with Docker compute_config examples
- [x] 13.2 Document pagination behavior and defaults
- [x] 13.3 Document validation rules for tenant names
- [x] 13.4 Add curl examples for each endpoint with compute_config
- [x] 13.5 Document compute_config structure for Docker provider
- [x] 13.6 Document error response format including provider validation errors
- [x] 13.7 Update CONTRIBUTING.md with tenant API examples
- [x] 13.8 Note that endpoints are currently unauthenticated
- [x] 13.9 Document that provider type is configured at landlord startup
- [x] 13.10 Document any known limitations

## 14. Code Review and Refinement

- [x] 14.1 Review all handler code for consistency
- [x] 14.2 Ensure error messages are clear and helpful
- [x] 14.3 Verify logging is appropriate and not excessive
- [x] 14.4 Check for code duplication and refactor if needed
- [x] 14.5 Verify all exported functions have comments
- [x] 14.6 Run go fmt on all new files
- [x] 14.7 Run go vet to check for issues
- [x] 14.8 Run golangci-lint if available
- [x] 14.9 Verify test coverage is adequate

## 15. Deployment Preparation

- [x] 15.1 Ensure all tests pass locally
- [x] 15.2 Build the application successfully
- [x] 15.3 Create pull request with all changes
- [x] 15.4 Review changes with team
- [x] 15.5 Address review feedback
- [x] 15.6 Merge to main branch
- [x] 15.7 Deploy to staging/test environment
- [x] 15.8 Verify tenant API works in staging
- [x] 15.9 Monitor logs for errors after deployment
