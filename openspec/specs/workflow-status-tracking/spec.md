## MODIFIED Requirements

### Requirement: Tenant records store workflow execution ID
The tenant database record SHALL store the workflow execution ID in the workflow_execution_id column and update it when workflows are triggered by the reconciler.

#### Scenario: Workflow execution ID set on reconciler trigger
- **WHEN** the reconciler triggers a workflow and receives an execution ID
- **THEN** the tenant record MUST be updated with the workflow_execution_id value alongside the status update

#### Scenario: Workflow execution ID preserved across status transitions
- **WHEN** a tenant transitions through multiple statuses (planning → provisioning → ready)
- **THEN** the workflow_execution_id MUST remain set until the workflow completes or fails

#### Scenario: Workflow execution ID cleared on completion
- **WHEN** a workflow completes successfully and tenant reaches "ready" status
- **THEN** the workflow_execution_id MAY be cleared or retained for audit purposes
