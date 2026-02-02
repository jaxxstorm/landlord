# Design: Database Connections and Tenant Schema

## Architecture Overview

```
┌─────────────────────────────────────────┐
│         Domain Layer                     │
│  (Tenant business logic, state machine) │
└─────────────────┬───────────────────────┘
                  │
                  ↓
┌─────────────────────────────────────────┐
│      Repository Interface (Port)         │
│  - CreateTenant()                        │
│  - GetTenant()                           │
│  - UpdateTenantState()                   │
│  - ListTenants()                         │
│  - RecordStateTransition()               │
└─────────────────┬───────────────────────┘
                  │
                  ↓
┌─────────────────────────────────────────┐
│    PostgreSQL Adapter (Adapter)         │
│  - Implements Repository interface       │
│  - SQL queries and transactions          │
│  - Connection pooling via pgxpool        │
└─────────────────┬───────────────────────┘
                  │
                  ↓
┌─────────────────────────────────────────┐
│           PostgreSQL Database            │
│  - tenants table (current state)         │
│  - tenant_state_history (audit log)      │
└─────────────────────────────────────────┘
```

**Key Principle**: The domain layer never imports database-specific packages. It depends on the Repository interface, not the concrete implementation.

## Database Schema

### Table: `tenants`

Primary table storing current tenant state.

```sql
CREATE TABLE tenants (
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

CREATE INDEX idx_tenants_status ON tenants(status);
CREATE INDEX idx_tenants_created_at ON tenants(created_at);
CREATE INDEX idx_tenants_deleted_at ON tenants(deleted_at) WHERE deleted_at IS NOT NULL;
```

**Design Decisions**:
- **UUID id + tenant_id**: Internal UUID for database relations, tenant_id for user-facing identifier
- **JSONB for config**: Flexible schema for tenant-specific configuration
- **observed_resource_ids**: Stores provider-specific IDs (ECS task ARN, etc.)
- **version field**: Enables optimistic locking to prevent concurrent update conflicts
- **Soft deletes**: deleted_at allows retaining history of deleted tenants

### Table: `tenant_state_history`

Audit log of all state transitions.

```sql
CREATE TABLE tenant_state_history (
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

CREATE INDEX idx_tenant_history_tenant_id ON tenant_state_history(tenant_id);
CREATE INDEX idx_tenant_history_created_at ON tenant_state_history(created_at);
CREATE INDEX idx_tenant_history_to_status ON tenant_state_history(to_status);
```

**Design Decisions**:
- **Immutable append-only log**: Never update or delete history records
- **from_status nullable**: Initial creation has no previous state
- **reason required**: Every transition must be explained
- **State snapshots**: Captures full context at transition time for debugging

### Table: `tenant_events` (Future)

Optional event log for detailed observability.

```sql
CREATE TABLE tenant_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    event_type VARCHAR(100) NOT NULL,
    event_data JSONB NOT NULL,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tenant_events_tenant_id ON tenant_events(tenant_id);
CREATE INDEX idx_tenant_events_created_at ON tenant_events(created_at);
CREATE INDEX idx_tenant_events_type ON tenant_events(event_type);
```

**Use Cases**:
- Workflow execution events
- Reconciliation attempts
- Health check results
- Resource provisioning steps

*Note: Implementation deferred to future change focused on observability.*

## Repository Interface

```go
package tenant

import (
    "context"
    "time"
)

// Repository defines the port for tenant persistence
type Repository interface {
    // CreateTenant creates a new tenant with desired state
    CreateTenant(ctx context.Context, tenant *Tenant) error
    
    // GetTenant retrieves a tenant by tenant_id
    GetTenant(ctx context.Context, tenantID string) (*Tenant, error)
    
    // GetTenantByID retrieves a tenant by internal UUID
    GetTenantByID(ctx context.Context, id string) (*Tenant, error)
    
    // UpdateTenant updates tenant state with optimistic locking
    UpdateTenant(ctx context.Context, tenant *Tenant) error
    
    // ListTenants returns all tenants matching filters
    ListTenants(ctx context.Context, filters ListFilters) ([]*Tenant, error)
    
    // DeleteTenant performs soft delete
    DeleteTenant(ctx context.Context, tenantID string) error
    
    // RecordStateTransition appends to audit log
    RecordStateTransition(ctx context.Context, transition *StateTransition) error
    
    // GetStateHistory returns audit log for a tenant
    GetStateHistory(ctx context.Context, tenantID string) ([]*StateTransition, error)
}

// ListFilters defines query filters
type ListFilters struct {
    Status      []string
    CreatedAfter *time.Time
    CreatedBefore *time.Time
    IncludeDeleted bool
    Limit       int
    Offset      int
}
```

**Design Decisions**:
- **Context-first**: All methods accept context for cancellation/timeout
- **Pointer returns**: Avoid copying large structs
- **Optimistic locking**: UpdateTenant checks version field
- **Soft delete**: DeleteTenant sets deleted_at, doesn't remove row

## Domain Types

### Tenant

