# Specification: Repository Interface

## Overview

This specification defines the Repository interface (port) for tenant persistence. This is the boundary between domain logic and database implementation, enabling pluggable storage backends while maintaining clean architecture principles.

## Interface Definition

### Location
`internal/tenant/repository.go`

### Repository Interface

```go
package tenant

import (
    "context"
    "time"
)

// Repository defines the port for tenant persistence operations
// Implementations must be thread-safe and support context cancellation
type Repository interface {
    // CreateTenant persists a new tenant with initial desired state
    // Returns ErrTenantExists if tenant_id already exists
    CreateTenant(ctx context.Context, tenant *Tenant) error
    
    // GetTenant retrieves a tenant by user-facing tenant_id
    // Returns ErrTenantNotFound if tenant doesn't exist
    GetTenant(ctx context.Context, tenantID string) (*Tenant, error)
    
    // GetTenantByID retrieves a tenant by internal database ID (UUID)
    // Returns ErrTenantNotFound if tenant doesn't exist
    GetTenantByID(ctx context.Context, id string) (*Tenant, error)
    
    // UpdateTenant updates tenant state using optimistic locking
    // Returns ErrVersionConflict if version doesn't match (concurrent update detected)
    // Returns ErrTenantNotFound if tenant doesn't exist
    UpdateTenant(ctx context.Context, tenant *Tenant) error
    
    // ListTenants returns tenants matching the provided filters
    // Returns empty slice if no matches (not an error)
    ListTenants(ctx context.Context, filters ListFilters) ([]*Tenant, error)
    
    // DeleteTenant performs soft delete by setting deleted_at timestamp
    // Returns ErrTenantNotFound if tenant doesn't exist
    // Idempotent: deleting an already-deleted tenant succeeds
    DeleteTenant(ctx context.Context, tenantID string) error
    
    // RecordStateTransition appends a state transition to audit log
    // Never returns error for duplicate transitions (append-only, idempotent)
    RecordStateTransition(ctx context.Context, transition *StateTransition) error
    
    // GetStateHistory returns audit log entries for a tenant
    // Ordered by created_at DESC (newest first)
    // Returns empty slice if no history (not an error)
    GetStateHistory(ctx context.Context, tenantID string) ([]*StateTransition, error)
}
```

## Filter Types

### ListFilters

```go
// ListFilters defines query parameters for listing tenants
type ListFilters struct {
    // Status filters by lifecycle state (empty = all statuses)
    Status []Status
    
    // CreatedAfter filters tenants created after this time
    CreatedAfter *time.Time
    
    // CreatedBefore filters tenants created before this time
    CreatedBefore *time.Time
    
    // IncludeDeleted includes soft-deleted tenants in results
    IncludeDeleted bool
    
    // Limit caps the number of results returned (0 = no limit)
    Limit int
    
    // Offset skips the first N results (for pagination)
    Offset int
}
```

## Method Contracts

### CreateTenant

**Preconditions:**
- `tenant.TenantID` must be non-empty
- `tenant.DesiredImage` must be non-empty
- `tenant.Status` must be valid Status constant

**Postconditions:**
- Tenant is persisted with generated UUID `id`
- `tenant.Version` is set to 1
- `tenant.CreatedAt` and `tenant.UpdatedAt` are set to current time
- Tenant is retrievable via GetTenant()

**Error Cases:**
- `ErrTenantExists`: tenant_id already exists in database
- Context error: timeout or cancellation

**Idempotency:** Not idempotent. Repeated calls with same tenant_id return ErrTenantExists.

### GetTenant / GetTenantByID

**Preconditions:**
- `tenantID` or `id` must be non-empty

**Postconditions:**
- Returns fully populated Tenant struct
- Includes all fields (desired state, observed state, metadata)

**Error Cases:**
- `ErrTenantNotFound`: tenant doesn't exist
- Context error: timeout or cancellation

**Idempotency:** Read-only, inherently idempotent.

### UpdateTenant

**Preconditions:**
- `tenant.ID` must match existing tenant
- `tenant.Version` must match current database version
- All required fields must be present

**Postconditions:**
- Tenant state updated in database
- `tenant.Version` incremented by 1
- `tenant.UpdatedAt` set to current time
- Tenant struct updated with new version

**Error Cases:**
- `ErrVersionConflict`: version mismatch (concurrent update detected)
- `ErrTenantNotFound`: tenant ID doesn't exist
- Context error: timeout or cancellation

**Idempotency:** Not idempotent due to version increment. Retry after conflict requires reloading tenant.

**Optimistic Locking Flow:**
```go
// Caller must handle conflicts
tenant, err := repo.GetTenant(ctx, tenantID)
if err != nil {
    return err
}

tenant.Status = StatusReady
tenant.ObservedImage = "app:v1.0.0"

err = repo.UpdateTenant(ctx, tenant)
if errors.Is(err, ErrVersionConflict) {
    // Another process updated this tenant
    // Must reload and retry
    return handleConflict(ctx, tenantID)
}
```

### ListTenants

**Preconditions:**
- Filters must be valid (e.g., Limit >= 0)

**Postconditions:**
- Returns slice of tenants matching filters
- Results ordered by created_at DESC (newest first)
- Pagination applied if Limit/Offset specified

**Error Cases:**
- Context error: timeout or cancellation
- Empty results are not an error (returns empty slice)

**Idempotency:** Read-only, inherently idempotent.

### DeleteTenant

**Preconditions:**
- `tenantID` must be non-empty

