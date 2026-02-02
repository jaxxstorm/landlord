# Specification: PostgreSQL Repository Implementation

## Overview

This specification defines the PostgreSQL implementation of the Repository interface, including schema design, SQL queries, transaction handling, and optimistic locking.

## Database Schema

### Tenants Table

```sql
CREATE TABLE IF NOT EXISTS tenants (
    -- Identity
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) UNIQUE NOT NULL,
    
    -- Current lifecycle state
    status VARCHAR(50) NOT NULL,
    status_message TEXT,
    
    -- Desired state (what user wants)
    desired_image VARCHAR(500) NOT NULL,
    desired_config JSONB NOT NULL DEFAULT '{}',
    
    -- Observed state (what actually exists)
    observed_image VARCHAR(500),
    observed_config JSONB,
    observed_resource_ids JSONB,
    
    -- Metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    -- Versioning for optimistic locking
    version INTEGER NOT NULL DEFAULT 1,
    
    -- Additional metadata
    labels JSONB NOT NULL DEFAULT '{}',
    annotations JSONB NOT NULL DEFAULT '{}'
);

-- Indexes for common queries
CREATE INDEX idx_tenants_status ON tenants(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_tenants_created_at ON tenants(created_at);
CREATE INDEX idx_tenants_deleted_at ON tenants(deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_tenants_labels ON tenants USING GIN (labels);
```

**Design Decisions:**
- **UUID for id**: Prevents enumeration, globally unique
- **tenant_id UNIQUE**: Enforces business constraint
- **JSONB for flexible data**: Config and metadata can evolve without migrations
- **Partial indexes**: Only index non-deleted tenants for status queries
- **GIN index on labels**: Enables efficient label-based filtering

### Tenant State History Table

```sql
CREATE TABLE IF NOT EXISTS tenant_state_history (
    -- Identity
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Transition details
    from_status VARCHAR(50),
    to_status VARCHAR(50) NOT NULL,
    
    -- Context
    reason TEXT NOT NULL,
    triggered_by VARCHAR(255),
    
    -- State snapshots (optional, for detailed audit)
    desired_state_snapshot JSONB,
    observed_state_snapshot JSONB,
    
    -- Metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for audit queries
CREATE INDEX idx_tenant_history_tenant_id ON tenant_state_history(tenant_id);
CREATE INDEX idx_tenant_history_created_at ON tenant_state_history(created_at DESC);
CREATE INDEX idx_tenant_history_to_status ON tenant_state_history(to_status);
```

**Design Decisions:**
- **Append-only**: No updates or deletes (immutable audit log)
- **Foreign key with CASCADE**: Delete history when tenant hard-deleted
- **Descending index on created_at**: Optimize for "recent history" queries

## Repository Implementation

### Package Location
`internal/tenant/postgres/repository.go`

### Repository Struct

```go
package postgres

import (
    "context"
    "errors"
    "fmt"
    
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "go.uber.org/zap"
    
    "github.com/jaxxstorm/landlord/internal/tenant"
)

// Repository implements tenant.Repository for PostgreSQL
type Repository struct {
    pool   *pgxpool.Pool
    logger *zap.Logger
}

// New creates a PostgreSQL repository
func New(pool *pgxpool.Pool, logger *zap.Logger) *Repository {
    return &Repository{
        pool:   pool,
        logger: logger.With(zap.String("component", "tenant-postgres-repository")),
    }
}
```

## CRUD Operations

### CreateTenant

