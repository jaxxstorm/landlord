## Why

The CLI can list tenants but cannot fetch a single tenant or update a tenant's desired state, forcing users to fall back to raw API calls. Adding these commands improves day-to-day operability and keeps the CLI aligned with supported tenant lifecycle actions.

## What Changes

- Add a `get` command to retrieve a single tenant by ID/name.
- Add an `update` command to modify a tenant's image and/or configuration.
- Extend CLI output formatting for the new commands.
- Update CLI configuration and tests to cover new tenant interactions.

## Capabilities

### New Capabilities
- (none)

### Modified Capabilities
- `landlord-cli`: extend CLI requirements to support `get` and `update` tenant operations with appropriate flags and output.

## Impact

- CLI command surface and help output.
- CLI client interactions with existing tenant read/update API endpoints.
- Tests for CLI commands and client behavior.
