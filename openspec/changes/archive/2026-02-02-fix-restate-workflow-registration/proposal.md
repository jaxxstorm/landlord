## Why

Currently, when provisioning a compute tenant using the Restate workflow provider, the workflow engine returns a "workflow not found" error. This is because workflows are not actually registered with the Restate backend, preventing the system from executing tenant lifecycle operations. Fixing this is critical for enabling real, end-to-end tenant provisioning and lifecycle management.

## What Changes

- Register tenant provisioning workflows with the Restate backend at startup
- Ensure workflow registration is idempotent and robust
- Add integration tests for the Restate provider with the Docker compute provider
- Validate that a tenant can be provisioned end-to-end using the API and Restate
- Improve error handling and logging for workflow registration and execution

## Capabilities

### New Capabilities
- `restate-workflow-registration`: Register and manage tenant lifecycle workflows with the Restate backend, enabling actual workflow execution for tenant provisioning.
- `restate-integration-test`: End-to-end integration test for Restate provider with Docker compute, validating tenant provisioning.

### Modified Capabilities


## Impact

- Affects the Restate workflow provider implementation
- Changes to workflow registration logic and startup
- Adds or modifies integration tests (especially for Docker compute provider)
- May impact API error handling and user experience for tenant provisioning
- Dependencies: Restate backend, Docker compute provider, integration test framework
