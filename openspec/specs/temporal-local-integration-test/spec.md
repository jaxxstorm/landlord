# Specification: Temporal Local Integration Test

## Purpose

Define local Docker Compose and integration-test requirements for Temporal-backed tenant lifecycle validation.

## ADDED Requirements

### Requirement: Local docker-compose environment supports Temporal workflow testing
The system SHALL provide a local Docker Compose topology that can run Landlord with Temporal workflow components for integration testing.

#### Scenario: Bring up Temporal local stack
- **WHEN** a developer starts the local integration stack for Temporal
- **THEN** Docker Compose SHALL start required Temporal services and Landlord components needed for lifecycle execution
- **AND** service health checks SHALL indicate readiness before tests run

#### Scenario: Temporal worker available in local stack
- **WHEN** the Temporal local stack is healthy
- **THEN** a Temporal worker process SHALL be connected to the configured namespace and task queue
- **AND** lifecycle workflows SHALL be discoverable and executable

### Requirement: Integration tests validate end-to-end tenant lifecycle with Temporal
The system SHALL include integration tests that validate create, update, delete, and cancellation behavior through Temporal-backed workflows.

#### Scenario: End-to-end lifecycle execution
- **WHEN** the Temporal integration suite creates, updates, and deletes a tenant via API
- **THEN** workflow executions SHALL complete and update tenant status through existing status APIs
- **AND** no provider-specific workflow-not-found or registration errors SHALL occur

#### Scenario: Lifecycle cancellation in integration test
- **WHEN** a test cancels an in-flight tenant lifecycle workflow
- **THEN** the workflow SHALL transition to a terminal cancelled or failed state consistent with canonical mapping
- **AND** the tenant status response SHALL include cancellation-related status context

### Requirement: Temporal tests cover compute-provider compatibility and idempotency
The system SHALL verify that Temporal workflow execution remains compatible with every supported compute provider and with retry/idempotent execution behavior.

#### Scenario: Compute provider matrix validation
- **WHEN** integration tests run in environments with multiple compute providers enabled
- **THEN** Temporal lifecycle tests SHALL run against each enabled compute provider
- **AND** provider-specific failures SHALL report clear, provider-attributed errors

#### Scenario: Idempotent execution and registration behavior
- **WHEN** integration or unit tests repeat workflow registration, start, or cancellation operations
- **THEN** repeated operations SHALL be accepted without duplicate side effects
- **AND** tests SHALL verify deterministic outcomes under retries
