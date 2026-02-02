# Restate Provider - Getting Started

Landlord supports [Restate.dev](https://restate.dev) as a workflow provider, enabling durable execution with strong consistency guarantees. Restate works seamlessly for both local development and production deployments.

## What is Restate?

Restate is a durable execution platform that provides:
- **Strong Consistency**: ACID properties with durable state management
- **Local Development**: Run workflows locally with production semantics via Docker
- **Multiple Deployment Options**: Lambda, ECS Fargate, Kubernetes, or self-hosted
- **Automatic Recovery**: Workflows recover automatically from failures with state intact
- **Developer Experience**: Type-safe SDK and straightforward error handling

## Quick Start - Local Development

### Prerequisites
- Docker and Docker Compose
- Go 1.24.7 or later

### 1. Start Restate Locally

```bash
cd /path/to/landlord
docker-compose up restate
```

This starts a Restate server on `http://localhost:8080`.

### 2. Configure Landlord

Edit your `config.yaml`:

```yaml
workflow:
  default_provider: "restate"
  restate:
    endpoint: "http://localhost:8080"
    execution_mechanism: "local"
    auth_type: "none"
    timeout: 30m
    retry_attempts: 3
```

Or use environment variables:

```bash
export LANDLORD_WORKFLOW_DEFAULT_PROVIDER=restate
export LANDLORD_WORKFLOW_RESTATE_ENDPOINT=http://localhost:8080
export LANDLORD_WORKFLOW_RESTATE_EXECUTION_MECHANISM=local
export LANDLORD_WORKFLOW_RESTATE_AUTH_TYPE=none
```

### 3. Run Landlord

```bash
landlord server
```

Landlord will connect to Restate and register as a provider.

During startup, Landlord also registers the tenant lifecycle workflows with the Restate backend. Registration is idempotent, so restarts safely re-register existing workflows. If Restate is unavailable, Landlord logs a warning and continues startup; workflow execution will return a "workflow not found" error until registration succeeds.

### 4. Create and Execute Workflows

Workflows work exactly the same as with other providers:

```bash
# Create a workflow
landlord workflow create my-workflow \
  --definition "workflow.yaml"

# Start execution
landlord workflow execute my-workflow \
  --input '{"key": "value"}'

# Check status
landlord workflow status my-workflow
```

## Key Concepts

### Workflows and Services

- **Workflow** (in Landlord): A blueprint for execution defined in workflow configuration
- **Service** (in Restate): The actual deployed unit that handles workflow execution
- **Service Name**: Derived from workflow ID (e.g., `my-workflow` → `MyWorkflow`)

### Execution Mechanisms

The execution mechanism tells Restate where and how to invoke your services:

- **`local`** (default): Use Restate's local compute for development
- **`lambda`**: Invoke AWS Lambda functions
- **`fargate`**: Invoke AWS ECS Fargate tasks
- **`kubernetes`**: Invoke Kubernetes services
- **`self-hosted`**: Direct HTTP calls to Restate instances

### State Management

Restate handles all state durability:
- Workflow state is automatically persisted
- Failed executions resume from the last successful step
- State is fully recoverable even if Restate goes down and restarts

## Next Steps

- [Configuration Reference](restate-configuration.md) - All available options
- [Local Development Setup](restate-local-development.md) - Detailed Docker setup
- [Production Deployment](restate-production-deployment.md) - Deploy to Lambda, Fargate, or Kubernetes
- [Authentication](restate-authentication.md) - Secure your Restate deployment
- [Troubleshooting](restate-troubleshooting.md) - Common issues and solutions
