package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/jaxxstorm/landlord/internal/api/models"
	"github.com/jaxxstorm/landlord/internal/compute"
	computemock "github.com/jaxxstorm/landlord/internal/compute/providers/mock"
	"github.com/jaxxstorm/landlord/internal/tenant"
	"github.com/jaxxstorm/landlord/internal/workflow"
)

// mockWorkflowClient implements WorkflowClient interface for testing
type mockWorkflowClient struct {
	triggerFunc           func(ctx context.Context, t *tenant.Tenant, action string) (string, error)
	triggerWithSourceFunc func(ctx context.Context, t *tenant.Tenant, action, source string) (string, error)
	determineActionFunc   func(status tenant.Status) (string, error)
	getStatusFunc         func(ctx context.Context, executionID string) (*workflow.ExecutionStatus, error)
}

func (m *mockWorkflowClient) TriggerWorkflow(ctx context.Context, t *tenant.Tenant, action string) (string, error) {
	if m.triggerFunc != nil {
		return m.triggerFunc(ctx, t, action)
	}
	return "exec-123", nil
}

func (m *mockWorkflowClient) TriggerWorkflowWithSource(ctx context.Context, t *tenant.Tenant, action, source string) (string, error) {
	if m.triggerWithSourceFunc != nil {
		return m.triggerWithSourceFunc(ctx, t, action, source)
	}
	return "exec-123", nil
}

func (m *mockWorkflowClient) DetermineAction(status tenant.Status) (string, error) {
	if m.determineActionFunc != nil {
		return m.determineActionFunc(status)
	}
	return "plan", nil
}

func (m *mockWorkflowClient) GetExecutionStatus(ctx context.Context, executionID string) (*workflow.ExecutionStatus, error) {
	if m.getStatusFunc != nil {
		return m.getStatusFunc(ctx, executionID)
	}
	return &workflow.ExecutionStatus{
		ExecutionID: executionID,
		State:       workflow.StateRunning,
	}, nil
}

func (m *mockWorkflowClient) StopExecution(ctx context.Context, t *tenant.Tenant, executionID string, reason string) error {
	return nil
}

// mockTenantRepo implements tenant.Repository for testing
type mockTenantRepo struct {
	createFunc           func(ctx context.Context, t *tenant.Tenant) error
	updateFunc           func(ctx context.Context, t *tenant.Tenant) error
	getByIDFunc          func(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error)
	getByNameFunc        func(ctx context.Context, name string) (*tenant.Tenant, error)
	listFunc             func(ctx context.Context, filters tenant.ListFilters) ([]*tenant.Tenant, error)
	listForReconcileFunc func(ctx context.Context) ([]*tenant.Tenant, error)
}

func (m *mockTenantRepo) CreateTenant(ctx context.Context, t *tenant.Tenant) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, t)
	}
	return nil
}

func (m *mockTenantRepo) UpdateTenant(ctx context.Context, t *tenant.Tenant) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, t)
	}
	return nil
}

func newTestComputeRegistry() *compute.Registry {
	registry := compute.NewRegistry(zap.NewNop())
	_ = registry.Register(computemock.New())
	return registry
}

func (m *mockTenantRepo) GetTenantByID(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return &tenant.Tenant{
		ID:     id,
		Name:   "test-tenant",
		Status: tenant.StatusRequested,
	}, nil
}

func (m *mockTenantRepo) GetTenantByName(ctx context.Context, name string) (*tenant.Tenant, error) {
	if m.getByNameFunc != nil {
		return m.getByNameFunc(ctx, name)
	}
	return nil, tenant.ErrTenantNotFound
}

func (m *mockTenantRepo) ListTenants(ctx context.Context, filters tenant.ListFilters) ([]*tenant.Tenant, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filters)
	}
	return nil, nil
}

func (m *mockTenantRepo) DeleteTenant(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockTenantRepo) ListTenantsForReconciliation(ctx context.Context) ([]*tenant.Tenant, error) {
	if m.listForReconcileFunc != nil {
		return m.listForReconcileFunc(ctx)
	}
	return nil, nil
}

func (m *mockTenantRepo) RecordStateTransition(ctx context.Context, transition *tenant.StateTransition) error {
	return nil
}

func (m *mockTenantRepo) GetStateHistory(ctx context.Context, tenantID uuid.UUID) ([]*tenant.StateTransition, error) {
	return nil, nil
}

