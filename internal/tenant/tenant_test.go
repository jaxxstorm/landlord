package tenant

import (
	"testing"

	"github.com/google/uuid"
)

func TestStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{"requested", StatusRequested, true},
		{"planning", StatusPlanning, true},
		{"provisioning", StatusProvisioning, true},
		{"ready", StatusReady, true},
		{"updating", StatusUpdating, true},
		{"deleting", StatusDeleting, true},
		{"archived", StatusArchived, true},
		{"failed", StatusFailed, true},
		{"invalid", Status("invalid"), false},
		{"empty", Status(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("Status.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatus_IsTerminal(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{"archived is terminal", StatusArchived, true},
		{"ready is not terminal", StatusReady, false},
		{"failed is not terminal", StatusFailed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsTerminal(); got != tt.want {
				t.Errorf("Status.IsTerminal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatus_IsHealthy(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{"ready is healthy", StatusReady, true},
		{"provisioning is not healthy", StatusProvisioning, false},
		{"failed is not healthy", StatusFailed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsHealthy(); got != tt.want {
				t.Errorf("Status.IsHealthy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatus_CanTransition(t *testing.T) {
	tests := []struct {
		name string
		from Status
		to   Status
		want bool
	}{
		{"requested -> provisioning", StatusRequested, StatusProvisioning, true},
		{"requested -> failed", StatusRequested, StatusFailed, true},
		{"requested -> ready (invalid)", StatusRequested, StatusReady, false},
		{"ready -> updating", StatusReady, StatusUpdating, true},
		{"ready -> deleting", StatusReady, StatusDeleting, true},
		{"ready -> archiving", StatusReady, StatusArchiving, true},
		{"archived -> anything (invalid)", StatusArchived, StatusReady, false},
		{"failed -> deleting", StatusFailed, StatusDeleting, true},
		{"failed -> archiving", StatusFailed, StatusArchiving, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.from.CanTransition(tt.to); got != tt.want {
				t.Errorf("Status.CanTransition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTenant_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tenant  *Tenant
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid tenant",
			tenant: &Tenant{
				ID:            uuid.New(),
				Name:          "valid-tenant-id",
				Status:        StatusRequested,
				DesiredImage:  "docker.io/app:v1.0.0",
				DesiredConfig: map[string]interface{}{"replicas": "2"},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			tenant: &Tenant{
				ID:           uuid.New(),
				Status:       StatusRequested,
				DesiredImage: "docker.io/app:v1.0.0",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "name too long",
			tenant: &Tenant{
				ID:           uuid.New(),
				Name:         string(make([]byte, 256)),
				Status:       StatusRequested,
				DesiredImage: "docker.io/app:v1.0.0",
			},
			wantErr: true,
			errMsg:  "name must be <= 255 characters",
		},
		{
			name: "invalid name format",
			tenant: &Tenant{
				ID:           uuid.New(),
				Name:         "Invalid_Tenant_ID",
				Status:       StatusRequested,
				DesiredImage: "docker.io/app:v1.0.0",
			},
			wantErr: true,
			errMsg:  "name must be lowercase alphanumeric with hyphens",
		},
		{
			name: "missing desired_image",
			tenant: &Tenant{
				ID:     uuid.New(),
				Name:   "valid-tenant",
				Status: StatusRequested,
			},
			wantErr: true,
			errMsg:  "desired_image is required",
		},
		{
			name: "missing status",
			tenant: &Tenant{
				ID:           uuid.New(),
				Name:         "valid-tenant",
				DesiredImage: "docker.io/app:v1.0.0",
			},
			wantErr: true,
			errMsg:  "status is required",
		},
		{
			name: "invalid status",
			tenant: &Tenant{
				ID:           uuid.New(),
				Name:         "valid-tenant",
				Status:       Status("invalid"),
				DesiredImage: "docker.io/app:v1.0.0",
			},
			wantErr: true,
			errMsg:  "invalid status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.tenant.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Tenant.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg && len(tt.errMsg) > 0 {
					// Check if error message contains expected substring
					contains := false
					if len(err.Error()) >= len(tt.errMsg) {
						for i := 0; i <= len(err.Error())-len(tt.errMsg); i++ {
							if err.Error()[i:i+len(tt.errMsg)] == tt.errMsg {
								contains = true
								break
							}
						}
					}
					// Also check reverse
					if !contains && len(tt.errMsg) >= len(err.Error()) {
						for i := 0; i <= len(tt.errMsg)-len(err.Error()); i++ {
							if tt.errMsg[i:i+len(err.Error())] == err.Error() {
								contains = true
								break
							}
						}
					}
					if !contains {
						t.Errorf("Tenant.Validate() error message = %q, want substring %q", err.Error(), tt.errMsg)
					}
				}
			}
		})
	}
}

func TestTenant_IsArchived(t *testing.T) {
	tests := []struct {
		name   string
		tenant *Tenant
		want   bool
	}{
		{
			name: "not archived",
			tenant: &Tenant{
				ID:     uuid.New(),
				Name:   "test",
				Status: StatusReady,
			},
			want: false,
		},
		{
			name: "archived",
			tenant: &Tenant{
				ID:     uuid.New(),
				Name:   "test",
				Status: StatusArchived,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tenant.IsArchived(); got != tt.want {
				t.Errorf("Tenant.IsArchived() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTenant_IsDrifted(t *testing.T) {
	tests := []struct {
		name   string
		tenant *Tenant
		want   bool
	}{
		{
			name: "not ready - no drift",
			tenant: &Tenant{
				ID:            uuid.New(),
				Status:        StatusProvisioning,
				DesiredImage:  "app:v1",
				ObservedImage: "app:v2",
			},
			want: false,
		},
		{
			name: "ready and in sync",
			tenant: &Tenant{
				ID:            uuid.New(),
				Status:        StatusReady,
				DesiredImage:  "app:v1",
				ObservedImage: "app:v1",
			},
			want: false,
		},
		{
			name: "ready and drifted",
			tenant: &Tenant{
				ID:            uuid.New(),
				Status:        StatusReady,
				DesiredImage:  "app:v2",
				ObservedImage: "app:v1",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tenant.IsDrifted(); got != tt.want {
				t.Errorf("Tenant.IsDrifted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTenant_Clone(t *testing.T) {
	original := &Tenant{
		ID:           uuid.New(),
		Name:         "test-tenant",
		Status:       StatusReady,
		DesiredImage: "app:v1",
		Labels: map[string]string{
			"env": "prod",
		},
		Annotations: map[string]string{
			"owner": "team-a",
		},
	}

	clone := original.Clone()

	// Verify clone is equal but separate
	if clone.ID != original.ID {
		t.Error("Clone ID mismatch")
	}
	if clone.Name != original.Name {
		t.Error("Clone Name mismatch")
	}

	// Modify clone and verify original unchanged
	clone.Labels["env"] = "dev"
	if original.Labels["env"] != "prod" {
		t.Error("Modifying clone Labels affected original")
	}

	clone.Annotations["owner"] = "team-b"
	if original.Annotations["owner"] != "team-a" {
		t.Error("Modifying clone Annotations affected original")
	}
}

func TestStateTransition_Validate(t *testing.T) {
	tenantID := uuid.New()
	fromStatus := StatusRequested

	tests := []struct {
		name       string
		transition *StateTransition
		wantErr    bool
		errMsg     string
	}{
		{
			name: "valid transition",
			transition: &StateTransition{
				ID:         uuid.New(),
				TenantID:   tenantID,
				FromStatus: &fromStatus,
				ToStatus:   StatusProvisioning,
				Reason:     "User requested",
			},
			wantErr: false,
		},
		{
			name: "missing tenant_id",
			transition: &StateTransition{
				ID:       uuid.New(),
				ToStatus: StatusProvisioning,
				Reason:   "Test",
			},
			wantErr: true,
			errMsg:  "tenant_id is required",
		},
		{
			name: "missing to_status",
			transition: &StateTransition{
				ID:       uuid.New(),
				TenantID: tenantID,
				Reason:   "Test",
			},
			wantErr: true,
			errMsg:  "to_status is required",
		},
		{
			name: "invalid to_status",
			transition: &StateTransition{
				ID:       uuid.New(),
				TenantID: tenantID,
				ToStatus: Status("invalid"),
				Reason:   "Test",
			},
			wantErr: true,
			errMsg:  "invalid to_status",
		},
		{
			name: "missing reason",
			transition: &StateTransition{
				ID:       uuid.New(),
				TenantID: tenantID,
				ToStatus: StatusProvisioning,
			},
			wantErr: true,
			errMsg:  "reason is required",
		},
		{
			name: "invalid transition",
			transition: &StateTransition{
				ID:         uuid.New(),
				TenantID:   tenantID,
				FromStatus: &fromStatus,
				ToStatus:   StatusReady, // Can't go directly from requested to ready
				Reason:     "Test",
			},
			wantErr: true,
			errMsg:  "invalid transition",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.transition.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("StateTransition.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				// Check if error message contains expected substring
				contains := false
				errStr := err.Error()
				if len(errStr) >= len(tt.errMsg) {
					for i := 0; i <= len(errStr)-len(tt.errMsg); i++ {
						if errStr[i:i+len(tt.errMsg)] == tt.errMsg {
							contains = true
							break
						}
					}
				}
				if !contains && len(tt.errMsg) >= len(errStr) {
					for i := 0; i <= len(tt.errMsg)-len(errStr); i++ {
						if tt.errMsg[i:i+len(errStr)] == errStr {
							contains = true
							break
						}
					}
				}
				if !contains {
					t.Errorf("StateTransition.Validate() error = %q, want substring %q", errStr, tt.errMsg)
				}
			}
		})
	}
}

func TestNewStateTransition(t *testing.T) {
	tenant := &Tenant{
		ID:             uuid.New(),
		Name:           "test-tenant",
		Status:         StatusRequested,
		DesiredConfig:  map[string]interface{}{"replicas": "2"},
		ObservedConfig: map[string]interface{}{"replicas": "1"},
	}

	transition := NewStateTransition(tenant, StatusProvisioning, "Starting provisioning", "user@example.com")

	if transition.ID == uuid.Nil {
		t.Error("Transition ID should be generated")
	}
	if transition.TenantID != tenant.ID {
		t.Error("TenantID mismatch")
	}
	if transition.ToStatus != StatusProvisioning {
		t.Error("ToStatus mismatch")
	}
	if transition.Reason != "Starting provisioning" {
		t.Error("Reason mismatch")
	}
	if transition.TriggeredBy != "user@example.com" {
		t.Error("TriggeredBy mismatch")
	}
	if transition.FromStatus == nil || *transition.FromStatus != StatusRequested {
		t.Error("FromStatus should be set to current status")
	}
	if transition.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}
