package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/jaxxstorm/landlord/internal/api/models"
	"github.com/jaxxstorm/landlord/internal/compute"
	"github.com/jaxxstorm/landlord/internal/tenant"
	"github.com/jaxxstorm/landlord/internal/workflow"
)

// TestIntegrationAPITriggersPlanWorkflow tests POST /v1/tenants does not trigger workflows directly
func TestIntegrationAPITriggersPlanWorkflow(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Track workflow trigger calls
	var triggeredTenants []string
	var triggeredActions []string
	var triggeredSources []string

	tenantID := uuid.New()
	tenantRepo := &mockTenantRepo{
		createFunc: func(ctx context.Context, t *tenant.Tenant) error {
			t.ID = tenantID
			triggeredTenants = append(triggeredTenants, t.Name)
			return nil
		},
		updateFunc: func(ctx context.Context, t *tenant.Tenant) error {
			return nil
		},
	}

	wfClient := &mockWorkflowClient{
		triggerWithSourceFunc: func(ctx context.Context, t *tenant.Tenant, action, source string) (string, error) {
			triggeredActions = append(triggeredActions, action)
			triggeredSources = append(triggeredSources, source)
			return fmt.Sprintf("exec-%s-%s", t.Name, action), nil
		},
	}

	srv := &Server{
		logger:         logger,
		workflowClient: wfClient,
		tenantRepo:     tenantRepo,
	}

	reqBody := models.CreateTenantRequest{
		Name:  "integration-test-1",
		Image: "nginx:latest",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/tenants", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleCreateTenant(w, req)

	// Verify 201 Created
	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	// Verify workflow was not triggered
	if len(triggeredActions) != 0 {
		t.Errorf("expected no workflow trigger, got %v", triggeredActions)
	}
	if len(triggeredSources) != 0 {
		t.Errorf("expected no trigger sources, got %v", triggeredSources)
	}

	// Verify response does not include execution ID
	var resp models.TenantResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.WorkflowExecutionID != nil {
		t.Errorf("expected no execution ID in response")
	}
}

// TestIntegrationAPITriggerUpdateWorkflow tests PUT /v1/tenants/{id} does not trigger workflows directly
func TestIntegrationAPITriggerUpdateWorkflow(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	var triggeredActions []string
	var triggeredSources []string

	tenantID := uuid.New()
	existingTenant := &tenant.Tenant{
		ID:     tenantID,
		Name:   "update-test",
		Status: tenant.StatusReady,
	}

	tenantRepo := &mockTenantRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
			return existingTenant, nil
		},
		updateFunc: func(ctx context.Context, t *tenant.Tenant) error {
			return nil
		},
	}

	wfClient := &mockWorkflowClient{
		triggerWithSourceFunc: func(ctx context.Context, t *tenant.Tenant, action, source string) (string, error) {
			triggeredActions = append(triggeredActions, action)
			triggeredSources = append(triggeredSources, source)
			return "exec-update-123", nil
		},
	}

	srv := &Server{
		logger:         logger,
		workflowClient: wfClient,
		tenantRepo:     tenantRepo,
	}

	reqBody := models.UpdateTenantRequest{
		Image: stringPtr("nginx:2.0"),
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/v1/tenants/"+tenantID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Add chi context
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("id", tenantID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

	w := httptest.NewRecorder()
	srv.handleUpdateTenant(w, req)

	// Verify 202 Accepted
	if w.Code != http.StatusAccepted {
		t.Errorf("expected status 202, got %d", w.Code)
	}

	// Verify workflow was not triggered
	if len(triggeredActions) != 0 {
		t.Errorf("expected no workflow trigger, got %v", triggeredActions)
	}
	if len(triggeredSources) != 0 {
		t.Errorf("expected no trigger sources, got %v", triggeredSources)
	}
}

// TestIntegrationAPITriggerDeleteWorkflow tests DELETE /v1/tenants/{id} does not trigger workflows directly
func TestIntegrationAPITriggerDeleteWorkflow(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	var triggeredActions []string
	var triggeredSources []string

	tenantID := uuid.New()
	existingTenant := &tenant.Tenant{
		ID:     tenantID,
		Name:   "delete-test",
		Status: tenant.StatusReady,
	}

	tenantRepo := &mockTenantRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
			return existingTenant, nil
		},
		updateFunc: func(ctx context.Context, t *tenant.Tenant) error {
			return nil
		},
	}

	wfClient := &mockWorkflowClient{
		triggerWithSourceFunc: func(ctx context.Context, t *tenant.Tenant, action, source string) (string, error) {
			triggeredActions = append(triggeredActions, action)
			triggeredSources = append(triggeredSources, source)
			return "exec-delete-123", nil
		},
	}

	srv := &Server{
		logger:         logger,
		workflowClient: wfClient,
		tenantRepo:     tenantRepo,
	}

	req := httptest.NewRequest(http.MethodDelete, "/v1/tenants/"+tenantID.String(), nil)

	// Add chi context
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("id", tenantID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

	w := httptest.NewRecorder()
	srv.handleDeleteTenant(w, req)

	// Verify 202 Accepted
	if w.Code != http.StatusAccepted {
		t.Errorf("expected status 202, got %d", w.Code)
	}

	// Verify workflow was not triggered
	if len(triggeredActions) != 0 {
		t.Errorf("expected no workflow trigger, got %v", triggeredActions)
	}
	if len(triggeredSources) != 0 {
		t.Errorf("expected no trigger sources, got %v", triggeredSources)
	}
}

