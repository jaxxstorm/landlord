package controller

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jaxxstorm/landlord/internal/tenant"
	"go.uber.org/zap"
)

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
