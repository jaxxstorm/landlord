## Context

Landlord is a Go-based workflow and compute orchestration platform that currently lacks container image support. The project:
- Uses Go 1.24+ with standard library tooling
- Builds via `make build` producing a binary executable
- Runs on local development machines and servers via direct binary execution
- Has dependencies on Docker SDK (for Docker compute provider), PostgreSQL, SQLite, and AWS SDK
- Exposes an HTTP API with Swagger documentation at runtime

Current deployment limitation: operators must manually compile the Go binary on the target machine or use complex CI/CD scripting to build and distribute pre-compiled binaries.

## Goals / Non-Goals

**Goals:**
- Create a production-ready multi-stage Dockerfile that produces optimized container images
- Enable deployment to Kubernetes, Docker Compose, ECS, and other container orchestration platforms
- Support both local development and production environments via environment-based configuration
- Ensure images are minimal (small footprint) using distroless base for runtime
- Enable automated builds in CI/CD pipelines (GitHub Actions, GitLab CI, etc.)
- Support health checks for container orchestrators

**Non-Goals:**
- Docker Compose orchestration templates (separate from base Dockerfile)
- Container registry hosting/authentication setup (operator responsibility)
- Kubernetes manifests or Helm charts (separate from base Dockerfile)
- Multi-architecture (arm64) builds in initial version
- Custom build-time configuration beyond standard Go build flags

## Decisions

**Decision 1: Multi-stage Build (Builder + Runtime)**
- **Choice**: Use two stages—builder stage for compilation, minimal distroless base for runtime
- **Rationale**: Reduces final image size (no build tools in production image). Builder stage ensures consistent compilation environment; distroless base minimizes attack surface and eliminates unnecessary OS overhead.
- **Alternative considered**: Single-stage Alpine-based image. Trade-off: Alpine adds ~5MB vs distroless, plus maintains package manager and shell, increasing security footprint.

**Decision 2: Configuration via Environment Variables**
- **Choice**: Read Landlord configuration from environment variables at runtime (via existing Viper support)
- **Rationale**: Aligns with container best practices (12-factor app). Viper already supports env var binding; no code changes needed.
- **Alternative considered**: Mount config files via volumes. Trade-off: More complex deployment; environment variables easier for orchestrators.

**Decision 3: Health Check Endpoint**
- **Choice**: Use existing `/health` or similar HTTP endpoint if available; otherwise add minimal health check endpoint
- **Rationale**: Container orchestrators require health checks for reliable restarts and load balancing
- **Alternative considered**: Process-level checks (liveness of binary). Trade-off: HTTP endpoint is more robust and observable.

**Decision 4: Build Arguments for Base Image Version**
- **Choice**: Use distroless base image with pinned version (e.g., `gcr.io/distroless/base:nonroot-latest`)
- **Rationale**: Nonroot user improves security; pinning version ensures reproducible builds
- **Alternative considered**: Alpine base. Trade-off: Distroless is preferred for production security posture.

## Risks / Trade-offs

**[Risk] distroless base lacks debugging tools**
→ *Mitigation*: Operators can build separate debug image using Alpine when needed; use `docker exec` with kubectl for K8s debugging

**[Risk] Go binary becomes larger without build optimization**
→ *Mitigation*: Use standard Go build flags (`-ldflags="-s -w"`) to strip symbols if needed; multi-stage build already minimizes image overhead

**[Risk] No shell in distroless image breaks some deployment scripts**
→ *Mitigation*: Document this limitation; operators using custom init logic should use builder stage or Alpine variant

**[Trade-off] Environment variable configuration vs. config files**
→ Simpler deployment model but less flexible for complex multi-tenant scenarios; Landlord's multi-tenant support (tenant schema) is already database-driven, so env vars sufficient

## Migration Plan

1. Create Dockerfile at repository root with multi-stage build
2. Add `.dockerignore` to exclude build artifacts and documentation
3. Document build commands in README (`docker build -t landlord:latest .`)
4. Verify image builds successfully and runs with basic configuration
5. Add CI/CD build job (optional; documented but not required in this change)
6. Existing binary build process (`make build`, direct execution) remains unchanged—Dockerfile is purely additive

Rollback: Simply continue using binary build process; Dockerfile has no runtime dependencies.

## Open Questions

- Should we provide Alpine variant for debugging scenarios?
- Should initial image support both amd64 and arm64 architectures?
- What health check endpoint should be used? (Assumes existing `/health` or similar in API layer)
