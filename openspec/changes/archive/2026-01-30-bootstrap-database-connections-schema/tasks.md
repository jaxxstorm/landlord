# Tasks: Bootstrap Database Connections and Schema

## Group 1: Domain Types (6 tasks)

### 1.1 Create tenant package structure
**Files**: `internal/tenant/`

Create the package structure:
- `internal/tenant/`
- `internal/tenant/postgres/`

### 1.2 Define Status enumeration
**File**: `internal/tenant/tenant.go`

Implement Status type and constants:
- Status type as string
- All 8 status constants (requested, planning, provisioning, ready, updating, deleting, deleted, failed)
- ValidTransitions map
- Helper methods: IsValid(), IsTerminal(), IsHealthy(), CanTransition()

### 1.3 Define Tenant struct
**File**: `internal/tenant/tenant.go`

Implement Tenant type with all fields:
- Identity fields (ID, TenantID)
- Status fields (Status, StatusMessage)
- Desired state fields (DesiredImage, DesiredConfig)
- Observed state fields (ObservedImage, ObservedConfig, ObservedResourceIDs)
- Metadata fields (CreatedAt, UpdatedAt, DeletedAt)
- Version field for optimistic locking
- Labels and Annotations

### 1.4 Define StateTransition struct
**File**: `internal/tenant/tenant.go`

Implement StateTransition type:
- Identity and linkage (ID, TenantID)
- Transition details (FromStatus, ToStatus)
- Context (Reason, TriggeredBy)
- State snapshots (DesiredStateSnapshot, ObservedStateSnapshot)
- Metadata (CreatedAt)
- NewStateTransition helper function

### 1.5 Implement validation functions
**File**: `internal/tenant/tenant.go`

Implement validation:
- Tenant.Validate() method
- Status validation methods
- StateTransition.Validate() method
- tenantIDPattern regex

### 1.6 Implement tenant helper methods
**File**: `internal/tenant/tenant.go`

Implement helper methods:
- Tenant.IsDeleted()
- Tenant.IsDrifted()
- Tenant.Clone()

## Group 2: Repository Interface (2 tasks)

### 2.1 Define Repository interface
**File**: `internal/tenant/repository.go`

Define Repository interface with all methods:
- CreateTenant()
- GetTenant()
- GetTenantByID()
- UpdateTenant()
- ListTenants()
- DeleteTenant()
- RecordStateTransition()
- GetStateHistory()

### 2.2 Define error types and filters
**File**: `internal/tenant/repository.go`

Define supporting types:
- Error variables (ErrTenantNotFound, ErrTenantExists, ErrVersionConflict)
- ListFilters struct with all fields

## Group 3: Database Schema (2 tasks)

### 3.1 Create tenants table migration
**File**: `internal/database/migrations/003_create_tenant_schema.up.sql`

Create migration for tenants table:
- All columns as specified
- Primary key and unique constraints
- Indexes (status, created_at, deleted_at, labels GIN)

### 3.2 Create tenant_state_history table migration
**File**: `internal/database/migrations/003_create_tenant_schema.up.sql`

Add to migration:
- tenant_state_history table
- Foreign key reference to tenants
- Indexes (tenant_id, created_at DESC, to_status)
- Down migration file (003_create_tenant_schema.down.sql)

## Group 4: PostgreSQL Repository Implementation (8 tasks)

### 4.1 Create PostgreSQL repository struct
**File**: `internal/tenant/postgres/repository.go`

Implement:
- Repository struct with pool and logger
- New() constructor function
- Package imports

### 4.2 Implement CreateTenant
**File**: `internal/tenant/postgres/repository.go`

Implement CreateTenant method:
- INSERT query with RETURNING clause
- Map unique violation to ErrTenantExists
- Proper logging
- Return populated tenant with ID, timestamps, version

### 4.3 Implement GetTenant and GetTenantByID
**File**: `internal/tenant/postgres/repository.go`

