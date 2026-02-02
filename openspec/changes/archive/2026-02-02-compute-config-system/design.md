## Context

Compute configuration is currently passed as opaque JSON and fails in create/set flows with “Unsupported Config Type.” The system supports only one compute provider at a time (e.g., Docker locally), but there is no authoritative, provider-declared schema for that configuration and no discovery endpoint for clients. This blocks validation, UX, and future providers (ECS/Kubernetes).

## Goals / Non-Goals

**Goals:**
- Introduce a compute config discovery endpoint that returns the active provider, its JSON schema, and optional defaults.
- Validate compute_config in create/set requests against the active provider schema before persistence.
- Keep compute_config flexible enough for future statefulset/task-definition style payloads.
- Add CLI support (`compute` command) and ensure create/set send compute_config correctly.

**Non-Goals:**
- Supporting multiple compute providers simultaneously in a single landlord instance.
- Enforcing a universal schema across providers (each provider owns its schema).
- Introducing breaking database migrations for config storage.

## Decisions

1) **Provider-owned JSON Schema**
- **Decision:** Each compute provider exposes a JSON Schema document for its config payload (draft 2020-12), plus an optional default config object.
- **Why:** Avoids hardcoding provider config in the API/CLI; allows ECS/Kubernetes to define richer schemas later.
- **Alternatives considered:** Hardcoded schema in API (rejected: duplicates provider knowledge and breaks extensibility).

2) **Single discovery endpoint**
- **Decision:** Add a compute config discovery endpoint (e.g., `GET /api/compute/config`) returning `{provider, schema, defaults?}`.
- **Why:** Keeps client behavior deterministic and aligned with the active provider configuration.
- **Alternatives considered:** Client-side heuristics or CLI-only config (rejected: API should be source of truth).

3) **Schema validation at API ingress**
- **Decision:** Validate compute_config in create/set using the active provider schema before DB writes.
- **Why:** Fail fast with actionable errors; guarantees stored desired_config aligns with provider requirements.
- **Alternatives considered:** Deferred validation in workflow/compute (rejected: delays errors and complicates reconciliation).

4) **Flexible JSON payloads**
- **Decision:** Treat compute_config as generic JSON object (`map[string]any` / raw JSON) with schema validation; store as JSONB unchanged.
- **Why:** Supports complex nested configs (ECS task definitions, Kubernetes manifests) without future schema migrations.
- **Alternatives considered:** Strongly typed structs for each provider (rejected: reduces flexibility and increases coupling).

5) **CLI mirrors API semantics**
- **Decision:** CLI `compute` command retrieves schema from API; create/set pass `--config` as JSON object directly.
- **Why:** Ensures CLI uses the same validation rules as API.
- **Alternatives considered:** CLI-side schema validation only (rejected: risks drift from server schema).

## Risks / Trade-offs

- **Schema drift** → Mitigation: schema exposed by the active provider at runtime; CLI always fetches from API.
- **Large/complex configs** → Mitigation: accept raw JSON and store as JSONB; avoid deep server-side coercion.
- **Ambiguity around defaults** → Mitigation: defaults are optional; API indicates absence explicitly.
- **Validation strictness may reject unknown fields** → Mitigation: schema should allow provider-defined extensions; document strict vs permissive behavior.

## Migration Plan

- No DB migration required (config stored as JSONB already).
- Deploy API changes first (new endpoint + validation). Then update CLI to use the endpoint and pass config payloads.
- Rollback: remove endpoint and revert validation to current permissive behavior; stored configs remain valid JSON.

## Open Questions

- Should the schema be strict (disallow unknown fields) by default for Docker, or permissive?
- Do we want to expose example configs or templates alongside defaults?
- Should the compute config endpoint be versioned for future schema evolution?