```go
package tenant

import (
    "encoding/json"
    "time"
)

// Tenant represents a logical tenant instance
type Tenant struct {
    // Identity
    ID       string `json:"id"`        // Internal UUID
    TenantID string `json:"tenant_id"` // User-facing identifier
    
    // Current state
    Status        Status `json:"status"`
    StatusMessage string `json:"status_message,omitempty"`
    
    // Desired state (declarative)
    DesiredImage  string          `json:"desired_image"`
    DesiredConfig json.RawMessage `json:"desired_config"`
    
    // Observed state (actual)
    ObservedImage       string          `json:"observed_image,omitempty"`
    ObservedConfig      json.RawMessage `json:"observed_config,omitempty"`
    ObservedResourceIDs json.RawMessage `json:"observed_resource_ids,omitempty"`
    
    // Metadata
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
    DeletedAt *time.Time `json:"deleted_at,omitempty"`
    
    // Concurrency control
    Version int `json:"version"`
    
    // Labels and annotations
    Labels      map[string]string `json:"labels,omitempty"`
    Annotations map[string]string `json:"annotations,omitempty"`
}

// Status represents tenant lifecycle state
type Status string

const (
    StatusRequested     Status = "requested"
    StatusPlanning      Status = "planning"
    StatusProvisioning  Status = "provisioning"
    StatusReady         Status = "ready"
    StatusUpdating      Status = "updating"
    StatusDeleting      Status = "deleting"
    StatusDeleted       Status = "deleted"
    StatusFailed        Status = "failed"
)

// StateTransition represents an audit log entry
type StateTransition struct {
    ID         string    `json:"id"`
    TenantID   string    `json:"tenant_id"`
    FromStatus *Status   `json:"from_status,omitempty"`
    ToStatus   Status    `json:"to_status"`
    Reason     string    `json:"reason"`
    TriggeredBy string   `json:"triggered_by,omitempty"`
    CreatedAt  time.Time `json:"created_at"`
    
    // Optional snapshots
    DesiredStateSnapshot  json.RawMessage `json:"desired_state_snapshot,omitempty"`
    ObservedStateSnapshot json.RawMessage `json:"observed_state_snapshot,omitempty"`
}
```

## PostgreSQL Adapter Implementation

### Structure

```go
package postgres

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
    "go.uber.org/zap"
)

// Repository implements tenant.Repository for PostgreSQL
type Repository struct {
    pool   *pgxpool.Pool
    logger *zap.Logger
}

// New creates a new PostgreSQL repository
func New(pool *pgxpool.Pool, logger *zap.Logger) *Repository {
    return &Repository{
        pool:   pool,
        logger: logger.With(zap.String("component", "postgres-repository")),
    }
}
```

### Key Operations

#### CreateTenant
- INSERT into tenants table
- Set version = 1
- Return error if tenant_id already exists (unique constraint)

#### UpdateTenant
- Use optimistic locking: `WHERE id = $1 AND version = $2`
- Increment version: `SET version = version + 1`
- If no rows affected, return ErrVersionConflict
- Update updated_at automatically

#### RecordStateTransition
- INSERT into tenant_state_history
- Always succeeds (append-only)
- Capture snapshots of desired/observed state

### Transaction Support

```go
// WithTx executes a function within a transaction
func (r *Repository) WithTx(ctx context.Context, fn func(*TxRepository) error) error {
    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)
    
    txRepo := &TxRepository{tx: tx, logger: r.logger}
    if err := fn(txRepo); err != nil {
        return err
    }
    
    return tx.Commit(ctx)
}
```

**Use Case**: Atomic tenant creation + state transition record.

## Error Handling

```go
package tenant

import "errors"

var (
    ErrTenantNotFound    = errors.New("tenant not found")
    ErrTenantExists      = errors.New("tenant already exists")
    ErrVersionConflict   = errors.New("tenant version conflict")
    ErrInvalidStatus     = errors.New("invalid status transition")
)
```

## Migration Strategy

### Migration File: `003_create_tenant_schema.up.sql`

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

CREATE INDEX idx_tenants_status ON tenants(status);
CREATE INDEX idx_tenants_created_at ON tenants(created_at);
CREATE INDEX idx_tenants_deleted_at ON tenants(deleted_at) WHERE deleted_at IS NOT NULL;

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
CREATE INDEX idx_tenant_history_created_at ON tenant_state_history(created_at);
CREATE INDEX idx_tenant_history_to_status ON tenant_state_history(to_status);
```

### Migration File: `003_create_tenant_schema.down.sql`

```sql
DROP TABLE IF EXISTS tenant_state_history;
DROP TABLE IF EXISTS tenants;
```

## Testing Strategy

### Unit Tests
- Repository interface mocked for domain logic testing
- PostgreSQL adapter tested against real database (testcontainers)

### Integration Tests
- Full CRUD lifecycle
- Concurrent updates (version conflict handling)
- State transition recording
- Query filtering

### Test Database Setup
```go
func setupTestDB(t *testing.T) *pgxpool.Pool {
    // Use testcontainers to spin up PostgreSQL
    // Run migrations
    // Return pool
}
```

## Configuration

Add to `internal/config/config.go`:

```go
type TenantConfig struct {
    Repository string `env:"TENANT_REPOSITORY" default:"postgres"`
}
```

## Wiring in main.go

```go
// Initialize tenant repository
var tenantRepo tenant.Repository
switch cfg.Tenant.Repository {
case "postgres":
    tenantRepo = postgres.New(db, log)
default:
    return fmt.Errorf("unknown repository type: %s", cfg.Tenant.Repository)
}

// Future: pass tenantRepo to tenant service/handler
```

## Future Considerations

1. **Read replicas**: Separate read/write connections for scaling
2. **Partitioning**: Partition tenant_state_history by time for performance
3. **Archival**: Move old history records to cold storage
4. **Caching**: Add Redis cache for frequently accessed tenants
5. **Multi-tenancy**: Schema per tenant vs shared schema with tenant_id filtering

These are explicitly deferred to maintain focus on MVP functionality.