**Postconditions:**
- Tenant's `deleted_at` set to current time
- Tenant still exists in database (soft delete)
- Tenant excluded from ListTenants() unless IncludeDeleted=true

**Error Cases:**
- `ErrTenantNotFound`: tenant doesn't exist
- Context error: timeout or cancellation

**Idempotency:** Idempotent. Deleting an already-deleted tenant succeeds (no-op).

### RecordStateTransition

**Preconditions:**
- `transition.TenantID` must reference existing tenant
- `transition.ToStatus` must be non-empty
- `transition.Reason` must be non-empty

**Postconditions:**
- Audit log entry persisted
- Entry has generated UUID `id`
- Entry has `created_at` timestamp

**Error Cases:**
- Foreign key violation if tenant_id doesn't exist
- Context error: timeout or cancellation

**Idempotency:** Append-only, inherently idempotent (duplicates are acceptable).

### GetStateHistory

**Preconditions:**
- `tenantID` must be non-empty

**Postconditions:**
- Returns slice of state transitions ordered by created_at DESC
- Includes all transitions for the tenant

**Error Cases:**
- Context error: timeout or cancellation
- Empty history is not an error (returns empty slice)

**Idempotency:** Read-only, inherently idempotent.

## Error Types

### Standard Errors

```go
package tenant

import "errors"

var (
    // ErrTenantNotFound indicates tenant doesn't exist
    ErrTenantNotFound = errors.New("tenant not found")
    
    // ErrTenantExists indicates tenant_id already exists
    ErrTenantExists = errors.New("tenant already exists")
    
    // ErrVersionConflict indicates concurrent update detected
    ErrVersionConflict = errors.New("tenant version conflict")
)
```

## Implementation Requirements

### Thread Safety
All Repository implementations MUST be safe for concurrent use by multiple goroutines.

### Context Handling
All methods MUST respect context cancellation and timeouts:
```go
select {
case <-ctx.Done():
    return ctx.Err()
default:
    // continue with operation
}
```

### Transaction Support (Optional)
Implementations MAY provide transaction support for atomic operations:
```go
type TxRepository interface {
    Repository
    
    // WithTx executes fn within a transaction
    // Commits if fn returns nil, rolls back otherwise
    WithTx(ctx context.Context, fn func(Repository) error) error
}
```

## Testing Requirements

### Interface Compliance Tests
All Repository implementations must pass a standard test suite:

```go
func TestRepositoryCompliance(t *testing.T, newRepo func() Repository) {
    // Test CreateTenant
    // Test GetTenant
    // Test UpdateTenant with optimistic locking
    // Test ListTenants with filters
    // Test DeleteTenant idempotency
    // Test RecordStateTransition
    // Test GetStateHistory ordering
    // Test concurrent updates (version conflicts)
}
```

### Mock Implementation
A mock repository must be provided for domain logic testing:

```go
type MockRepository struct {
    tenants map[string]*Tenant
    history map[string][]*StateTransition
    mu      sync.RWMutex
}
```

## Examples

### Create and Update Tenant

```go
// Create new tenant
tenant := &Tenant{
    TenantID:      "acme-corp",
    Status:        StatusRequested,
    DesiredImage:  "app:v1.0.0",
    DesiredConfig: json.RawMessage(`{"replicas": 2}`),
}

if err := repo.CreateTenant(ctx, tenant); err != nil {
    return fmt.Errorf("create failed: %w", err)
}

// Record initial state transition
transition := &StateTransition{
    TenantID:   tenant.ID,
    ToStatus:   StatusRequested,
    Reason:     "Tenant creation requested by user",
    TriggeredBy: "user@example.com",
}
if err := repo.RecordStateTransition(ctx, transition); err != nil {
    log.Error("failed to record transition", zap.Error(err))
}

// Later: update to ready state
tenant, err := repo.GetTenant(ctx, "acme-corp")
if err != nil {
    return err
}

tenant.Status = StatusReady
tenant.ObservedImage = "app:v1.0.0"
tenant.ObservedResourceIDs = json.RawMessage(`{"task_arn": "arn:aws:ecs:..."}`)

if err := repo.UpdateTenant(ctx, tenant); err != nil {
    return fmt.Errorf("update failed: %w", err)
}
```

### List Tenants with Filtering

```go
// Get all ready tenants from last 24 hours
since := time.Now().Add(-24 * time.Hour)
filters := ListFilters{
    Status:       []Status{StatusReady},
    CreatedAfter: &since,
    Limit:        100,
}

tenants, err := repo.ListTenants(ctx, filters)
if err != nil {
    return fmt.Errorf("list failed: %w", err)
}

for _, tenant := range tenants {
    fmt.Printf("Tenant %s is ready\n", tenant.TenantID)
}
```

### Handle Version Conflicts

```go
func updateTenantWithRetry(ctx context.Context, repo Repository, tenantID string) error {
    maxRetries := 3
    for i := 0; i < maxRetries; i++ {
        tenant, err := repo.GetTenant(ctx, tenantID)
        if err != nil {
            return err
        }
        
        tenant.Status = StatusUpdating
        
        err = repo.UpdateTenant(ctx, tenant)
        if err == nil {
            return nil // Success
        }
        
        if !errors.Is(err, ErrVersionConflict) {
            return err // Permanent error
        }
        
        // Conflict - retry
        log.Warn("version conflict, retrying", zap.Int("attempt", i+1))
    }
    
    return fmt.Errorf("update failed after %d retries", maxRetries)
}
```
