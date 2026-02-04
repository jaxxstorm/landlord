## ADDED Requirements

### Requirement: System identifies workflows in degraded states
The system SHALL identify workflow executions in degraded states that indicate provisioning failures suitable for restart.

#### Scenario: Backing-off state is degraded
- **WHEN** workflow execution has SubState == SubStateBackingOff
- **THEN** the system SHALL classify the workflow as degraded
- **AND** eligible for restart on config change

#### Scenario: Retrying state is degraded
- **WHEN** workflow execution has SubState == SubStateRetrying
- **THEN** the system SHALL classify the workflow as degraded
- **AND** eligible for restart on config change

#### Scenario: Running state is not degraded
- **WHEN** workflow execution has SubState == SubStateRunning
- **THEN** the system SHALL NOT classify the workflow as degraded
- **AND** NOT eligible for restart on config change

#### Scenario: Succeeded state is not degraded
- **WHEN** workflow execution has State == StateDone and SubState == SubStateSuccess
- **THEN** the system SHALL NOT classify the workflow as degraded
- **AND** NOT eligible for restart

#### Scenario: Failed state handled separately
- **WHEN** workflow execution has State == StateDone and SubState == SubStateError
- **THEN** the system SHALL NOT classify as degraded (terminal failure)
- **AND** handle through normal failure reconciliation flow

### Requirement: Reconciler checks degraded state before config-based restart
The reconciler SHALL only trigger workflow restart for degraded workflows when config changes.

#### Scenario: Config change with degraded workflow triggers restart
- **WHEN** reconciler detects config hash mismatch
- **AND** workflow is classified as degraded (backing-off or retrying)
- **THEN** the reconciler SHALL trigger workflow stop and restart sequence

#### Scenario: Config change with healthy workflow preserves execution
- **WHEN** reconciler detects config hash mismatch
- **AND** workflow is NOT classified as degraded (running, succeeded, failed)
- **THEN** the reconciler SHALL NOT trigger workflow restart
- **AND** allow current execution to continue

#### Scenario: No config change skips degraded check
- **WHEN** config hash matches workflow metadata config_hash
- **THEN** the reconciler SHALL NOT check degraded state for restart
- **AND** continue normal status polling

### Requirement: Degraded state detection uses canonical sub-states
The system SHALL use canonical workflow sub-states for consistent degraded detection across providers.

#### Scenario: Map provider-specific states to canonical
- **WHEN** workflow provider returns execution status
- **THEN** the system SHALL map provider-specific sub-state strings to canonical SubState enum
- **AND** use canonical values for degraded detection

#### Scenario: Backing-off sub-state from multiple provider formats
- **WHEN** provider returns "backing-off", "backoff", "retrying", or "retry" state strings
- **THEN** the system SHALL map to canonical SubStateBackingOff or SubStateRetrying
- **AND** consistently classify as degraded

#### Scenario: Unknown sub-states not classified as degraded
- **WHEN** provider returns unknown or unmapped sub-state
- **THEN** the system SHALL NOT classify as degraded by default
- **AND** log warning about unmapped state for investigation
