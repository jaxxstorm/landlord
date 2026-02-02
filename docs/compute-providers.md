# Compute Providers

Compute providers provision the runtime resources for tenants (containers, networking, and related infrastructure). Landlord ships with a small set of in-tree providers and supports adding new ones.

## Supported providers

| Provider | Use case | Notes |
| --- | --- | --- |
| docker | Local development and simple deployments | Uses Docker Engine for container provisioning |
| mock | Tests and demos | In-memory, no real infrastructure |

## Configuration example

```yaml
compute:
  default_provider: docker
  docker:
    host: ""
    network_name: bridge
    network_driver: bridge
    label_prefix: landlord
```

## Provider interface

All compute providers implement the `compute.Provider` interface:

```go
type Provider interface {
    Name() string
    Provision(ctx context.Context, spec *TenantComputeSpec) (*ProvisionResult, error)
    Update(ctx context.Context, tenantID string, spec *TenantComputeSpec) (*UpdateResult, error)
    Destroy(ctx context.Context, tenantID string) error
    GetStatus(ctx context.Context, tenantID string) (*ComputeStatus, error)
    Validate(ctx context.Context, spec *TenantComputeSpec) error
    ValidateConfig(config json.RawMessage) error
    ConfigSchema() json.RawMessage
    ConfigDefaults() json.RawMessage
}
```

## Adding a new provider

1. Create a package under `internal/compute/providers/<name>/`.
2. Implement the provider interface.
3. Register the provider in `cmd/landlord/main.go`.
4. Add tests for the provider behavior.

## Tenant compute specification

Compute providers receive a `TenantComputeSpec` describing containers, resources, and provider-specific config.

## Error handling

Use the standard error types from `internal/compute/errors.go` so the API can surface consistent responses.
