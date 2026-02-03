## Why

The current compute configuration requires `default_provider` and `defaults`, which forces users to define provider settings they do not intend to use and makes startup configuration noisy. We need a provider-keyed compute config so providers are enabled only when configured, and clients can request a specific provider schema from the compute discovery API.

## What Changes

- **BREAKING** Redefine the `compute` config schema to be provider-keyed (providers are enabled by presence) and remove `compute.default_provider` and `compute.defaults`.
- **BREAKING** Update compute config discovery API and CLI to accept a provider identifier and return schema/defaults for that provider, with clear errors for unknown or disabled providers.
- Allow single-provider configurations without additional flags or boilerplate; update config validation accordingly.
- Update documentation, config examples, and generated API schemas to reflect the new compute config structure and discovery behavior.

## Capabilities

### New Capabilities
- `compute-provider-configuration`: Define the compute stanza shape, provider enablement semantics, and validation rules for provider-specific configuration blocks.

### Modified Capabilities
- `compute-config-discovery`: Discovery endpoint/CLI now returns schema/defaults for a requested provider rather than only the active provider.
- `http-api-server`: Compute config discovery endpoint contract and schema updated to include a provider parameter.

## Impact

- Configuration structs, validation, and loading (YAML/JSON/env/flags) for compute providers.
- API handlers and CLI commands for compute config discovery.
- Documentation and example configs (`config.example.yaml`, `docker.config.yaml`, `ecs.config.yaml`, tests).
