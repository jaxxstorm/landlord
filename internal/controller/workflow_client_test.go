package controller

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jaxxstorm/landlord/internal/compute"
	"github.com/jaxxstorm/landlord/internal/tenant"
	"github.com/jaxxstorm/landlord/internal/workflow"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type captureWorkflowProvider struct {
	name              string
	invokedWorkflow   string
	lastProvisionReq  *workflow.ProvisionRequest
	provisionRequests []*workflow.ProvisionRequest
}

func (p *captureWorkflowProvider) Name() string { return p.name }

func (p *captureWorkflowProvider) Invoke(ctx context.Context, workflowID string, request *workflow.ProvisionRequest) (*workflow.ExecutionResult, error) {
	_ = ctx
	p.invokedWorkflow = workflowID
	p.lastProvisionReq = request
	p.provisionRequests = append(p.provisionRequests, request)
	return &workflow.ExecutionResult{
		ExecutionID:  "exec-1",
		WorkflowID:   workflowID,
		ProviderType: p.name,
		State:        workflow.StateRunning,
		StartedAt:    time.Now(),
	}, nil
}

func (p *captureWorkflowProvider) GetWorkflowStatus(ctx context.Context, executionID string) (*workflow.WorkflowStatus, error) {
	_ = ctx
	_ = executionID
	return &workflow.WorkflowStatus{ExecutionID: executionID, State: workflow.StateRunning}, nil
}
func (p *captureWorkflowProvider) CreateWorkflow(ctx context.Context, spec *workflow.WorkflowSpec) (*workflow.CreateWorkflowResult, error) {
	_ = ctx
	return &workflow.CreateWorkflowResult{
		WorkflowID:   spec.WorkflowID,
		ProviderType: p.name,
		CreatedAt:    time.Now(),
	}, nil
}
func (p *captureWorkflowProvider) StartExecution(ctx context.Context, workflowID string, input *workflow.ExecutionInput) (*workflow.ExecutionResult, error) {
	_ = ctx
	_ = input
	return &workflow.ExecutionResult{ExecutionID: "exec-1", WorkflowID: workflowID, ProviderType: p.name, State: workflow.StateRunning, StartedAt: time.Now()}, nil
}
func (p *captureWorkflowProvider) GetExecutionStatus(ctx context.Context, executionID string) (*workflow.ExecutionStatus, error) {
	_ = ctx
	return &workflow.ExecutionStatus{ExecutionID: executionID, ProviderType: p.name, State: workflow.StateRunning, StartTime: time.Now()}, nil
}
func (p *captureWorkflowProvider) StopExecution(ctx context.Context, executionID string, reason string) error {
	_ = ctx
	_ = executionID
	_ = reason
	return nil
}
func (p *captureWorkflowProvider) DeleteWorkflow(ctx context.Context, workflowID string) error {
	_ = ctx
	_ = workflowID
	return nil
}
func (p *captureWorkflowProvider) Validate(ctx context.Context, spec *workflow.WorkflowSpec) error {
	_ = ctx
	_ = spec
	return nil
}
func (p *captureWorkflowProvider) PostComputeCallback(ctx context.Context, executionID string, payload *compute.CallbackPayload, opts *compute.CallbackOptions) error {
	_ = ctx
	_ = executionID
	_ = payload
	_ = opts
	return nil
}

func newTestWorkflowClient() *WorkflowClient {
	logger, _ := zap.NewDevelopment()
	return &WorkflowClient{
		manager: nil,
		logger:  logger,
		timeout: 5 * time.Second,
	}
}

func TestDetermineAction_RequestedStatus(t *testing.T) {
	wc := newTestWorkflowClient()
	action, err := wc.DetermineAction(tenant.StatusRequested)
	if err != nil {
		t.Errorf("DetermineAction() error = %v, want nil", err)
	}
	if action != "provision" {
		t.Errorf("DetermineAction() = %s, want provision", action)
	}
}

func TestDetermineAction_PlanningStatus(t *testing.T) {
	wc := newTestWorkflowClient()
	action, err := wc.DetermineAction(tenant.StatusPlanning)
	if err != nil {
		t.Errorf("DetermineAction() error = %v, want nil", err)
	}
	if action != "provision" {
		t.Errorf("DetermineAction() = %s, want provision", action)
	}
}

