# Workflow Providers

Workflow providers execute provisioning plans and manage long-running tenant workflows. Landlord ships with multiple provider implementations so you can pick the best fit for your environment.

## Supported providers

| Provider | Use case | Notes |
| --- | --- | --- |
| restate | Local development and production | Durable execution with local-friendly semantics |
| step-functions | AWS-native orchestration | Managed service with AWS integrations |
| mock | Tests and local experimentation | In-memory, non-durable execution |

## Restate

Restate provides durable workflow execution with strong consistency guarantees and a developer-friendly local setup.

Public documentation:
- https://restate.dev/
- https://docs.restate.dev/

Configuration example:

```yaml
workflow:
  default_provider: restate
  restate:
    endpoint: http://localhost:8080
    admin_endpoint: http://localhost:9070
    execution_mechanism: direct
```

## AWS Step Functions

Step Functions is a fully managed AWS workflow service. This provider integrates with AWS Step Functions state machines.

Public documentation:
- https://docs.aws.amazon.com/step-functions/

Configuration example:

```yaml
workflow:
  default_provider: step_functions
  step_functions:
    region: us-west-2
    role_arn: arn:aws:iam::123456789012:role/LandlordStepFunctionsRole
```

## Mock provider

The mock provider executes workflows in memory. It is intended for tests and local usage when durability is not required.

Configuration example:

```yaml
workflow:
  default_provider: mock
```

## Switching providers

Switching providers is a configuration change only. Update `workflow.default_provider` and the provider-specific config block.

## Worker integration

Workflow providers rely on worker types to execute compute actions. See `workers.md` for worker types and configuration.
