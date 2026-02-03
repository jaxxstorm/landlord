## 1. ECS Provider Scaffolding

- [x] 1.1 Create `internal/compute/providers/ecs` package with provider skeleton implementing `compute.Provider`
- [x] 1.2 Define ECS provider config struct and JSON schema for `compute_config` validation
- [x] 1.3 Add ECS provider tests for config validation and schema defaults

## 2. AWS Client & Credential Helpers

- [x] 2.1 Add shared AWS config helper to load SDK default credential chain
- [x] 2.2 Implement optional assume-role parameters in helper (role ARN, external ID, session name)
- [x] 2.3 Add unit tests for AWS config helper behavior without real AWS calls

## 3. ECS Provisioning Operations

- [x] 3.1 Implement Provision: create/ensure ECS service per tenant using task definition and cluster
- [x] 3.2 Implement Update: update ECS service task definition / desired count idempotently
- [x] 3.3 Implement Destroy: delete ECS service idempotently (ignore not found)
- [x] 3.4 Implement GetStatus: query ECS service status and map to compute status

## 4. Workflow Worker Integration

- [x] 4.1 Register ECS provider in worker compute registry
- [x] 4.2 Ensure workflow compute resolver can select ECS provider via config/labels
- [x] 4.3 Add tests for workflow worker invoking ECS provider path (mocked)

## 5. Documentation & Examples

- [x] 5.1 Update docs with ECS compute provider configuration example
- [x] 5.2 Document credential resolution behavior (env, shared config/SSO, metadata)
