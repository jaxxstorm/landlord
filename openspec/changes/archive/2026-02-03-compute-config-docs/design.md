## Context

Compute provider documentation currently lists only partial compute_config fields and lacks a single, readable JSON/YAML reference per provider. CLI `--config` input only supports inline JSON or a plain file path, which makes it harder to share canonical configs. We want docs to be authoritative and the CLI to accept YAML and `file://` URIs.

## Goals / Non-Goals

**Goals:**
- Add complete compute_config reference docs for each provider with all supported fields and readable multi-line JSON/YAML examples.
- Update CLI `--config` parsing to accept YAML and local `file://` URIs.
- Ensure docs provide enough information for a new user to construct a full compute_config payload.

**Non-Goals:**
- Changing provider runtime behavior or adding new compute_config fields.
- Supporting remote file retrieval (`http://`, `https://`, etc.).
- Adding schema validation beyond current provider validation.

## Decisions

- **Single-source documentation per provider**: Consolidate full compute_config option lists in provider docs under `docs/compute/<provider>/`. This keeps options discoverable alongside provider behavior.
  - *Alternative*: Centralize in a single compute-config doc. Rejected because provider docs already exist and are the most relevant entry point for users.

- **Readable JSON/YAML examples**: Provide multi-line examples for each provider, showing optional fields explicitly with descriptions where needed.
  - *Alternative*: Rely only on schema discovery endpoint. Rejected because docs should be static and accessible without running the API.

- **CLI parsing supports YAML and file://**: Extend `parseConfigInput` to detect `file://` URIs, read local files, and parse YAML as well as JSON.
  - *Alternative*: Require users to convert YAML to JSON externally. Rejected due to unnecessary friction.

## Risks / Trade-offs

- [Risk] Docs drift from actual provider schema as fields evolve → Mitigation: tie docs to provider schema files and update during provider changes; add checklist task in provider changes.
- [Risk] YAML parsing introduces ambiguity (e.g., numeric vs string) → Mitigation: rely on Go YAML parser and document expected types.
- [Risk] Users assume file:// supports remote URLs → Mitigation: document that only local file paths are supported.

## Migration Plan

- Update docs first (no runtime impact).
- Update CLI parsing and add tests for JSON, YAML, and file:// sources.
- No migration required for existing users; inline JSON still supported.

## Open Questions

- Should YAML input require a `.yaml/.yml` extension, or should we auto-detect by content for inline values?
- Do we need to support `file://` for tenant config updates in non-CLI entry points (API clients)?
