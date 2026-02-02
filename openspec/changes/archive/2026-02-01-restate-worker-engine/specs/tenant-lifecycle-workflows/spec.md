## ADDED Requirements

### Requirement: Tenant create workflow
The system SHALL implement a tenant create workflow that provisions compute and records lifecycle transitions.

#### Scenario: Successful create
- **WHEN** a tenant create request is accepted
- **THEN** the create workflow SHALL provision compute for the tenant
- **AND** SHALL record a successful lifecycle transition

#### Scenario: Create idempotency
- **WHEN** a create workflow is invoked for an already-provisioned tenant
- **THEN** the workflow SHALL complete without error
- **AND** SHALL leave the tenant in the provisioned state

### Requirement: Tenant update workflow
The system SHALL implement a tenant update workflow that reconciles desired and observed state.

#### Scenario: Successful update
- **WHEN** a tenant update request is accepted
- **THEN** the update workflow SHALL reconcile compute to the desired configuration
- **AND** SHALL record a successful lifecycle transition

#### Scenario: Update with no changes
- **WHEN** a tenant update workflow runs with no drift detected
- **THEN** the workflow SHALL complete without changes
- **AND** SHALL record a no-op transition

### Requirement: Tenant delete workflow
The system SHALL implement a tenant delete workflow that deprovisions compute and removes tenant resources.

#### Scenario: Successful delete
- **WHEN** a tenant delete request is accepted
- **THEN** the delete workflow SHALL deprovision compute for the tenant
- **AND** SHALL record a successful lifecycle transition

#### Scenario: Delete idempotency
- **WHEN** a delete workflow is invoked for a tenant that is already deleted
- **THEN** the workflow SHALL complete without error
- **AND** SHALL leave the tenant in the deleted state

### Requirement: Workflow status reporting
The system SHALL publish workflow status updates for create, update, and delete operations.

#### Scenario: Status updates during execution
- **WHEN** a lifecycle workflow changes state
- **THEN** the workflow engine SHALL record status updates in the tenant history
- **AND** the status SHALL be retrievable via the tenant status API
