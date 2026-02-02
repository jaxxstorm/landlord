## Why

Compute configuration is currently opaque and failing at runtime (e.g., create/set returning "Unsupported Config Type"), which blocks real tenant provisioning and makes future providers (ECS/Kubernetes) hard to validate. We need a provider-declared config schema and a consistent way for API/CLI to accept, validate, and display compute config.

## What Changes

- Add a compute config discovery endpoint that returns the active compute provider and its JSON schema (and optional defaults if available).
- Add a CLI `compute` command to fetch and display the active compute config schema from the API.
- Ensure tenant create/set accept flexible compute config payloads, validate them against the active provider schema, and persist them as desired config.
- Improve error messaging when compute config is missing, malformed, or uses the wrong provider type.

## Capabilities

### New Capabilities
- `compute-config-discovery`: Expose provider-specific compute config schema (and optional defaults) via API and CLI.

### Modified Capabilities
- `tenant-create-api`: Accept and validate provider-specific compute_config payloads; return stored config.
- `tenant-update-api`: Accept and validate compute_config updates; preserve flexible JSON structure.
- `tenant-request-validation`: Validate compute_config against the active provider schema and report actionable errors.
- `landlord-cli`: Add `compute` command and ensure create/set pass compute_config payloads.
- `http-api-server`: Document and serve the compute config discovery endpoint.

## Impact

- API: new compute config discovery endpoint; create/update request validation behavior tightened.
- CLI: new command and clearer config error reporting.
- Compute providers: must surface a JSON schema (and optional default config) for their config payloads.
- Persistence: desired_config continues to be stored as JSONB; no schema migration expected.
