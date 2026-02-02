## MODIFIED Requirements

### Requirement: Workflow status reporting
The system SHALL publish workflow status updates for create, update, archive, and delete operations via the reconciler and workflow provider status polling.

#### Scenario: Status updates during execution
- **WHEN** a lifecycle workflow changes state
- **THEN** the reconciler SHALL update tenant status and status message based on provider status
- **AND** status history MAY be recorded when supported
- **AND** the status SHALL be retrievable via the tenant status API
