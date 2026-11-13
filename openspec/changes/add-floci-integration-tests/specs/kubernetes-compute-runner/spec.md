## ADDED Requirements

### Requirement: Kubernetes runner provisions tenant workloads
The system SHALL provide a `kubernetes` compute provider that creates a
tenant-owned namespace and Deployment through a configured Kubernetes API using
the tenant container configuration.

#### Scenario: Provision Kubernetes tenant
- **WHEN** a lifecycle request selects `kubernetes` with a valid container configuration
- **THEN** the provider SHALL create the tenant namespace and Deployment
- **AND** it SHALL return namespace and workload identifiers in the provision result

### Requirement: Kubernetes runner lifecycle operations are idempotent
The Kubernetes provider SHALL query workload status, apply updates to the
tenant-owned Deployment, and delete tenant-owned resources idempotently.

#### Scenario: Update Kubernetes workload
- **WHEN** a lifecycle update changes the tenant container configuration
- **THEN** the provider SHALL apply the updated Deployment configuration
- **AND** it SHALL report an update result without creating a second tenant namespace

#### Scenario: Destroy Kubernetes tenant
- **WHEN** a lifecycle request destroys a Kubernetes tenant
- **THEN** the provider SHALL delete the tenant namespace and its owned workload
- **AND** it SHALL succeed when the namespace is already absent

### Requirement: Floci EKS validates Kubernetes runner behavior
The integration suite SHALL provision an ephemeral Kubernetes cluster through
Floci EKS and exercise Kubernetes runner provision, status, update, and destroy
against its Kubernetes API.

#### Scenario: Kubernetes runner integration lifecycle
- **WHEN** the Floci EKS integration test runs on a Docker-capable host
- **THEN** it SHALL create and clean up a tenant namespace and workload through the Kubernetes runner
- **AND** it SHALL verify workload status and update behavior before cleanup
