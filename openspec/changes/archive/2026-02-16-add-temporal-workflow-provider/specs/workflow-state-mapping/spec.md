## MODIFIED Requirements

### Requirement: Canonical workflow sub-states are provider-agnostic
The system SHALL define canonical workflow sub-states that abstract provider-specific execution states.

#### Scenario: Running state represents active execution
- **WHEN** a workflow is actively executing steps
- **THEN** the canonical sub-state SHALL be "running"
- **AND** this maps to Step Functions "RUNNING", Restate "running" or "active", and Temporal "RUNNING" or "CONTINUED_AS_NEW"

#### Scenario: Waiting state represents suspended execution
- **WHEN** a workflow is paused waiting for external input or callback
- **THEN** the canonical sub-state SHALL be "waiting"
- **AND** this maps to Restate "suspended" and Temporal activity or timer wait states exposed as non-running waits

#### Scenario: Backing-off state represents retry delay
- **WHEN** a workflow is in exponential backoff between retry attempts
- **THEN** the canonical sub-state SHALL be "backing-off"
- **AND** this is inferred from provider retry metadata, including Temporal attempt and retry state information

#### Scenario: Error state represents transient failure
- **WHEN** a workflow encountered an error but may retry
- **THEN** the canonical sub-state SHALL be "error"
- **AND** this represents a non-terminal failure condition

#### Scenario: Succeeded state represents successful completion
- **WHEN** a workflow completed successfully
- **THEN** the canonical sub-state SHALL be "succeeded"
- **AND** this maps to Step Functions "SUCCEEDED", Restate "completed" or "succeeded", and Temporal "COMPLETED"

#### Scenario: Failed state represents terminal failure
- **WHEN** a workflow failed with no further retries or was cancelled or terminated
- **THEN** the canonical sub-state SHALL be "failed"
- **AND** this maps to Step Functions "FAILED", "TIMED_OUT", "ABORTED", Restate "failed" or "error", and Temporal "FAILED", "TIMED_OUT", "CANCELED", or "TERMINATED"

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

#### Scenario: Temporal state mapping
- **WHEN** querying Temporal workflow execution status
- **THEN** the provider SHALL map Temporal status to canonical sub-states
- **AND** "RUNNING", "CONTINUED_AS_NEW" map to "running"
- **AND** "COMPLETED" maps to "succeeded"
- **AND** "FAILED", "TIMED_OUT", "CANCELED", "TERMINATED" map to "failed"

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

#### Scenario: Temporal retry detection
- **WHEN** Temporal execution metadata indicates retries or scheduled retry backoff
- **THEN** the system SHALL extract retry count from Temporal execution metadata
- **AND** set sub-state to "backing-off" while a retry is pending

#### Scenario: No retry metadata available
- **WHEN** provider does not expose retry metadata
- **THEN** the system SHALL use "running" as default sub-state
- **AND** retry count SHALL be 0

