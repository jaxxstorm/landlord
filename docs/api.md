# API Browser

This page embeds a standalone Swagger UI hosted at `api.html`. If the embedded view does not load, open `api.html` directly. The API is versioned under `/v1`.

## Authentication and Authorization

By default, Landlord does not require authentication for API access.

When `http.tailscale_auth.enabled` is enabled in configuration, Landlord can serve the API over `tsnet` and enforce Tailscale capability checks on selected endpoints.

Behavior in Tailscale auth mode:

- Authorization is opt-in and only applies to endpoints listed in `http.tailscale_auth.protected_endpoints`.
- Protected endpoints are matched by HTTP method and path pattern.
- Requests to protected endpoints are allowed only when the caller's Tailscale identity has the configured capability.
- Capability values must use `lbrlabs.com/cap/landlord` or a nested capability such as `lbrlabs.com/cap/landlord/read`.
- If identity lookup fails, the capability is missing, or capability evaluation cannot complete, Landlord denies the request before any tenant, workflow, or compute side effects occur.

Default unprotected endpoints:

- `/health`
- `/ready`

Those endpoints stay open unless you explicitly add them to the protected endpoint policy.

Common response behavior for protected endpoints:

- `403 Forbidden` when the caller is identified but does not satisfy the required capability, or when no Tailscale identity is available for the request.
- `503 Service Unavailable` when Landlord cannot complete authorization safely, such as an internal Tailscale authorization failure.

Operational notes:

- Authorization decisions are logged with method, path, remote address, decision outcome, and denial reason.
- Successful requests can carry Tailscale caller identity through request context for downstream logging.
- Swagger describes the API surface, but the actual authorization policy comes from runtime configuration.

<style>
  #swagger-frame {
    width: 100%;
    height: 85vh;
    border: 0;
    border-radius: 12px;
    background: #ffffff;
  }
</style>

<iframe id="swagger-frame" title="Landlord API Browser" src="api.html"></iframe>
