package compute

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"

	"github.com/jaxxstorm/landlord/internal/database"
	"github.com/jaxxstorm/landlord/internal/logger"
)

func TestExecutionRepository(t *testing.T) {
	// Setup database connection
	ctx := context.Background()
	log, _ := logger.New("development", "debug")
	defer log.Sync()

	// Connect to test database
	pool, err := pgxpool.New(ctx, "postgres://landlord:landlord@localhost:5432/landlord_test")
	if err != nil {
		t.Skipf("skipping execution repository tests: could not connect to database: %v", err)
	}
	defer pool.Close()

	// Run migrations
	connectionString := "postgres://landlord:landlord@localhost:5432/landlord_test"
	err = database.RunMigrations(connectionString, log)
	if err != nil {
		t.Skipf("skipping execution repository tests: migrations failed: %v", err)
	}

	// Create repository
	repo := NewPgExecutionRepository(pool, log)

	// Clean up any existing data before tests
	t.Cleanup(func() {
		pool.Exec(ctx, "DELETE FROM compute_execution_history")
		pool.Exec(ctx, "DELETE FROM compute_executions")
	})

	t.Run("CreateComputeExecution", func(t *testing.T) {
		exec := &ComputeExecution{
			ExecutionID:         "exec-001",
			TenantID:            "tenant-001",
			WorkflowExecutionID: "wf-exec-001",
			OperationType:       OperationTypeProvision,
			Status:              ExecutionStatusPending,
			ResourceIDs:         nil,
			ErrorCode:           nil,
			ErrorMessage:        nil,
		}

		err := repo.CreateComputeExecution(ctx, exec)
		assert.NoError(t, err)

		// Verify it was created
		retrieved, err := repo.GetComputeExecution(ctx, exec.ExecutionID)
		assert.NoError(t, err)
		assert.Equal(t, exec.ExecutionID, retrieved.ExecutionID)
		assert.Equal(t, exec.TenantID, retrieved.TenantID)
		assert.Equal(t, exec.OperationType, retrieved.OperationType)
		assert.Equal(t, exec.Status, retrieved.Status)
	})

	t.Run("UpdateComputeExecution", func(t *testing.T) {
		exec := &ComputeExecution{
			ExecutionID:         "exec-002",
			TenantID:            "tenant-002",
			WorkflowExecutionID: "wf-exec-002",
			OperationType:       OperationTypeProvision,
			Status:              ExecutionStatusPending,
			ResourceIDs:         nil,
			ErrorCode:           nil,
			ErrorMessage:        nil,
		}

		err := repo.CreateComputeExecution(ctx, exec)
		assert.NoError(t, err)

		// Update the execution
		exec.Status = ExecutionStatusRunning
		resourceJSON := `["resource-001", "resource-002"]`
		exec.ResourceIDs = []byte(resourceJSON)

		err = repo.UpdateComputeExecution(ctx, exec)
		assert.NoError(t, err)

		// Verify update
		retrieved, err := repo.GetComputeExecution(ctx, exec.ExecutionID)
		assert.NoError(t, err)
		assert.Equal(t, ExecutionStatusRunning, retrieved.Status)
		assert.Equal(t, []byte(resourceJSON), retrieved.ResourceIDs)
	})

	t.Run("ListComputeExecutions", func(t *testing.T) {
		tenantID := "tenant-list-test"

		// Create multiple executions for the same tenant
		for i := 0; i < 3; i++ {
			exec := &ComputeExecution{
				ExecutionID:         operationTypeToString(OperationTypeProvision) + "-" + time.Now().Format("150405") + "-" + string(rune(i+'0')),
				TenantID:            tenantID,
				WorkflowExecutionID: "wf-" + time.Now().Format("150405"),
				OperationType:       OperationTypeProvision,
				Status:              ExecutionStatusPending,
				ResourceIDs:         nil,
				ErrorCode:           nil,
				ErrorMessage:        nil,
			}
			err := repo.CreateComputeExecution(ctx, exec)
			assert.NoError(t, err)
		}

		// List all executions for the tenant
		filters := ExecutionListFilters{
			Limit:  10,
			Offset: 0,
		}
		executions, err := repo.ListComputeExecutions(ctx, tenantID, filters)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(executions), 3)

		// List with status filter
		// filter by pending status
		s := ExecutionStatusPending
		filters.Status = &s
		executions, err = repo.ListComputeExecutions(ctx, tenantID, filters)
		assert.NoError(t, err)
		for _, e := range executions {
			assert.Equal(t, ExecutionStatusPending, e.Status)
		}
	})

	t.Run("AddExecutionHistory", func(t *testing.T) {
		exec := &ComputeExecution{
			ExecutionID:         "exec-history-001",
			TenantID:            "tenant-history",
			WorkflowExecutionID: "wf-exec-history-001",
			OperationType:       OperationTypeProvision,
			Status:              ExecutionStatusPending,
			ResourceIDs:         nil,
			ErrorCode:           nil,
			ErrorMessage:        nil,
		}

		err := repo.CreateComputeExecution(ctx, exec)
		assert.NoError(t, err)

		// Add history record
		history := &ComputeExecutionHistory{
			ComputeExecutionID: exec.ExecutionID,
			Status:             ExecutionStatusRunning,
			Details:            json.RawMessage(`{"phase":"provisioning"}`),
		}

		err = repo.AddExecutionHistory(ctx, history)
		assert.NoError(t, err)

		// Retrieve history
		histories, err := repo.GetExecutionHistory(ctx, exec.ExecutionID)
		assert.NoError(t, err)
		assert.Greater(t, len(histories), 0)
	})

	t.Run("GetExecutionHistory", func(t *testing.T) {
		exec := &ComputeExecution{
			ExecutionID:         "exec-hist-retrieve-001",
			TenantID:            "tenant-hist-retrieve",
			WorkflowExecutionID: "wf-exec-hist-retrieve-001",
			OperationType:       OperationTypeUpdate,
			Status:              ExecutionStatusPending,
			ResourceIDs:         nil,
			ErrorCode:           nil,
			ErrorMessage:        nil,
		}

		err := repo.CreateComputeExecution(ctx, exec)
		assert.NoError(t, err)

		// Add multiple history records
		statuses := []ComputeExecutionStatus{
			ExecutionStatusRunning,
			ExecutionStatusSucceeded,
		}

		for _, status := range statuses {
			details := json.RawMessage([]byte(fmt.Sprintf(`{"status":"%s"}`, string(status))))
			history := &ComputeExecutionHistory{
				ComputeExecutionID: exec.ExecutionID,
				Status:             status,
				Details:            details,
			}
			err := repo.AddExecutionHistory(ctx, history)
			assert.NoError(t, err)
			time.Sleep(10 * time.Millisecond) // Ensure timestamp ordering
		}

		// Retrieve full history
		histories, err := repo.GetExecutionHistory(ctx, exec.ExecutionID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(histories), 2)

		// Verify ordering (should be by timestamp ASC)
		for i := 0; i < len(histories)-1; i++ {
			assert.True(t, histories[i].Timestamp.Before(histories[i+1].Timestamp) || histories[i].Timestamp.Equal(histories[i+1].Timestamp))
		}
	})

	t.Run("UpdateNonexistentExecution", func(t *testing.T) {
		exec := &ComputeExecution{
			ExecutionID: "nonexistent-update",
			TenantID:    "tenant-nonexistent",
			Status:      ExecutionStatusRunning,
		}

		err := repo.UpdateComputeExecution(ctx, exec)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "execution not found")
	})

	t.Run("ListWithPagination", func(t *testing.T) {
		tenantID := "tenant-pagination"

		// Create 5 executions
		for i := 0; i < 5; i++ {
			exec := &ComputeExecution{
				ExecutionID:         "paging-exec-" + time.Now().Format("150405.000000") + "-" + string(rune(i+'0')),
				TenantID:            tenantID,
				WorkflowExecutionID: "wf-paging",
				OperationType:       OperationTypeDelete,
				Status:              ExecutionStatusPending,
				ResourceIDs:         nil,
				ErrorCode:           nil,
				ErrorMessage:        nil,
			}
			err := repo.CreateComputeExecution(ctx, exec)
			assert.NoError(t, err)
			time.Sleep(1 * time.Millisecond)
		}

		// Test pagination: limit 2, offset 0
		filters := ExecutionListFilters{Limit: 2, Offset: 0}
		execs1, err := repo.ListComputeExecutions(ctx, tenantID, filters)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(execs1))

		// Test pagination: limit 2, offset 2
		filters.Offset = 2
		execs2, err := repo.ListComputeExecutions(ctx, tenantID, filters)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(execs2))

		// Verify no duplicates
		for _, e1 := range execs1 {
			for _, e2 := range execs2 {
				assert.NotEqual(t, e1.ExecutionID, e2.ExecutionID)
			}
		}
	})
}

// Helper function to convert operation type to string
func operationTypeToString(op ComputeOperationType) string {
	switch op {
	case OperationTypeProvision:
		return "provision"
	case OperationTypeUpdate:
		return "update"
	case OperationTypeDelete:
		return "delete"
	default:
		return "unknown"
	}
}
