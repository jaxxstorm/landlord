## Why

The landlord API currently only provides health check endpoints, lacking essential tenant management functionality. To enable full multi-tenant orchestration, the API must support complete CRUD operations for tenants, allowing clients to create, retrieve, update, and delete tenant resources through RESTful HTTP endpoints with proper Swagger documentation.

## What Changes

- Add HTTP POST endpoint to create new tenants
- Add HTTP GET endpoint to retrieve a specific tenant by ID
- Add HTTP GET endpoint to list all tenants with pagination support
- Add HTTP PUT/PATCH endpoint to update tenant status and configuration
- Add HTTP DELETE endpoint to remove tenants
- Implement request validation for tenant operations
- Add Swagger/OpenAPI annotations to all new tenant endpoints
- Define tenant request/response models with JSON schema documentation
- Integrate tenant endpoints with existing database layer
- Add error handling for tenant operations (404, 400, 409, 500)

## Capabilities

### New Capabilities
- `tenant-create-api`: HTTP endpoint for creating new tenants with validation
- `tenant-get-api`: HTTP endpoint for retrieving a single tenant by ID
- `tenant-list-api`: HTTP endpoint for listing all tenants with optional pagination and filtering
- `tenant-update-api`: HTTP endpoint for updating tenant status and properties
- `tenant-delete-api`: HTTP endpoint for deleting tenants with proper cleanup
- `tenant-request-validation`: Input validation for tenant operations (name, configuration, IDs)
- `tenant-api-errors`: Standardized error responses for tenant operations

### Modified Capabilities
<!-- No existing capabilities have requirement changes -->

## Impact

- **Code**: New HTTP handlers in `internal/api/` for tenant operations
- **APIs**: New REST endpoints under `/api/tenants` path
  - `POST /api/tenants` - Create tenant
  - `GET /api/tenants/:id` - Get tenant by ID
  - `GET /api/tenants` - List tenants
  - `PUT /api/tenants/:id` - Update tenant
  - `DELETE /api/tenants/:id` - Delete tenant
- **Domain Models**: New request/response structs in `internal/domain/` for tenant API operations
- **Database**: Integration with existing tenant repository from `internal/tenant/`
- **Documentation**: Swagger annotations for all tenant endpoints extending existing API docs
- **Testing**: New test cases for tenant CRUD operations
- **Validation**: Request body validation, UUID validation, tenant name validation
