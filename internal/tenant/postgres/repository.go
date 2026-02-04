package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
// Accepts interface{} to satisfy provider abstraction, type asserts to *pgxpool.Pool
func New(pool interface{}, logger *zap.Logger) (*Repository, error) {
	pgPool, ok := pool.(*pgxpool.Pool)
	if !ok {
		return nil, fmt.Errorf("expected *pgxpool.Pool, got %T", pool)
	}
	return &Repository{
		pool:   pgPool,
		logger: logger.With(zap.String("component", "tenant-postgres-repository")),
	}, nil
}

const createTenantQuery = `
INSERT INTO tenants (
    id, name, status, status_message,
    desired_config,
    labels, annotations, workflow_config_hash
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING created_at, updated_at, version
`

func (r *Repository) CreateTenant(ctx context.Context, t *tenant.Tenant) error {
	// Generate UUID for ID if not already set
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}

	r.logger.Debug("creating tenant",
		zap.String("name", t.Name),
		zap.String("id", t.ID.String()),
		zap.String("status", string(t.Status)))

	row := r.pool.QueryRow(ctx, createTenantQuery,
		t.ID.String(),
		t.Name,
		t.Status,
		t.StatusMessage,
		jsonbOrEmptyInterfaceMap(t.DesiredConfig),
		jsonbOrEmptyStringMap(t.Labels),
		jsonbOrEmptyStringMap(t.Annotations),
		t.WorkflowConfigHash,
	)

	err := row.Scan(&t.CreatedAt, &t.UpdatedAt, &t.Version)
	if err != nil {
		if isUniqueViolation(err) {
			return tenant.ErrTenantExists
		}
		return fmt.Errorf("create tenant: %w", err)
	}

	r.logger.Info("tenant created",
		zap.String("id", t.ID.String()),
		zap.String("name", t.Name))

	return nil
}

const getTenantQuery = `
SELECT
    id, name, status, status_message,
    desired_config,
    observed_config, observed_resource_ids,
    created_at, updated_at,
	version, labels, annotations, workflow_execution_id,
	workflow_sub_state, workflow_retry_count, workflow_error_message,
	workflow_config_hash
FROM tenants
WHERE name = $1
`

func (r *Repository) GetTenantByName(ctx context.Context, name string) (*tenant.Tenant, error) {
	r.logger.Debug("getting tenant", zap.String("name", name))

	t := &tenant.Tenant{}
	var desiredConfigJSON, observedConfigJSON, observedResourceIDsJSON, labelsJSON, annotationsJSON []byte

	err := r.pool.QueryRow(ctx, getTenantQuery, name).Scan(
		&t.ID,
		&t.Name,
		&t.Status,
		&t.StatusMessage,
		&desiredConfigJSON,
		&observedConfigJSON,
		&observedResourceIDsJSON,
		&t.CreatedAt,
		&t.UpdatedAt,
		&t.Version,
		&labelsJSON,
		&annotationsJSON,
		&t.WorkflowExecutionID,
		&t.WorkflowSubState,
		&t.WorkflowRetryCount,
		&t.WorkflowErrorMessage,
		&t.WorkflowConfigHash,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, tenant.ErrTenantNotFound
		}
		return nil, fmt.Errorf("get tenant: %w", err)
	}

	// Unmarshal JSONB fields
	if err := unmarshalInterfaceMap(desiredConfigJSON, &t.DesiredConfig); err != nil {
		return nil, fmt.Errorf("unmarshal desired_config: %w", err)
	}
	if err := unmarshalInterfaceMap(observedConfigJSON, &t.ObservedConfig); err != nil {
		return nil, fmt.Errorf("unmarshal observed_config: %w", err)
	}
	if err := unmarshalStringMap(observedResourceIDsJSON, &t.ObservedResourceIDs); err != nil {
		return nil, fmt.Errorf("unmarshal observed_resource_ids: %w", err)
	}
	if err := unmarshalStringMap(labelsJSON, &t.Labels); err != nil {
		return nil, fmt.Errorf("unmarshal labels: %w", err)
	}
	if err := unmarshalStringMap(annotationsJSON, &t.Annotations); err != nil {
		return nil, fmt.Errorf("unmarshal annotations: %w", err)
	}

	return t, nil
}

const getTenantByIDQuery = `
SELECT
    id, name, status, status_message,
    desired_config,
    observed_config, observed_resource_ids,
    created_at, updated_at,
	version, labels, annotations, workflow_execution_id,
	workflow_sub_state, workflow_retry_count, workflow_error_message,
	workflow_config_hash
FROM tenants
WHERE id = $1
`

