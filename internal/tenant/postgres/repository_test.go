package postgres

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jaxxstorm/landlord/internal/tenant"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

// getMigrationsPath returns the path to the database migrations directory
func getMigrationsPath() string {
	// Get the directory of this file
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	// Navigate from internal/tenant/postgres to internal/database/migrations
	// First go to internal/tenant, then database/migrations
	parentDir := filepath.Dir(dir)      // internal/tenant
	parentDir = filepath.Dir(parentDir) // internal
	migrationsDir := filepath.Join(parentDir, "database", "migrations")
	return migrationsDir
}

func setupTestRepo(t *testing.T) (*Repository, func()) {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testcontainers.SkipIfProviderIsNotHealthy(t)

	ctx := context.Background()

	// Start PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get container host: %s", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("failed to get container port: %s", err)
	}

	dsn := "postgres://testuser:testpass@" + host + ":" + port.Port() + "/testdb?sslmode=disable"

	// Run migrations
	migrationPath := "file://" + getMigrationsPath()
	m, err := migrate.New(
		migrationPath,
		dsn,
	)
	if err != nil {
		t.Fatalf("failed to create migrate instance: %s", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("failed to run migrations: %s", err)
	}

	// Create connection pool
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to create pool: %s", err)
	}

	logger, _ := zap.NewDevelopment()
	repo, err := New(pool, logger)
	if err != nil {
		t.Fatalf("failed to create repository: %s", err)
	}

	cleanup := func() {
		pool.Close()
		if err := container.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %s", err)
		}
	}

	return repo, cleanup
}

func createTestTenant(t *testing.T, tenantName string) *tenant.Tenant {
	t.Helper()
	return &tenant.Tenant{
		Name:          tenantName,
		Status:        tenant.StatusRequested,
		StatusMessage: "awaiting planning",
		DesiredConfig: map[string]interface{}{
			"image":    "myapp:v1",
			"replicas": "3",
			"region":   "us-west-2",
		},
		Labels: map[string]string{
			"env": "test",
		},
		Annotations: map[string]string{
			"owner": "test-suite",
		},
	}
}

func TestRepository_CreateTenant(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()
	tn := createTestTenant(t, "test-tenant")

	// Create tenant
	err := repo.CreateTenant(ctx, tn)
	if err != nil {
		t.Fatalf("CreateTenant() error = %v", err)
	}

	// Verify fields were populated
	if tn.ID == uuid.Nil {
		t.Error("CreateTenant() did not set ID")
	}
	if tn.CreatedAt.IsZero() {
		t.Error("CreateTenant() did not set CreatedAt")
	}
	if tn.UpdatedAt.IsZero() {
		t.Error("CreateTenant() did not set UpdatedAt")
	}
	if tn.Version != 1 {
		t.Errorf("CreateTenant() Version = %d, want 1", tn.Version)
	}
}

func TestRepository_CreateTenant_Duplicate(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()
	tenant1 := createTestTenant(t, "duplicate-tenant")
	tenant2 := createTestTenant(t, "duplicate-tenant")

	// Create first tenant
	if err := repo.CreateTenant(ctx, tenant1); err != nil {
		t.Fatalf("CreateTenant() first insert error = %v", err)
	}

	// Try to create duplicate
	err := repo.CreateTenant(ctx, tenant2)
	if err != tenant.ErrTenantExists {
		t.Errorf("CreateTenant() duplicate error = %v, want %v", err, tenant.ErrTenantExists)
	}
}

func TestRepository_GetTenant(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()
	original := createTestTenant(t, "get-tenant")
	if err := repo.CreateTenant(ctx, original); err != nil {
		t.Fatalf("CreateTenant() error = %v", err)
	}

	// Retrieve tenant
	retrieved, err := repo.GetTenantByName(ctx, "get-tenant")
	if err != nil {
		t.Fatalf("GetTenantByName() error = %v", err)
	}

	// Verify fields
	if retrieved.ID != original.ID {
		t.Errorf("GetTenant() ID = %v, want %v", retrieved.ID, original.ID)
	}
	if retrieved.Name != original.Name {
		t.Errorf("GetTenantByName() Name = %v, want %v", retrieved.Name, original.Name)
	}
	if retrieved.Status != original.Status {
		t.Errorf("GetTenantByName() Status = %v, want %v", retrieved.Status, original.Status)
	}
	if value, ok := retrieved.DesiredConfig["replicas"].(string); !ok || value != "3" {
		t.Errorf("GetTenantByName() DesiredConfig[replicas] = %v, want 3", retrieved.DesiredConfig["replicas"])
	}
}

func TestRepository_GetTenant_NotFound(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()
	_, err := repo.GetTenantByName(ctx, "nonexistent")
	if err != tenant.ErrTenantNotFound {
		t.Errorf("GetTenantByName() error = %v, want %v", err, tenant.ErrTenantNotFound)
	}
}

