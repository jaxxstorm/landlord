## 1. Dependencies and Setup

- [x] 1.1 Add swaggo/swag to go.mod (github.com/swaggo/swag)
- [x] 1.2 Install swag CLI tool (swag init)
- [x] 1.3 Create docs/ directory for generated swagger files
- [x] 1.4 Verify swag init command works in the project

## 2. Add Swagger Annotations to HTTP Handlers

- [x] 2.1 Annotate main API handler with package-level swagger comments (title, description, version)
- [x] 2.2 Add swagger annotations to all GET endpoints in internal/api/
- [x] 2.3 Add swagger annotations to all POST endpoints in internal/api/
- [x] 2.4 Add swagger annotations to all PUT endpoints in internal/api/
- [x] 2.5 Add swagger annotations to all DELETE endpoints in internal/api/
- [x] 2.6 Document request body schemas for each endpoint
- [x] 2.7 Document response schemas for success cases (200/201/204)
- [x] 2.8 Document error response schemas (400, 401, 403, 404, 500)
- [x] 2.9 Document URL path parameters and query parameters for all endpoints
- [x] 2.10 Add authentication/authorization scope documentation where applicable

## 3. Generate OpenAPI Specification

- [x] 3.1 Run swag init to generate initial swagger.json and docs files
- [x] 3.2 Verify swagger.json is valid OpenAPI 3.0 specification
- [x] 3.3 Verify all endpoints are present in generated spec
- [x] 3.4 Commit docs/ directory to version control (or exclude generated files if preferred)
- [x] 3.5 Add generated files to .gitignore or update documentation strategy

## 4. Implement API Documentation Web Server Routes

- [x] 4.1 Create HTTP handler for /api/swagger.json endpoint (serve generated spec)
- [x] 4.2 Create HTTP handler for /api/docs endpoint (serve Redoc HTML page)
- [x] 4.3 Embed or serve Redoc standalone JavaScript from HTML template
- [x] 4.4 Configure Redoc to load swagger.json from /api/swagger.json endpoint
- [x] 4.5 Test that /api/docs loads successfully in browser
- [x] 4.6 Test that /api/swagger.json returns valid JSON

## 5. Documentation Page and UI

- [x] 5.1 Create api.html template with Redoc integration (similar to reference provided)
- [x] 5.2 Set Redoc options (scrollYOffset, hideLoading, etc.)
- [x] 5.3 Ensure Redoc page is responsive on mobile/tablet/desktop
- [x] 5.4 Test Redoc search functionality
- [x] 5.5 Test schema expansion/collapse in Redoc UI
- [x] 5.6 Verify dark mode support (if needed)

## 6. Integration and Build Process

- [x] 6.1 Add swag init to Makefile build target
- [ ] 6.2 Add spec generation to CI/CD pipeline
- [ ] 6.3 Add pre-commit hook to regenerate specs on code changes
- [ ] 6.4 Ensure builds fail if swagger annotations are invalid
- [ ] 6.5 Configure linting rules for swagger annotation consistency
- [ ] 6.6 Update go.mod and go.sum with dependencies

## 7. Documentation and Examples

- [x] 7.1 Update project README with API documentation endpoint URL
- [x] 7.2 Add section to README explaining how to add annotations to new handlers
- [x] 7.3 Create example of annotated handler in docs/ or CONTRIBUTING.md
- [x] 7.4 Document swagger annotation syntax and conventions used
- [x] 7.5 Add troubleshooting guide for common annotation issues
- [x] 7.6 Document how to update spec during local development

## 8. Testing and Validation

- [x] 8.1 Verify all existing endpoints are documented in spec
- [x] 8.2 Test spec validity with OpenAPI validator
- [x] 8.3 Test /api/docs endpoint returns 200 OK
- [x] 8.4 Test /api/swagger.json endpoint returns valid JSON
- [x] 8.5 Manually test a few endpoints through Redoc UI
- [x] 8.6 Verify no authentication is required to view documentation (if applicable)
- [x] 8.7 Test documentation generation after code changes
- [x] 8.8 Check spec file size and performance impact

## 9. Deployment and Release

- [x] 9.1 Create pull request with all changes
- [ ] 9.2 Review annotations and spec for accuracy with team
- [ ] 9.3 Merge to main branch
- [ ] 9.4 Build and test in staging environment
- [ ] 9.5 Deploy to production
- [ ] 9.6 Verify /api/docs is accessible in production
- [ ] 9.7 Update API documentation link in public resources (website, etc.)
- [ ] 9.8 Announce documentation endpoint availability to users
