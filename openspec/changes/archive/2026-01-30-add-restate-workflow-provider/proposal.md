## Why

Landlord currently supports Step Functions for production workflows and a mock provider for testing, but lacks a solution that works seamlessly for both local development and production. Restate.dev provides durable execution with strong consistency guarantees, works locally without cloud dependencies, and scales to production with multiple execution mechanisms (Lambda, ECS Fargate, Kubernetes). This enables developers to test workflows locally with the same execution semantics they'll have in production, while giving operators flexibility to choose the deployment model that fits their infrastructure.

## What Changes

- Add restate.dev as a third workflow provider option alongside step-functions and mock
- Implement workflow provisioning using the restate.dev Go SDK
- Add configuration for restate server endpoint (for local development and production deployments)
- Support restate's multiple execution mechanisms (Lambda, ECS Fargate, Kubernetes) as deployment options as well as using restate's default compute options as well
- Create restate workflow provider implementation in `internal/workflow/providers/restate/`
- Add restate-specific configuration to workflow config with server endpoint, execution mechanism, and authentication options
- Document restate provider setup for local development and production deployment patterns
- Add restate.dev Go SDK as a project dependency

## Capabilities

### New Capabilities
- `restate-workflow-provider`: Complete integration with restate.dev for workflow orchestration, including service registration, workflow execution, state management, and support for multiple execution mechanisms

### Modified Capabilities
- `workflow-provisioning`: Add restate as a third provider option (alongside step-functions and mock) with provider-specific configuration for server endpoint and execution mechanism

## Impact

**Code:**
- New package: `internal/workflow/providers/restate/` with restate provider implementation
- Modified: `internal/workflow/registry.go` to register restate provider
- Modified: `internal/config/workflow.go` to add restate configuration options

**Configuration:**
- New workflow provider option: `workflow.default_provider = "restate"`
- New config section: `workflow.restate` with fields:
  - `endpoint`: Restate server endpoint (e.g., `http://localhost:8080` for local, production URL for deployed)
  - `execution_mechanism`: Deployment target (lambda, fargate, kubernetes, self-hosted)
  - `service_name`: Service identifier for workflow registration
  - Authentication options (API keys, IAM roles depending on execution mechanism)

**Dependencies:**
- Add restate.dev Go SDK (`github.com/restatedev/sdk-go` or similar)
- No changes to existing step-functions or mock providers

**Documentation:**
- Setup guide for local development with restate (Docker Compose example)
- Production deployment patterns for each execution mechanism
- Configuration reference for all restate options
- Migration guide from mock/step-functions to restate

**Testing:**
- Local development workflow becomes primary use case
- Developers can test durable execution locally without AWS credentials
- Production deployments gain flexibility to use Lambda, Fargate, or Kubernetes based on operational preferences
