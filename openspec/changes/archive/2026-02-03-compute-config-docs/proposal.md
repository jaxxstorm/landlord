## Why

Compute provider configuration options are currently scattered and incomplete, making it hard for new users to assemble a correct compute_config payload. We need a single, readable reference that shows every option and also make CLI config input more flexible (YAML and file-based) to reduce friction.

## What Changes

- Publish full compute_config reference docs for each provider (docker, ecs, mock), listing every supported option with clear descriptions.
- Add multi-line, readable JSON and YAML examples for each provider’s compute_config in the docs.
- Extend CLI compute_config input to accept YAML as well as JSON.
- Support `file://` URIs for `--config` so users can load compute_config from local files.

## Capabilities

### New Capabilities
- `compute-config-reference-docs`: Provide complete, provider-specific compute_config reference documentation with full JSON/YAML examples.

### Modified Capabilities
- `landlord-cli`: Accept YAML and `file://` sources for `--config` compute_config input.
- `public-docs`: Require provider docs to include comprehensive compute_config option listings and JSON/YAML examples.

## Impact

- Docs: `docs/compute/*`, `docs/cli/README.md`, `docs/configuration.md`, `docs/quickstart.md`, root `README.md` as needed.
- CLI parsing: `cmd/cli` config parsing for `--config` to support YAML and `file://` URIs.
- Tests: CLI parsing tests for YAML and file URI inputs.
