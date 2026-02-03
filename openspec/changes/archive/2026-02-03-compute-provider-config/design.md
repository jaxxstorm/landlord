## Context

Compute configuration currently relies on a global `default_provider` plus `defaults` and provider blocks, which forces users to configure providers they may not use. The change introduces provider-keyed configuration where providers are enabled by presence under `compute`, and discovery endpoints/CLI must return schema/defaults for a requested provider.

## Goals / Non-Goals

**Goals:**
- Replace `compute.default_provider`/`compute.defaults` with provider-keyed configuration blocks.
- Treat presence of a provider block as enablement; allow single-provider configs without boilerplate.
- Update compute config discovery to accept a provider identifier and return schema/defaults for that provider.
- Keep existing provider-specific schema validation behavior intact (just keyed differently).

**Non-Goals:**
- Changing compute provider runtime behavior beyond configuration loading and discovery.
- Introducing provider auto-selection heuristics beyond explicit configuration.
- Redesigning tenant API compute_config payloads beyond required schema/validation updates.

## Decisions

- **Provider-keyed compute config**: Move provider config under `compute.<provider>` and remove `default_provider`/`defaults` to avoid enabling unused providers. Alternatives considered: keep `default_provider` but allow override via provider presence; rejected due to ambiguity and continued boilerplate.
- **Provider-aware discovery API/CLI**: Accept a provider identifier and return schema/defaults for that provider. Alternative: return all providers in a single response; rejected due to larger payloads and unclear default selection.
- **Validation rules**: Provider is enabled if its config block exists. Missing block means provider is disabled and discovery requests should return a clear error. Alternative: enable providers via explicit list; rejected as it duplicates the configuration data.

## Risks / Trade-offs

- [Risk] Breaking config files that rely on `default_provider` and `defaults` → Mitigation: document breaking change and update examples/tests.
- [Risk] Clients calling discovery without provider parameter may fail → Mitigation: return a clear validation error and document new required parameter.
- [Risk] Providers with partial config blocks could enable unintentionally → Mitigation: schema validation should fail fast on missing required fields.
