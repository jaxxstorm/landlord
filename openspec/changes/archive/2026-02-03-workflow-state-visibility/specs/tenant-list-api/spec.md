## MODIFIED Requirements

### Requirement: List all tenants
The system SHALL provide a GET endpoint at `/v1/tenants` that returns a list of all tenants including workflow execution status summary.

#### Scenario: Retrieve all tenants
- **WHEN** a client sends a GET request to `/v1/tenants`
- **THEN** the system returns an array of tenant resources
- **AND** each tenant includes workflow execution status fields when applicable

#### Scenario: Tenant list includes workflow sub-state
- **WHEN** listing tenants with active workflows
- **THEN** each tenant response SHALL include `workflow_sub_state` field
- **AND** the field shows current execution sub-state (e.g., "running", "backing-off", "error")

#### Scenario: Tenant list includes retry count
- **WHEN** listing tenants with active workflows experiencing retries
- **THEN** each tenant response SHALL include `workflow_retry_count` field
- **AND** the field shows number of retry attempts (e.g., 0, 1, 2)

#### Scenario: Tenant list includes error summary
- **WHEN** listing tenants with workflow errors
- **THEN** each tenant response SHALL include `workflow_error_message` field
- **AND** the field contains the error message from the workflow provider

#### Scenario: No tenants exist
- **WHEN** there are no tenants in the system
- **THEN** the system returns HTTP 200 with an empty array

### Requirement: Support filtering by workflow sub-state
The system SHALL allow filtering tenants by workflow sub-state.

#### Scenario: Filter by workflow sub-state
- **WHEN** a `workflow_sub_state` query parameter is provided
- **THEN** the system returns only tenants with that sub-state
- **AND** valid values include "running", "backing-off", "waiting", "error", "succeeded", "failed"

#### Scenario: Filter for tenants with errors
- **WHEN** a `has_workflow_error=true` query parameter is provided
- **THEN** the system returns only tenants with non-null workflow_error_message
- **AND** results show tenants in error or backing-off states

#### Scenario: Filter for tenants with high retry counts
- **WHEN** a `min_retry_count` query parameter is provided
- **THEN** the system returns only tenants with retry count >= threshold
- **AND** results can be sorted by retry count descending
