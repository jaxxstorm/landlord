package postgres

import (
	"context"
	"testing"

	"github.com/jaxxstorm/landlord/internal/tenant"
)

func TestRepository_ListTenantsForReconciliation_Empty(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	// List with no tenants
	tenants, err := repo.ListTenantsForReconciliation(ctx)
	if err != nil {
		t.Fatalf("ListTenantsForReconciliation() error = %v", err)
	}

	if len(tenants) != 0 {
		t.Errorf("ListTenantsForReconciliation() len = %d, want 0", len(tenants))
	}
}

func TestRepository_ListTenantsForReconciliation_NonTerminalStates(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create tenants in non-terminal states that should be returned
	nonTerminalStates := []tenant.Status{
		tenant.StatusRequested,
		tenant.StatusPlanning,
		tenant.StatusProvisioning,
		tenant.StatusUpdating,
		tenant.StatusDeleting,
	}

	for _, status := range nonTerminalStates {
		tn := createTestTenant(t, "tenant-non-terminal-"+string(status))
		tn.Status = status
		if err := repo.CreateTenant(ctx, tn); err != nil {
			t.Fatalf("CreateTenant() error = %v", err)
		}
	}

	// List tenants for reconciliation
	tenants, err := repo.ListTenantsForReconciliation(ctx)
	if err != nil {
		t.Fatalf("ListTenantsForReconciliation() error = %v", err)
	}

	if len(tenants) != 5 {
		t.Errorf("ListTenantsForReconciliation() len = %d, want 5", len(tenants))
	}

	// Verify all returned tenants are in non-terminal states
	statusMap := make(map[tenant.Status]bool)
	for _, tn := range tenants {
		statusMap[tn.Status] = true
	}

	for _, status := range nonTerminalStates {
		if !statusMap[status] {
			t.Errorf("ListTenantsForReconciliation() missing status = %s", status)
		}
	}
}

func TestRepository_ListTenantsForReconciliation_ExcludesTerminalStates(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create tenants in terminal states that should NOT be returned
	terminalStates := []tenant.Status{
		tenant.StatusReady,
		tenant.StatusFailed,
		tenant.StatusArchived,
	}

	for _, status := range terminalStates {
		tn := createTestTenant(t, "tenant-terminal-"+string(status))
		tn.Status = status
		if err := repo.CreateTenant(ctx, tn); err != nil {
			t.Fatalf("CreateTenant() error = %v", err)
		}
	}

	// List tenants for reconciliation
	tenants, err := repo.ListTenantsForReconciliation(ctx)
	if err != nil {
		t.Fatalf("ListTenantsForReconciliation() error = %v", err)
	}

	// Should return empty list since all created tenants are in terminal states
	if len(tenants) != 0 {
		t.Errorf("ListTenantsForReconciliation() returned %d tenants, want 0 (all terminal)", len(tenants))
	}
}

func TestRepository_ListTenantsForReconciliation_ExcludesDeleted(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create a non-terminal tenant
	tn := createTestTenant(t, "tenant-to-delete")
	tn.Status = tenant.StatusRequested
	if err := repo.CreateTenant(ctx, tn); err != nil {
		t.Fatalf("CreateTenant() error = %v", err)
	}

	// Hard delete it
	if err := repo.DeleteTenant(ctx, tn.ID); err != nil {
		t.Fatalf("DeleteTenant() error = %v", err)
	}

	// List tenants for reconciliation
	tenants, err := repo.ListTenantsForReconciliation(ctx)
	if err != nil {
		t.Fatalf("ListTenantsForReconciliation() error = %v", err)
	}

	// Should not return the deleted tenant
	if len(tenants) != 0 {
		t.Errorf("ListTenantsForReconciliation() returned %d tenants, want 0 (all deleted)", len(tenants))
	}
}

func TestRepository_ListTenantsForReconciliation_Mixed(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create a mix of tenant states
	testCases := []struct {
		tenantID      string
		status        tenant.Status
		shouldInclude bool
	}{
		{"requested1", tenant.StatusRequested, true},
		{"planning1", tenant.StatusPlanning, true},
		{"provisioning1", tenant.StatusProvisioning, true},
		{"updating1", tenant.StatusUpdating, true},
		{"deleting1", tenant.StatusDeleting, true},
		{"ready1", tenant.StatusReady, false},
		{"failed1", tenant.StatusFailed, false},
		{"archived1", tenant.StatusArchived, false},
	}

	for _, tc := range testCases {
		tn := createTestTenant(t, tc.tenantID)
		tn.Status = tc.status
		if err := repo.CreateTenant(ctx, tn); err != nil {
			t.Fatalf("CreateTenant() for %s error = %v", tc.tenantID, err)
		}
	}

	// List tenants for reconciliation
	tenants, err := repo.ListTenantsForReconciliation(ctx)
	if err != nil {
		t.Fatalf("ListTenantsForReconciliation() error = %v", err)
	}

	// Should only return the 5 non-terminal tenants
	if len(tenants) != 5 {
		t.Errorf("ListTenantsForReconciliation() len = %d, want 5", len(tenants))
	}

	// Verify all returned tenants match expected non-terminal states
	returnedStatuses := make(map[tenant.Status]int)
	for _, tn := range tenants {
		returnedStatuses[tn.Status]++
	}

	expectedStatuses := map[tenant.Status]int{
		tenant.StatusRequested:    1,
		tenant.StatusPlanning:     1,
		tenant.StatusProvisioning: 1,
		tenant.StatusUpdating:     1,
		tenant.StatusDeleting:     1,
	}

	for status, count := range expectedStatuses {
		if returnedStatuses[status] != count {
			t.Errorf("ListTenantsForReconciliation() status %s count = %d, want %d", status, returnedStatuses[status], count)
		}
	}
}

