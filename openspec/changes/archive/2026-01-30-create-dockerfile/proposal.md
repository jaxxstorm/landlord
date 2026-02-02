## Why

Landlord is currently built and run locally via `make build` and `./landlord`, but lacks containerized deployment support. A Dockerfile enables consistent, reproducible builds across environments and facilitates deployment to container orchestration platforms (Kubernetes, Docker Swarm, ECS), reducing "works on my machine" issues and enabling seamless CI/CD integration.

## What Changes

- Add production-ready Dockerfile with multi-stage build for optimized image size
- Support multiple runtime environments (local development, Kubernetes, Docker Compose)
- Enable container-native deployment workflows and orchestration
- Include health check configuration and optimized container runtime settings

## Capabilities

### New Capabilities
- `docker-image-build`: Multi-stage Dockerfile for building and running Landlord as a container image
- `container-runtime-config`: Container runtime settings including health checks, resource limits, and environment variable handling
- `ci-cd-container-integration`: Dockerfile compatibility with CI/CD pipelines for automated image builds and registry pushes

### Modified Capabilities

<!-- No existing specs require changes for basic Dockerfile support -->

## Impact

- **Code**: No changes to Go codebase; Dockerfile is additive
- **Dependencies**: Docker/container runtime becomes required for containerized deployments (optional for local dev)
- **Build Process**: `docker build` becomes an alternative to `make build`
- **Deployment**: Enables cloud-native deployment patterns (K8s, ECS, Docker Compose)
- **CI/CD**: Integrates with container registry workflows (Docker Hub, ECR, GHCR)
