## ADDED Requirements

### Requirement: Dockerfile exists at repository root
The system SHALL include a production-ready Dockerfile at the repository root that can build a container image of Landlord.

#### Scenario: Build succeeds with default arguments
- **WHEN** a user runs `docker build -t landlord:latest .` from the repository root
- **THEN** the build completes successfully and produces a container image tagged `landlord:latest`

#### Scenario: Multi-stage build produces optimized image
- **WHEN** the Dockerfile completes building
- **THEN** the final image uses a distroless base (nonroot variant) and is under 200MB in size

### Requirement: Go binary is compiled during build
The system SHALL compile the Landlord Go binary during the Docker build process using a builder stage.

#### Scenario: Binary compilation in builder stage
- **WHEN** the Dockerfile builder stage executes
- **THEN** the Go compiler runs with the correct Go version and produces an executable binary at `/app/landlord`

#### Scenario: Build artifacts excluded from final image
- **WHEN** the final runtime stage is created
- **THEN** Go build tools, source files, and intermediate build artifacts are not present in the final image

### Requirement: .dockerignore file filters build context
The system SHALL include a `.dockerignore` file to exclude unnecessary files from the Docker build context.

#### Scenario: Unnecessary files excluded from context
- **WHEN** the Docker build reads the build context
- **THEN** files matching patterns in `.dockerignore` (e.g., `*.md`, `.git/`, build artifacts) are excluded from the builder stage
