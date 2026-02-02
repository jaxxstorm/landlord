## ADDED Requirements

### Requirement: Default compute metadata keys
The system SHALL define a standard set of default metadata keys for compute resources that includes an owner label for Landlord and the tenant identifier.

#### Scenario: Default metadata key set exists
- **WHEN** compute metadata is constructed for a tenant
- **THEN** it includes an owner label indicating Landlord ownership
- **AND** it includes the tenant ID

### Requirement: Apply default metadata across compute resources
Compute providers SHALL apply the default metadata to all resources they create for a tenant.

#### Scenario: Provider creates compute resources
- **WHEN** a compute provider provisions resources for a tenant
- **THEN** all created resources include the default metadata labels or tags

### Requirement: Docker provider applies default labels
The Docker compute provider MUST apply the default metadata as Docker labels on all created resources.

#### Scenario: Docker provider provisions a tenant
- **WHEN** the Docker provider creates containers for a tenant
- **THEN** each container includes the default labels including owner and tenant ID

### Requirement: Metadata keys are namespaced
Default metadata keys MUST be namespaced under a Landlord-specific prefix to avoid collisions.

#### Scenario: Metadata keys are emitted
- **WHEN** the system emits default metadata
- **THEN** keys use a `landlord.*` namespace prefix

### Requirement: Provider-specific mapping compatibility
The default metadata set SHALL be compatible with providers that use labels or tags (e.g., Kubernetes labels, ECS tags).

#### Scenario: Provider maps metadata to its labeling system
- **WHEN** a provider maps default metadata to its native labels or tags
- **THEN** the mapping preserves owner and tenant identifiers without loss
