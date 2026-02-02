## MODIFIED Requirements

### Requirement: Tenant entities have unique user-facing identifiers
The system SHALL store tenant entities with unique user-facing tenant names separate from internal database IDs.

#### Scenario: Create tenant with unique tenant name
- **WHEN** a tenant is created with a unique name
- **THEN** the tenant is persisted with a generated internal UUID
- **AND** the name is unique across all tenants
- **AND** the tenant can be retrieved by its name

#### Scenario: Create tenant with duplicate tenant name
- **WHEN** a tenant is created with a name that already exists
- **THEN** the operation fails with ErrTenantExists
- **AND** no new tenant record is created

### Requirement: Tenant lifecycle follows defined status transitions
The system SHALL manage tenant lifecycle through statuses that include archived and deleted with valid state transitions.

#### Scenario: Tenant progresses through provisioning lifecycle
- **WHEN** a tenant is created with status "requested"
- **THEN** the tenant can transition to "planning" or "failed"
- **AND** from "planning" can transition to "provisioning" or "failed"
- **AND** from "provisioning" can transition to "ready" or "failed"
- **AND** from "ready" can transition to "updating", "deleting", or "archiving"

#### Scenario: Archived status is terminal for compute
- **WHEN** a tenant reaches "archived" status
- **THEN** no further compute lifecycle transitions are allowed
- **AND** the tenant remains queryable in the database

#### Scenario: Deleted status removes tenant record
- **WHEN** a tenant reaches "deleted" status
- **THEN** the tenant record is removed from the database
- **AND** subsequent queries return not found

#### Scenario: Invalid status transition is rejected
- **WHEN** a status transition is attempted that is not defined as valid
- **THEN** the operation fails with validation error
- **AND** the tenant status remains unchanged

### Requirement: Tenant deletions are soft deletes
The system SHALL distinguish archived tenants from deleted tenants and remove records on delete.

#### Scenario: Archive tenant
- **WHEN** a tenant delete workflow completes compute teardown
- **THEN** the tenant status is set to archived
- **AND** the tenant record remains in the database

#### Scenario: Hard delete tenant record
- **WHEN** a tenant is deleted after being archived
- **THEN** the tenant record is removed from the database
- **AND** it is excluded from all queries

#### Scenario: Include archived tenants in queries
- **WHEN** listing tenants with IncludeDeleted flag set to true
- **THEN** archived tenants are included in results
- **AND** deleted tenants are not returned because they no longer exist
