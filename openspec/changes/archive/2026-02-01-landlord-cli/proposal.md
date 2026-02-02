## Why

Landlord lacks a first-party CLI for interacting with its API, making local development, demos, and operational workflows slower and inconsistent. A dedicated CLI with consistent output and configuration improves usability and reduces friction for common tenant operations.

## What Changes

- Add a new `landlord-cli` command in `cmd/cli` with Cobra command structure and Viper-based configuration.
- Implement verb-based commands for API interaction: `create`, `list`, `delete` (with room to add `get`/`update` later).
- Use Lip Gloss for styled output and Fang for CLI theming and help presentation.
- Provide configuration via config file, env vars, and flags (Viper), including API base URL and auth settings if needed.
- Add `go run` developer workflow for testing CLI commands against a running API.

## Capabilities

### New Capabilities
- `landlord-cli`: CLI client for Landlord API operations with styled output and configurable settings.

### Modified Capabilities
<!-- None -->

## Impact

- New CLI entrypoint under `cmd/cli` and supporting packages for API client, formatting, and config.
- New dependencies: `github.com/charmbracelet/lipgloss` and `github.com/charmbracelet/fang`.
- Documentation updates for CLI usage and local development/testing workflow.