// TestCreateTenantWithWorkflowTrigger tests successful tenant creation with workflow triggering
func TestCreateTenantWithWorkflowTrigger(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	wfClient := &mockWorkflowClient{}
	tenantRepo := &mockTenantRepo{}

	srv := &Server{
		logger:          logger,
		workflowClient:  wfClient,
		tenantRepo:      tenantRepo,
		computeRegistry: newTestComputeRegistry(),
		defaultComputeProvider: "mock",
		router:          nil, // Not needed for direct handler testing
	}

	reqBody := models.CreateTenantRequest{
		Name: "test-tenant",
		ComputeConfig: map[string]interface{}{
			"image": "nginx:latest",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/tenants", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleCreateTenant(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Should return 201 Created for new tenant
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}

	var respBody models.TenantResponse
	json.NewDecoder(resp.Body).Decode(&respBody)

	// Should not have workflow_execution_id yet
	if respBody.WorkflowExecutionID != nil {
		t.Error("expected workflow_execution_id to be nil")
	}

	// Status should be requested
	if respBody.Status != string(tenant.StatusRequested) {
		t.Errorf("expected status 'requested', got %s", respBody.Status)
	}
}

// TestCreateTenantWorkflowTriggerFailure tests error handling when workflow trigger fails
func TestCreateTenantWorkflowTriggerFailure(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	wfClient := &mockWorkflowClient{
		triggerWithSourceFunc: func(ctx context.Context, tenant *tenant.Tenant, action, source string) (string, error) {
			return "", ErrWorkflowProviderUnavailable
		},
	}
	tenantRepo := &mockTenantRepo{}

	srv := &Server{
		logger:          logger,
		workflowClient:  wfClient,
		tenantRepo:      tenantRepo,
		computeRegistry: newTestComputeRegistry(),
		defaultComputeProvider: "mock",
	}

	reqBody := models.CreateTenantRequest{
		Name: "test-tenant",
		ComputeConfig: map[string]interface{}{
			"image": "nginx:latest",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/tenants", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleCreateTenant(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Should still return 201 since workflows are triggered by reconciler
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}
}

func TestCreateTenantRequiresComputeConfig(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tenantRepo := &mockTenantRepo{}

	srv := &Server{
		logger:          logger,
		tenantRepo:      tenantRepo,
		computeRegistry: newTestComputeRegistry(),
		defaultComputeProvider: "mock",
	}

	reqBody := models.CreateTenantRequest{
		Name: "test-tenant",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/tenants", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleCreateTenant(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var errResp models.ErrorResponse
	json.NewDecoder(w.Body).Decode(&errResp)
	if errResp.Error != "compute_config is required" {
		t.Fatalf("expected compute_config required error, got %s", errResp.Error)
	}
}

// TestUpdateTenantWithWorkflowTrigger tests successful tenant update with workflow triggering
func TestUpdateTenantWithWorkflowTrigger(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	wfClient := &mockWorkflowClient{}
	tenantID := uuid.New()
	existingTenant := &tenant.Tenant{
		ID:     tenantID,
		Name:   "test-tenant",
		Status: tenant.StatusReady,
	}

	tenantRepo := &mockTenantRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
			return existingTenant, nil
		},
	}

	srv := &Server{
		logger:          logger,
		workflowClient:  wfClient,
		tenantRepo:      tenantRepo,
		computeRegistry: newTestComputeRegistry(),
		defaultComputeProvider: "mock",
	}

	reqBody := models.UpdateTenantRequest{
		ComputeConfig: map[string]interface{}{
			"image": "nginx:1.0",
		},
	}

	body, _ := json.Marshal(reqBody)

	// Create request with URL parameter using chi context
	req := httptest.NewRequest(http.MethodPut, "/v1/tenants/"+tenantID.String(), strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")

	// Manually set URL param since we're not using full router
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{Keys: []string{"id"}, Values: []string{tenantID.String()}},
	}))

	w := httptest.NewRecorder()
	srv.handleUpdateTenant(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Should return 202 Accepted
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("expected status 202, got %d", resp.StatusCode)
	}

	var respBody models.TenantResponse
	json.NewDecoder(resp.Body).Decode(&respBody)

	// Should not have workflow_execution_id yet
	if respBody.WorkflowExecutionID != nil {
		t.Error("expected workflow_execution_id to be nil")
	}

	// Status should be updating
	if respBody.Status != string(tenant.StatusUpdating) {
		t.Errorf("expected status 'updating', got %s", respBody.Status)
	}
}

