# database-persistence Specification

## Purpose
Defines the tenant persistence layer requirements including repository interface, domain types, lifecycle status management, state history audit trail, and PostgreSQL implementation with optimistic locking.

## Requirements
### Requirement: Application connects to PostgreSQL database

The system SHALL establish a connection pool to a PostgreSQL database using configuration parameters.

#### Scenario: Successful database connection
- **WHEN** the application starts with valid database configuration
- **THEN** a connection pool is established to PostgreSQL
- **AND** the pool is ready to execute queries

#### Scenario: Connection fails due to invalid credentials
- **WHEN** the application starts with invalid database credentials
- **THEN** the application fails to start with a clear authentication error
- **AND** the error message indicates credential issues

#### Scenario: Connection fails due to unreachable host
- **WHEN** the database host is unreachable
- **THEN** the application retries connection with exponential backoff
- **AND** fails after maximum retry attempts with a clear error

### Requirement: Database connection pool is configurable

The system SHALL allow configuration of connection pool parameters.

#### Scenario: Connection pool sizing
- **WHEN** max connections and min connections are configured
- **THEN** the pool maintains at least min connections
- **AND** the pool grows up to max connections under load

#### Scenario: Connection timeout configuration
- **WHEN** connection timeout is configured
- **THEN** connection attempts fail after the timeout period
- **AND** the timeout error is logged appropriately

### Requirement: Database health checks verify connectivity

The system SHALL provide a mechanism to check database connectivity status.

#### Scenario: Health check with healthy database
- **WHEN** database health check is performed
- **THEN** a ping query succeeds within the timeout
- **AND** the health check returns success

#### Scenario: Health check with disconnected database
- **WHEN** database health check is performed and database is unavailable
- **THEN** the health check returns failure
- **AND** the failure includes connection error details

### Requirement: Database migrations are applied on startup

The system SHALL automatically apply pending database migrations on application startup.

#### Scenario: Migrations applied successfully
- **WHEN** the application starts and pending migrations exist
- **THEN** all pending migrations are applied in order
- **AND** the schema version is updated to reflect applied migrations

#### Scenario: Migration fails during application
- **WHEN** a migration fails to apply
- **THEN** the migration is rolled back if possible
- **AND** the application fails to start with migration error details

#### Scenario: No pending migrations
- **WHEN** the application starts and all migrations are current
- **THEN** no migrations are applied
- **AND** the application continues startup normally

### Requirement: Migration files are embedded in binary

The system SHALL embed migration files in the compiled binary.

#### Scenario: Migrations available without external files
- **WHEN** the application binary is deployed
- **THEN** migration files are accessible from the embedded filesystem
- **AND** no external migration file directory is required

### Requirement: Database connections are properly closed on shutdown

The system SHALL close all database connections during graceful shutdown.

#### Scenario: Clean connection closure
- **WHEN** the application shuts down gracefully
- **THEN** all active database connections are closed
- **AND** the connection pool is drained cleanly

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
- **AND** from "ready" can transition to "updating" or "deleting"

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

### Requirement: Tenants maintain desired and observed state separation

The system SHALL store both desired state (what should exist) and observed state (what actually exists) for each tenant.

#### Scenario: Tenant created with desired state
- **WHEN** a tenant is created with desired_image and desired_config
- **THEN** the desired state is persisted
- **AND** the observed state fields are initially empty
- **AND** the tenant can be updated to reflect observed state

#### Scenario: Detect state drift
- **WHEN** a tenant has desired_image different from observed_image
- **THEN** the tenant is identified as drifted
- **AND** reconciliation logic can detect the mismatch

### Requirement: Tenant updates use optimistic locking for concurrency control

The system SHALL use version numbers for optimistic locking to prevent lost updates from concurrent modifications.

#### Scenario: Successful update with correct version
- **WHEN** a tenant is updated with the current version number
- **THEN** the update succeeds
- **AND** the version number is incremented by 1
- **AND** the updated_at timestamp is set to current time

#### Scenario: Concurrent update detection
- **WHEN** a tenant is updated with an outdated version number
- **THEN** the operation fails with ErrVersionConflict
- **AND** no changes are applied to the tenant
- **AND** the caller must reload and retry

#### Scenario: Optimistic lock retry pattern
- **WHEN** a version conflict occurs
- **THEN** the caller reloads the latest tenant state
- **AND** reapplies the changes to the fresh copy
- **AND** retries the update with the new version

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

### Requirement: State transitions are recorded in immutable audit log

