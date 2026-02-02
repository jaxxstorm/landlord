# Restate Workflow Provider

Durable execution with strong consistency guarantees. Works locally with Docker or deploys to production (Lambda, Fargate, Kubernetes).

## Quick Links

- **[Getting Started](getting-started.md)** - 5 minute quick start for local development
- **[Configuration Reference](configuration.md)** - All configuration options documented
- **[Local Development](local-development.md)** - Docker Compose setup for development
- **[Production Deployment](production-deployment.md)** - Deploy to Lambda, Fargate, Kubernetes, or self-hosted
- **[Authentication](authentication.md)** - Authentication types: none, api_key, iam
- **[Troubleshooting](troubleshooting.md)** - Common issues and solutions
- **[Example Configs](.)** - Local and production configuration examples

## Features

- **Durable Execution**: Strong consistency with ACID guarantees
- **Local Development**: Docker container with production semantics
- **Multiple Deployments**: Lambda, Fargate, Kubernetes, self-hosted
- **State Management**: Automatic state persistence and recovery
- **Error Recovery**: Workflows resume from last successful step
- **Type-Safe SDK**: Comprehensive Go SDK with good error handling

## Start Here

### For Local Development

1. [Getting Started](getting-started.md) - Overview and quick start
2. [Local Development](local-development.md) - Docker Compose setup
3. [Configuration Reference](configuration.md) - Customize settings

### For Production

1. [Production Deployment](production-deployment.md) - Choose your deployment model
2. [Authentication](authentication.md) - Setup security
3. [Troubleshooting](troubleshooting.md) - Common production issues

## Configuration Examples

### Local Development

```yaml
workflow:
  default_provider: "restate"
  restate:
    endpoint: "http://localhost:8080"
    execution_mechanism: "local"
    auth_type: "none"
```

See: [examples-config-local.yaml](examples-config-local.yaml)

### Production (AWS Lambda)

```yaml
workflow:
  default_provider: "restate"
  restate:
    endpoint: "https://restate.prod.internal"
    execution_mechanism: "lambda"
    auth_type: "iam"
```

See: [examples-config-production-lambda.yaml](examples-config-production-lambda.yaml)

## Workflow Providers

See [../README.md](../README.md) for comparison with other workflow providers (mock, Step Functions).

---

**Documentation Version**: 1.0  
**Last Updated**: January 30, 2026