func (r *Repository) GetTenantByID(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
	r.logger.Debug("getting tenant by ID", zap.String("id", id.String()))

	t := &tenant.Tenant{}
	var desiredConfigJSON, observedConfigJSON, observedResourceIDsJSON, labelsJSON, annotationsJSON []byte

	err := r.pool.QueryRow(ctx, getTenantByIDQuery, id).Scan(
		&t.ID,
		&t.Name,
		&t.Status,
		&t.StatusMessage,
		&desiredConfigJSON,
		&observedConfigJSON,
		&observedResourceIDsJSON,
		&t.CreatedAt,
		&t.UpdatedAt,
		&t.Version,
		&labelsJSON,
		&annotationsJSON,
		&t.WorkflowExecutionID,
		&t.WorkflowSubState,
		&t.WorkflowRetryCount,
		&t.WorkflowErrorMessage,
		&t.WorkflowConfigHash,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, tenant.ErrTenantNotFound
		}
		return nil, fmt.Errorf("get tenant by ID: %w", err)
	}

	// Unmarshal JSONB fields
	if err := unmarshalInterfaceMap(desiredConfigJSON, &t.DesiredConfig); err != nil {
		return nil, fmt.Errorf("unmarshal desired_config: %w", err)
	}
	if err := unmarshalInterfaceMap(observedConfigJSON, &t.ObservedConfig); err != nil {
		return nil, fmt.Errorf("unmarshal observed_config: %w", err)
	}
	if err := unmarshalStringMap(observedResourceIDsJSON, &t.ObservedResourceIDs); err != nil {
		return nil, fmt.Errorf("unmarshal observed_resource_ids: %w", err)
	}
	if err := unmarshalStringMap(labelsJSON, &t.Labels); err != nil {
		return nil, fmt.Errorf("unmarshal labels: %w", err)
	}
	if err := unmarshalStringMap(annotationsJSON, &t.Annotations); err != nil {
		return nil, fmt.Errorf("unmarshal annotations: %w", err)
	}

	return t, nil
}

const updateTenantQuery = `
UPDATE tenants SET
    name = $2,
    status = $3,
    status_message = $4,
    desired_config = $5,
    observed_config = $6,
    observed_resource_ids = $7,
    updated_at = NOW(),
    version = version + 1,
    labels = $8,
    annotations = $9,
	workflow_execution_id = $10,
	workflow_sub_state = $11,
	workflow_retry_count = $12,
	workflow_error_message = $13,
	workflow_config_hash = $15
WHERE id = $1 AND version = $14
RETURNING version, updated_at
`

func (r *Repository) UpdateTenant(ctx context.Context, t *tenant.Tenant) error {
	r.logger.Debug("updating tenant",
		zap.String("id", t.ID.String()),
		zap.Int("version", t.Version))

	row := r.pool.QueryRow(ctx, updateTenantQuery,
		t.ID,
		t.Name,
		t.Status,
		t.StatusMessage,
		jsonbOrEmptyInterfaceMap(t.DesiredConfig),
		jsonbOrEmptyInterfaceMap(t.ObservedConfig),
		jsonbOrEmptyStringMap(t.ObservedResourceIDs),
		jsonbOrEmptyStringMap(t.Labels),
		jsonbOrEmptyStringMap(t.Annotations),
		t.WorkflowExecutionID,
		t.WorkflowSubState,
		t.WorkflowRetryCount,
		t.WorkflowErrorMessage,
		t.Version, // Optimistic locking check
		t.WorkflowConfigHash,
	)

	err := row.Scan(&t.Version, &t.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return tenant.ErrTenantExists
		}
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
		zap.String("id", t.ID.String()),
		zap.Int("new_version", t.Version))

	return nil
}

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
		var desiredConfigJSON, observedConfigJSON, observedResourceIDsJSON, labelsJSON, annotationsJSON []byte

		err := rows.Scan(
			&t.ID, &t.Name, &t.Status, &t.StatusMessage,
			&desiredConfigJSON,
			&observedConfigJSON, &observedResourceIDsJSON,
			&t.CreatedAt, &t.UpdatedAt,
			&t.Version, &labelsJSON, &annotationsJSON,
			&t.WorkflowExecutionID,
			&t.WorkflowSubState,
			&t.WorkflowRetryCount,
			&t.WorkflowErrorMessage,
			&t.WorkflowConfigHash,
		)
		if err != nil {
			return nil, fmt.Errorf("scan tenant: %w", err)
		}

		// Unmarshal JSONB fields
		if err := unmarshalInterfaceMap(desiredConfigJSON, &t.DesiredConfig); err != nil {
			return nil, fmt.Errorf("unmarshal desired_config: %w", err)
		}
		if err := unmarshalInterfaceMap(observedConfigJSON, &t.ObservedConfig); err != nil {
			return nil, fmt.Errorf("unmarshal observed_config: %w", err)
		}
		if err := unmarshalStringMap(observedResourceIDsJSON, &t.ObservedResourceIDs); err != nil {
			return nil, fmt.Errorf("unmarshal observed_resource_ids: %w", err)
		}
		if err := unmarshalStringMap(labelsJSON, &t.Labels); err != nil {
			return nil, fmt.Errorf("unmarshal labels: %w", err)
		}
		if err := unmarshalStringMap(annotationsJSON, &t.Annotations); err != nil {
			return nil, fmt.Errorf("unmarshal annotations: %w", err)
		}

		tenants = append(tenants, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tenants: %w", err)
	}

	return tenants, nil
}