The system SHALL record all tenant status transitions in an append-only audit log.

#### Scenario: Record state transition
- **WHEN** a tenant status changes from one state to another
- **THEN** a state transition record is created
- **AND** the record includes from_status, to_status, reason, and triggered_by
- **AND** the record is timestamped with created_at

#### Scenario: Retrieve tenant state history
- **WHEN** state history is requested for a tenant
- **THEN** all transition records are returned
- **AND** the records are ordered by created_at descending (newest first)
- **AND** empty history returns an empty slice without error

#### Scenario: State transitions are append-only
- **WHEN** state transitions are recorded
- **THEN** existing records are never modified or deleted
- **AND** duplicate transitions are allowed (idempotent)

#### Scenario: State transition includes optional snapshots
- **WHEN** a state transition is recorded with state snapshots
- **THEN** desired_state_snapshot and observed_state_snapshot are stored
- **AND** the snapshots preserve the tenant state at transition time
- **AND** snapshots aid in debugging and audit analysis

### Requirement: Repository interface supports filtering and pagination

The system SHALL provide filtering and pagination capabilities for listing tenants.

#### Scenario: Filter tenants by status
- **WHEN** tenants are listed with status filter
- **THEN** only tenants matching the specified statuses are returned
- **AND** multiple status values can be combined

#### Scenario: Filter tenants by creation time range
- **WHEN** tenants are listed with CreatedAfter and CreatedBefore filters
- **THEN** only tenants created within the time range are returned
- **AND** results exclude tenants outside the range

#### Scenario: Paginate tenant results
- **WHEN** tenants are listed with Limit and Offset parameters
- **THEN** the first Offset results are skipped
- **AND** at most Limit results are returned
- **AND** results enable pagination through large datasets

#### Scenario: List returns empty results
- **WHEN** no tenants match the filter criteria
- **THEN** an empty slice is returned
- **AND** this is not considered an error condition

### Requirement: Repository operations respect context cancellation

The system SHALL respect context cancellation and timeouts for all repository operations.

#### Scenario: Operation cancelled by context
- **WHEN** a repository operation is in progress and context is cancelled
- **THEN** the operation terminates promptly
- **AND** context.Canceled error is returned
- **AND** no partial state changes are committed

#### Scenario: Operation times out
- **WHEN** a repository operation exceeds context deadline
- **THEN** the operation is terminated
- **AND** context.DeadlineExceeded error is returned

### Requirement: Tenant configuration and metadata use flexible JSON storage

The system SHALL store tenant configuration and metadata as JSONB for schema flexibility.

#### Scenario: Store structured configuration
- **WHEN** a tenant has desired_config with nested JSON
- **THEN** the configuration is stored as JSONB
- **AND** the configuration can include arbitrary key-value pairs
- **AND** no schema migration is required for config changes

#### Scenario: Store labels and annotations
- **WHEN** a tenant has labels and annotations
- **THEN** labels are stored as JSONB for filtering and grouping
- **AND** annotations are stored as JSONB for metadata
- **AND** labels can be queried efficiently with GIN indexes

### Requirement: Repository implementation is thread-safe

The system SHALL provide thread-safe repository implementations for concurrent access.

#### Scenario: Concurrent repository operations
- **WHEN** multiple goroutines access the repository simultaneously
- **THEN** all operations complete without data races
- **AND** optimistic locking prevents lost updates
- **AND** connection pooling handles concurrent queries

### Requirement: Database schema uses UUID primary keys

The system SHALL use UUID primary keys for tenant and state transition records.

#### Scenario: Generate unique tenant IDs
- **WHEN** a new tenant is created
- **THEN** a UUID is generated for the internal id field
- **AND** the UUID is globally unique
- **AND** the UUID prevents enumeration attacks

### Requirement: Repository returns standard domain errors

The system SHALL return well-defined error types for common failure scenarios.

#### Scenario: Tenant not found
- **WHEN** a tenant lookup fails because the tenant doesn't exist
- **THEN** ErrTenantNotFound is returned
- **AND** the caller can distinguish this from other errors

#### Scenario: Tenant already exists
- **WHEN** tenant creation fails due to duplicate tenant name
- **THEN** ErrTenantExists is returned
- **AND** the caller knows the tenant name is taken

#### Scenario: Version conflict
- **WHEN** an update fails due to optimistic lock mismatch
- **THEN** ErrVersionConflict is returned
- **AND** the caller knows to reload and retry
