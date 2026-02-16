# Specification: Temporal Worker Engine

## Purpose

Define worker-engine behavior for Temporal workflow execution, lifecycle activities, and compute-provider integration.

## ADDED Requirements

### Requirement: Temporal worker engine registers lifecycle workflows at startup
The Temporal worker engine SHALL register required workflows and activities during worker startup before serving lifecycle jobs.

#### Scenario: Successful worker startup registration
- **WHEN** the Temporal worker process starts with valid configuration
- **THEN** it SHALL register tenant lifecycle workflows and required activities with its configured task queue
- **AND** it SHALL emit a startup log confirming registration readiness
- **AND** lifecycle executions for workflow ID `tenant-provisioning` SHALL be routable to registered workflow/activity handlers

#### Scenario: Temporal backend unavailable during startup
- **WHEN** worker startup cannot connect to Temporal services
- **THEN** the worker SHALL retry registration with backoff for a bounded interval
- **AND** return a startup error if readiness cannot be achieved

### Requirement: Temporal worker engine executes create, update, and delete jobs from payload
The Temporal worker engine SHALL execute tenant create, update, and delete workflows using payload-provided tenant state and shared lifecycle logic.

#### Scenario: Execute create job
- **WHEN** a create lifecycle job is dispatched to the Temporal worker
- **THEN** the worker SHALL execute create orchestration using tenant payload fields
- **AND** persist status transitions through existing lifecycle APIs or adapters

#### Scenario: Execute update job
- **WHEN** an update lifecycle job is dispatched to the Temporal worker
- **THEN** the worker SHALL execute update orchestration using tenant payload fields
- **AND** preserve idempotency across retries and worker restarts

#### Scenario: Execute delete job
- **WHEN** a delete lifecycle job is dispatched to the Temporal worker
- **THEN** the worker SHALL execute delete orchestration using tenant payload fields
- **AND** complete cleanup and terminal status transitions through existing lifecycle adapters

### Requirement: Temporal worker engine supports all registered compute providers
The Temporal worker engine SHALL resolve compute execution through Landlord compute abstractions so workflow behavior remains provider-agnostic.

#### Scenario: Compute provider selected from payload
- **WHEN** lifecycle payload includes an explicit compute provider selection
- **THEN** the worker SHALL use that provider through compute manager abstractions
- **AND** execute lifecycle actions without Temporal-specific compute branching

#### Scenario: Compute provider selected from defaults
- **WHEN** lifecycle payload omits explicit compute provider selection
- **THEN** the worker SHALL resolve compute provider using existing default resolver logic
- **AND** execute against any compute provider that is registered and enabled

#### Scenario: Unknown compute provider in payload
- **WHEN** payload references a compute provider that is not registered
- **THEN** the worker SHALL fail the activity with an actionable validation error
- **AND** expose the error through workflow status and logs

### Requirement: Temporal worker engine handles cancellation safely
The Temporal worker engine SHALL process cancellation signals in a deterministic and idempotent way.

#### Scenario: Cancellation for running lifecycle workflow
- **WHEN** a running Temporal lifecycle execution receives cancellation
- **THEN** the worker SHALL stop in-flight activity execution at safe boundaries
- **AND** persist a terminal status and reason compatible with existing tenant status reporting

#### Scenario: Duplicate cancellation signal
- **WHEN** cancellation is requested more than once for the same execution
- **THEN** cancellation handling MUST remain idempotent
- **AND** duplicate cancellation handling MUST NOT produce duplicate cleanup side effects
