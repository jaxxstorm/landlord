## Why

Landlord currently has no first-class way to restrict which callers may invoke which API operations, which makes it difficult to safely expose administrative endpoints. Adding opt-in Tailscale authentication backed by tsnet capabilities allows operators to enforce per-endpoint authorization using their existing tailnet ACLs without changing the default unauthenticated deployment model.

## What Changes

- Add an opt-in authentication and authorization mode that runs Landlord API access through `tsnet` instead of relying on the public listener alone.
- Define configuration for enabling Tailscale-backed auth and mapping API endpoints to required Tailscale capabilities.
- Introduce authorization behavior that requires callers to present the `lbrlabs.com/cap/landlord` capability for configured endpoints before the API handler runs.
- Require authorization failures to be explicit, auditable, and non-destructive, with no partial API side effects before access is granted.
- Define idempotent request handling so retries after transient auth or network failures do not create duplicate mutations.
- Add observability for auth decisions, denied requests, capability mismatches, and fallback behavior when Tailscale auth is disabled.
- Add unit-test-first coverage for config parsing, endpoint matching, auth enforcement, and failure handling.

## Capabilities

### New Capabilities
- `tailscale-api-auth`: Opt-in Tailscale-based authentication and capability enforcement for Landlord API endpoints.

### Modified Capabilities
- None.

## Impact

- Affected code: API server startup, request middleware, configuration loading, and audit/logging paths.
- Affected APIs: HTTP endpoints gain optional authorization requirements when configured; unconfigured deployments keep existing behavior.
- Dependencies and systems: `tsnet`, Tailscale capability checks, and deployment configuration for enabling auth and defining protected endpoints.
- Failure and recovery expectations: if capability evaluation fails or Tailscale identity is unavailable, protected endpoints must fail closed with observable denial responses, while disabled auth mode continues to operate unchanged.