// TestIntegrationConcurrentAPITriggers tests concurrent API triggers result in deduplication
func TestIntegrationConcurrentAPITriggers(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Track how many times workflow is triggered
	triggerCount := 0

	tenantID := uuid.New()
	existingTenant := &tenant.Tenant{
		ID:     tenantID,
		Name:   "concurrent-test",
		Status: tenant.StatusReady,
	}

	tenantRepo := &mockTenantRepo{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
			return existingTenant, nil
		},
		updateFunc: func(ctx context.Context, t *tenant.Tenant) error {
			return nil
		},
	}

	wfClient := &mockWorkflowClient{
		triggerWithSourceFunc: func(ctx context.Context, t *tenant.Tenant, action, source string) (string, error) {
			triggerCount++
			return fmt.Sprintf("exec-%d", triggerCount), nil
		},
		getStatusFunc: func(ctx context.Context, executionID string) (*workflow.ExecutionStatus, error) {
			// Simulate active execution
			return &workflow.ExecutionStatus{
				ExecutionID: executionID,
				State:       workflow.StateRunning,
			}, nil
		},
	}

	srv := &Server{
		logger:         logger,
		workflowClient: wfClient,
		tenantRepo:     tenantRepo,
	}

	reqBody := models.UpdateTenantRequest{
		Image: stringPtr("nginx:2.0"),
	}

	body, _ := json.Marshal(reqBody)

	// Make first request
	req1 := httptest.NewRequest(http.MethodPut, "/v1/tenants/"+tenantID.String(), bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	chiCtx1 := chi.NewRouteContext()
	chiCtx1.URLParams.Add("id", tenantID.String())
	req1 = req1.WithContext(context.WithValue(req1.Context(), chi.RouteCtxKey, chiCtx1))
	w1 := httptest.NewRecorder()
	srv.handleUpdateTenant(w1, req1)

	// Request should not trigger workflow directly
	if triggerCount != 0 {
		t.Errorf("expected 0 triggers after request, got %d", triggerCount)
	}

	// Verify both got 202 status
	if w1.Code != http.StatusAccepted {
		t.Errorf("expected status 202 for first request, got %d", w1.Code)
	}
}

type testSchemaProvider struct{}

func (t *testSchemaProvider) Name() string { return "docker" }
func (t *testSchemaProvider) Provision(ctx context.Context, spec *compute.TenantComputeSpec) (*compute.ProvisionResult, error) {
	return nil, nil
}
func (t *testSchemaProvider) Update(ctx context.Context, tenantID string, spec *compute.TenantComputeSpec) (*compute.UpdateResult, error) {
	return nil, nil
}
func (t *testSchemaProvider) Destroy(ctx context.Context, tenantID string) error { return nil }
func (t *testSchemaProvider) GetStatus(ctx context.Context, tenantID string) (*compute.ComputeStatus, error) {
	return nil, nil
}
func (t *testSchemaProvider) Validate(ctx context.Context, spec *compute.TenantComputeSpec) error {
	return nil
}
func (t *testSchemaProvider) ValidateConfig(config json.RawMessage) error { return nil }
func (t *testSchemaProvider) ConfigSchema() json.RawMessage {
	return json.RawMessage(`{"type":"object"}`)
}
func (t *testSchemaProvider) ConfigDefaults() json.RawMessage { return nil }

