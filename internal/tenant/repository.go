package tenant

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrTenantNotFound is returned when a tenant doesn't exist
	ErrTenantNotFound = errors.New("tenant not found")

	// ErrTenantExists is returned when trying to create a tenant with a duplicate name
	ErrTenantExists = errors.New("tenant already exists")

	// ErrVersionConflict is returned when an optimistic locking conflict occurs
	ErrVersionConflict = errors.New("version conflict: tenant was modified by another operation")
)

// ListFilters contains optional filters for listing tenants
type ListFilters struct {
	// Status filtering
	Statuses []Status // If empty, match all statuses

	// Time range filtering
	CreatedAfter  *time.Time // If nil, no lower bound
	CreatedBefore *time.Time // If nil, no upper bound

	// Pagination
	Limit  int // Maximum number of results (0 = no limit)
	Offset int // Number of results to skip

	// IncludeDeleted includes archived tenants in results when true
	IncludeDeleted bool

	// Label filtering (optional future enhancement)
	Labels map[string]string // Match all specified labels
}

// Repository defines the persistence layer for tenant resources
type Repository interface {
	// CreateTenant persists a new tenant
	// Returns ErrTenantExists if name already exists
	// Populates ID, CreatedAt, UpdatedAt, and Version fields
	CreateTenant(ctx context.Context, tenant *Tenant) error

	// GetTenantByName retrieves a tenant by business identifier (name)
	// Returns ErrTenantNotFound if not found
	GetTenantByName(ctx context.Context, name string) (*Tenant, error)

	// GetTenantByID retrieves a tenant by database primary key (id)
	// Returns ErrTenantNotFound if not found
	GetTenantByID(ctx context.Context, id uuid.UUID) (*Tenant, error)

	// UpdateTenant modifies an existing tenant using optimistic locking
	// Returns ErrTenantNotFound if not found
	// Returns ErrVersionConflict if version doesn't match (concurrent modification)
	// Updates UpdatedAt and increments Version
	UpdateTenant(ctx context.Context, tenant *Tenant) error

	// ListTenants retrieves multiple tenants with optional filtering
	// Returns empty slice if no matches, never returns error for no results
	ListTenants(ctx context.Context, filters ListFilters) ([]*Tenant, error)

	// ListTenantsForReconciliation retrieves tenants in non-terminal states requiring reconciliation
	// Specifically returns tenants with status: requested, planning, provisioning, updating, or deleting
	// Returns empty slice if no tenants need reconciliation
	ListTenantsForReconciliation(ctx context.Context) ([]*Tenant, error)

	// DeleteTenant permanently removes a tenant record
	// Returns ErrTenantNotFound if tenant never existed
	DeleteTenant(ctx context.Context, id uuid.UUID) error

	// RecordStateTransition appends an audit record to the state history
	// Populates ID and CreatedAt fields
	RecordStateTransition(ctx context.Context, transition *StateTransition) error

	// GetStateHistory retrieves all state transitions for a tenant, newest first
	// Returns empty slice if no history, never returns error for empty results
	GetStateHistory(ctx context.Context, tenantID uuid.UUID) ([]*StateTransition, error)
}
