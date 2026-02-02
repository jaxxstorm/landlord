package controller

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jaxxstorm/landlord/internal/config"
	"github.com/jaxxstorm/landlord/internal/tenant"
	"github.com/jaxxstorm/landlord/internal/workflow"
	workflowmock "github.com/jaxxstorm/landlord/internal/workflow/providers/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

type memoryTenantRepo struct {
	mu      sync.Mutex
	tenants map[uuid.UUID]*tenant.Tenant
	names   map[string]uuid.UUID
	history map[uuid.UUID][]*tenant.StateTransition
	nowFunc func() time.Time
	version map[uuid.UUID]int
}

func newMemoryTenantRepo() *memoryTenantRepo {
	return &memoryTenantRepo{
		tenants: make(map[uuid.UUID]*tenant.Tenant),
		names:   make(map[string]uuid.UUID),
		history: make(map[uuid.UUID][]*tenant.StateTransition),
		nowFunc: time.Now,
		version: make(map[uuid.UUID]int),
	}
}

func (m *memoryTenantRepo) CreateTenant(ctx context.Context, t *tenant.Tenant) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.names[t.Name]; exists {
		return tenant.ErrTenantExists
	}
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}

	now := m.nowFunc()
	t.CreatedAt = now
	t.UpdatedAt = now
	if t.Version == 0 {
		t.Version = 1
	}
	m.version[t.ID] = t.Version

	m.tenants[t.ID] = cloneTenant(t)
	m.names[t.Name] = t.ID
	return nil
}

func (m *memoryTenantRepo) GetTenantByName(ctx context.Context, name string) (*tenant.Tenant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id, ok := m.names[name]
	if !ok {
		return nil, tenant.ErrTenantNotFound
	}
	t, ok := m.tenants[id]
	if !ok {
		return nil, tenant.ErrTenantNotFound
	}
	return cloneTenant(t), nil
}

func (m *memoryTenantRepo) GetTenantByID(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, ok := m.tenants[id]
	if !ok {
		return nil, tenant.ErrTenantNotFound
	}
	return cloneTenant(t), nil
}

func (m *memoryTenantRepo) UpdateTenant(ctx context.Context, t *tenant.Tenant) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	existing, ok := m.tenants[t.ID]
	if !ok {
		return tenant.ErrTenantNotFound
	}
	if t.Version != 0 && t.Version != existing.Version {
		return tenant.ErrVersionConflict
	}

	updated := cloneTenant(t)
	updated.UpdatedAt = m.nowFunc()
	updated.Version = existing.Version + 1
	m.version[t.ID] = updated.Version
	m.tenants[t.ID] = updated
	return nil
}

func (m *memoryTenantRepo) ListTenants(ctx context.Context, filters tenant.ListFilters) ([]*tenant.Tenant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	statusFilter := make(map[tenant.Status]bool)
	for _, status := range filters.Statuses {
		statusFilter[status] = true
	}

	var results []*tenant.Tenant
	for _, t := range m.tenants {
		if !filters.IncludeDeleted && t.Status == tenant.StatusArchived {
			continue
		}
		if len(statusFilter) > 0 && !statusFilter[t.Status] {
			continue
		}
		results = append(results, cloneTenant(t))
	}

	return results, nil
}

func (m *memoryTenantRepo) ListTenantsForReconciliation(ctx context.Context) ([]*tenant.Tenant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var results []*tenant.Tenant
	for _, t := range m.tenants {
		if tenant.ShouldReconcile(t.Status) {
			results = append(results, cloneTenant(t))
		}
	}
	return results, nil
}

func (m *memoryTenantRepo) DeleteTenant(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, ok := m.tenants[id]
	if !ok {
		return tenant.ErrTenantNotFound
	}
	delete(m.tenants, id)
	delete(m.names, t.Name)
	return nil
}

func (m *memoryTenantRepo) RecordStateTransition(ctx context.Context, transition *tenant.StateTransition) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if transition.ID == uuid.Nil {
		transition.ID = uuid.New()
	}
	if transition.CreatedAt.IsZero() {
		transition.CreatedAt = m.nowFunc()
	}
	m.history[transition.TenantID] = append(m.history[transition.TenantID], transition)
	return nil
}

func (m *memoryTenantRepo) GetStateHistory(ctx context.Context, tenantID uuid.UUID) ([]*tenant.StateTransition, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	history := m.history[tenantID]
	copied := make([]*tenant.StateTransition, 0, len(history))
	for _, item := range history {
		copyItem := *item
		copied = append(copied, &copyItem)
	}
	return copied, nil
}

func cloneTenant(t *tenant.Tenant) *tenant.Tenant {
	if t == nil {
		return nil
	}
	clone := *t
	if t.WorkflowExecutionID != nil {
		id := *t.WorkflowExecutionID
		clone.WorkflowExecutionID = &id
	}
	clone.DesiredConfig = copyInterfaceMap(t.DesiredConfig)
	clone.ObservedConfig = copyInterfaceMap(t.ObservedConfig)
	clone.Labels = copyStringMap(t.Labels)
	clone.Annotations = copyStringMap(t.Annotations)
	return &clone
}

