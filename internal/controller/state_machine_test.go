package controller

import (
	"testing"

	"github.com/jaxxstorm/landlord/internal/tenant"
)

func TestNextStatus(t *testing.T) {
	tests := []struct {
		name        string
		current     tenant.Status
		wantNext    tenant.Status
		wantErr     bool
		errContains string
	}{
		{"requested to provisioning", tenant.StatusRequested, tenant.StatusProvisioning, false, ""},
		{"planning to provisioning", tenant.StatusPlanning, tenant.StatusProvisioning, false, ""},
		{"provisioning to ready", tenant.StatusProvisioning, tenant.StatusReady, false, ""},
		{"updating to ready", tenant.StatusUpdating, tenant.StatusReady, false, ""},
		{"deleting to archived", tenant.StatusDeleting, tenant.StatusArchived, false, ""},
		{"archiving to archived", tenant.StatusArchiving, tenant.StatusArchived, false, ""},
		{"ready is terminal", tenant.StatusReady, "", true, "terminal state"},
		{"archived is terminal", tenant.StatusArchived, "", true, "terminal state"},
		{"failed is terminal", tenant.StatusFailed, "", true, "terminal state"},
		{"unknown status", tenant.Status("unknown"), "", true, "unknown status"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := nextStatus(tt.current)
			if (err != nil) != tt.wantErr {
				t.Errorf("nextStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.wantNext {
				t.Errorf("nextStatus() = %v, want %v", got, tt.wantNext)
			}
		})
	}
}

func TestShouldReconcile(t *testing.T) {
	tests := []struct {
		name   string
		status tenant.Status
		want   bool
	}{
		{"requested", tenant.StatusRequested, true},
		{"planning", tenant.StatusPlanning, true},
		{"provisioning", tenant.StatusProvisioning, true},
		{"updating", tenant.StatusUpdating, true},
		{"deleting", tenant.StatusDeleting, true},
		{"archiving", tenant.StatusArchiving, true},
		{"ready", tenant.StatusReady, false},
		{"archived", tenant.StatusArchived, false},
		{"failed", tenant.StatusFailed, false},
		{"unknown", tenant.Status("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldReconcile(tt.status); got != tt.want {
				t.Errorf("shouldReconcile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateTransition(t *testing.T) {
	tests := []struct {
		name    string
		from    tenant.Status
		to      tenant.Status
		wantErr bool
	}{
		{"req to prov", tenant.StatusRequested, tenant.StatusProvisioning, false},
		{"req to fail", tenant.StatusRequested, tenant.StatusFailed, false},
		{"req to plan", tenant.StatusRequested, tenant.StatusPlanning, true},
		{"plan to prov", tenant.StatusPlanning, tenant.StatusProvisioning, false},
		{"plan to fail", tenant.StatusPlanning, tenant.StatusFailed, false},
		{"prov to ready", tenant.StatusProvisioning, tenant.StatusReady, false},
		{"ready to upd", tenant.StatusReady, tenant.StatusUpdating, false},
		{"ready to del", tenant.StatusReady, tenant.StatusDeleting, false},
		{"ready to archive", tenant.StatusReady, tenant.StatusArchiving, false},
		{"req to ready", tenant.StatusRequested, tenant.StatusReady, true},
		{"plan to ready", tenant.StatusPlanning, tenant.StatusReady, true},
		{"deleting to archived", tenant.StatusDeleting, tenant.StatusArchived, false},
		{"archiving to archived", tenant.StatusArchiving, tenant.StatusArchived, false},
		{"failed to archive", tenant.StatusFailed, tenant.StatusArchiving, false},
		{"failed no trans", tenant.StatusFailed, tenant.StatusRequested, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTransition(tt.from, tt.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTransition() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
