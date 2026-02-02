## Context

The CLI currently supports list/create/delete but does not expose single-tenant retrieval or updates, forcing users to call the HTTP API directly. We need to extend the existing Cobra/Viper/Fang CLI and the internal CLI client to support get and update operations, while keeping output consistent with the existing list command and tests.

## Goals / Non-Goals

**Goals:**
- Add `get` and `update` subcommands to the CLI with clear, validated flags.
- Reuse the existing internal CLI client for API interaction, expanding it to include get/update calls.
- Provide consistent, styled output for single-tenant and updated-tenant responses.
- Extend unit/integration-style tests to cover new commands and client methods.

**Non-Goals:**
- Changing API behavior or adding new API endpoints.
- Adding interactive prompts or TUI features beyond existing styling.
- Redesigning the CLI config system or output format for existing commands.

## Decisions

- **Command surface**: add `landlord-cli get` and `landlord-cli update` as peer commands to `list/create/delete` for clarity and discoverability. Alternatives considered: nesting under `tenant` or overloading `list` with filters; rejected to keep existing verbs and avoid breaking changes.
- **Identifier input**: accept `--tenant-id` for both commands, with `get` allowing either UUID or tenant name to match server lookup behavior. Alternative: separate `--id` and `--name` flags; rejected to avoid extra complexity and user confusion.
- **Update payload**: support optional `--image` and `--config` (path or inline JSON) flags, passing only provided fields to the API to avoid unintentional resets. Alternative: require full tenant spec on update; rejected because it is error-prone for small edits.
- **Output**: reuse existing lipgloss styles and add a single-tenant table or key/value view, matching list output where possible. Alternative: raw JSON output; rejected for default flow but could be added later via a `--json` flag.
- **Testing strategy**: extend existing `cmd/cli` tests for command parsing and `internal/cli` tests for HTTP interactions using `httptest`. Alternative: end-to-end tests with real API; rejected for speed and determinism.

## Risks / Trade-offs

- [Partial updates may be ambiguous] → Only send fields explicitly provided and document behavior in command help.
- [Server expects full desired state] → Confirm update endpoint semantics in specs; if needed, fetch current tenant before patching and merge locally.
- [Output consistency drift] → Centralize formatting helpers and reuse existing table rendering patterns.

## Migration Plan

- Implement client and command changes in a backward-compatible way; no config changes required.
- No data migrations. Rollback by reverting CLI changes; API remains unchanged.

## Open Questions

- Should `update` accept config as a file path, inline JSON, or both?
- Should `get` allow lookup by tenant name when names are not unique?
- Do we need a `--json` output flag now or defer to later enhancement?
