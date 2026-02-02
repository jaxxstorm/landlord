## Why

Compute resources currently lack consistent ownership and tenant metadata, making it harder to discover, filter, and manage tenant infrastructure across providers. Adding default metadata now establishes a stable convention for all current and future compute backends.

## What Changes

- Define a standard set of default compute metadata applied to all provisioned resources (owner label for Landlord, tenant ID, and other useful identifiers).
- Apply default metadata in the Docker compute provider using labels.
- Establish the requirement for equivalent metadata on future compute providers (Kubernetes labels, ECS tags).
- Ensure metadata is applied broadly across compute resources created for a tenant.

## Capabilities

### New Capabilities
- `compute-default-metadata`: Standard default metadata applied to compute resources across providers (labels/tags) including owner and tenant identifiers.

### Modified Capabilities
- (none)

## Impact

- Compute provider implementations under `internal/compute/providers/`
- Tenant compute provisioning flows where metadata is injected
- Tests for compute providers and provisioning
