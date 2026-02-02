## MODIFIED Requirements

### Requirement: Tenant delete workflow
The system SHALL implement a tenant delete workflow that deprovisions compute and archives tenants.

#### Scenario: Successful delete
- **WHEN** a tenant delete request is accepted
- **THEN** the delete workflow SHALL deprovision compute for the tenant
- **AND** SHALL record a successful lifecycle transition to archived

#### Scenario: Delete idempotency
- **WHEN** a delete workflow is invoked for a tenant that is already archived
- **THEN** the workflow SHALL complete without error
- **AND** SHALL leave the tenant in the archived state

### Requirement: Workflow status reporting
The system SHALL publish workflow status updates for create, update, archive, and delete operations.

#### Scenario: Status updates during execution
- **WHEN** a lifecycle workflow changes state
- **THEN** the workflow engine SHALL record status updates in the tenant history
- **AND** the status SHALL be retrievable via the tenant status API