func TestDetermineAction_ProvisioningStatus(t *testing.T) {
	wc := newTestWorkflowClient()
	action, err := wc.DetermineAction(tenant.StatusProvisioning)
	if err != nil {
		t.Errorf("DetermineAction() error = %v, want nil", err)
	}
	if action != "provision" {
		t.Errorf("DetermineAction() = %s, want provision", action)
	}
}

func TestDetermineAction_UpdatingStatus(t *testing.T) {
	wc := newTestWorkflowClient()
	action, err := wc.DetermineAction(tenant.StatusUpdating)
	if err != nil {
		t.Errorf("DetermineAction() error = %v, want nil", err)
	}
	if action != "update" {
		t.Errorf("DetermineAction() = %s, want update", action)
	}
}

func TestDetermineAction_DeletingStatus(t *testing.T) {
	wc := newTestWorkflowClient()
	action, err := wc.DetermineAction(tenant.StatusDeleting)
	if err != nil {
		t.Errorf("DetermineAction() error = %v, want nil", err)
	}
	if action != "delete" {
		t.Errorf("DetermineAction() = %s, want delete", action)
	}
}

func TestDetermineAction_TerminalStatus(t *testing.T) {
	tests := []struct {
		name       string
		status     tenant.Status
		wantErr    bool
		wantAction string
	}{
		{"ready status", tenant.StatusReady, true, ""},
		{"failed status", tenant.StatusFailed, true, ""},
		{"archived status", tenant.StatusArchived, true, ""},
	}

	wc := newTestWorkflowClient()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, err := wc.DetermineAction(tt.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetermineAction() error = %v, wantErr %v", err, tt.wantErr)
			}
			if action != tt.wantAction {
				t.Errorf("DetermineAction() = %s, want %s", action, tt.wantAction)
			}
		})
	}
}

func TestDetermineAction_UnknownStatus(t *testing.T) {
	wc := newTestWorkflowClient()
	action, err := wc.DetermineAction(tenant.Status("unknown"))
	if err == nil {
		t.Error("DetermineAction() error = nil, want error for unknown status")
	}
	if action != "" {
		t.Errorf("DetermineAction() = %s, want empty string", action)
	}
}

func TestIsRetryableError_NilError(t *testing.T) {
	if IsRetryableError(nil) {
		t.Error("IsRetryableError(nil) = true, want false")
	}
}

func TestIsRetryableError_ContextDeadlineExceeded(t *testing.T) {
	err := context.DeadlineExceeded
	if !IsRetryableError(err) {
		t.Error("IsRetryableError(DeadlineExceeded) = false, want true")
	}
}

func TestIsRetryableError_ContextCanceled(t *testing.T) {
	err := context.Canceled
	if IsRetryableError(err) {
		t.Error("IsRetryableError(Canceled) = true, want false")
	}
}

func TestIsRetryableError_GenericError(t *testing.T) {
	err := errors.New("some error")
	if !IsRetryableError(err) {
		t.Error("IsRetryableError(generic error) = false, want true (defaults to retryable)")
	}
}

func TestDetermineAction_AllNonTerminalStates(t *testing.T) {
	wc := newTestWorkflowClient()

	tests := []struct {
		status         tenant.Status
		expectedAction string
	}{
		{tenant.StatusRequested, "provision"},
		{tenant.StatusPlanning, "provision"},
		{tenant.StatusProvisioning, "provision"},
		{tenant.StatusUpdating, "update"},
		{tenant.StatusDeleting, "delete"},
	}

	for _, tt := range tests {
		action, err := wc.DetermineAction(tt.status)
		if err != nil {
			t.Errorf("DetermineAction(%s) error = %v", tt.status, err)
		}
		if action != tt.expectedAction {
			t.Errorf("DetermineAction(%s) = %s, want %s", tt.status, action, tt.expectedAction)
		}
	}
}