func TestUpdateTenantRequiresComputeConfig(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tenantID := uuid.New()
	existingTenant := &tenant.Tenant{
		ID:     tenantID,
		Name:   "test-tenant",
		Status: tenant.StatusReady,
	}

	tenantRepo := &mockTenantRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
			return existingTenant, nil
		},
	}

	srv := &Server{
		logger:          logger,
		tenantRepo:      tenantRepo,
		computeRegistry: newTestComputeRegistry(),
		defaultComputeProvider: "mock",
	}

	reqBody := models.UpdateTenantRequest{
		Name: stringPtr("updated-name"),
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/v1/tenants/"+tenantID.String(), strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{Keys: []string{"id"}, Values: []string{tenantID.String()}},
	}))

	w := httptest.NewRecorder()
	srv.handleUpdateTenant(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var errResp models.ErrorResponse
	json.NewDecoder(w.Body).Decode(&errResp)
	if errResp.Error != "compute_config is required" {
		t.Fatalf("expected compute_config required error, got %s", errResp.Error)
	}
}

// TestUpdateArchivedTenant tests updating an archived tenant returns 409 Conflict
func TestUpdateArchivedTenant(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	wfClient := &mockWorkflowClient{}
	tenantID := uuid.New()
	existingTenant := &tenant.Tenant{
		ID:     tenantID,
		Name:   "test-tenant",
		Status: tenant.StatusArchived,
	}

	tenantRepo := &mockTenantRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
			return existingTenant, nil
		},
	}

	srv := &Server{
		logger:          logger,
		workflowClient:  wfClient,
		tenantRepo:      tenantRepo,
		computeRegistry: newTestComputeRegistry(),
		defaultComputeProvider: "mock",
	}

	reqBody := models.UpdateTenantRequest{
		ComputeConfig: map[string]interface{}{
			"image": "nginx:1.0",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/v1/tenants/"+tenantID.String(), strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{Keys: []string{"id"}, Values: []string{tenantID.String()}},
	}))
	w := httptest.NewRecorder()

	srv.handleUpdateTenant(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Should return 409 Conflict
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("expected status 409, got %d", resp.StatusCode)
	}
}

// TestDeleteTenantWithWorkflowTrigger tests successful tenant deletion with workflow triggering
func TestDeleteTenantWithWorkflowTrigger(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	wfClient := &mockWorkflowClient{}
	tenantID := uuid.New()
	existingTenant := &tenant.Tenant{
		ID:     tenantID,
		Name:   "test-tenant",
		Status: tenant.StatusReady,
	}

	tenantRepo := &mockTenantRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
			return existingTenant, nil
		},
	}

	srv := &Server{
		logger:          logger,
		workflowClient:  wfClient,
		tenantRepo:      tenantRepo,
		computeRegistry: newTestComputeRegistry(),
		defaultComputeProvider: "mock",
	}

	req := httptest.NewRequest(http.MethodDelete, "/v1/tenants/"+tenantID.String(), nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{Keys: []string{"id"}, Values: []string{tenantID.String()}},
	}))
	w := httptest.NewRecorder()

	srv.handleDeleteTenant(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Should return 202 Accepted
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("expected status 202, got %d", resp.StatusCode)
	}

	var respBody models.TenantResponse
	json.NewDecoder(resp.Body).Decode(&respBody)

	// Should not have workflow_execution_id yet
	if respBody.WorkflowExecutionID != nil {
		t.Error("expected workflow_execution_id to be nil")
	}

	// Status should be archiving
	if respBody.Status != string(tenant.StatusArchiving) {
		t.Errorf("expected status 'archiving', got %s", respBody.Status)
	}
}

