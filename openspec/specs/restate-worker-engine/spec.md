## MODIFIED Requirements

### Requirement: Worker registration on startup
The restate worker engine SHALL register its services and workflows with Restate when the worker process starts.

#### Scenario: Successful registration
- **WHEN** the restate worker engine starts
- **THEN** it SHALL register required services and workflows with the Restate admin API

#### Scenario: Registration retry
- **WHEN** the Restate admin API is temporarily unavailable at startup
- **THEN** the worker engine SHALL retry registration with backoff
- **AND** SHALL surface a startup error if registration cannot be completed

### Requirement: Tenant lifecycle execution
The restate worker engine SHALL execute tenant create, update, and delete workflows using payload-provided tenant state without direct database access.

#### Scenario: Execute create workflow
- **WHEN** a create job is received
- **THEN** the worker engine SHALL use the payload-provided tenant state

#### Scenario: Execute update workflow
- **WHEN** an update job is received
- **THEN** the worker engine SHALL use the payload-provided tenant state

#### Scenario: Execute delete workflow
- **WHEN** a delete job is received
- **THEN** the worker engine SHALL use the payload-provided tenant state
