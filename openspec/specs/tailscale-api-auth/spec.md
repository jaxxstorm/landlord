## ADDED Requirements

### Requirement: Tailscale API authorization is opt-in
The system MUST provide configuration that enables Tailscale-backed API authorization explicitly. When that configuration is disabled or omitted, the API SHALL preserve its current listener, routing, workflow, persistence, and compute behavior with no endpoint capability checks.

#### Scenario: Authorization disabled
- **WHEN** Landlord starts without Tailscale authorization enabled
- **THEN** requests SHALL be handled exactly as they are today without Tailscale identity or capability evaluation

#### Scenario: Authorization enabled
- **WHEN** Landlord starts with Tailscale authorization enabled
- **THEN** the API server SHALL initialize the Tailscale auth path and evaluate configured protected endpoints before invoking their handlers

### Requirement: Protected endpoints require configured Tailscale capabilities
The system MUST allow operators to configure protected HTTP endpoints by method and path pattern, and each protected endpoint entry MUST declare the required capability value under `lbrlabs.com/cap/landlord`. For a matching request, the system SHALL authorize the caller only when Tailscale identity is present and the caller has the configured capability.

Impact on API behavior: matching protected endpoints SHALL return an authorization failure instead of invoking the handler when the capability is missing.
Impact on state model and transitions: authorization checks SHALL occur before any tenant state transition logic runs.
Impact on persistence schema: no persistence schema changes are required for endpoint policy configuration.
Impact on workflow semantics: unauthorized requests SHALL NOT enqueue, trigger, resume, or stop workflows.
Impact on compute semantics: unauthorized requests SHALL NOT provision, update, inspect, or delete compute resources.

#### Scenario: Caller has required capability
- **WHEN** a request matches a protected endpoint rule and the caller has the configured `lbrlabs.com/cap/landlord` capability
- **THEN** the system SHALL pass the request to the target handler

#### Scenario: Caller lacks required capability
- **WHEN** a request matches a protected endpoint rule and the caller lacks the configured `lbrlabs.com/cap/landlord` capability
- **THEN** the system SHALL reject the request with an authorization error before any handler side effects occur

### Requirement: Authorization failures fail closed and preserve idempotency
For any protected endpoint, the system MUST fail closed when Tailscale identity cannot be determined, when capability evaluation errors, or when auth middleware cannot complete safely. These failures SHALL be returned before request mutation logic begins, so retried requests do not create duplicate tenant records, duplicate workflow executions, or partial state transitions.

Impact on API behavior: protected endpoints SHALL return a deterministic non-success response for denied or indeterminate auth results.
Impact on state model and transitions: denied or indeterminate requests SHALL leave tenant desired and observed state unchanged.
Impact on persistence schema: no new persistence writes are required to preserve idempotency for denied requests.
Impact on workflow semantics: retries after auth denial or transient auth infrastructure failure SHALL remain safe because no workflow side effects are started before authorization completes.
Impact on compute semantics: no compute action SHALL begin until authorization succeeds.

#### Scenario: Identity unavailable
- **WHEN** a protected request arrives and Tailscale identity data is unavailable
- **THEN** the system SHALL deny the request and SHALL NOT execute handler logic

#### Scenario: Capability evaluation error
- **WHEN** a protected request arrives and capability evaluation returns an internal error
- **THEN** the system SHALL deny the request, emit an observable error signal, and leave application state unchanged

### Requirement: Authorization decisions are observable and auditable
The system MUST emit structured observability data for protected endpoint authorization decisions, including the request path, HTTP method, decision outcome, and denial reason. When request correlation metadata is available, the authorization logs SHALL include it so operators can trace denied and allowed requests through the rest of the API processing path.

Impact on API behavior: auth responses SHALL be explainable through logs without exposing sensitive internals to callers.
Impact on state model and transitions: authorization metadata MAY be attached to in-process request context but SHALL NOT alter tenant lifecycle semantics.
Impact on persistence schema: durable audit persistence is not required for the initial version.
Impact on workflow semantics: allowed requests SHOULD preserve request identity metadata for downstream audit logging where existing workflow interfaces already support it.
Impact on compute semantics: none.

#### Scenario: Request denied
- **WHEN** the system denies a protected request
- **THEN** it SHALL log a structured authorization denial event with the path, method, and denial reason

#### Scenario: Request allowed
- **WHEN** the system allows a protected request
- **THEN** it SHALL emit a structured authorization success event or metric that can be correlated with the request

### Requirement: Configuration validation rejects unusable authorization policy
The system MUST validate Tailscale authorization configuration at startup. Invalid configuration, including malformed endpoint rules, unsupported HTTP methods, missing required capability values, or inconsistent enablement settings, SHALL prevent the server from starting in authorization-enabled mode.

#### Scenario: Invalid protected endpoint rule
- **WHEN** startup configuration enables Tailscale authorization and includes an invalid endpoint rule
- **THEN** configuration validation SHALL fail before the API server begins serving requests

#### Scenario: Missing capability value
- **WHEN** startup configuration enables Tailscale authorization for an endpoint but omits the required capability value
- **THEN** configuration validation SHALL fail with an error that identifies the invalid policy entry