// TestDeleteAlreadyDeletedTenant tests deleting archived tenant returns 202 Accepted
func TestDeleteAlreadyDeletedTenant(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	wfClient := &mockWorkflowClient{}
	tenantID := uuid.New()
	existingTenant := &tenant.Tenant{
		ID:     tenantID,
		Name:   "test-tenant",
		Status: tenant.StatusArchived,
	}

	tenantRepo := &mockTenantRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
			return existingTenant, nil
		},
	}

	srv := &Server{
		logger:          logger,
		workflowClient:  wfClient,
		tenantRepo:      tenantRepo,
		computeRegistry: newTestComputeRegistry(),
		defaultComputeProvider: "mock",
	}

	req := httptest.NewRequest(http.MethodDelete, "/v1/tenants/"+tenantID.String(), nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{Keys: []string{"id"}, Values: []string{tenantID.String()}},
	}))
	w := httptest.NewRecorder()

	srv.handleDeleteTenant(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Should return 202 Accepted for deletion request
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("expected status 202, got %d", resp.StatusCode)
	}
}

// TestWorkflowExecutionIDInResponse tests that workflow_execution_id is returned in GET responses
func TestWorkflowExecutionIDInResponse(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	execID := "exec-test-123"
	tenantID := uuid.New()
	existingTenant := &tenant.Tenant{
		ID:                  tenantID,
		Name:                "test-tenant",
		Status:              tenant.StatusPlanning,
		WorkflowExecutionID: &execID,
	}

	tenantRepo := &mockTenantRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
			return existingTenant, nil
		},
	}

	srv := &Server{
		logger:          logger,
		tenantRepo:      tenantRepo,
		computeRegistry: newTestComputeRegistry(),
		defaultComputeProvider: "mock",
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/tenants/"+tenantID.String(), nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{Keys: []string{"id"}, Values: []string{tenantID.String()}},
	}))
	w := httptest.NewRecorder()

	srv.handleGetTenant(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	var respBody models.TenantResponse
	json.NewDecoder(resp.Body).Decode(&respBody)

	// Should include execution ID
	if respBody.WorkflowExecutionID == nil || *respBody.WorkflowExecutionID != execID {
		t.Errorf("expected execution ID %s in response", execID)
	}
}

