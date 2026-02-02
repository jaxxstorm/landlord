## ADDED Requirements

### Requirement: Dockerfile is compatible with CI/CD build systems
The system SHALL support automated image builds in CI/CD pipelines without requiring special modification or wrapper scripts.

#### Scenario: CI/CD pipeline builds image successfully
- **WHEN** a CI/CD system (e.g., GitHub Actions, GitLab CI) clones the repository and runs `docker build`
- **THEN** the build succeeds and produces a valid container image

#### Scenario: Build args support custom tags and registries
- **WHEN** a CI/CD system passes build arguments (e.g., `--build-arg IMAGE_TAG=v1.2.3`)
- **THEN** the Dockerfile either uses these args for flexible tagging or ignores them without error

### Requirement: Multi-architecture build support
The system SHALL be compatible with multi-architecture Docker build tools (e.g., buildx) for building amd64 and arm64 images.

#### Scenario: Buildx compatible Dockerfile syntax
- **WHEN** a CI/CD system uses `docker buildx build --platform linux/amd64,linux/arm64`
- **THEN** the Dockerfile syntax is compatible and builds succeed for both architectures

#### Scenario: Base images support target architecture
- **WHEN** the build targets arm64 architecture
- **THEN** the selected base image (distroless) provides arm64 variant

### Requirement: Container image can be pushed to registries
The system SHALL produce images compatible with standard container registries (Docker Hub, GitHub Container Registry, Amazon ECR, etc.).

#### Scenario: Image tag format is registry-compatible
- **WHEN** the built image is tagged with registry format (e.g., `ghcr.io/owner/landlord:v1.0`)
- **THEN** the image can be pushed to the specified registry without re-tagging

#### Scenario: Image pushed to registry is executable
- **WHEN** an image is pulled from a registry and started as a container
- **THEN** the container runs correctly with the same behavior as locally-built images

### Requirement: Build reproducibility and caching
The system SHALL support Docker layer caching to enable fast incremental rebuilds during development.

#### Scenario: Unchanged code layers are cached
- **WHEN** a second build is run with identical Go source code
- **THEN** the builder stage uses cached layers and completes faster than the initial build

#### Scenario: Dependency layer caching works
- **WHEN** Go module dependencies (go.mod/go.sum) are unchanged
- **THEN** the `go mod download` layer is cached and reused across builds
