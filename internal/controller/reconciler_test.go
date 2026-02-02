package controller

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jaxxstorm/landlord/internal/compute"
	"github.com/jaxxstorm/landlord/internal/config"
	"github.com/jaxxstorm/landlord/internal/tenant"
	tenantpg "github.com/jaxxstorm/landlord/internal/tenant/postgres"
	"github.com/jaxxstorm/landlord/internal/workflow"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

// getMigrationsPath returns the path to the database migrations directory
func getMigrationsPathForTests() string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	parentDir := filepath.Dir(dir) // internal
	migrationsDir := filepath.Join(parentDir, "database", "migrations")
	return migrationsDir
}

// setupTestReconciler creates a reconciler with a test database
func setupTestReconciler(t *testing.T) (*Reconciler, *tenantpg.Repository, func()) {
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
		t.Skipf("skipping integration test (container start failed): %s", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		t.Skipf("skipping integration test (container host unavailable): %s", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		_ = container.Terminate(ctx)
		t.Skipf("skipping integration test (container port unavailable): %s", err)
	}

	dsn := "postgres://testuser:testpass@" + host + ":" + port.Port() + "/testdb?sslmode=disable"

	// Run migrations
	migrationPath := "file://" + getMigrationsPathForTests()
	m, err := migrate.New(migrationPath, dsn)
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
	repo, err := tenantpg.New(pool, logger)
	if err != nil {
		t.Fatalf("failed to create repository: %s", err)
	}

	// Create workflow manager with mock provider
	registry := workflow.NewRegistry(logger)
	mockProvider := &MockWorkflowProvider{
		logger: logger,
	}
	if err := registry.Register(mockProvider); err != nil {
		t.Fatalf("failed to register mock provider: %s", err)
	}
	wfManager := workflow.New(registry, logger)

	// Create workflow client
	wfClient := NewWorkflowClient(wfManager, logger, 5*time.Second, "mock")

	// Create controller config
	cfg := config.ControllerConfig{
		Enabled:                true,
		ReconciliationInterval: 1 * time.Second,
		Workers:                1,
		WorkflowTriggerTimeout: 5 * time.Second,
		ShutdownTimeout:        10 * time.Second,
		MaxRetries:             3,
	}

	// Create reconciler
	reconciler := NewReconciler(repo, wfClient, cfg, logger)

	cleanup := func() {
		pool.Close()
		if err := container.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %s", err)
		}
	}

	return reconciler, repo, cleanup
}

// MockWorkflowProvider for testing
type MockWorkflowProvider struct {
	logger                *zap.Logger
	createWorkflowCalls   int
	startExecutionCalls   int
	lastStartExecutionID  string
	simulateError         bool
	simulateSlowExecution bool
}

func (m *MockWorkflowProvider) Name() string {
	return "mock"
}

func (m *MockWorkflowProvider) CreateWorkflow(ctx context.Context, spec *workflow.WorkflowSpec) (*workflow.CreateWorkflowResult, error) {
	m.createWorkflowCalls++
	return &workflow.CreateWorkflowResult{
		WorkflowID:   spec.WorkflowID,
		ProviderType: "mock",
		CreatedAt:    time.Now(),
	}, nil
}

func (m *MockWorkflowProvider) Invoke(ctx context.Context, workflowID string, request *workflow.ProvisionRequest) (*workflow.ExecutionResult, error) {
	return m.StartExecution(ctx, workflowID, &workflow.ExecutionInput{ExecutionName: "exec-" + workflowID})
}

func (m *MockWorkflowProvider) GetWorkflowStatus(ctx context.Context, executionID string) (*workflow.WorkflowStatus, error) {
	status, err := m.GetExecutionStatus(ctx, executionID)
	if err != nil {
		return nil, err
	}
	return &workflow.WorkflowStatus{
		ExecutionID: status.ExecutionID,
		State:       status.State,
	}, nil
}