// TestAPITriggersIncludeTriggerSource tests that API does not trigger workflows directly
func TestAPITriggersIncludeTriggerSource(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	capturedSource := ""
	wfClient := &mockWorkflowClient{
		triggerWithSourceFunc: func(ctx context.Context, tenant *tenant.Tenant, action, source string) (string, error) {
			capturedSource = source
			return "exec-123", nil
		},
	}
	tenantRepo := &mockTenantRepo{}

	srv := &Server{
		logger:          logger,
		workflowClient:  wfClient,
		tenantRepo:      tenantRepo,
		computeRegistry: newTestComputeRegistry(),
		defaultComputeProvider: "mock",
	}

	reqBody := models.CreateTenantRequest{
		Name: "test-tenant",
		ComputeConfig: map[string]interface{}{
			"image": "nginx:latest",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/tenants", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleCreateTenant(w, req)

	// Verify workflow client was not invoked
	if capturedSource != "" {
		t.Errorf("expected no workflow trigger, got source '%s'", capturedSource)
	}
}

// TestCreateTenantWorkflowProviderUnavailable tests workflow trigger failure (500 error)
func TestCreateTenantWorkflowProviderUnavailable(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	wfClient := &mockWorkflowClient{
		triggerWithSourceFunc: func(ctx context.Context, tenant *tenant.Tenant, action, source string) (string, error) {
			return "", ErrWorkflowProviderUnavailable
		},
	}

	tenantID := uuid.New()
	tenantRepo := &mockTenantRepo{
		createFunc: func(ctx context.Context, t *tenant.Tenant) error {
			t.ID = tenantID
			return nil
		},
		updateFunc: func(ctx context.Context, t *tenant.Tenant) error {
			return nil
		},
	}

	srv := &Server{
		logger:          logger,
		workflowClient:  wfClient,
		tenantRepo:      tenantRepo,
		computeRegistry: newTestComputeRegistry(),
		defaultComputeProvider: "mock",
	}

	reqBody := models.CreateTenantRequest{
		Name: "test-tenant",
		ComputeConfig: map[string]interface{}{
			"image": "nginx:latest",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/tenants", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleCreateTenant(w, req)

	// Verify 201 status code
	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}
}

// TestUpdateTenantArchivedReturns409 tests updating archived tenant returns 409 Conflict
func TestUpdateTenantArchivedReturns409(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	tenantID := uuid.New()
	archivedTenant := &tenant.Tenant{
		ID:     tenantID,
		Name:   "archived-tenant",
		Status: tenant.StatusArchived,
	}

	tenantRepo := &mockTenantRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
			return archivedTenant, nil
		},
	}

	srv := &Server{
		logger:          logger,
		tenantRepo:      tenantRepo,
		computeRegistry: newTestComputeRegistry(),
		defaultComputeProvider: "mock",
	}

	reqBody := models.UpdateTenantRequest{
		ComputeConfig: map[string]interface{}{
			"image": "nginx:2.0",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/v1/tenants/"+tenantID.String(), strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")

	// Add chi context for URL parameters
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("id", tenantID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

	w := httptest.NewRecorder()
	srv.handleUpdateTenant(w, req)

	// Verify 409 Conflict status code
	if w.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", w.Code)
	}

	// Verify error response
	var errResp models.ErrorResponse
	json.NewDecoder(w.Body).Decode(&errResp)
	if errResp.Error != "Tenant is archived" {
		t.Errorf("expected 'Tenant is archived', got %s", errResp.Error)
	}
}

// TestDeleteTenantArchivedReturns200 tests deleting archived tenant returns 202
func TestDeleteTenantArchivedReturns200(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	tenantID := uuid.New()
	archivedTenant := &tenant.Tenant{
		ID:     tenantID,
		Name:   "already-archived",
		Status: tenant.StatusArchived,
	}

	tenantRepo := &mockTenantRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
			return archivedTenant, nil
		},
	}

	srv := &Server{
		logger:          logger,
		tenantRepo:      tenantRepo,
		computeRegistry: newTestComputeRegistry(),
		defaultComputeProvider: "mock",
	}

	req := httptest.NewRequest(http.MethodDelete, "/v1/tenants/"+tenantID.String(), nil)

	// Add chi context for URL parameters
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("id", tenantID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

	w := httptest.NewRecorder()
	srv.handleDeleteTenant(w, req)

	// Verify 202 Accepted status code
	if w.Code != http.StatusAccepted {
		t.Errorf("expected status 202, got %d", w.Code)
	}
}

// TestUpdateTenantInvalidStateTransition tests invalid state transition returns 409
func TestUpdateTenantInvalidStateTransition(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	tenantID := uuid.New()
	failedTenant := &tenant.Tenant{
		ID:     tenantID,
		Name:   "failed-tenant",
		Status: tenant.StatusFailed,
	}

	tenantRepo := &mockTenantRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
			return failedTenant, nil
		},
	}

	srv := &Server{
		logger:          logger,
		tenantRepo:      tenantRepo,
		computeRegistry: newTestComputeRegistry(),
		defaultComputeProvider: "mock",
	}

	reqBody := models.UpdateTenantRequest{
		ComputeConfig: map[string]interface{}{
			"image": "nginx:2.0",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/v1/tenants/"+tenantID.String(), strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")

	// Add chi context for URL parameters
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("id", tenantID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

	w := httptest.NewRecorder()
	srv.handleUpdateTenant(w, req)

	// Verify 409 Conflict status code
	if w.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", w.Code)
	}

	// Verify error response
	var errResp models.ErrorResponse
	json.NewDecoder(w.Body).Decode(&errResp)
	if errResp.Error != "Cannot update tenant in failed state" {
		t.Errorf("expected conflict error, got %s", errResp.Error)
	}
}

