## 1. Configuration and auth adapter setup

- [x] 1.1 Add a Tailscale auth config block under application config with enablement, listener settings, and protected endpoint capability rules.
- [x] 1.2 Add startup validation for protected endpoint rules, supported HTTP methods, and required `lbrlabs.com/cap/landlord` capability values.
- [x] 1.3 Introduce a Tailscale auth adapter interface around `tsnet` identity and capability evaluation so middleware can be unit tested without a live tailnet.

## 2. API authorization enforcement

- [x] 2.1 Add API middleware that matches requests against configured protected endpoints and enforces Tailscale capability checks before handlers run.
- [x] 2.2 Wire the middleware into `internal/api/server.go` so auth remains disabled by default and preserves existing routing when not configured.
- [x] 2.3 Ensure denied or indeterminate auth results fail closed and return a consistent non-success HTTP response without invoking tenant, workflow, or compute operations.

## 3. Observability and request metadata

- [x] 3.1 Emit structured logs for authorization allow and deny decisions including method, path, request correlation metadata, and denial reason.
- [x] 3.2 Propagate any useful caller identity metadata through request context only where existing logging or workflow hooks can safely consume it.
- [x] 3.3 Document or encode the default expectation that health and readiness endpoints remain unprotected unless explicitly configured.

## 4. Verification

- [x] 4.1 Add unit tests for config parsing and validation covering disabled mode, valid rules, invalid methods, and missing capability values.
- [x] 4.2 Add middleware tests covering allowed requests, missing capability denials, identity lookup failures, and capability evaluation errors with no handler side effects.
- [x] 4.3 Add server wiring tests to verify protected endpoint matching and unchanged behavior when Tailscale auth is disabled.
