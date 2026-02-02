## ADDED Requirements

### Requirement: Integration test validates Restate with Docker compute
The system SHALL provide integration tests that validate end-to-end tenant provisioning using the Restate workflow provider with the Docker compute provider.

#### Scenario: End-to-end tenant provisioning test
- **WHEN** an integration test provisions a tenant using Restate and Docker compute
- **THEN** the tenant is created successfully
- **AND** the workflow executes without errors
- **AND** the Docker container is started

#### Scenario: Test setup initializes Restate backend
- **WHEN** the integration test suite runs
- **THEN** a Restate backend instance is available for testing
- **AND** workflows are registered before tests execute

### Requirement: Integration test validates workflow registration
The system SHALL include integration tests that verify workflow registration occurs correctly.

#### Scenario: Workflow registration verification
- **WHEN** the integration test checks workflow registration
- **THEN** all expected tenant lifecycle workflows are registered with Restate
- **AND** registered workflows can be invoked successfully

#### Scenario: Registration state after provider restart
- **WHEN** the integration test restarts the workflow provider
- **THEN** workflows remain registered or are re-registered
- **AND** tenant provisioning continues to work

### Requirement: Integration test validates error handling
The system SHALL include integration tests that verify error handling for workflow operations.

#### Scenario: Workflow not found error is eliminated
- **WHEN** the integration test invokes tenant provisioning
- **THEN** no "workflow not found" errors occur
- **AND** the workflow executes successfully

#### Scenario: Clear error messages for failures
- **WHEN** a workflow execution fails for a legitimate reason
- **THEN** the error message is clear and actionable
- **AND** the error does not indicate workflow registration issues

### Requirement: Integration test validates idempotency
The system SHALL include integration tests that verify workflow registration idempotency.

#### Scenario: Multiple registration attempts succeed
- **WHEN** the integration test registers workflows multiple times
- **THEN** all registration attempts succeed
- **AND** workflows remain functional after repeated registration

### Requirement: Integration test uses realistic configuration
The system SHALL use realistic configuration in integration tests that matches production setup.

#### Scenario: Test uses Docker compute provider
- **WHEN** the integration test runs tenant provisioning
- **THEN** the Docker compute provider is used for actual container operations
- **AND** containers are created and cleaned up properly

#### Scenario: Test uses Restate workflow provider
- **WHEN** the integration test runs tenant provisioning
- **THEN** the Restate workflow provider is used for workflow execution
- **AND** workflows execute in the Restate backend
