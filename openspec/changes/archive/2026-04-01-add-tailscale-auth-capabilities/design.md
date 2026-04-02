## Context

Landlord's API server currently exposes all registered routes through a single chi router with shared logging, correlation, recovery, and timeout middleware. There is no first-class authentication or authorization layer on the HTTP API, so every deployment must rely on external network placement to decide who can invoke mutating or administrative endpoints.

This change introduces an opt-in Tailscale-backed access model using `tsnet`. The design needs to preserve Landlord's default behavior when disabled, fail closed for configured protected endpoints when enabled, and avoid changing tenant lifecycle, persistence, or workflow behavior except to prevent unauthorized requests from reaching handlers.

## Goals / Non-Goals

**Goals:**
- Add a config-driven way to enable Tailscale-backed API authentication without changing the default listener behavior when disabled.
- Allow operators to declare which HTTP endpoints require Tailscale capability checks.
- Enforce the `lbrlabs.com/cap/landlord` capability before protected handlers execute.
- Ensure denied or indeterminate authorization results do not trigger tenant state transitions, persistence writes, or workflow execution.
- Produce audit and log signals that explain allow or deny decisions and why they occurred.
- Keep the implementation testable with isolated unit tests around config validation, route matching, and middleware behavior.

**Non-Goals:**
- General-purpose RBAC, user management, or non-Tailscale auth providers.
- Per-tenant or per-resource authorization rules.
- Changing Landlord's public API shapes beyond auth-related responses.
- Persisting auth policy in the database.

## Decisions

### Decision: model this as HTTP auth middleware with config-owned policy
Landlord already centralizes request handling in `internal/api/server.go`, so the smallest change is to add an authorization middleware into the existing chi stack or route groups. A config-owned policy keeps endpoint protection declarative and consistent with the rest of the application's env and config-file model.

Alternatives considered:
- Handler-local checks: rejected because auth behavior would be duplicated and easy to miss on new endpoints.
- Reverse-proxy-only protection: rejected because the proposal requires Landlord itself to integrate with `tsnet` and Tailscale capabilities.

### Decision: keep Tailscale auth strictly opt-in
The server SHALL continue to operate exactly as it does today unless a dedicated config flag enables Tailscale auth. This avoids a breaking rollout, preserves local development defaults, and lets operators phase in protection per environment.

Alternatives considered:
- Enable Tailscale auth automatically when `tsnet` config exists: rejected because partial config could silently change exposure.

### Decision: protect endpoints with explicit method-and-path match rules
Protected endpoints SHALL be declared as route match entries, such as method plus path or chi-style route pattern, with each entry mapped to a required capability value. This keeps policy reviewable and avoids inferring authorization from handler names.

Alternatives considered:
- Protect entire route prefixes only: rejected because it is too coarse for mixed public/private routes such as health checks and docs.
- Hardcode protected mutating endpoints: rejected because the proposal requires config-driven opt-in behavior.

### Decision: use a single capability namespace rooted at `lbrlabs.com/cap/landlord`
The first implementation SHALL evaluate the caller against Tailscale ACL-granted capabilities under `lbrlabs.com/cap/landlord`. Config can specify the required value for an endpoint, but all values stay under that namespace so operators have one predictable ACL surface.

Alternatives considered:
- Free-form capability domains: rejected because it complicates docs and interoperability.
- One implicit capability for every route: rejected because operators may want a smaller set of grouped permissions.

### Decision: fail closed for protected endpoints on identity or capability lookup errors
If Tailscale identity cannot be established, capability evaluation fails, or middleware configuration is invalid for a request, the request SHALL be denied before any handler side effects occur. This is safer than attempting best-effort bypass for protected routes.

Alternatives considered:
- Fail open on transient Tailscale errors: rejected because it turns auth outages into privilege escalation.

### Decision: no persistence or workflow schema changes in the first increment
Authorization decisions happen at request ingress, so protected request denials SHALL not create tenant records, workflow executions, or audit rows in persistence unless the existing audit subsystem already records denied requests out of band. Structured logs are the minimum required observability surface for the initial increment.

Alternatives considered:
- Persist denied auth attempts in the database: deferred because it introduces a new audit schema decision not required to start implementation.

## Risks / Trade-offs

- [Tailscale capability APIs may expose less request context than expected] -> Mitigation: isolate capability extraction behind a narrow interface and cover it with adapter tests.
- [Route matching drift could leave new endpoints unprotected] -> Mitigation: keep policy explicit, validate config entries against known methods, and add tests covering protected and unprotected routes.
- [Fail-closed behavior can make outages visible to operators] -> Mitigation: emit structured denial reasons and document safe rollout with health/readiness endpoints left unprotected by default.
- [Running through `tsnet` may introduce listener or deployment complexity] -> Mitigation: keep the feature behind a separate config block and preserve existing HTTP startup when disabled.

## Migration Plan

1. Add config types and validation for Tailscale auth enablement and endpoint capability rules.
2. Introduce a Tailscale auth adapter and middleware that can be injected into the API server.
3. Keep `/health` and `/ready` unprotected unless an operator explicitly opts in.
4. Roll out with auth disabled by default.
5. Enable auth in a target environment, define protected endpoints, and apply matching tailnet ACLs for `lbrlabs.com/cap/landlord`.
6. Roll back by disabling the auth config block, which returns the server to existing listener and authorization behavior.

## Open Questions

- Whether `tsnet` should replace the primary listener entirely when enabled or run alongside the existing listener for phased adoption.
- Whether endpoint policy should support chi patterns only or exact path matching plus method.
- Whether denied authorization attempts need durable persistence beyond structured logs in the first release.