func (m *MockWorkflowProvider) StartExecution(ctx context.Context, workflowID string, input *workflow.ExecutionInput) (*workflow.ExecutionResult, error) {
	m.startExecutionCalls++

	if m.simulateSlowExecution {
		select {
		case <-time.After(200 * time.Millisecond):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if m.simulateError {
		return nil, workflow.ErrProviderConflict
	}

	executionID := "exec-" + workflowID
	m.lastStartExecutionID = executionID

	return &workflow.ExecutionResult{
		ExecutionID:  executionID,
		WorkflowID:   workflowID,
		ProviderType: "mock",
		State:        workflow.StateRunning,
		StartedAt:    time.Now(),
	}, nil
}

func (m *MockWorkflowProvider) GetExecutionStatus(ctx context.Context, executionID string) (*workflow.ExecutionStatus, error) {
	return &workflow.ExecutionStatus{
		ExecutionID:  executionID,
		State:        workflow.StateSucceeded,
		StartTime:    time.Now(),
		ProviderType: "mock",
	}, nil
}

func (m *MockWorkflowProvider) StopExecution(ctx context.Context, executionID string, reason string) error {
	return nil
}

func (m *MockWorkflowProvider) DeleteWorkflow(ctx context.Context, workflowID string) error {
	return nil
}

func (m *MockWorkflowProvider) Validate(ctx context.Context, spec *workflow.WorkflowSpec) error {
	return nil
}

func (m *MockWorkflowProvider) PostComputeCallback(ctx context.Context, executionID string, payload *compute.CallbackPayload, opts *compute.CallbackOptions) error {
	// Stub implementation for tests
	return nil
}

func TestReconcilerIntegration_SingleTenantProvisioning(t *testing.T) {
	reconciler, repo, cleanup := setupTestReconciler(t)
	defer cleanup()

	ctx := context.Background()

	// Create a tenant in requested status
	tn := &tenant.Tenant{
		Name:          "test-tenant-1",
		Status:        tenant.StatusRequested,
		StatusMessage: "awaiting provisioning",
		DesiredImage:  "myapp:v1",
		DesiredConfig: map[string]interface{}{
			"replicas": "2",
		},
	}
	if err := repo.CreateTenant(ctx, tn); err != nil {
		t.Fatalf("CreateTenant() error = %v", err)
	}

	// Run reconciliation once
	reconciler.queue.Add(tn.ID.String())
	if err := reconciler.reconcile(tn.ID.String()); err != nil {
		t.Logf("reconcile() error = %v (may be expected)", err)
	}

	// Verify tenant was updated
	updated, err := repo.GetTenantByID(ctx, tn.ID)
	if err != nil {
		t.Fatalf("GetTenantByID() error = %v", err)
	}

	// Status transitions to provisioning after workflow invocation
	if updated.Status != tenant.StatusProvisioning {
		t.Errorf("Tenant status = %s, want %s", updated.Status, tenant.StatusProvisioning)
	}
}

func TestReconcilerIntegration_EndToEndProvisioning(t *testing.T) {
	reconciler, repo, cleanup := setupTestReconciler(t)
	defer cleanup()

	ctx := context.Background()
	tenantID := uuid.New()

	err := repo.CreateTenant(ctx, &tenant.Tenant{
		ID:           tenantID,
		Name:         "e2e-tenant",
		Status:       tenant.StatusRequested,
		DesiredImage: "nginx:latest",
	})
	require.NoError(t, err)

	reconciler.queue.Add(tenantID.String())
	err = reconciler.reconcile(tenantID.String())
	require.NoError(t, err)

	updated, err := repo.GetTenantByID(ctx, tenantID)
	require.NoError(t, err)
	require.Equal(t, tenant.StatusProvisioning, updated.Status)
	require.NotNil(t, updated.WorkflowExecutionID)

	reconciler.queue.Add(tenantID.String())
	err = reconciler.reconcile(tenantID.String())
	require.NoError(t, err)

	final, err := repo.GetTenantByID(ctx, tenantID)
	require.NoError(t, err)
	require.Equal(t, tenant.StatusReady, final.Status)
}

func TestReconcilerIntegration_ListingForReconciliation(t *testing.T) {
	_, repo, cleanup := setupTestReconciler(t)
	defer cleanup()

	ctx := context.Background()

	// Create tenants in different states
	testCases := []struct {
		tenantID      string
		status        tenant.Status
		shouldInclude bool
	}{
		{"req-1", tenant.StatusRequested, true},
		{"plan-1", tenant.StatusPlanning, true},
		{"ready-1", tenant.StatusReady, false},
		{"failed-1", tenant.StatusFailed, false},
	}

	for _, tc := range testCases {
		tn := &tenant.Tenant{
			Name:          tc.tenantID,
			Status:        tc.status,
			DesiredImage:  "app:v1",
			DesiredConfig: map[string]interface{}{},
		}
		if err := repo.CreateTenant(ctx, tn); err != nil {
			t.Fatalf("CreateTenant() error = %v", err)
		}
	}

	// List tenants for reconciliation
	tenants, err := repo.ListTenantsForReconciliation(ctx)
	if err != nil {
		t.Fatalf("ListTenantsForReconciliation() error = %v", err)
	}

	// Should only have 2 non-terminal tenants
	if len(tenants) != 2 {
		t.Errorf("ListTenantsForReconciliation() returned %d tenants, want 2", len(tenants))
	}

	// Verify the returned tenants are non-terminal
	for _, tn := range tenants {
		if tn.Status == tenant.StatusReady || tn.Status == tenant.StatusFailed {
			t.Errorf("ListTenantsForReconciliation() returned terminal status: %s", tn.Status)
		}
	}
}

func TestReconcilerIntegration_DeletedTenantsIgnored(t *testing.T) {
	_, repo, cleanup := setupTestReconciler(t)
	defer cleanup()

	ctx := context.Background()

	// Create two tenants
	tn1 := &tenant.Tenant{
		Name:          "active-tenant",
		Status:        tenant.StatusRequested,
		DesiredImage:  "app:v1",
		DesiredConfig: map[string]interface{}{},
	}
	tn2 := &tenant.Tenant{
		Name:          "deleted-tenant",
		Status:        tenant.StatusRequested,
		DesiredImage:  "app:v1",
		DesiredConfig: map[string]interface{}{},
	}

	if err := repo.CreateTenant(ctx, tn1); err != nil {
		t.Fatalf("CreateTenant() error = %v", err)
	}
	if err := repo.CreateTenant(ctx, tn2); err != nil {
		t.Fatalf("CreateTenant() error = %v", err)
	}

	// Hard delete the second tenant
	if err := repo.DeleteTenant(ctx, tn2.ID); err != nil {
		t.Fatalf("DeleteTenant() error = %v", err)
	}

	// List tenants for reconciliation
	tenants, err := repo.ListTenantsForReconciliation(ctx)
	if err != nil {
		t.Fatalf("ListTenantsForReconciliation() error = %v", err)
	}

	// Only the active tenant should be returned
	if len(tenants) != 1 {
		t.Errorf("ListTenantsForReconciliation() returned %d tenants, want 1", len(tenants))
	}
	if len(tenants) > 0 && tenants[0].Name != "active-tenant" {
		t.Errorf("ListTenantsForReconciliation() returned %s, want active-tenant", tenants[0].Name)
	}
}

func TestReconcilerIntegration_QueueDeduplication(t *testing.T) {
	reconciler, repo, cleanup := setupTestReconciler(t)
	defer cleanup()

	ctx := context.Background()

	// Create a tenant
	tn := &tenant.Tenant{
		Name:          "dedup-test",
		Status:        tenant.StatusRequested,
		DesiredImage:  "app:v1",
		DesiredConfig: map[string]interface{}{},
	}
	if err := repo.CreateTenant(ctx, tn); err != nil {
		t.Fatalf("CreateTenant() error = %v", err)
	}

	// Add to queue multiple times (should deduplicate)
	reconciler.queue.Add(tn.ID.String())
	reconciler.queue.Add(tn.ID.String())
	reconciler.queue.Add(tn.ID.String())

	// Should only have one item in queue
	if reconciler.queue.Len() != 1 {
		t.Errorf("Queue length = %d, want 1 (deduplication)", reconciler.queue.Len())
	}
}

func TestReconcilerIntegration_MultipleTenantsInQueue(t *testing.T) {
	reconciler, repo, cleanup := setupTestReconciler(t)
	defer cleanup()

	ctx := context.Background()

	// Create multiple tenants
	tenantCount := 3
	tenantIDs := make([]string, tenantCount)
	for i := 0; i < tenantCount; i++ {
		tenantName := "tenant-" + string(rune(48+i)) // "tenant-0", "tenant-1", etc
		tn := &tenant.Tenant{
			Name:          tenantName,
			Status:        tenant.StatusRequested,
			StatusMessage: "test",
			DesiredImage:  "app:v1",
			DesiredConfig: map[string]interface{}{},
		}
		if err := repo.CreateTenant(ctx, tn); err != nil {
			t.Fatalf("CreateTenant() error = %v", err)
		}
		tenantIDs[i] = tn.ID.String()
		reconciler.queue.Add(tn.ID.String())
	}

	// Verify queue length
	if reconciler.queue.Len() != tenantCount {
		t.Errorf("Queue length = %d, want %d", reconciler.queue.Len(), tenantCount)
	}

	// Process items
	processed := 0
	for {
		item, shutdown := reconciler.queue.Get()
		if shutdown {
			break
		}
		tenantID := item.(string)
		if err := reconciler.reconcile(tenantID); err != nil {
			t.Logf("reconcile(%s) error = %v", tenantID, err)
		}
		reconciler.queue.Done(item)
		processed++
		if processed == tenantCount {
			break
		}
	}

	if processed != tenantCount {
		t.Errorf("Processed %d items, want %d", processed, tenantCount)
	}
}

func TestReconcilerIntegration_UpdateTenantPreservesData(t *testing.T) {
	reconciler, repo, cleanup := setupTestReconciler(t)
	defer cleanup()

	ctx := context.Background()

	// Create a tenant with detailed data
	tn := &tenant.Tenant{
		Name:          "data-tenant",
		Status:        tenant.StatusRequested,
		StatusMessage: "Initial setup",
		DesiredImage:  "myapp:v1.2.3",
		DesiredConfig: map[string]interface{}{
			"replicas":    "3",
			"cpu":         "512m",
			"memory":      "1Gi",
			"environment": "production",
		},
		Labels: map[string]string{
			"team": "platform",
			"cost": "backend",
		},
		Annotations: map[string]string{
			"slack-channel": "#alerts",
			"wiki-link":     "https://wiki.example.com/tenant",
		},
	}
	if err := repo.CreateTenant(ctx, tn); err != nil {
		t.Fatalf("CreateTenant() error = %v", err)
	}

	// Reconcile the tenant
	reconciler.queue.Add(tn.ID.String())
	if err := reconciler.reconcile(tn.ID.String()); err != nil {
		t.Logf("reconcile() error = %v", err)
	}
	reconciler.queue.Done(tn.ID.String())

	// Verify data is preserved
	updated, err := repo.GetTenantByID(ctx, tn.ID)
	if err != nil {
		t.Fatalf("GetTenantByID() error = %v", err)
	}

	// Check preserved fields
	if updated.DesiredImage != tn.DesiredImage {
		t.Errorf("DesiredImage = %s, want %s", updated.DesiredImage, tn.DesiredImage)
	}
	if value, ok := updated.DesiredConfig["replicas"].(string); !ok || value != "3" {
		t.Errorf("DesiredConfig[replicas] = %v, want 3", updated.DesiredConfig["replicas"])
	}
	if updated.Labels["team"] != "platform" {
		t.Errorf("Labels[team] = %s, want platform", updated.Labels["team"])
	}
	if updated.Annotations["slack-channel"] != "#alerts" {
		t.Errorf("Annotations[slack-channel] = %s, want #alerts", updated.Annotations["slack-channel"])
	}
}

func TestReconcilerIntegration_TerminalStatusNotReconciled(t *testing.T) {
	_, repo, cleanup := setupTestReconciler(t)
	defer cleanup()

	ctx := context.Background()

	// Create a tenant in ready status
	tn := &tenant.Tenant{
		Name:          "ready-tenant",
		Status:        tenant.StatusReady,
		DesiredImage:  "app:v1",
		DesiredConfig: map[string]interface{}{},
	}
	if err := repo.CreateTenant(ctx, tn); err != nil {
		t.Fatalf("CreateTenant() error = %v", err)
	}

	// List tenants for reconciliation
	tenants, err := repo.ListTenantsForReconciliation(ctx)
	if err != nil {
		t.Fatalf("ListTenantsForReconciliation() error = %v", err)
	}

	// Terminal status tenant should not be in the list
	if len(tenants) != 0 {
		t.Errorf("ListTenantsForReconciliation() returned %d tenants, want 0", len(tenants))
	}
}