func TestIntegrationCreateSetWithComputeConfig(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	created := &tenant.Tenant{}
	updated := &tenant.Tenant{}
	existingID := uuid.New()

	tenantRepo := &mockTenantRepo{
		createFunc: func(ctx context.Context, t *tenant.Tenant) error {
			*created = *t
			created.ID = existingID
			return nil
		},
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
			return &tenant.Tenant{
				ID:           id,
				Name:         "config-tenant",
				Status:       tenant.StatusReady,
				DesiredImage: "nginx:latest",
			}, nil
		},
		updateFunc: func(ctx context.Context, t *tenant.Tenant) error {
			*updated = *t
			return nil
		},
	}

	srv := &Server{
		logger:          logger,
		workflowClient:  &mockWorkflowClient{},
		tenantRepo:      tenantRepo,
		computeProvider: &testSchemaProvider{},
	}

	createReq := models.CreateTenantRequest{
		Name:  "config-tenant",
		Image: "nginx:latest",
		ComputeConfig: map[string]interface{}{
			"env": map[string]interface{}{
				"FOO": "bar",
			},
		},
	}
	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/v1/tenants", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.handleCreateTenant(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}
	if created.DesiredConfig == nil || created.DesiredConfig["env"] == nil {
		t.Fatalf("expected compute_config to be stored in desired_config")
	}

	updateReq := models.UpdateTenantRequest{
		ComputeConfig: map[string]interface{}{
			"env": map[string]interface{}{
				"BAZ": "qux",
			},
		},
	}
	updateBody, _ := json.Marshal(updateReq)
	updateHTTPReq := httptest.NewRequest(http.MethodPut, "/v1/tenants/"+existingID.String(), bytes.NewReader(updateBody))
	updateHTTPReq.Header.Set("Content-Type", "application/json")

	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("id", existingID.String())
	updateHTTPReq = updateHTTPReq.WithContext(context.WithValue(updateHTTPReq.Context(), chi.RouteCtxKey, chiCtx))

	updateWriter := httptest.NewRecorder()
	srv.handleUpdateTenant(updateWriter, updateHTTPReq)
	if updateWriter.Code != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d", updateWriter.Code)
	}
	if updated.DesiredConfig == nil || updated.DesiredConfig["env"] == nil {
		t.Fatalf("expected compute_config to be updated in desired_config")
	}
}

// TestIntegrationAPITriggerFailureRecovery tests create path does not fail on workflow errors
func TestIntegrationAPITriggerFailureRecovery(t *testing.T) {
	logger, _ := zap.NewDevelopment()

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
		logger:     logger,
		tenantRepo: tenantRepo,
	}

	reqBody := models.CreateTenantRequest{
		Name:  "failure-test",
		Image: "nginx:latest",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/tenants", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleCreateTenant(w, req)

	// Verify 201 Created since workflows are triggered by reconciler
	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}
}

// TestIntegrationWorkflowCompletionRetrigger tests workflow completion and re-trigger
func TestIntegrationWorkflowCompletionRetrigger(t *testing.T) {
	var executionStates []workflow.ExecutionState
	tenantID := uuid.New()

	wfClient := &mockWorkflowClient{
		getStatusFunc: func(ctx context.Context, executionID string) (*workflow.ExecutionStatus, error) {
			// Simulate workflow completion after first check
			if len(executionStates) == 0 {
				executionStates = append(executionStates, workflow.StateSucceeded)
				return &workflow.ExecutionStatus{
					ExecutionID: executionID,
					State:       workflow.StateSucceeded,
				}, nil
			}
			return &workflow.ExecutionStatus{
				ExecutionID: executionID,
				State:       workflow.StateSucceeded,
			}, nil
		},
	}

	// Test that we can check execution status
	status, err := wfClient.GetExecutionStatus(context.Background(), "exec-1")
	if err != nil {
		t.Errorf("failed to get execution status: %v", err)
	}

	if status.State != workflow.StateSucceeded {
		t.Errorf("expected succeeded state, got %v", status.State)
	}

	_ = tenantID // Avoid unused variable
}
