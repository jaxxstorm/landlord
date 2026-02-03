## MODIFIED Requirements

### Requirement: Return complete tenant data
The system SHALL return all tenant fields including workflow execution status fields in the response.

#### Scenario: Response includes tenant details
- **WHEN** a tenant is retrieved successfully
- **THEN** the response includes `id`, `name`, `created_at`, `updated_at`, and all other tenant properties
- **AND** the response includes `workflow_execution_id` when a workflow is active
- **AND** the response includes `workflow_sub_state` when a workflow is active
- **AND** the response includes `workflow_retry_count` when a workflow is active
- **AND** the response includes `workflow_error_message` when a workflow has encountered an error

#### Scenario: Workflow status fields are null when no active workflow
- **WHEN** a tenant has no active workflow execution
- **THEN** `workflow_execution_id` SHALL be null or omitted
- **AND** `workflow_sub_state` SHALL be null or omitted
- **AND** `workflow_retry_count` SHALL be null or omitted
- **AND** `workflow_error_message` SHALL be null or omitted

#### Scenario: Workflow status fields populated during provisioning
- **WHEN** a tenant is in "provisioning" status with active workflow
- **THEN** the response SHALL include `workflow_sub_state` (e.g., "running", "backing-off")
- **AND** the response SHALL include `workflow_retry_count` (e.g., 0, 1, 2)
- **AND** the response MAY include `workflow_error_message` if workflow has errors
