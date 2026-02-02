## Context

Landlord's HTTP API surface is currently unversioned, and both the CLI and docs reference unversioned paths. We need explicit API versions (starting with v1) to allow future v2+ changes without breaking existing clients while keeping routing, validation, and error handling consistent across endpoints.

## Goals / Non-Goals

**Goals:**
- Establish a single, explicit versioning scheme for all HTTP APIs (path prefix like `/v1`).
- Keep v1 behavior functionally equivalent to current endpoints, only changing the URL structure.
- Ensure unsupported versions are handled consistently with clear error responses.
- Update CLI defaults, docs, and examples to use the versioned base path.
- Preserve a clean path to introduce v2 in the future without large refactors.

**Non-Goals:**
- Designing or implementing any v2 behavior in this change.
- Adding new authentication, authorization, or other API features unrelated to versioning.
- Introducing a new API client library or SDK.

## Decisions

- **Path-prefix versioning (`/v1/...`) is the canonical scheme.**
  - Rationale: easiest to route in chi, explicit in URLs and docs, and aligns with most existing tooling (curl, proxies, gateways).
  - Alternatives considered: header-based versioning (e.g., `Accept-Version`) and query param versioning. Rejected to keep URLs explicit and avoid hidden behavior.

- **Router structure: mount versioned subrouters and share handlers.**
  - Implement `/v1` as a chi subrouter, reusing existing handlers and services to minimize behavioral drift.
  - Future `/v2` can swap in adapters/DTOs at the routing layer while sharing domain logic.

- **Unsupported/unknown versions return a consistent error response.**
  - Use a single error code (e.g., `unsupported_version`) and message that points to supported versions.
  - Alternatives considered: 404 with no body or redirect. Rejected to ensure clients receive actionable errors and to keep semantics consistent across endpoints.

- **Unversioned paths are rejected by default.**
  - Policy: return an explicit error that the version is required and list supported versions.
  - Alternatives considered: temporary alias/redirect to `/v1`. Rejected to avoid hidden compatibility behavior and to keep versioning explicit from the first release.

## Risks / Trade-offs

- **Breaking existing clients** → Mitigation: update CLI defaults, docs, examples, and release notes; provide a clear error response for unversioned requests.
- **Route duplication across versions** → Mitigation: keep handlers shared and version-specific differences isolated at the routing/DTO layer.
- **Future version drift and increased test surface** → Mitigation: establish versioned routing tests and shared contract tests for common behavior.

## Migration Plan

1. Add `/v1` routes for all HTTP API surfaces in the server router.
2. Update CLI base URLs and docs/examples to use `/v1`.
3. Add error handling for unsupported/unversioned requests.
4. Release with explicit versioning and publish upgrade guidance.

## Open Questions

- Should the CLI expose an explicit `--api-version` flag or rely solely on base URL configuration?
- Is a 400 or 404 response preferable for unsupported versions?
- Should we add a discoverability endpoint for supported versions (e.g., `/version` or `/` metadata)?