const listTenantsForReconciliationQuery = `
SELECT
    id, name, status, status_message,
    desired_config,
    observed_config, observed_resource_ids,
    created_at, updated_at,
	version, labels, annotations, workflow_execution_id,
	workflow_sub_state, workflow_retry_count, workflow_error_message,
	workflow_config_hash
FROM tenants
WHERE status IN ('requested', 'planning', 'provisioning', 'updating', 'deleting', 'archiving')
ORDER BY created_at ASC
`

func (r *Repository) ListTenantsForReconciliation(ctx context.Context) ([]*tenant.Tenant, error) {
	r.logger.Debug("listing tenants for reconciliation")

	rows, err := r.pool.Query(ctx, listTenantsForReconciliationQuery)
	if err != nil {
		return nil, fmt.Errorf("list tenants for reconciliation: %w", err)
	}
	defer rows.Close()

	var tenants []*tenant.Tenant
	for rows.Next() {
		t := &tenant.Tenant{}
		var desiredConfigJSON, observedConfigJSON, observedResourceIDsJSON, labelsJSON, annotationsJSON []byte

		err := rows.Scan(
			&t.ID, &t.Name, &t.Status, &t.StatusMessage,
			&desiredConfigJSON,
			&observedConfigJSON, &observedResourceIDsJSON,
			&t.CreatedAt, &t.UpdatedAt,
			&t.Version, &labelsJSON, &annotationsJSON,
			&t.WorkflowExecutionID,
			&t.WorkflowSubState,
			&t.WorkflowRetryCount,
			&t.WorkflowErrorMessage,
			&t.WorkflowConfigHash,
		)
		if err != nil {
			return nil, fmt.Errorf("scan tenant: %w", err)
		}

		// Unmarshal JSONB fields
		if err := unmarshalInterfaceMap(desiredConfigJSON, &t.DesiredConfig); err != nil {
			return nil, fmt.Errorf("unmarshal desired_config: %w", err)
		}
		if err := unmarshalInterfaceMap(observedConfigJSON, &t.ObservedConfig); err != nil {
			return nil, fmt.Errorf("unmarshal observed_config: %w", err)
		}
		if err := unmarshalStringMap(observedResourceIDsJSON, &t.ObservedResourceIDs); err != nil {
			return nil, fmt.Errorf("unmarshal observed_resource_ids: %w", err)
		}
		if err := unmarshalStringMap(labelsJSON, &t.Labels); err != nil {
			return nil, fmt.Errorf("unmarshal labels: %w", err)
		}
		if err := unmarshalStringMap(annotationsJSON, &t.Annotations); err != nil {
			return nil, fmt.Errorf("unmarshal annotations: %w", err)
		}

		tenants = append(tenants, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tenants for reconciliation: %w", err)
	}

	r.logger.Debug("found tenants for reconciliation", zap.Int("count", len(tenants)))
	return tenants, nil
}

func (r *Repository) buildListQuery(filters tenant.ListFilters) (string, []interface{}) {
	query := `
        SELECT
            id, name, status, status_message,
            desired_config,
            observed_config, observed_resource_ids,
            created_at, updated_at,
			version, labels, annotations, workflow_execution_id,
			workflow_sub_state, workflow_retry_count, workflow_error_message,
	workflow_config_hash
        FROM tenants
        WHERE 1=1
    `
	args := []interface{}{}
	argPos := 1

	// Filter by status
	if !filters.IncludeDeleted {
		query += " AND status != 'archived'"
	}
	if len(filters.Statuses) > 0 {
		query += fmt.Sprintf(" AND status = ANY($%d)", argPos)
		statusStrings := make([]string, len(filters.Statuses))
		for i, s := range filters.Statuses {
			statusStrings[i] = string(s)
		}
		args = append(args, statusStrings)
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

	// Filter by workflow sub-state
	if len(filters.WorkflowSubStates) > 0 {
		query += fmt.Sprintf(" AND workflow_sub_state = ANY($%d)", argPos)
		args = append(args, filters.WorkflowSubStates)
		argPos++
	}

	// Filter by workflow error presence
	if filters.HasWorkflowError != nil {
		if *filters.HasWorkflowError {
			query += " AND workflow_error_message IS NOT NULL"
		} else {
			query += " AND workflow_error_message IS NULL"
		}
	}

	// Filter by minimum retry count
	if filters.MinRetryCount != nil {
		query += fmt.Sprintf(" AND COALESCE(workflow_retry_count, 0) >= $%d", argPos)
		args = append(args, *filters.MinRetryCount)
		argPos++
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

const deleteTenantQuery = `
DELETE FROM tenants
WHERE id = $1
RETURNING id
`

func (r *Repository) DeleteTenant(ctx context.Context, id uuid.UUID) error {
	r.logger.Debug("deleting tenant", zap.String("id", id.String()))

	var deletedID uuid.UUID
	err := r.pool.QueryRow(ctx, deleteTenantQuery, id).Scan(&deletedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return tenant.ErrTenantNotFound
		}
		return fmt.Errorf("delete tenant: %w", err)
	}

	r.logger.Info("tenant deleted", zap.String("id", id.String()))
	return nil
}

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
		zap.String("tenant_id", st.TenantID.String()),
		zap.String("to_status", string(st.ToStatus)))

	row := r.pool.QueryRow(ctx, recordTransitionQuery,
		st.TenantID,
		st.FromStatus,
		st.ToStatus,
		st.Reason,
		st.TriggeredBy,
		jsonbOrEmptyInterfaceMap(st.DesiredStateSnapshot),
		jsonbOrEmptyInterfaceMap(st.ObservedStateSnapshot),
	)

	err := row.Scan(&st.ID, &st.CreatedAt)
	if err != nil {
		return fmt.Errorf("record transition: %w", err)
	}

	return nil
}

const getHistoryQuery = `
SELECT
    id, tenant_id, from_status, to_status,
    reason, triggered_by,
    desired_state_snapshot, observed_state_snapshot,
    created_at
FROM tenant_state_history
WHERE tenant_id = $1
ORDER BY created_at DESC
`

func (r *Repository) GetStateHistory(ctx context.Context, tenantID uuid.UUID) ([]*tenant.StateTransition, error) {
	r.logger.Debug("getting state history", zap.String("tenant_id", tenantID.String()))

	rows, err := r.pool.Query(ctx, getHistoryQuery, tenantID)
	if err != nil {
		return nil, fmt.Errorf("get history: %w", err)
	}
	defer rows.Close()

	var history []*tenant.StateTransition
	for rows.Next() {
		st := &tenant.StateTransition{}
		var desiredSnapshotJSON, observedSnapshotJSON []byte

		err := rows.Scan(
			&st.ID,
			&st.TenantID,
			&st.FromStatus,
			&st.ToStatus,
			&st.Reason,
			&st.TriggeredBy,
			&desiredSnapshotJSON,
			&observedSnapshotJSON,
			&st.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan transition: %w", err)
		}

		// Unmarshal JSONB snapshot fields
		if err := unmarshalInterfaceMap(desiredSnapshotJSON, &st.DesiredStateSnapshot); err != nil {
			return nil, fmt.Errorf("unmarshal desired_state_snapshot: %w", err)
		}
		if err := unmarshalInterfaceMap(observedSnapshotJSON, &st.ObservedStateSnapshot); err != nil {
			return nil, fmt.Errorf("unmarshal observed_state_snapshot: %w", err)
		}

		history = append(history, st)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate history: %w", err)
	}

	return history, nil
}

// jsonbOrEmpty converts map to JSONB, returns empty object if nil
func jsonbOrEmptyStringMap(m map[string]string) interface{} {
	if len(m) == 0 {
		return "{}"
	}
	return m
}

func jsonbOrEmptyInterfaceMap(m map[string]interface{}) interface{} {
	if len(m) == 0 {
		return "{}"
	}
	return m
}

// unmarshalStringMap unmarshals JSONB bytes into a map[string]string
func unmarshalStringMap(data []byte, m *map[string]string) error {
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, m)
}

// unmarshalInterfaceMap unmarshals JSONB bytes into a map[string]interface{}
func unmarshalInterfaceMap(data []byte, m *map[string]interface{}) error {
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, m)
}

// isUniqueViolation checks if error is unique constraint violation
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