func TestListTenantsIncludesWorkflowStatusFields(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	subState := "backing-off"
	retryCount := 2
	errMsg := "transient"
	execID := "exec-123"

	tenantRepo := &mockTenantRepo{
		listFunc: func(ctx context.Context, filters tenant.ListFilters) ([]*tenant.Tenant, error) {
			return []*tenant.Tenant{
				{
					ID:                  uuid.New(),
					Name:                "workflow-tenant",
					Status:              tenant.StatusProvisioning,
					WorkflowExecutionID: &execID,
					WorkflowSubState:    &subState,
					WorkflowRetryCount:  &retryCount,
					WorkflowErrorMessage: &errMsg,
				},
			}, nil
		},
	}

	srv := &Server{
		logger:          logger,
		tenantRepo:      tenantRepo,
		computeRegistry: newTestComputeRegistry(),
		defaultComputeProvider: "mock",
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/tenants", nil)
	w := httptest.NewRecorder()

	srv.handleListTenants(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp models.ListTenantsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Tenants) != 1 {
		t.Fatalf("expected 1 tenant, got %d", len(resp.Tenants))
	}

	got := resp.Tenants[0]
	if got.WorkflowExecutionID == nil || *got.WorkflowExecutionID != execID {
		t.Fatalf("workflow_execution_id = %v, want %v", got.WorkflowExecutionID, execID)
	}
	if got.WorkflowSubState == nil || *got.WorkflowSubState != subState {
		t.Fatalf("workflow_sub_state = %v, want %v", got.WorkflowSubState, subState)
	}
	if got.WorkflowRetryCount == nil || *got.WorkflowRetryCount != retryCount {
		t.Fatalf("workflow_retry_count = %v, want %v", got.WorkflowRetryCount, retryCount)
	}
	if got.WorkflowErrorMessage == nil || *got.WorkflowErrorMessage != errMsg {
		t.Fatalf("workflow_error_message = %v, want %v", got.WorkflowErrorMessage, errMsg)
	}
}

func TestListTenantsWorkflowFilters(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	var captured []tenant.ListFilters

	tenantRepo := &mockTenantRepo{
		listFunc: func(ctx context.Context, filters tenant.ListFilters) ([]*tenant.Tenant, error) {
			captured = append(captured, filters)
			return []*tenant.Tenant{}, nil
		},
	}

	srv := &Server{
		logger:          logger,
		tenantRepo:      tenantRepo,
		computeRegistry: newTestComputeRegistry(),
		defaultComputeProvider: "mock",
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/tenants?workflow_sub_state=backing-off,waiting&has_workflow_error=true&min_retry_count=2", nil)
	w := httptest.NewRecorder()

	srv.handleListTenants(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if len(captured) == 0 {
		t.Fatalf("expected filters to be captured")
	}

	filters := captured[0]
	if len(filters.WorkflowSubStates) != 2 {
		t.Fatalf("expected 2 workflow_sub_state values, got %d", len(filters.WorkflowSubStates))
	}
	if filters.HasWorkflowError == nil || *filters.HasWorkflowError != true {
		t.Fatalf("expected has_workflow_error true")
	}
	if filters.MinRetryCount == nil || *filters.MinRetryCount != 2 {
		t.Fatalf("expected min_retry_count 2")
	}
}

func TestGetTenantIncludesWorkflowStatusFields(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tenantID := uuid.New()
	subState := "running"
	retryCount := 1
	errMsg := "transient"
	execID := "exec-456"

	tenantRepo := &mockTenantRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
			return &tenant.Tenant{
				ID:                  tenantID,
				Name:                "tenant-with-workflow",
				Status:              tenant.StatusProvisioning,
				WorkflowExecutionID: &execID,
				WorkflowSubState:    &subState,
				WorkflowRetryCount:  &retryCount,
				WorkflowErrorMessage: &errMsg,
			}, nil
		},
	}

	srv := &Server{
		logger:          logger,
		tenantRepo:      tenantRepo,
		computeRegistry: newTestComputeRegistry(),
		defaultComputeProvider: "mock",
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/tenants/"+tenantID.String(), nil)
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("id", tenantID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

	w := httptest.NewRecorder()
	srv.handleGetTenant(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var respBody models.TenantResponse
	if err := json.NewDecoder(w.Body).Decode(&respBody); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if respBody.WorkflowExecutionID == nil || *respBody.WorkflowExecutionID != execID {
		t.Fatalf("workflow_execution_id = %v, want %v", respBody.WorkflowExecutionID, execID)
	}
	if respBody.WorkflowSubState == nil || *respBody.WorkflowSubState != subState {
		t.Fatalf("workflow_sub_state = %v, want %v", respBody.WorkflowSubState, subState)
	}
	if respBody.WorkflowRetryCount == nil || *respBody.WorkflowRetryCount != retryCount {
		t.Fatalf("workflow_retry_count = %v, want %v", respBody.WorkflowRetryCount, retryCount)
	}
	if respBody.WorkflowErrorMessage == nil || *respBody.WorkflowErrorMessage != errMsg {
		t.Fatalf("workflow_error_message = %v, want %v", respBody.WorkflowErrorMessage, errMsg)
	}
}

// Helper function for creating string pointers
func stringPtr(s string) *string {
	return &s
}

// Mock error for testing
var ErrWorkflowProviderUnavailable = &WorkflowProviderError{message: "provider unavailable"}

type WorkflowProviderError struct {
	message string
}

func (e *WorkflowProviderError) Error() string {
	return e.message
}