func TestIsRetryableError_MultipleErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"deadline exceeded", context.DeadlineExceeded, true},
		{"canceled", context.Canceled, false},
		{"generic error", errors.New("test"), true},
		{"wrapped error", errors.New("wrapped: test"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRetryableError(tt.err); got != tt.want {
				t.Errorf("IsRetryableError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTriggerWorkflow_ComputesConfigHash(t *testing.T) {
	// This test verifies that config hash is computed when triggering workflow
	// The hash computation itself is tested in tenant package

	testTenant := &tenant.Tenant{
		Name:   "test-tenant",
		Status: tenant.StatusRequested,
		DesiredConfig: map[string]interface{}{
			"image": "nginx:1.25",
			"env": map[string]string{
				"FOO": "bar",
			},
		},
	}

	// Compute expected hash
	expectedHash, err := tenant.ComputeConfigHash(testTenant.DesiredConfig)
	if err != nil {
		t.Fatalf("Failed to compute expected hash: %v", err)
	}

	if expectedHash == "" {
		t.Error("Expected non-empty config hash for non-empty config")
	}

	// Verify hash is deterministic
	hash2, err := tenant.ComputeConfigHash(testTenant.DesiredConfig)
	if err != nil {
		t.Fatalf("Failed to compute second hash: %v", err)
	}

	if expectedHash != hash2 {
		t.Errorf("Config hash not deterministic: %s != %s", expectedHash, hash2)
	}
}

func TestTriggerWorkflow_UsesStaticWorkflowIDForSharedWorkflowProviders(t *testing.T) {
	for _, providerName := range []string{"temporal", "step-functions"} {
		t.Run(providerName, func(t *testing.T) {
			logger, _ := zap.NewDevelopment()
			provider := &captureWorkflowProvider{name: providerName}
			registry := workflow.NewRegistry(logger)
			require.NoError(t, registry.Register(provider))

			manager := workflow.New(registry, logger)
			wc := NewWorkflowClient(manager, logger, 5*time.Second, providerName)

			testTenant := &tenant.Tenant{
				ID:     uuid.New(),
				Name:   "shared-workflow-app",
				Status: tenant.StatusRequested,
				DesiredConfig: map[string]interface{}{
					"image":            "nginx:latest",
					"compute_provider": "ecs",
				},
			}

			_, err := wc.TriggerWorkflow(context.Background(), testTenant, "provision")
			require.NoError(t, err)
			require.Equal(t, tenantProvisioningWorkflowID, provider.invokedWorkflow)
			require.Equal(t, testTenant.ID.String(), provider.lastProvisionReq.TenantUUID)
			require.Equal(t, "provision", provider.lastProvisionReq.Operation)
			require.Equal(t, "ecs", provider.lastProvisionReq.ComputeProvider)
			require.NotEmpty(t, provider.lastProvisionReq.Metadata["config_hash"])
		})
	}
}

func TestTriggerWorkflow_UsesPerTenantWorkflowIDForNonTemporalProviders(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	provider := &captureWorkflowProvider{name: "mock"}
	registry := workflow.NewRegistry(logger)
	require.NoError(t, registry.Register(provider))

	manager := workflow.New(registry, logger)
	wc := NewWorkflowClient(manager, logger, 5*time.Second, "mock")

	tenantID := uuid.New()
	testTenant := &tenant.Tenant{
		ID:     tenantID,
		Name:   "mock-app",
		Status: tenant.StatusRequested,
		DesiredConfig: map[string]interface{}{
			"image": "nginx:latest",
		},
	}

	_, err := wc.TriggerWorkflow(context.Background(), testTenant, "provision")
	require.NoError(t, err)
	require.Equal(t, "tenant-"+tenantID.String()+"-provision", provider.invokedWorkflow)
}

func TestTriggerWorkflow_LifecycleRevisionChangesWithDesiredConfig(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	provider := &captureWorkflowProvider{name: "step-functions"}
	registry := workflow.NewRegistry(logger)
	require.NoError(t, registry.Register(provider))

	wc := NewWorkflowClient(workflow.New(registry, logger), logger, 5*time.Second, "step-functions")
	testTenant := &tenant.Tenant{
		ID:   uuid.New(),
		Name: "revision-app",
		DesiredConfig: map[string]interface{}{
			"image": "nginx:1.25",
		},
	}

	_, err := wc.TriggerWorkflow(context.Background(), testTenant, "provision")
	require.NoError(t, err)
	_, err = wc.TriggerWorkflow(context.Background(), testTenant, "provision")
	require.NoError(t, err)
	require.Len(t, provider.provisionRequests, 2)
	require.Equal(t, provider.provisionRequests[0].Metadata["config_hash"], provider.provisionRequests[1].Metadata["config_hash"])

	testTenant.DesiredConfig["image"] = "nginx:1.26"
	_, err = wc.TriggerWorkflow(context.Background(), testTenant, "update")
	require.NoError(t, err)
	require.Len(t, provider.provisionRequests, 3)
	require.NotEqual(t, provider.provisionRequests[0].Metadata["config_hash"], provider.provisionRequests[2].Metadata["config_hash"])
}

const tenantProvisioningWorkflowID = "tenant-provisioning"