```go
const createTenantQuery = `
INSERT INTO tenants (
    tenant_id, status, status_message,
    desired_image, desired_config,
    labels, annotations
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING id, created_at, updated_at, version
`

func (r *Repository) CreateTenant(ctx context.Context, t *tenant.Tenant) error {
    r.logger.Debug("creating tenant",
        zap.String("tenant_id", t.TenantID),
        zap.String("status", string(t.Status)))
    
    row := r.pool.QueryRow(ctx, createTenantQuery,
        t.TenantID,
        t.Status,
        t.StatusMessage,
        t.DesiredImage,
        t.DesiredConfig,
        jsonbOrEmpty(t.Labels),
        jsonbOrEmpty(t.Annotations),
    )
    
    err := row.Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt, &t.Version)
    if err != nil {
        if isUniqueViolation(err) {
            return tenant.ErrTenantExists
        }
        return fmt.Errorf("create tenant: %w", err)
    }
    
    r.logger.Info("tenant created",
        zap.String("id", t.ID),
        zap.String("tenant_id", t.TenantID))
    
    return nil
}
```

**Key Points:**
- RETURNING clause populates generated fields
- Unique violation maps to ErrTenantExists
- Logs at DEBUG for operation, INFO for completion

### GetTenant

```go
const getTenantQuery = `
SELECT
    id, tenant_id, status, status_message,
    desired_image, desired_config,
    observed_image, observed_config, observed_resource_ids,
    created_at, updated_at, deleted_at,
    version, labels, annotations
FROM tenants
WHERE tenant_id = $1
`

func (r *Repository) GetTenant(ctx context.Context, tenantID string) (*tenant.Tenant, error) {
    r.logger.Debug("getting tenant", zap.String("tenant_id", tenantID))
    
    t := &tenant.Tenant{}
    err := r.pool.QueryRow(ctx, getTenantQuery, tenantID).Scan(
        &t.ID,
        &t.TenantID,
        &t.Status,
        &t.StatusMessage,
        &t.DesiredImage,
        &t.DesiredConfig,
        &t.ObservedImage,
        &t.ObservedConfig,
        &t.ObservedResourceIDs,
        &t.CreatedAt,
        &t.UpdatedAt,
        &t.DeletedAt,
        &t.Version,
        &t.Labels,
        &t.Annotations,
    )
    
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, tenant.ErrTenantNotFound
        }
        return nil, fmt.Errorf("get tenant: %w", err)
    }
    
    return t, nil
}
```

**Key Points:**
- ErrNoRows maps to ErrTenantNotFound
- Retrieves all fields including deleted_at
- Nullable fields use sql.Null* types or pointers

### UpdateTenant

```go
const updateTenantQuery = `
UPDATE tenants SET
    status = $2,
    status_message = $3,
    desired_image = $4,
    desired_config = $5,
    observed_image = $6,
    observed_config = $7,
    observed_resource_ids = $8,
    updated_at = NOW(),
    version = version + 1,
    labels = $9,
    annotations = $10
WHERE id = $1 AND version = $11
RETURNING version, updated_at
`

func (r *Repository) UpdateTenant(ctx context.Context, t *tenant.Tenant) error {
    r.logger.Debug("updating tenant",
        zap.String("id", t.ID),
        zap.Int("version", t.Version))
    
    row := r.pool.QueryRow(ctx, updateTenantQuery,
        t.ID,
        t.Status,
        t.StatusMessage,
        t.DesiredImage,
        t.DesiredConfig,
        t.ObservedImage,
        t.ObservedConfig,
        t.ObservedResourceIDs,
        jsonbOrEmpty(t.Labels),
        jsonbOrEmpty(t.Annotations),
        t.Version, // Optimistic locking check
    )
    
    err := row.Scan(&t.Version, &t.UpdatedAt)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            // Either tenant doesn't exist or version mismatch
            // Check which one
            _, getErr := r.GetTenantByID(ctx, t.ID)
            if getErr != nil {
                return tenant.ErrTenantNotFound
            }
            return tenant.ErrVersionConflict
        }
        return fmt.Errorf("update tenant: %w", err)
    }
    
    r.logger.Info("tenant updated",
        zap.String("id", t.ID),
        zap.Int("new_version", t.Version))
    
    return nil
}
```

**Key Points:**
- WHERE clause includes version for optimistic locking
- version auto-incremented in UPDATE
- ErrNoRows could mean not found OR version conflict - disambiguate with GET

### ListTenants

```go
func (r *Repository) ListTenants(ctx context.Context, filters tenant.ListFilters) ([]*tenant.Tenant, error) {
    query, args := r.buildListQuery(filters)
    
    r.logger.Debug("listing tenants", zap.Any("filters", filters))
    
    rows, err := r.pool.Query(ctx, query, args...)
    if err != nil {
        return nil, fmt.Errorf("list tenants: %w", err)
    }
    defer rows.Close()
    
    var tenants []*tenant.Tenant
    for rows.Next() {
        t := &tenant.Tenant{}
        err := rows.Scan(
            &t.ID, &t.TenantID, &t.Status, &t.StatusMessage,
            &t.DesiredImage, &t.DesiredConfig,
            &t.ObservedImage, &t.ObservedConfig, &t.ObservedResourceIDs,
            &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
            &t.Version, &t.Labels, &t.Annotations,
        )
        if err != nil {
            return nil, fmt.Errorf("scan tenant: %w", err)
        }
        tenants = append(tenants, t)
    }
    
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("iterate tenants: %w", err)
    }
    
    return tenants, nil
}

func (r *Repository) buildListQuery(filters tenant.ListFilters) (string, []interface{}) {
    query := `
        SELECT
            id, tenant_id, status, status_message,
            desired_image, desired_config,
            observed_image, observed_config, observed_resource_ids,
            created_at, updated_at, deleted_at,
            version, labels, annotations
        FROM tenants
        WHERE 1=1
    `
    args := []interface{}{}
    argPos := 1
    
    // Filter by status
    if len(filters.Status) > 0 {
        query += fmt.Sprintf(" AND status = ANY($%d)", argPos)
        args = append(args, filters.Status)
        argPos++
    }
    
    // Filter by created_at range
    if filters.CreatedAfter != nil {
        query += fmt.Sprintf(" AND created_at > $%d", argPos)
        args = append(args, *filters.CreatedAfter)
        argPos++
    }
    if filters.CreatedBefore != nil {
        query += fmt.Sprintf(" AND created_at < $%d", argPos)
        args = append(args, *filters.CreatedBefore)
        argPos++
    }
    
    // Filter deleted
    if !filters.IncludeDeleted {
        query += " AND deleted_at IS NULL"
    }
    
    // Order and pagination
    query += " ORDER BY created_at DESC"
    
    if filters.Limit > 0 {
        query += fmt.Sprintf(" LIMIT $%d", argPos)
        args = append(args, filters.Limit)
        argPos++
    }
    
    if filters.Offset > 0 {
        query += fmt.Sprintf(" OFFSET $%d", argPos)
        args = append(args, filters.Offset)
    }
    
    return query, args
}
```

**Key Points:**
- Dynamic query building based on filters
- Uses ANY for status array matching
- Ordered by created_at DESC (newest first)
- Supports pagination with LIMIT/OFFSET

### DeleteTenant

```go
const deleteTenantQuery = `
UPDATE tenants
SET deleted_at = NOW()
WHERE tenant_id = $1 AND deleted_at IS NULL
RETURNING id
`

func (r *Repository) DeleteTenant(ctx context.Context, tenantID string) error {
    r.logger.Debug("deleting tenant", zap.String("tenant_id", tenantID))
    
    var id string
    err := r.pool.QueryRow(ctx, deleteTenantQuery, tenantID).Scan(&id)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            // Either not found or already deleted
            // Check if exists
            t, getErr := r.GetTenant(ctx, tenantID)
            if getErr != nil {
                return tenant.ErrTenantNotFound
            }
            // Already deleted - idempotent, return success
            if t.DeletedAt != nil {
                return nil
            }
        }
        return fmt.Errorf("delete tenant: %w", err)
    }
    
    r.logger.Info("tenant deleted", zap.String("tenant_id", tenantID))
    return nil
}
```

**Key Points:**
- Soft delete (sets deleted_at)
- Idempotent: deleting already-deleted tenant succeeds
- WHERE clause prevents double-delete

## State History Operations

### RecordStateTransition

```go
const recordTransitionQuery = `
INSERT INTO tenant_state_history (
    tenant_id, from_status, to_status,
    reason, triggered_by,
    desired_state_snapshot, observed_state_snapshot
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING id, created_at
`

func (r *Repository) RecordStateTransition(ctx context.Context, st *tenant.StateTransition) error {
    r.logger.Debug("recording state transition",
        zap.String("tenant_id", st.TenantID),
        zap.String("to_status", string(st.ToStatus)))
    
    row := r.pool.QueryRow(ctx, recordTransitionQuery,
        st.TenantID,
        st.FromStatus,
        st.ToStatus,
        st.Reason,
        st.TriggeredBy,
        st.DesiredStateSnapshot,
        st.ObservedStateSnapshot,
    )
    
    err := row.Scan(&st.ID, &st.CreatedAt)
    if err != nil {
        return fmt.Errorf("record transition: %w", err)
    }
    
    return nil
}
```

**Key Points:**
- Append-only, never fails on duplicates
- Captures snapshots for detailed audit
- Foreign key ensures tenant exists

### GetStateHistory

```go
const getHistoryQuery = `
SELECT
    id, tenant_id, from_status, to_status,
    reason, triggered_by,
    desired_state_snapshot, observed_state_snapshot,
    created_at
FROM tenant_state_history
WHERE tenant_id = (SELECT id FROM tenants WHERE tenant_id = $1)
ORDER BY created_at DESC
`

func (r *Repository) GetStateHistory(ctx context.Context, tenantID string) ([]*tenant.StateTransition, error) {
    r.logger.Debug("getting state history", zap.String("tenant_id", tenantID))
    
    rows, err := r.pool.Query(ctx, getHistoryQuery, tenantID)
    if err != nil {
        return nil, fmt.Errorf("get history: %w", err)
    }
    defer rows.Close()
    
    var history []*tenant.StateTransition
    for rows.Next() {
        st := &tenant.StateTransition{}
        err := rows.Scan(
            &st.ID,
            &st.TenantID,
            &st.FromStatus,
            &st.ToStatus,
            &st.Reason,
            &st.TriggeredBy,
            &st.DesiredStateSnapshot,
            &st.ObservedStateSnapshot,
            &st.CreatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("scan transition: %w", err)
        }
        history = append(history, st)
    }
    
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("iterate history: %w", err)
    }
    
    return history, nil
}
```

**Key Points:**
- Ordered DESC (newest first)
- Subquery converts tenant_id to internal UUID
- Returns empty slice if no history (not an error)

## Helper Functions

### JSONB Handling

```go
// jsonbOrEmpty converts map to JSONB, returns empty object if nil
func jsonbOrEmpty(m map[string]string) interface{} {
    if m == nil || len(m) == 0 {
        return "{}"
    }
    return m
}

// isUniqueViolation checks if error is unique constraint violation
func isUniqueViolation(err error) bool {
    var pgErr *pgconn.PgError
    return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
```

## Migration Files

### Location
`internal/database/migrations/003_create_tenant_schema.up.sql`

```sql
-- Tenants table
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL,
    status_message TEXT,
    desired_image VARCHAR(500) NOT NULL,
    desired_config JSONB NOT NULL DEFAULT '{}',
    observed_image VARCHAR(500),
    observed_config JSONB,
    observed_resource_ids JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    version INTEGER NOT NULL DEFAULT 1,
    labels JSONB NOT NULL DEFAULT '{}',
    annotations JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX idx_tenants_status ON tenants(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_tenants_created_at ON tenants(created_at);
CREATE INDEX idx_tenants_deleted_at ON tenants(deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_tenants_labels ON tenants USING GIN (labels);

-- State history table
CREATE TABLE IF NOT EXISTS tenant_state_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    from_status VARCHAR(50),
    to_status VARCHAR(50) NOT NULL,
    reason TEXT NOT NULL,
    triggered_by VARCHAR(255),
    desired_state_snapshot JSONB,
    observed_state_snapshot JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tenant_history_tenant_id ON tenant_state_history(tenant_id);
CREATE INDEX idx_tenant_history_created_at ON tenant_state_history(created_at DESC);
CREATE INDEX idx_tenant_history_to_status ON tenant_state_history(to_status);
```

### Down Migration
`internal/database/migrations/003_create_tenant_schema.down.sql`

```sql
DROP TABLE IF EXISTS tenant_state_history;
DROP TABLE IF EXISTS tenants;
```

## Testing

### Test Setup

```go
func setupTestRepo(t *testing.T) (*Repository, func()) {
    pool := testcontainers.StartPostgres(t)
    
    // Run migrations
    migrationPath := "file://../../database/migrations"
    err := database.RunMigrations(pool.Config().ConnString(), zap.NewNop())
    require.NoError(t, err)
    
    repo := New(pool, zap.NewNop())
    
    cleanup := func() {
        pool.Close()
    }
    
    return repo, cleanup
}
```

### Test Cases

```go
func TestRepository_CreateTenant(t *testing.T) {
    repo, cleanup := setupTestRepo(t)
    defer cleanup()
    
    tenant := &tenant.Tenant{
        TenantID:      "test-tenant",
        Status:        tenant.StatusRequested,
        DesiredImage:  "app:v1.0.0",
        DesiredConfig: json.RawMessage(`{"replicas": 2}`),
    }
    
    err := repo.CreateTenant(context.Background(), tenant)
    assert.NoError(t, err)
    assert.NotEmpty(t, tenant.ID)
    assert.Equal(t, 1, tenant.Version)
    
    // Duplicate should fail
    err = repo.CreateTenant(context.Background(), tenant)
    assert.ErrorIs(t, err, tenant.ErrTenantExists)
}

func TestRepository_UpdateTenant_OptimisticLocking(t *testing.T) {
    repo, cleanup := setupTestRepo(t)
    defer cleanup()
    
    // Create tenant
    tenant := createTestTenant(t, repo)
    
    // Get two copies
    t1, _ := repo.GetTenant(context.Background(), tenant.TenantID)
    t2, _ := repo.GetTenant(context.Background(), tenant.TenantID)
    
    // Update first copy - succeeds
    t1.Status = tenant.StatusReady
    err := repo.UpdateTenant(context.Background(), t1)
    assert.NoError(t, err)
    assert.Equal(t, 2, t1.Version)
    
    // Update second copy - fails (stale version)
    t2.Status = tenant.StatusFailed
    err = repo.UpdateTenant(context.Background(), t2)
    assert.ErrorIs(t, err, tenant.ErrVersionConflict)
}
```
