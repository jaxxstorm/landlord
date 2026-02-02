## ADDED Requirements

### Requirement: Provisioning request payloads for Restate
The Restate workflow provider SHALL accept execution inputs that include tenant identity, action, and desired configuration without requiring direct database access.

#### Scenario: Provisioning input for plan execution
- **WHEN** StartExecution is called for a plan action
- **THEN** the execution input SHALL include tenant ID, action, and desired configuration
- **AND** the worker SHALL use the request payload without querying the database

#### Scenario: Provisioning input for update execution
- **WHEN** StartExecution is called for an update action
- **THEN** the execution input SHALL include tenant ID, action, and desired configuration
- **AND** the worker SHALL use the request payload without querying the database
