## 1. Dockerfile Creation

- [x] 1.1 Create multi-stage Dockerfile with builder stage (Go compilation) and runtime stage (distroless base)
- [x] 1.2 Configure builder stage to use Go 1.24+ and build from go.mod
- [x] 1.3 Configure runtime stage to use distroless base:nonroot image
- [x] 1.4 Set EXPOSE 8080 for the HTTP API port
- [x] 1.5 Set ENTRYPOINT to run the Landlord binary

## 2. Health Check and Runtime Configuration

- [x] 2.1 Add HEALTHCHECK instruction to Dockerfile (using curl or similar to /health endpoint)
- [x] 2.2 Document environment variable configuration requirements in Dockerfile comments
- [x] 2.3 Set appropriate working directory and user permissions in Dockerfile

## 3. Build Optimization

- [x] 3.1 Add .dockerignore file to exclude unnecessary files (*.md, .git/, vendor/, docs/, etc.)
- [x] 3.2 Configure builder stage to cache Go modules layer (go mod download before COPY source)
- [x] 3.3 Use Go build flags to optimize binary size if needed (e.g., -ldflags="-s -w")

## 4. Documentation and Integration

- [x] 4.1 Update README.md with Docker build instructions (`docker build -t landlord:latest .`)
- [x] 4.2 Document how to run container with environment variable configuration examples
- [x] 4.3 Add Docker run examples in README (e.g., with database connectivity, port binding)
- [x] 4.4 Document health check verification in README

## 5. Testing and Validation

- [x] 5.1 Build Dockerfile successfully from repository root
- [x] 5.2 Verify built image size is reasonable (under 200MB)
- [x] 5.3 Run container and verify API is accessible at port 8080
- [x] 5.4 Run container with environment variable configuration and verify settings applied
- [x] 5.5 Test health check endpoint responds correctly
- [x] 5.6 Verify graceful shutdown (SIGTERM handling) works in container
- [x] 5.7 Test container image can be tagged and would be compatible with registries

## 6. CI/CD Compatibility (Optional)

- [x] 6.1 Verify Dockerfile is compatible with docker buildx for multi-platform builds
- [x] 6.2 Document CI/CD build and push workflow in project docs
