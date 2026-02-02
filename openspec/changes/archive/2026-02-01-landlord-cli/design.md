## Context

Landlord currently exposes only an HTTP API for tenant lifecycle operations. There is no first-party CLI, so common actions require manual curl calls or ad hoc scripts. The proposal introduces a new CLI under `cmd/cli` that uses Cobra for commands/flags, Viper for configuration, and Charm libraries for presentation. The CLI must be usable via `go run` in local development and should be configurable through config file, env vars, and flags.

## Goals / Non-Goals

**Goals:**
- Provide a new `landlord-cli` command with verb-based subcommands (create, list, delete).
- Use Cobra for command structure and Viper for configuration (flags, env vars, config file).
- Use Lip Gloss for styled output and Fang for CLI theming/help.
- Keep the CLI small and focused on API interactions, with a clean API client boundary.
- Support `go run ./cmd/cli` as the primary local dev invocation.

**Non-Goals:**
- Implement full CRUD coverage beyond create/list/delete in the first iteration.
- Add interactive TUI workflows or advanced auth beyond basic token/header support.
- Replace or refactor the existing API server or workflow subsystems.

## Decisions

- **Cobra + Viper integration for config loading.**
  - *Why*: Standard Go CLI pattern; fits requirement for config via file/env/flags.
  - *Alternative*: Kong (used elsewhere) was not chosen because the requirement specifies Cobra/Viper.

- **CLI command layout under `cmd/cli` with an internal client package.**
  - *Why*: Keeps CLI entrypoint isolated and allows simple testing/mocking of API calls.
  - *Alternative*: Implementing in `cmd/landlord` or adding flags to existing binaries would conflate concerns.

- **Lip Gloss for output rendering, Fang for CLI styling.**
  - *Why*: Matches requirement and provides consistent formatting for lists and status output.
  - *Alternative*: Plain text only; rejected due to requirement for styled output.

- **Minimal API surface in v1: create/list/delete tenants.**
  - *Why*: Focus on core workflows and match requested verbs; keep initial scope tight.
  - *Alternative*: Add get/update from day one; deferred for follow-up change.

## Risks / Trade-offs

- [Risk] Adding new dependencies (Lip Gloss, Fang) increases maintenance surface. → Mitigation: keep usage minimal and confined to CLI package.
- [Risk] Cobra/Viper config precedence can be confusing. → Mitigation: document config order and defaults in README/CLI help.
- [Trade-off] Keeping CLI focused on tenants limits immediate usefulness for other resources. → Mitigation: design command structure to allow expansion.
