## Context

The landlord API currently lacks developer-friendly API documentation. The existing codebase in `internal/api/` defines HTTP handlers but has no automated documentation generation or interactive web UI for exploring endpoints. Developers must read source code or incomplete docs to understand available endpoints, parameters, and response schemas. This friction reduces developer experience and limits API adoption.

The project uses:
- Go HTTP handlers with net/http or similar framework
- REST API endpoints for core functionality
- A desire to serve documentation alongside the API

The team wants to adopt industry-standard tools: swaggo/swag for annotation-driven spec generation and Redoc for rendering interactive documentation.

## Goals / Non-Goals

**Goals:**
- Enable automatic OpenAPI 3.0 spec generation from Go code annotations using swaggo/swag
- Serve an interactive API documentation page using Redoc at a dedicated endpoint
- Make the API self-documenting through annotations in handler source code
- Maintain spec files in version control for distribution and CI/CD integration
- Support documentation discovery and searchability

**Non-Goals:**
- Full client code generation (swagger-codegen integration is out of scope)
- Custom documentation theming beyond Redoc's built-in options
- GraphQL or non-REST API documentation (REST-only for this phase)
- API versioning strategy (specs will support multiple versions but strategy is separate)

## Decisions

### Decision: Use swaggo/swag for spec generation
**Choice**: Adopt swaggo/swag over manual OpenAPI authoring or other tools.

**Rationale**: 
- Annotations live directly in Go source code, keeping documentation co-located with implementation
- Zero-runtime overhead—specs are pre-generated during build
- Mature tooling with strong Go community support
- Reduces documentation drift by enforcing spec freshness via build process

**Alternatives considered:**
- Manual OpenAPI authoring: High maintenance burden, specs drift from code
- Other Go tools (e.g., spectacle): Smaller ecosystem, less mature

### Decision: Serve documentation via HTML with Redoc
**Choice**: Use Redoc as the interactive documentation renderer.

**Rationale**:
- Redoc is purpose-built for OpenAPI specs and provides excellent UX
- Can be embedded in a static HTML file or served from CDN
- Responsive design works across desktop, tablet, and mobile
- Supports schema expansion, search, and dark mode
- Minimal JavaScript footprint and no backend dependencies

**Alternatives considered**:
- Swagger UI: More interactive but heavier JavaScript bundle
- Static HTML generation: Limited interactivity and search
- Custom documentation: High maintenance burden

### Decision: Generate specs during build, not runtime
**Choice**: Run `swag init` as part of the build process to pre-generate specs.

**Rationale**:
- No runtime overhead or dependencies in the running API
- Specs are deterministic and reproducible across builds
- Easy to version control and audit
- CI/CD can validate spec quality before deployment

**Alternative considered**:
- Runtime spec generation: Adds latency and complexity; less suitable for static assets

### Decision: Serve spec and docs from dedicated HTTP endpoints
**Choice**: Expose `/api/swagger.json` for the spec and `/api/docs` for the HTML documentation page.

**Rationale**:
- Clear, discoverable endpoints following REST conventions
- Allows documentation to be browsed without impacting API performance
- Easy to route through reverse proxies and load balancers
- Separates concerns: spec availability and rendering

### Decision: Use Go comments for Swagger annotations
**Choice**: Follow swaggo/swag convention of placing annotations in Go comment blocks above handlers.

**Rationale**:
- Annotations stay close to implementation, improving maintainability
- No separate annotation files or DSL to learn
- Go-native approach familiar to Go developers
- Tooling automatically parses Go comments

## Risks / Trade-offs

### Risk: Documentation drift
**Mitigation**: Enforce spec generation in CI/CD; fail builds if annotations are missing or invalid. Add pre-commit hooks to regenerate specs before commits.

### Risk: Large spec file growth
**Mitigation**: Monitor spec size; consider API versioning or endpoint grouping if spec becomes unwieldy. Specs are typically < 1MB even for large APIs.

### Risk: Annotation complexity
**Mitigation**: Provide clear examples and documentation in the project README. Establish linting rules for swaggo annotation format consistency.

### Risk: Redoc CDN dependency
**Mitigation**: Redoc can be self-hosted or bundled. For now, use CDN; add self-hosting option if needed for air-gapped deployments.

### Trade-off: Spec-first vs. code-first approach
**Choice**: Code-first (annotations in handlers). Spec is generated from code, not the reverse.

**Rationale**: Aligns with Go development practices and keeps spec and implementation in sync. Spec-first would require separate OpenAPI authoring and code generation.

## Migration Plan

1. **Prepare**: Add swaggo/swag dependency to `go.mod`
2. **Annotate**: Add Swagger annotations to existing HTTP handlers in `internal/api/`
3. **Generate**: Run `swag init` to generate initial specs
4. **Serve**: Add HTTP route for `/api/swagger.json` and `/api/docs` serving static HTML with Redoc
5. **Integrate**: Add spec generation to build process (Makefile, CI/CD) and pre-commit hooks
6. **Document**: Update project README with link to docs endpoint and guide for adding annotations to new endpoints
7. **Validate**: Test that all endpoints are documented and spec is accurate
8. **Deploy**: Release updated API with documentation live

## Open Questions

- Should docs be available in production or dev/staging only? (Recommendation: Available in all environments for transparency)
- Do we need spec versioning for backward-compatible API updates? (Out of scope for this phase; specs can be updated in-place)
- Should endpoint pagination and filtering be documented? (Yes, as part of handler annotation requirements)
- Do we need to generate client libraries from the spec? (Out of scope; spec is available for third-party tools)