func TestRepository_UpdateTenant(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()
	tn := createTestTenant(t, "update-tenant")
	if err := repo.CreateTenant(ctx, tn); err != nil {
		t.Fatalf("CreateTenant() error = %v", err)
	}

	// Update tenant
	originalVersion := tn.Version
	tn.Status = tenant.StatusPlanning
	tn.StatusMessage = "planning complete"
	tn.DesiredConfig["image"] = "myapp:v2"

	err := repo.UpdateTenant(ctx, tn)
	if err != nil {
		t.Fatalf("UpdateTenant() error = %v", err)
	}

	// Verify version incremented
	if tn.Version != originalVersion+1 {
		t.Errorf("UpdateTenant() Version = %d, want %d", tn.Version, originalVersion+1)
	}

	// Retrieve and verify
	retrieved, err := repo.GetTenantByName(ctx, "update-tenant")
	if err != nil {
		t.Fatalf("GetTenantByName() error = %v", err)
	}

	if retrieved.Status != tenant.StatusPlanning {
		t.Errorf("UpdateTenant() Status = %v, want %v", retrieved.Status, tenant.StatusPlanning)
	}
	if value, ok := retrieved.DesiredConfig["image"].(string); !ok || value != "myapp:v2" {
		t.Errorf("UpdateTenant() DesiredConfig[image] = %v, want myapp:v2", retrieved.DesiredConfig["image"])
	}
}

func TestRepository_UpdateTenant_VersionConflict(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()
	tn := createTestTenant(t, "conflict-tenant")
	if err := repo.CreateTenant(ctx, tn); err != nil {
		t.Fatalf("CreateTenant() error = %v", err)
	}

	// Simulate concurrent modification: update directly
	tenant2 := tn.Clone()
	tenant2.Status = tenant.StatusPlanning
	if err := repo.UpdateTenant(ctx, tenant2); err != nil {
		t.Fatalf("UpdateTenant() first update error = %v", err)
	}

	// Try to update with stale version
	tn.Status = tenant.StatusFailed // Using old version
	err := repo.UpdateTenant(ctx, tn)
	if err != tenant.ErrVersionConflict {
		t.Errorf("UpdateTenant() error = %v, want %v", err, tenant.ErrVersionConflict)
	}
}

func TestRepository_PersistsWorkflowStatusFields(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()
	tn := createTestTenant(t, "workflow-status-tenant")
	if err := repo.CreateTenant(ctx, tn); err != nil {
		t.Fatalf("CreateTenant() error = %v", err)
	}

	subState := "backing-off"
	retryCount := 3
	errMsg := "transient error"
	execID := "exec-123"
	tn.WorkflowSubState = &subState
	tn.WorkflowRetryCount = &retryCount
	tn.WorkflowErrorMessage = &errMsg
	tn.WorkflowExecutionID = &execID

	if err := repo.UpdateTenant(ctx, tn); err != nil {
		t.Fatalf("UpdateTenant() error = %v", err)
	}

	updated, err := repo.GetTenantByID(ctx, tn.ID)
	if err != nil {
		t.Fatalf("GetTenantByID() error = %v", err)
	}

	if updated.WorkflowSubState == nil || *updated.WorkflowSubState != subState {
		t.Fatalf("WorkflowSubState = %v, want %v", updated.WorkflowSubState, subState)
	}
	if updated.WorkflowRetryCount == nil || *updated.WorkflowRetryCount != retryCount {
		t.Fatalf("WorkflowRetryCount = %v, want %v", updated.WorkflowRetryCount, retryCount)
	}
	if updated.WorkflowErrorMessage == nil || *updated.WorkflowErrorMessage != errMsg {
		t.Fatalf("WorkflowErrorMessage = %v, want %v", updated.WorkflowErrorMessage, errMsg)
	}
	if updated.WorkflowExecutionID == nil || *updated.WorkflowExecutionID != execID {
		t.Fatalf("WorkflowExecutionID = %v, want %v", updated.WorkflowExecutionID, execID)
	}
}

func TestRepository_DeleteTenant(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()
	tn := createTestTenant(t, "delete-tenant")
	if err := repo.CreateTenant(ctx, tn); err != nil {
		t.Fatalf("CreateTenant() error = %v", err)
	}

	// Delete tenant
	err := repo.DeleteTenant(ctx, tn.ID)
	if err != nil {
		t.Fatalf("DeleteTenant() error = %v", err)
	}

	// Verify hard delete
	if _, err := repo.GetTenantByID(ctx, tn.ID); err != tenant.ErrTenantNotFound {
		t.Fatalf("GetTenantByID() after delete error = %v, want %v", err, tenant.ErrTenantNotFound)
	}
}
