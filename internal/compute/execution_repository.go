package compute

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// ExecutionRepository provides data access for compute executions
type ExecutionRepository interface {
	// CreateComputeExecution inserts a new compute execution record
	CreateComputeExecution(ctx context.Context, exec *ComputeExecution) error

	// UpdateComputeExecution updates an existing execution record
	UpdateComputeExecution(ctx context.Context, exec *ComputeExecution) error

	// GetComputeExecution retrieves an execution by ID
	GetComputeExecution(ctx context.Context, executionID string) (*ComputeExecution, error)

	// ListComputeExecutions lists executions filtered by tenant ID
	ListComputeExecutions(ctx context.Context, tenantID string, filters ExecutionListFilters) ([]*ComputeExecution, error)

	// AddExecutionHistory appends a history record
	AddExecutionHistory(ctx context.Context, history *ComputeExecutionHistory) error

	// GetExecutionHistory retrieves history for an execution
	GetExecutionHistory(ctx context.Context, executionID string) ([]*ComputeExecutionHistory, error)
}

// ExecutionListFilters allows filtering execution queries
type ExecutionListFilters struct {
	// Status filter (empty = all statuses)
	Status *ComputeExecutionStatus

	// OperationType filter (empty = all types)
	OperationType *ComputeOperationType

	// Limit on number of results
	Limit int

	// Offset for pagination
	Offset int
}

// PgExecutionRepository implements ExecutionRepository using PostgreSQL
type PgExecutionRepository struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

// NewPgExecutionRepository creates a new PostgreSQL execution repository
func NewPgExecutionRepository(pool *pgxpool.Pool, logger *zap.Logger) *PgExecutionRepository {
	return &PgExecutionRepository{
		pool:   pool,
		logger: logger.With(zap.String("component", "execution-repository")),
	}
}

