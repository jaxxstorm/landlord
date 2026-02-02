## Why

The landlord API currently lacks comprehensive, discoverable API documentation. Developers integrating with the landlord API must manually explore the codebase or rely on incomplete documentation, creating friction and limiting adoption. Implementing Swagger/OpenAPI documentation with a web-based UI (Redoc) will improve developer experience, enable better API discoverability, and provide an interactive interface for testing endpoints.

## What Changes

- Add Swagger/OpenAPI spec generation using swaggo/swag
- Integrate swaggo/swag annotations into HTTP API handlers to auto-generate specs
- Implement an HTML-based API documentation web page using Redoc
- Serve API documentation from a dedicated HTTP endpoint
- Generate and maintain OpenAPI spec files as part of the build process

## Capabilities

### New Capabilities
- `swagger-spec-generation`: Auto-generate OpenAPI 3.0 specs from Go code annotations using swaggo/swag
- `api-docs-web-ui`: Serve an interactive API documentation web page using Redoc at a dedicated endpoint
- `swagger-annotations`: Add Swagger annotations to HTTP handler methods to document endpoints, request/response schemas, and parameters

### Modified Capabilities
<!-- No existing capabilities have requirement changes -->

## Impact

- **Code**: HTTP API handlers in `internal/api/` will need Swagger annotations added
- **APIs**: New HTTP endpoint for serving API documentation (e.g., `/api/docs`)
- **Dependencies**: Addition of `github.com/swaggo/swag` for spec generation and CLI tooling
- **Build Process**: Add swaggo spec generation to build/development workflow
- **Documentation**: Generated OpenAPI specs will serve as authoritative API documentation
