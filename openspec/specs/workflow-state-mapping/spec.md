# Specification: Workflow State Mapping

## Purpose

Define provider-agnostic workflow sub-state nomenclature and mapping rules from workflow provider-specific states to Landlord canonical states.

## ADDED Requirements

### Requirement: Canonical workflow sub-states are provider-agnostic
The system SHALL define canonical workflow sub-states that abstract provider-specific execution states.

#### Scenario: Running state represents active execution
- **WHEN** a workflow is actively executing steps
- **THEN** the canonical sub-state SHALL be "running"
- **AND** this maps to Step Functions "RUNNING" and Restate "running" or "active"

#### Scenario: Waiting state represents suspended execution
- **WHEN** a workflow is paused waiting for external input or callback
- **THEN** the canonical sub-state SHALL be "waiting"
- **AND** this maps to Restate "suspended" state

#### Scenario: Backing-off state represents retry delay
- **WHEN** a workflow is in exponential backoff between retry attempts
- **THEN** the canonical sub-state SHALL be "backing-off"
- **AND** this is inferred from provider retry metadata or execution history

#### Scenario: Error state represents transient failure
- **WHEN** a workflow encountered an error but may retry
- **THEN** the canonical sub-state SHALL be "error"
- **AND** this represents a non-terminal failure condition

#### Scenario: Succeeded state represents successful completion
- **WHEN** a workflow completed successfully
- **THEN** the canonical sub-state SHALL be "succeeded"
- **AND** this maps to Step Functions "SUCCEEDED" and Restate "completed"

#### Scenario: Failed state represents terminal failure
- **WHEN** a workflow failed with no further retries
- **THEN** the canonical sub-state SHALL be "failed"
- **AND** this maps to Step Functions "FAILED", "TIMED_OUT", "ABORTED" and Restate "failed"

### Requirement: Provider implementations map native states to canonical states
Workflow providers SHALL implement mapping logic from their native execution states to canonical sub-states.

#### Scenario: Step Functions state mapping
- **WHEN** querying Step Functions execution status
- **THEN** the provider SHALL map AWS state names to canonical sub-states
- **AND** "RUNNING" maps to "running"
- **AND** "SUCCEEDED" maps to "succeeded"
- **AND** "FAILED", "TIMED_OUT", "ABORTED" map to "failed"

#### Scenario: Restate state mapping
- **WHEN** querying Restate invocation status
- **THEN** the provider SHALL map Restate status to canonical sub-states
- **AND** "running", "active" map to "running"
- **AND** "suspended" maps to "waiting"
- **AND** "completed", "succeeded" map to "succeeded"
- **AND** "failed", "error" map to "failed"

#### Scenario: Unknown provider state defaults to running
- **WHEN** a provider returns an unrecognized state
- **THEN** the system SHALL default to "running" canonical sub-state
- **AND** log a warning about the unmapped state

### Requirement: Backing-off sub-state is inferred from retry metadata
The system SHALL determine backing-off sub-state by analyzing workflow execution retry behavior.

#### Scenario: Step Functions retry detection
- **WHEN** Step Functions execution history shows retry events
- **THEN** the system SHALL count retry attempts from event history
- **AND** set sub-state to "backing-off" if execution is between retry attempts

#### Scenario: Restate retry detection
- **WHEN** Restate invocation metadata indicates retry scheduling
- **THEN** the system SHALL extract retry count from invocation response
- **AND** set sub-state to "backing-off" if retry is scheduled

#### Scenario: No retry metadata available
- **WHEN** provider does not expose retry metadata
- **THEN** the system SHALL use "running" as default sub-state
- **AND** retry count SHALL be 0

### Requirement: State mapping is documented in provider interface
The workflow provider interface specification SHALL document the mapping from provider-specific states to canonical sub-states.

#### Scenario: Provider interface includes state mapping table
- **WHEN** implementing a new workflow provider
- **THEN** the provider documentation SHALL include a state mapping table
- **AND** the table shows provider states and their canonical equivalents

#### Scenario: State mapping is testable
- **WHEN** testing workflow provider implementations
- **THEN** tests SHALL verify correct mapping of all provider states
- **AND** tests SHALL verify handling of unknown states
