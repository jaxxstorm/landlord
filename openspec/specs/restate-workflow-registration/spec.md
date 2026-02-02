## ADDED Requirements

### Requirement: System registers workflows at startup
The system SHALL register all tenant lifecycle workflows with the Restate backend during the workflow provider initialization.

#### Scenario: Successful workflow registration on startup
- **WHEN** the Restate workflow provider initializes at application startup
- **THEN** all tenant lifecycle workflows are registered with the Restate backend
- **AND** registration completion is logged

#### Scenario: Startup with unavailable Restate backend
- **WHEN** the Restate workflow provider initializes but the Restate backend is unavailable
- **THEN** the system logs a warning about registration failure
- **AND** the system continues startup without failing

### Requirement: Workflow registration is idempotent
The system SHALL handle re-registration of workflows gracefully without errors.

#### Scenario: Re-registration of existing workflow
- **WHEN** a workflow that is already registered with Restate is registered again
- **THEN** the operation succeeds without error
- **AND** the workflow remains available for execution

#### Scenario: Partial registration recovery
- **WHEN** some workflows are already registered and the system restarts
- **THEN** all workflows are registered successfully
- **AND** already-registered workflows do not cause errors

### Requirement: Registration errors are logged and reported
The system SHALL provide clear logging and error messages for workflow registration failures.

#### Scenario: Registration failure is logged
- **WHEN** workflow registration fails for any reason
- **THEN** the error is logged with the workflow name and error details
- **AND** the error message clearly indicates the workflow is unavailable

#### Scenario: Successful registration is confirmed
- **WHEN** all workflows are successfully registered
- **THEN** a log entry confirms successful registration
- **AND** the log includes the count of registered workflows

### Requirement: Workflows are executable after registration
The system SHALL ensure that registered workflows can be invoked successfully through the workflow provider interface.

#### Scenario: Tenant provisioning workflow execution
- **WHEN** a registered tenant provisioning workflow is invoked
- **THEN** the workflow executes successfully in the Restate backend
- **AND** workflow execution completes without "workflow not found" errors

#### Scenario: Workflow invocation before registration completes
- **WHEN** a workflow invocation is attempted before registration is complete
- **THEN** the system returns an appropriate error indicating the workflow is not yet available
- **AND** the error message suggests retrying after startup

### Requirement: Registration integrates with provider initialization
The system SHALL perform workflow registration as part of the standard workflow provider initialization process.

#### Scenario: Provider initialization includes registration
- **WHEN** the Restate workflow provider is initialized
- **THEN** workflow registration is performed automatically
- **AND** registration completes before the provider is marked as ready

#### Scenario: Registration failure does not block provider creation
- **WHEN** workflow registration fails during provider initialization
- **THEN** the provider initialization completes
- **AND** the provider is created but with workflows unavailable