func copyInterfaceMap(input map[string]interface{}) map[string]interface{} {
	if input == nil {
		return nil
	}
	out := make(map[string]interface{}, len(input))
	for k, v := range input {
		out[k] = v
	}
	return out
}

func copyStringMap(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	out := make(map[string]string, len(input))
	for k, v := range input {
		out[k] = v
	}
	return out
}

func TestReconciler_InvokesWorkflowForRequestedTenant(t *testing.T) {
	repo := newMemoryTenantRepo()
	logger := zaptest.NewLogger(t)
	registry := workflow.NewRegistry(logger)
	provider := workflowmock.New(logger)
	require.NoError(t, registry.Register(provider))
	manager := workflow.New(registry, logger)

	tenantID := uuid.New()
	workflowID := "tenant-" + tenantID.String() + "-provision"
	_, err := manager.CreateWorkflow(context.Background(), &workflow.WorkflowSpec{
		WorkflowID:   workflowID,
		ProviderType: "mock",
		Name:         "Provision Tenant",
		Definition:   json.RawMessage(`{"test": true}`),
	})
	require.NoError(t, err)

	workflowClient := NewWorkflowClient(manager, logger, 5*time.Second, "mock")
	cfg := config.ControllerConfig{
		Enabled:                true,
		ReconciliationInterval: 100 * time.Millisecond,
		StatusPollInterval:     100 * time.Millisecond,
		Workers:                1,
		WorkflowTriggerTimeout: 5 * time.Second,
		ShutdownTimeout:        5 * time.Second,
		MaxRetries:             1,
	}
	reconciler := NewReconciler(repo, workflowClient, cfg, logger)

	err = repo.CreateTenant(context.Background(), &tenant.Tenant{
		ID:           tenantID,
		Name:         "requested-tenant",
		Status:       tenant.StatusRequested,
		DesiredImage: "nginx:latest",
	})
	require.NoError(t, err)

	err = reconciler.reconcile(tenantID.String())
	require.NoError(t, err)

	updated, err := repo.GetTenantByID(context.Background(), tenantID)
	require.NoError(t, err)
	require.Equal(t, tenant.StatusProvisioning, updated.Status)
	require.NotNil(t, updated.WorkflowExecutionID)
}

func TestReconciler_UpdatesTenantOnWorkflowSuccess(t *testing.T) {
	repo := newMemoryTenantRepo()
	logger := zaptest.NewLogger(t)
	registry := workflow.NewRegistry(logger)
	provider := workflowmock.New(logger)
	require.NoError(t, registry.Register(provider))
	manager := workflow.New(registry, logger)

	tenantID := uuid.New()
	workflowID := "tenant-" + tenantID.String() + "-provision"
	_, err := manager.CreateWorkflow(context.Background(), &workflow.WorkflowSpec{
		WorkflowID:   workflowID,
		ProviderType: "mock",
		Name:         "Provision Tenant",
		Definition:   json.RawMessage(`{"test": true}`),
	})
	require.NoError(t, err)

	request := &workflow.ProvisionRequest{
		TenantID:      "provisioned-tenant",
		TenantUUID:    tenantID.String(),
		Operation:     "provision",
		DesiredImage:  "nginx:latest",
		DesiredConfig: map[string]interface{}{"replicas": "1"},
	}
	execResult, err := manager.Invoke(context.Background(), workflowID, "mock", request)
	require.NoError(t, err)

	err = repo.CreateTenant(context.Background(), &tenant.Tenant{
		ID:                  tenantID,
		Name:                "provisioned-tenant",
		Status:              tenant.StatusProvisioning,
		WorkflowExecutionID: &execResult.ExecutionID,
		DesiredImage:        "nginx:latest",
	})
	require.NoError(t, err)

	workflowClient := NewWorkflowClient(manager, logger, 5*time.Second, "mock")
	cfg := config.ControllerConfig{
		Enabled:                true,
		ReconciliationInterval: 100 * time.Millisecond,
		StatusPollInterval:     100 * time.Millisecond,
		Workers:                1,
		WorkflowTriggerTimeout: 5 * time.Second,
		ShutdownTimeout:        5 * time.Second,
		MaxRetries:             1,
	}
	reconciler := NewReconciler(repo, workflowClient, cfg, logger)

	err = reconciler.reconcile(tenantID.String())
	require.NoError(t, err)

	updated, err := repo.GetTenantByID(context.Background(), tenantID)
	require.NoError(t, err)
	require.Equal(t, tenant.StatusReady, updated.Status)
	require.NotNil(t, updated.ObservedConfig)
	require.Equal(t, "success", updated.ObservedConfig["result"])
	require.Equal(t, true, updated.ObservedConfig["mock"])
}
