## ADDED Requirements

### Requirement: Provisioning request payloads for Step Functions
The Step Functions provider SHALL accept execution inputs that include tenant identity, action, and desired configuration without requiring direct database access.

#### Scenario: Provisioning input for plan execution
- **WHEN** StartExecution is called for a plan action
- **THEN** the execution input SHALL include tenant ID, action, and desired configuration
- **AND** the worker SHALL use the request payload without querying the database

#### Scenario: Provisioning input for delete execution
- **WHEN** StartExecution is called for a delete action
- **THEN** the execution input SHALL include tenant ID, action, and desired configuration
- **AND** the worker SHALL use the request payload without querying the database