Implement both getter methods:
- SELECT query with all fields
- Map ErrNoRows to ErrTenantNotFound
- Proper NULL handling for optional fields

### 4.4 Implement UpdateTenant
**File**: `internal/tenant/postgres/repository.go`

Implement UpdateTenant with optimistic locking:
- UPDATE query with version in WHERE clause
- Auto-increment version
- Disambiguate ErrNoRows (not found vs version conflict)
- Return updated version and timestamp

### 4.5 Implement ListTenants
**File**: `internal/tenant/postgres/repository.go`

Implement ListTenants with filtering:
- Dynamic query building (buildListQuery helper)
- Status array filtering with ANY
- Date range filtering
- Soft delete filtering
- Pagination (LIMIT/OFFSET)
- Order by created_at DESC

### 4.6 Implement DeleteTenant
**File**: `internal/tenant/postgres/repository.go`

Implement soft delete:
- UPDATE to set deleted_at
- Idempotent (already-deleted succeeds)
- Map ErrNoRows to ErrTenantNotFound

### 4.7 Implement RecordStateTransition
**File**: `internal/tenant/postgres/repository.go`

Implement audit log append:
- INSERT query for tenant_state_history
- Capture snapshots
- Return generated ID and timestamp

### 4.8 Implement GetStateHistory
**File**: `internal/tenant/postgres/repository.go`

Implement history retrieval:
- SELECT from tenant_state_history
- Order by created_at DESC
- Subquery to convert tenant_id to UUID
- Return empty slice if no history

## Group 5: Helper Functions and Testing (4 tasks)

### 5.1 Implement helper functions
**File**: `internal/tenant/postgres/repository.go`

Implement utility functions:
- jsonbOrEmpty() for nil map handling
- isUniqueViolation() for error checking

### 5.2 Write tenant domain tests
**File**: `internal/tenant/tenant_test.go`

Test domain types:
- Status validation and transitions
- Tenant.Validate()
- StateTransition.Validate()
- Helper methods (IsDeleted, IsDrifted, Clone)

### 5.3 Write repository interface tests
**File**: `internal/tenant/postgres/repository_test.go`

Test PostgreSQL repository:
- CreateTenant (success and duplicate)
- GetTenant (success and not found)
- UpdateTenant with optimistic locking
- ListTenants with various filters
- DeleteTenant (idempotency)
- RecordStateTransition
- GetStateHistory ordering
- Concurrent updates (version conflicts)

### 5.4 Setup test database infrastructure
**File**: `internal/tenant/postgres/repository_test.go`

Implement test helpers:
- setupTestRepo() using testcontainers
- Run migrations in tests
- Cleanup functions
- Helper functions for creating test data

## Group 6: Integration (2 tasks)

### 6.1 Wire repository into main.go
**File**: `cmd/landlord/main.go`

Initialize tenant repository:
- Create PostgreSQL repository
- Pass to future tenant service
- Log initialization

### 6.2 Add tenant configuration
**File**: `internal/config/config.go`

Add configuration:
- TenantConfig struct
- Repository type field (default "postgres")

## Acceptance Criteria

- [x] Status enumeration with 8 states defined
- [x] Tenant type with all fields (desired/observed state)
- [x] StateTransition audit type defined
- [x] Repository interface with 8 methods defined
- [x] PostgreSQL implementation complete
- [x] Database migrations create both tables
- [x] Optimistic locking works (version conflicts detected)
- [x] Soft deletes work correctly
- [x] State history is append-only and ordered
- [x] ListTenants filtering works (status, dates, deleted)
- [x] All error types defined and used correctly
- [x] Repository tests pass with testcontainers
- [x] Domain validation tests pass
- [x] `go test -short ./...` passes
- [x] Migrations run successfully

## Notes

- No tenant CRUD API endpoints in this change (deferred to next change)
- No tenant service layer yet (repository only)
- Focus on persistence layer foundation
- Test using testcontainers for real PostgreSQL
- Ensure migrations are reversible