func TestRepository_ListTenantsForReconciliation_OrderedByCreation(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create multiple tenants
	tenantIDs := []string{"first", "second", "third"}
	createdTenants := make([]*tenant.Tenant, len(tenantIDs))

	for i, id := range tenantIDs {
		tn := createTestTenant(t, id)
		tn.Status = tenant.StatusRequested
		if err := repo.CreateTenant(ctx, tn); err != nil {
			t.Fatalf("CreateTenant() error = %v", err)
		}
		createdTenants[i] = tn
	}

	// List tenants for reconciliation
	tenants, err := repo.ListTenantsForReconciliation(ctx)
	if err != nil {
		t.Fatalf("ListTenantsForReconciliation() error = %v", err)
	}

	// Verify order is by creation time
	if len(tenants) != 3 {
		t.Errorf("ListTenantsForReconciliation() len = %d, want 3", len(tenants))
	}

	for i := 0; i < len(tenants)-1; i++ {
		if tenants[i].CreatedAt.After(tenants[i+1].CreatedAt) {
			t.Errorf("ListTenantsForReconciliation() order incorrect: %s > %s", tenants[i].Name, tenants[i+1].Name)
		}
	}
}

func TestRepository_ListTenantsForReconciliation_PreservesTenantData(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create a tenant with all fields populated
	original := createTestTenant(t, "data-preservation")
	original.Status = tenant.StatusProvisioning
	original.StatusMessage = "In provisioning"
	original.DesiredConfig = map[string]interface{}{
		"image":    "myapp:v2",
		"replicas": "5",
		"zone":     "us-east-1",
	}
	original.Labels = map[string]string{
		"env":  "prod",
		"team": "platform",
	}
	original.Annotations = map[string]string{
		"slack-channel": "#alerts",
	}

	if err := repo.CreateTenant(ctx, original); err != nil {
		t.Fatalf("CreateTenant() error = %v", err)
	}

	// List and retrieve
	tenants, err := repo.ListTenantsForReconciliation(ctx)
	if err != nil {
		t.Fatalf("ListTenantsForReconciliation() error = %v", err)
	}

	if len(tenants) != 1 {
		t.Fatalf("ListTenantsForReconciliation() len = %d, want 1", len(tenants))
	}

	retrieved := tenants[0]

	// Verify all data is preserved
	if retrieved.Name != original.Name {
		t.Errorf("Name = %s, want %s", retrieved.Name, original.Name)
	}
	if retrieved.Status != original.Status {
		t.Errorf("Status = %s, want %s", retrieved.Status, original.Status)
	}
	if retrieved.StatusMessage != original.StatusMessage {
		t.Errorf("StatusMessage = %s, want %s", retrieved.StatusMessage, original.StatusMessage)
	}

	// Check config maps
	if len(retrieved.DesiredConfig) != len(original.DesiredConfig) {
		t.Errorf("DesiredConfig len = %d, want %d", len(retrieved.DesiredConfig), len(original.DesiredConfig))
	}
	for k, v := range original.DesiredConfig {
		if retrieved.DesiredConfig[k] != v {
			t.Errorf("DesiredConfig[%s] = %v, want %v", k, retrieved.DesiredConfig[k], v)
		}
	}

	// Check labels
	if len(retrieved.Labels) != len(original.Labels) {
		t.Errorf("Labels len = %d, want %d", len(retrieved.Labels), len(original.Labels))
	}
	for k, v := range original.Labels {
		if retrieved.Labels[k] != v {
			t.Errorf("Labels[%s] = %s, want %s", k, retrieved.Labels[k], v)
		}
	}

	// Check annotations
	if len(retrieved.Annotations) != len(original.Annotations) {
		t.Errorf("Annotations len = %d, want %d", len(retrieved.Annotations), len(original.Annotations))
	}
	for k, v := range original.Annotations {
		if retrieved.Annotations[k] != v {
			t.Errorf("Annotations[%s] = %s, want %s", k, retrieved.Annotations[k], v)
		}
	}
}
