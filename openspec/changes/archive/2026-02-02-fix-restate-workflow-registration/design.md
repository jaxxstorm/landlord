## Context

The landlord system uses workflow providers to manage tenant lifecycle operations, including provisioning. The Restate workflow provider is intended to handle these operations asynchronously using the Restate backend. However, currently, workflows are not registered with the Restate backend, resulting in "workflow not found" errors when attempting to execute tenant provisioning workflows. This prevents end-to-end tenant provisioning from working with the Restate provider.

The current state involves the Restate provider being initialized but not performing workflow registration. The system attempts to invoke workflows that don't exist in the Restate backend, leading to failures. Stakeholders include developers using the landlord system for multi-tenant applications and operators deploying the system with Restate.

Constraints include maintaining compatibility with existing workflow provider interfaces and ensuring the registration process is robust against Restate backend unavailability or restarts.

## Goals / Non-Goals

**Goals:**
- Register tenant provisioning workflows with the Restate backend during system startup
- Ensure workflow registration is idempotent and handles errors gracefully
- Enable end-to-end tenant provisioning using the Restate workflow provider
- Add integration tests validating the Restate provider with Docker compute
- Improve error handling and logging for workflow operations

**Non-Goals:**
- Modifying existing workflow definitions or logic
- Changing the workflow provider interface contracts
- Implementing workflow versioning or dynamic registration beyond startup

## Decisions

### Workflow Registration Timing
**Decision:** Register workflows during Restate provider initialization at application startup.

**Rationale:** Startup registration ensures workflows are available before any tenant operations are attempted. This is simpler than lazy registration, which could introduce race conditions or delays during first use.

**Alternatives Considered:**
- Lazy registration on first workflow invocation: Rejected due to potential delays and complexity in handling concurrent requests.
- Manual registration via CLI: Rejected as it adds operational overhead and risk of forgetting to register.

### Registration Location
**Decision:** Implement registration logic within the Restate workflow provider's initialization method.

**Rationale:** Keeps registration logic co-located with the provider, following the single responsibility principle. The provider already handles Restate client management.

**Alternatives Considered:**
- Separate registration service: Overkill for this use case, adds unnecessary complexity.
- Controller-level registration: Would require the controller to know provider-specific details, breaking abstraction.

### Idempotency Handling
**Decision:** Use Restate's built-in idempotency for registration operations, with error handling for already-registered workflows.

**Rationale:** Restate allows re-registering the same workflow without issues, and attempting to register an existing workflow typically succeeds or fails gracefully. This simplifies the implementation.

**Alternatives Considered:**
- Check existence before registering: Would require additional API calls and error handling, potentially more complex.
- Custom idempotency tracking: Unnecessary given Restate's behavior.

### Error Handling Strategy
**Decision:** Log registration failures but allow the system to start, with clear error messages indicating workflow unavailability.

**Rationale:** Failing fast on registration errors would prevent the system from starting if Restate is temporarily unavailable. Instead, log warnings and allow operations to fail gracefully with informative errors.

**Alternatives Considered:**
- Fail startup on registration errors: Too brittle for production deployments where services may have startup ordering issues.
- Retry registration indefinitely: Could delay startup unnecessarily.

## Risks / Trade-offs

- **Risk:** Registration failures at startup may not be immediately visible, leading to runtime failures.
  **Mitigation:** Comprehensive logging and health checks to verify workflow availability.

- **Risk:** Startup performance impact from registration operations.
  **Mitigation:** Registration is typically fast; if it becomes a bottleneck, consider async registration.

- **Risk:** Restate API changes could break registration logic.
  **Mitigation:** Use official Restate client libraries and keep registration logic simple and testable.

- **Trade-off:** Allowing startup with failed registration vs. strict validation.
  **Benefit:** Improved reliability in distributed environments.
  **Drawback:** Potential for confusing runtime errors if registration silently fails.

## Migration Plan

No migration is required as this is a startup-time change. Existing deployments will need to restart the landlord service to enable workflow registration. The change is backward compatible - systems without Restate will continue to work unchanged.

**Deployment Steps:**
1. Deploy updated landlord binary
2. Restart landlord service
3. Monitor logs for successful workflow registration
4. Test tenant provisioning end-to-end

**Rollback:** Revert to previous binary version and restart service.

## Open Questions

- What is the exact Restate API for workflow registration? Need to confirm the client library methods.
- How should we handle partial registration failures (some workflows register, others don't)?
- What logging level should be used for registration success/failure?
- Should we add metrics for registration status?