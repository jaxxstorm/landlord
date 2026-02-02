## ADDED Requirements

### Requirement: Workflows can invoke compute provisioning operations
The workflow system SHALL provide a mechanism for workflows to trigger compute provisioning operations (create, update, delete) through a workflow-accessible API.

#### Scenario: Workflow invokes tenant compute provisioning
- **WHEN** a workflow execution calls the compute provisioning API with tenant configuration
- **THEN** the compute engine SHALL receive the provisioning request
- **AND** the workflow SHALL receive a compute execution ID for tracking

#### Scenario: Workflow invokes tenant compute update
- **WHEN** a workflow execution calls the compute update API with modified tenant configuration
- **THEN** the compute engine SHALL receive the update request
- **AND** existing compute resources SHALL be updated or recreated as needed

#### Scenario: Workflow invokes tenant compute deletion
- **WHEN** a workflow execution calls the compute deletion API for a tenant
- **THEN** the compute engine SHALL receive the deletion request
- **AND** all associated compute resources SHALL be scheduled for cleanup

### Requirement: Compute provisioning API uses pluggable compute provider
The compute provisioning API invoked by workflows SHALL route through the pluggable compute provider abstraction, maintaining provider independence.

#### Scenario: Compute provider is injected into workflow system
- **WHEN** the workflow engine is initialized
- **THEN** it SHALL receive a ComputeManager instance as a dependency
- **AND** workflow operations SHALL delegate to this manager

#### Scenario: Different compute providers are transparently supported
- **WHEN** workflows invoke compute provisioning operations
- **THEN** the calls SHALL work with any registered compute provider (ECS, Kubernetes, etc.)
- **AND** the workflow logic SHALL not depend on provider-specific details

### Requirement: Compute provisioning requests include sufficient tenant context
Workflow invocations of compute provisioning operations SHALL include all necessary tenant configuration and runtime context.

#### Scenario: Create request includes tenant specification
- **WHEN** a workflow calls provision with a tenant ID
- **THEN** the compute provider SHALL receive the full tenant desired state (image, configuration, secrets)
- **AND** the provider SHALL have access to tenant metadata (labels, owner, environment)

#### Scenario: Update request preserves existing compute state
- **WHEN** a workflow calls update for an existing tenant
- **THEN** the compute provider SHALL receive both the current state and the desired state changes
- **AND** existing compute identifiers (task ARNs, container IDs, etc.) SHALL be available for updates

### Requirement: Compute operations are idempotent where possible
Compute provisioning operations invoked by workflows SHALL be idempotent to support retry scenarios.

#### Scenario: Duplicate create request returns existing compute
- **WHEN** a workflow retries a provision operation with the same tenant ID
- **THEN** if compute resources already exist, the operation SHALL return the existing resource identifiers
- **AND** no duplicate resources SHALL be created

#### Scenario: Update request is safe to retry
- **WHEN** a workflow retries an update operation
- **THEN** the operation SHALL produce the same final state as a single execution
- **AND** no partial or corrupted resources SHALL result from retries
