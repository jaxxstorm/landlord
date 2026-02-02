## ADDED Requirements

### Requirement: Worker registration on startup
The restate worker engine SHALL register its services and workflows with Restate when the worker process starts.

#### Scenario: Successful registration
- **WHEN** the restate worker engine starts
- **THEN** it SHALL register required services and workflows with the Restate admin API
- **AND** it SHALL log a confirmation message with the registered service names

#### Scenario: Registration retry
- **WHEN** the Restate admin API is temporarily unavailable at startup
- **THEN** the worker engine SHALL retry registration with backoff
- **AND** SHALL surface a startup error if registration cannot be completed

### Requirement: Worker readiness gating
The restate worker engine SHALL not report readiness until registration is complete.

#### Scenario: Ready after registration
- **WHEN** registration succeeds
- **THEN** the worker engine SHALL mark itself ready to accept jobs

#### Scenario: Not ready on failure
- **WHEN** registration fails
- **THEN** the worker engine SHALL remain not-ready and return an error to the caller

### Requirement: Tenant lifecycle execution
The restate worker engine SHALL execute tenant create, update, and delete workflows.

#### Scenario: Execute create workflow
- **WHEN** a create job is received
- **THEN** the worker engine SHALL execute the tenant provisioning workflow
- **AND** SHALL report completion or failure with a durable status

#### Scenario: Execute update workflow
- **WHEN** an update job is received
- **THEN** the worker engine SHALL execute the tenant update workflow
- **AND** SHALL report completion or failure with a durable status

#### Scenario: Execute delete workflow
- **WHEN** a delete job is received
- **THEN** the worker engine SHALL execute the tenant deletion workflow
- **AND** SHALL report completion or failure with a durable status

### Requirement: Restate worker configuration
The system SHALL provide Restate worker configuration settings for registration and connectivity.

#### Scenario: Required worker configuration
- **WHEN** restate worker configuration is loaded
- **THEN** it SHALL include:
  - admin_endpoint (string, required): Restate admin API URL
  - namespace (string, optional): Restate namespace for registrations
  - service_prefix (string, optional): prefix for registered services
  - auth_type (string, optional): authentication method
  - api_key (string, optional): API key for auth_type api_key

#### Scenario: Configuration validation
- **WHEN** the worker starts with invalid configuration
- **THEN** it SHALL fail fast with a clear configuration error
