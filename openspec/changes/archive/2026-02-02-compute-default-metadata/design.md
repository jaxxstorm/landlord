## Context

Compute providers currently provision resources without a consistent, tenant-scoped metadata standard. This makes it harder to discover resources, attribute ownership, and support operations across providers. Today only the Docker provider is implemented, but future providers (Kubernetes, ECS) will require labels or tags for the same purpose.

## Goals / Non-Goals

**Goals:**
- Define a standard set of default metadata keys applied to all compute resources.
- Ensure the Docker provider applies these labels consistently across created resources.
- Establish conventions that map cleanly to future providers (Kubernetes labels, ECS tags).
- Make metadata useful for discovery and ownership (owner and tenant identifiers).

**Non-Goals:**
- Changing the tenant API or spec format.
- Introducing new metadata configuration surfaces beyond the default set.
- Implementing non-Docker providers as part of this change.

## Decisions

1. **Standard metadata keys with provider-specific mapping**
   - **Decision:** Introduce a common metadata map in compute provisioning that is translated into provider-specific mechanisms (labels/tags).
   - **Alternatives:** Provider-specific ad hoc metadata or only apply to Docker.
   - **Rationale:** Ensures consistency and future compatibility across providers.

2. **Required keys: owner + tenant identifiers**
   - **Decision:** Include a stable owner label (e.g., `landlord.owner=landlord`) and tenant identifiers (tenant ID, and optionally tenant name if available).
   - **Alternatives:** Only tenant ID or only owner label.
   - **Rationale:** Ownership and tenant scope are the primary discovery keys.

3. **Apply metadata broadly to compute resources**
   - **Decision:** Default metadata is applied to all compute resources created by a provider (containers, tasks, services, etc.).
   - **Alternatives:** Only apply to top-level resources.
   - **Rationale:** Resource discovery should work for any resource depth.

## Risks / Trade-offs

- [Label/tag collisions with user-provided metadata] → Mitigation: Namespace keys under a `landlord.*` prefix and document precedence if a provider supports custom labels.
- [Provider limitations on label/tag lengths or characters] → Mitigation: Keep keys short, values simple, and validate where necessary.
- [Future provider mapping differences] → Mitigation: Define a canonical key set and map to provider constraints in each implementation.

## Migration Plan

1. Define canonical metadata keys and values.
2. Update Docker provider to attach default labels to created resources.
3. Add tests verifying labels are present for Docker compute resources.
4. Document the metadata conventions for future providers.

## Open Questions

- Should tenant name be included alongside tenant ID, and if so, is it always available at provisioning time?
- Do we need a configurable override for the owner label value (default to `landlord`)?
