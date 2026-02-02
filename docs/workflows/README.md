# Workflow Providers

Landlord supports multiple workflow providers, each suited for different use cases.

## Available Providers

### Restate

Durable execution with strong consistency guarantees.

- Local development with production semantics
- Multiple deployment options (Lambda, Fargate, Kubernetes, self-hosted)
- Automatic state recovery and error handling
- Works on premise or in cloud

Best for: Teams wanting local development to match production execution semantics.

Documentation:
- Getting Started: restate/restate-getting-started.md
- Configuration Reference: restate/restate-configuration.md
- Local Development Setup: restate/restate-local-development.md
- Production Deployment: restate/restate-production-deployment.md
- Authentication: restate/restate-authentication.md
- Troubleshooting: restate/restate-troubleshooting.md

---

### AWS Step Functions

Fully managed serverless workflow service from AWS.

- AWS native integration
- Managed service (no operational overhead)
- Pay-per-execution pricing
- Integrates with Lambda, SNS, SQS, and other AWS services

Best for: Teams already using AWS heavily.

Requires: AWS account, IAM credentials, configuration of region and role ARN.

---

### Mock Provider

In-memory provider for testing.

- No dependencies or setup required
- Fast for unit and integration tests
- Workflows run in-memory (no persistence)
- Ideal for CI and development environments

Best for: Testing and development environments.

---

## Comparison

| Feature | Restate | Step Functions | Mock |
| --- | --- | --- | --- |
| Local Development | Yes | No | Yes |
| Production Ready | Yes | Yes | No |
| Durable Execution | Yes | Yes | No |
| State Recovery | Yes | Yes | No |
| Multi-Cloud | Yes | No | N/A |
| AWS Integration | Yes | Yes | N/A |
| Setup Complexity | Medium | Low | Very Low |
| Operational Overhead | Medium | None | None |

---

## Quick Start

### Local Development (Recommended)

Start with Restate for local development:

```yaml
workflow:
  default_provider: "restate"
  restate:
    endpoint: "http://localhost:8080"
    execution_mechanism: "local"
    auth_type: "none"
```

See restate/restate-getting-started.md.

### Production Deployment

Choose based on your infrastructure:

- Restate + Lambda: restate/restate-production-deployment.md#aws-lambda-deployment
- Restate + Fargate: restate/restate-production-deployment.md#aws-ecs-fargate-deployment
- Restate + Kubernetes: restate/restate-production-deployment.md#kubernetes-deployment
- AWS Only: configuration.md

### Testing

Use the mock provider for tests:

```yaml
workflow:
  default_provider: "mock"
```

---

## Switching Providers

Switching between providers only requires changing configuration:

```yaml
workflow:
  default_provider: "mock"  # Change to: restate, step_functions, or mock
```

No code changes needed. Existing workflows remain compatible across providers.

---

## Need Help?

- Getting Started: provider-specific guides above
- Configuration: configuration.md
- Issues: provider troubleshooting guides
