## MODIFIED Requirements

### Requirement: Tenants maintain desired and observed state separation
The system SHALL store both desired state (what should exist), observed state (what actually exists), and workflow execution status for each tenant.

#### Scenario: Tenant created with desired state
- **WHEN** a tenant is created with desired_image and desired_config
- **THEN** the desired state is persisted
- **AND** the observed state fields are initially empty
- **AND** the workflow execution status fields are initially empty
- **AND** the tenant can be updated to reflect observed state and workflow status

#### Scenario: Workflow execution status fields are persisted
- **WHEN** reconciler updates workflow execution status
- **THEN** workflow_sub_state is stored in the database
- **AND** workflow_retry_count is stored in the database
- **AND** workflow_error_message is stored in the database
- **AND** these fields are queryable via repository methods

#### Scenario: Detect state drift
- **WHEN** a tenant has desired_image different from observed_image
- **THEN** the tenant is identified as drifted
- **AND** reconciliation logic can detect the mismatch
- **AND** workflow execution status provides insight into reconciliation progress

### Requirement: Database schema includes workflow execution status fields
The tenants table SHALL include columns for workflow sub-state, retry count, and error message.

#### Scenario: Workflow sub-state column exists
- **WHEN** querying tenants table schema
- **THEN** a workflow_sub_state VARCHAR(50) NULL column SHALL exist
- **AND** the column stores canonical sub-state values

#### Scenario: Workflow retry count column exists
- **WHEN** querying tenants table schema
- **THEN** a workflow_retry_count INTEGER NULL column SHALL exist
- **AND** the column defaults to 0 when workflow is active

#### Scenario: Workflow error message column exists
- **WHEN** querying tenants table schema
- **THEN** a workflow_error_message TEXT NULL column SHALL exist
- **AND** the column stores error messages from workflow providers

#### Scenario: Workflow status fields are nullable
- **WHEN** a tenant has no active workflow
- **THEN** all workflow status fields SHALL be NULL
- **AND** the fields are only populated when workflow_execution_id is set

### Requirement: Repository supports querying by workflow execution status
The tenant repository SHALL provide methods to query tenants by workflow sub-state, retry count, and error presence.

#### Scenario: Query tenants by workflow sub-state
- **WHEN** repository is queried for tenants with specific workflow_sub_state
- **THEN** only tenants matching that sub-state are returned
- **AND** query supports filtering by multiple sub-states

#### Scenario: Query tenants with workflow errors
- **WHEN** repository is queried for tenants with non-null workflow_error_message
- **THEN** only tenants with errors are returned
- **AND** results include the error message content

#### Scenario: Query tenants exceeding retry threshold
- **WHEN** repository is queried for tenants with workflow_retry_count > threshold
- **THEN** only tenants exceeding the threshold are returned
- **AND** results can be sorted by retry count

### Requirement: Migration adds workflow execution status columns
The database migration SHALL add three nullable columns to the tenants table for workflow execution status.

#### Scenario: Migration adds workflow_sub_state column
- **WHEN** migration is applied
- **THEN** workflow_sub_state VARCHAR(50) NULL column is added
- **AND** existing tenant records have NULL values for this column

#### Scenario: Migration adds workflow_retry_count column
- **WHEN** migration is applied
- **THEN** workflow_retry_count INTEGER NULL column is added
- **AND** existing tenant records have NULL values for this column

#### Scenario: Migration adds workflow_error_message column
- **WHEN** migration is applied
- **THEN** workflow_error_message TEXT NULL column is added
- **AND** existing tenant records have NULL values for this column

#### Scenario: Migration is non-blocking
- **WHEN** migration is applied
- **THEN** the operation does not lock the tenants table
- **AND** existing queries continue to function
- **AND** columns are added without constraints that would block writes

#### Scenario: Migration rollback removes columns
- **WHEN** migration is rolled back
- **THEN** all three workflow status columns are dropped
- **AND** existing data is preserved (tenant records remain intact)
