## Context

Workers currently connect directly to the database to read tenant state and write lifecycle updates. This couples workers to the persistence layer and makes workflow providers difficult to swap independently. We need a reconciliation-driven architecture where Landlord owns tenant state (Postgres), workflow providers orchestrate execution (Restate/Temporal/Step Functions), and workers are stateless HTTP handlers invoked by providers.

## Goals / Non-Goals

**Goals:**
- Make workers stateless by removing direct database access.
- Add reconciler in Landlord that polls DB state and invokes workflow providers.
- Define pluggable `workflow.Provider` interface (Invoke, GetStatus) supporting Restate, Temporal, Step Functions.
- Workers receive full execution context from providers (no API calls needed).
- Preserve existing lifecycle semantics while decoupling execution from persistence.

**Non-Goals:**
- Redesign tenant domain model or lifecycle states.
- Replace existing compute providers; this is a workflow orchestration refactor only.
- Introduce new auth/RBAC beyond what workflow providers need.

## Decisions

- **Extend existing reconciler**: The existing `internal/controller/reconciler.go` will be enhanced to invoke workflow providers and poll status, rather than creating a separate reconciler.
  - *Alternative:* Create new tenant-specific reconciler. Rejected to avoid duplicating reconciliation logic and maintain single control loop.

- **Add methods to existing Provider interface**: Extend `workflow.Provider` with `Invoke(workflowName, request)` and `GetWorkflowStatus(executionID)` methods alongside existing methods.
  - *Alternative:* Replace interface entirely. Rejected to maintain backward compatibility during migration; old methods can delegate to new ones.

- **Reuse tenant.Status for workflow state**: Use existing tenant lifecycle states (StatusRequested → StatusProvisioning → StatusReady/StatusFailed) instead of adding new workflow_status field.
  - *Alternative:* Add separate workflow_status field. Rejected because tenant.Status already represents workflow state; adding another field creates ambiguity.

- **Store compute config in tenant.DesiredConfig**: Use existing JSONB field for compute provider configuration instead of new compute_config field.
  - *Alternative:* Add new compute_config/compute_result fields. Rejected to avoid dual sources of truth; DesiredConfig/ObservedConfig already serve this purpose.

- **Workers are stateless handlers**: Workers receive everything needed in request payload (tenantID, compute provider, desiredConfig); no DB or Landlord API calls for tenant data.
  - *Alternative:* Workers query Landlord API for tenant data. Rejected to minimize coupling; provider passes what's needed. Workers may still call API for status updates/callbacks.

- **Reconciler uses workflow.Manager**: Reconciler calls workflow.Manager methods which abstract provider differences, rather than calling providers directly.
  - *Alternative:* Reconciler calls providers directly. Rejected because Manager provides retry logic, provider selection, and logging that shouldn't be duplicated.

- **API sets StatusRequested only**: API handlers set tenant status to `StatusRequested` and the reconciler performs workflow invocation.
  - *Alternative:* API triggers workflows immediately. Rejected to avoid duplicate triggers and keep reconciliation as the single source of orchestration.

## Risks / Trade-offs

- [Reconciler polling adds latency] → Mitigation: configurable interval (10-30s); optimize for eventual consistency.
- [Provider API differences] → Mitigation: adapter pattern in provider implementations abstracts differences.
- [Migration complexity] → Mitigation: keep DB schema backward compatible; run reconciler alongside existing flows initially.
- [Status polling overhead] → Mitigation: batch status queries where provider APIs support it.

## Migration Plan

- Extend `workflow.Provider` interface with Invoke() and GetWorkflowStatus(executionID) methods (keep existing methods for compatibility).
- Update Restate and Step Functions providers to implement new methods (map to existing internal calls).
- Enhance existing `internal/controller/reconciler.go` with workflow invocation and status polling logic.
- Remove DB access from Restate worker in `cmd/workers/restate/main.go`.
- Update worker handler to receive tenant context from workflow payload instead of querying DB.
- Update all workflow.Manager and provider call sites to use new methods.
- Verify tenant.DesiredConfig/ObservedConfig fields are used for compute configuration.
- Test end-to-end with reconciliation flow enabled.

## Compatibility Notes

- Existing `tenant.Status` states map to workflow lifecycle: StatusRequested (invoke workflow), StatusProvisioning (poll status), StatusReady/StatusFailed (terminal states).
- Existing `tenant.DesiredConfig` stores compute provider configuration; no new DB fields needed.
- Existing `tenant.ObservedConfig` stores compute results (IPs, container IDs, etc.) from workflows.
- Existing `reconciler.go` is enhanced, not replaced; controller reconciliation continues to work.
- Provider interface is extended, not replaced; existing CreateWorkflow/StartExecution methods remain for backward compatibility.

## Status Mapping

- `tenant.Status` is the source of truth for lifecycle state; workflow providers only report execution status.
- `StatusRequested`/`StatusPlanning` → reconciler invokes workflow and sets `StatusProvisioning`.
- `StatusProvisioning`/`StatusUpdating`/`StatusDeleting` → reconciler polls provider status.
- Provider `Succeeded` → reconciler advances `tenant.Status` to next terminal or steady state and stores outputs in `tenant.ObservedConfig`.
- Provider `Failed` → reconciler sets `tenant.Status = StatusFailed` with error details.

## Open Questions

- Should reconciler support event-driven triggers (webhook from provider) in addition to polling?
- What's the right reconciliation interval (10s? 30s?)?
- How should reconciler handle failed workflows (retry? manual intervention?)?
- Do we need separate intervals for invoke loop vs status polling loop?