// CreateComputeExecution inserts a new compute execution
func (r *PgExecutionRepository) CreateComputeExecution(ctx context.Context, exec *ComputeExecution) error {
	query := `
		INSERT INTO compute_executions 
		(execution_id, tenant_id, workflow_execution_id, operation_type, status, resource_ids, error_code, error_message, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	now := time.Now()
	_, err := r.pool.Exec(ctx, query,
		exec.ExecutionID,
		exec.TenantID,
		exec.WorkflowExecutionID,
		exec.OperationType,
		exec.Status,
		exec.ResourceIDs,
		exec.ErrorCode,
		exec.ErrorMessage,
		now,
		now,
	)

	if err != nil {
		r.logger.Error("failed to create compute execution",
			zap.String("execution_id", exec.ExecutionID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to create compute execution: %w", err)
	}

	r.logger.Debug("created compute execution",
		zap.String("execution_id", exec.ExecutionID),
		zap.String("tenant_id", exec.TenantID),
		zap.String("operation_type", string(exec.OperationType)),
	)

	return nil
}

// UpdateComputeExecution updates an existing compute execution
func (r *PgExecutionRepository) UpdateComputeExecution(ctx context.Context, exec *ComputeExecution) error {
	query := `
		UPDATE compute_executions
		SET status = $1, resource_ids = $2, error_code = $3, error_message = $4, updated_at = $5
		WHERE execution_id = $6
	`

	now := time.Now()
	result, err := r.pool.Exec(ctx, query,
		exec.Status,
		exec.ResourceIDs,
		exec.ErrorCode,
		exec.ErrorMessage,
		now,
		exec.ExecutionID,
	)

	if err != nil {
		r.logger.Error("failed to update compute execution",
			zap.String("execution_id", exec.ExecutionID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to update compute execution: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
	}

	if rowsAffected == 0 {
		return fmt.Errorf("execution not found: %s", exec.ExecutionID)
	}

	r.logger.Debug("updated compute execution",
		zap.String("execution_id", exec.ExecutionID),
		zap.String("status", string(exec.Status)),
	)

	return nil
}

// GetComputeExecution retrieves an execution by ID
func (r *PgExecutionRepository) GetComputeExecution(ctx context.Context, executionID string) (*ComputeExecution, error) {
	query := `
		SELECT id, execution_id, tenant_id, workflow_execution_id, operation_type, status, 
		       resource_ids, error_code, error_message, created_at, updated_at
		FROM compute_executions
		WHERE execution_id = $1
	`

	exec := &ComputeExecution{}
	err := r.pool.QueryRow(ctx, query, executionID).Scan(
		&exec.ID,
		&exec.ExecutionID,
		&exec.TenantID,
		&exec.WorkflowExecutionID,
		&exec.OperationType,
		&exec.Status,
		&exec.ResourceIDs,
		&exec.ErrorCode,
		&exec.ErrorMessage,
		&exec.CreatedAt,
		&exec.UpdatedAt,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, fmt.Errorf("execution not found: %s", executionID)
		}
		r.logger.Error("failed to get compute execution",
			zap.String("execution_id", executionID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get compute execution: %w", err)
	}

	return exec, nil
}

// ListComputeExecutions lists executions with optional filtering
func (r *PgExecutionRepository) ListComputeExecutions(ctx context.Context, tenantID string, filters ExecutionListFilters) ([]*ComputeExecution, error) {
	query := `
		SELECT id, execution_id, tenant_id, workflow_execution_id, operation_type, status, 
		       resource_ids, error_code, error_message, created_at, updated_at
		FROM compute_executions
		WHERE tenant_id = $1
	`
	args := []interface{}{tenantID}
	argNum := 2

	// Add optional filters
	if filters.Status != nil {
		query += fmt.Sprintf(` AND status = $%d`, argNum)
		args = append(args, *filters.Status)
		argNum++
	}

	if filters.OperationType != nil {
		query += fmt.Sprintf(` AND operation_type = $%d`, argNum)
		args = append(args, *filters.OperationType)
		argNum++
	}

	// Add ordering and pagination
	query += ` ORDER BY created_at DESC`

	if filters.Limit > 0 {
		query += fmt.Sprintf(` LIMIT $%d`, argNum)
		args = append(args, filters.Limit)
		argNum++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(` OFFSET $%d`, argNum)
		args = append(args, filters.Offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to list compute executions",
			zap.String("tenant_id", tenantID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to list compute executions: %w", err)
	}
	defer rows.Close()

	var executions []*ComputeExecution
	for rows.Next() {
		exec := &ComputeExecution{}
		err := rows.Scan(
			&exec.ID,
			&exec.ExecutionID,
			&exec.TenantID,
			&exec.WorkflowExecutionID,
			&exec.OperationType,
			&exec.Status,
			&exec.ResourceIDs,
			&exec.ErrorCode,
			&exec.ErrorMessage,
			&exec.CreatedAt,
			&exec.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan execution row",
				zap.String("tenant_id", tenantID),
				zap.Error(err),
			)
			return nil, fmt.Errorf("failed to scan execution: %w", err)
		}
		executions = append(executions, exec)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("error iterating execution rows",
			zap.String("tenant_id", tenantID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("error iterating executions: %w", err)
	}

	return executions, nil
}

// AddExecutionHistory appends a history record
func (r *PgExecutionRepository) AddExecutionHistory(ctx context.Context, history *ComputeExecutionHistory) error {
	query := `
		INSERT INTO compute_execution_history 
		(compute_execution_id, status, details, timestamp)
		VALUES ($1, $2, $3, $4)
	`

	now := time.Now()
	_, err := r.pool.Exec(ctx, query,
		history.ComputeExecutionID,
		history.Status,
		history.Details,
		now,
	)

	if err != nil {
		r.logger.Error("failed to add execution history",
			zap.String("execution_id", history.ComputeExecutionID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to add execution history: %w", err)
	}

	return nil
}

// GetExecutionHistory retrieves history records for an execution
func (r *PgExecutionRepository) GetExecutionHistory(ctx context.Context, executionID string) ([]*ComputeExecutionHistory, error) {
	query := `
		SELECT id, compute_execution_id, status, details, timestamp
		FROM compute_execution_history
		WHERE compute_execution_id = $1
		ORDER BY timestamp ASC
	`

	rows, err := r.pool.Query(ctx, query, executionID)
	if err != nil {
		r.logger.Error("failed to get execution history",
			zap.String("execution_id", executionID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get execution history: %w", err)
	}
	defer rows.Close()

	var history []*ComputeExecutionHistory
	for rows.Next() {
		h := &ComputeExecutionHistory{}
		err := rows.Scan(
			&h.ID,
			&h.ComputeExecutionID,
			&h.Status,
			&h.Details,
			&h.Timestamp,
		)
		if err != nil {
			r.logger.Error("failed to scan history row",
				zap.String("execution_id", executionID),
				zap.Error(err),
			)
			return nil, fmt.Errorf("failed to scan history: %w", err)
		}
		history = append(history, h)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("error iterating history rows",
			zap.String("execution_id", executionID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("error iterating history: %w", err)
	}

	return history, nil
}
