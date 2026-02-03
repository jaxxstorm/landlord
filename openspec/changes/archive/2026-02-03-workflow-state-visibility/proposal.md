## Why

Currently, tenants in the "provisioning" state provide no visibility into underlying workflow execution details. When a workflow is backing-off, retrying, or in an error state, this critical status information is not surfaced to API consumers. Users cannot distinguish between normal provisioning progress and failure conditions, leading to poor observability and delayed incident response.

## What Changes

- Introduce granular sub-states within the provisioning lifecycle phase that reflect workflow execution status
- Add workflow provider state mapping to translate provider-specific states to provider-agnostic Landlord states
- Extend GET /v1/tenants/{id} response to include workflow execution status details
- Enhance LIST /v1/tenants response to surface workflow state information for in-progress tenants
- Update reconciler to poll and persist workflow execution state from workflow providers
- Define canonical workflow states (e.g., running, backing-off, error, waiting) that are provider-agnostic

## Capabilities

### New Capabilities
- `workflow-state-mapping`: Provider-agnostic state nomenclature and mapping from workflow provider states to Landlord canonical states
- `workflow-execution-status`: Enriched workflow execution status retrieval and persistence including sub-state, retry count, and error details

### Modified Capabilities
- `tenant-get-api`: Extend response to include workflow execution status fields when tenant is in workflow-driven states
- `tenant-list-api`: Include workflow status summary in list responses for tenants with active workflows
- `tenant-lifecycle-workflows`: Update status reporting requirement to include workflow sub-state details
- `database-persistence`: Add workflow execution status fields to tenant persistence layer

## Impact

- Database schema: New fields for workflow_sub_state, workflow_retry_count, workflow_error_message
- API responses: GET and LIST endpoints return additional workflow status fields
- Reconciler: Enhanced to poll detailed workflow execution state from providers
- Workflow providers: Must implement GetExecutionStatus method returning provider-specific state that can be mapped
- CLI: Display enhanced status information in tenant get/list output
