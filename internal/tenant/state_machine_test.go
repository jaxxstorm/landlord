package tenant

import (
	"testing"
)

func TestNextStatus(t *testing.T) {
	tests := []struct {
		name        string
		current     Status
		expected    Status
		expectError bool
	}{
		{
			name:        "requested to provisioning",
			current:     StatusRequested,
			expected:    StatusProvisioning,
			expectError: false,
		},
		{
			name:        "planning to provisioning",
			current:     StatusPlanning,
			expected:    StatusProvisioning,
			expectError: false,
		},
		{
			name:        "provisioning to ready",
			current:     StatusProvisioning,
			expected:    StatusReady,
			expectError: false,
		},
		{
			name:        "updating to ready",
			current:     StatusUpdating,
			expected:    StatusReady,
			expectError: false,
		},
		{
			name:        "deleting to archived",
			current:     StatusDeleting,
			expected:    StatusArchived,
			expectError: false,
		},
		{
			name:        "archiving to archived",
			current:     StatusArchiving,
			expected:    StatusArchived,
			expectError: false,
		},
		{
			name:        "ready is terminal",
			current:     StatusReady,
			expected:    "",
			expectError: true,
		},
		{
			name:        "archived is terminal",
			current:     StatusArchived,
			expected:    "",
			expectError: true,
		},
		{
			name:        "failed is terminal",
			current:     StatusFailed,
			expected:    "",
			expectError: true,
		},
		{
			name:        "unknown status",
			current:     Status("unknown"),
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NextStatus(tt.current)
			if tt.expectError {
				if err == nil {
					t.Errorf("NextStatus() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("NextStatus() unexpected error: %v", err)
				}
				if got != tt.expected {
					t.Errorf("NextStatus() = %v, want %v", got, tt.expected)
				}
			}
		})
	}
}

func TestShouldReconcile(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
		expected bool
	}{
		{
			name:     "requested requires reconciliation",
			status:   StatusRequested,
			expected: true,
		},
		{
			name:     "planning requires reconciliation",
			status:   StatusPlanning,
			expected: true,
		},
		{
			name:     "provisioning requires reconciliation",
			status:   StatusProvisioning,
			expected: true,
		},
		{
			name:     "updating requires reconciliation",
			status:   StatusUpdating,
			expected: true,
		},
		{
			name:     "deleting requires reconciliation",
			status:   StatusDeleting,
			expected: true,
		},
		{
			name:     "archiving requires reconciliation",
			status:   StatusArchiving,
			expected: true,
		},
		{
			name:     "archived does not require reconciliation",
			status:   StatusArchived,
			expected: false,
		},
		{
			name:     "ready does not require reconciliation",
			status:   StatusReady,
			expected: false,
		},
		{
			name:     "failed does not require reconciliation",
			status:   StatusFailed,
			expected: false,
		},
		{
			name:     "unknown status does not require reconciliation",
			status:   Status("unknown"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldReconcile(tt.status)
			if got != tt.expected {
				t.Errorf("ShouldReconcile() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsTerminalStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
		expected bool
	}{
		{
			name:     "ready is terminal",
			status:   StatusReady,
			expected: true,
		},
		{
			name:     "archived is terminal",
			status:   StatusArchived,
			expected: true,
		},
		{
			name:     "failed is terminal",
			status:   StatusFailed,
			expected: true,
		},
		{
			name:     "requested is not terminal",
			status:   StatusRequested,
			expected: false,
		},
		{
			name:     "planning is not terminal",
			status:   StatusPlanning,
			expected: false,
		},
		{
			name:     "provisioning is not terminal",
			status:   StatusProvisioning,
			expected: false,
		},
		{
			name:     "updating is not terminal",
			status:   StatusUpdating,
			expected: false,
		},
		{
			name:     "deleting is not terminal",
			status:   StatusDeleting,
			expected: false,
		},
		{
			name:     "archiving is not terminal",
			status:   StatusArchiving,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsTerminalStatus(tt.status)
			if got != tt.expected {
				t.Errorf("IsTerminalStatus() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValidateTransition(t *testing.T) {
	tests := []struct {
		name        string
		from        Status
		to          Status
		expectError bool
	}{
		// Valid transitions
		{
			name:        "requested to provisioning",
			from:        StatusRequested,
			to:          StatusProvisioning,
			expectError: false,
		},
		{
			name:        "requested to failed",
			from:        StatusRequested,
			to:          StatusFailed,
			expectError: false,
		},
		{
			name:        "planning to provisioning",
			from:        StatusPlanning,
			to:          StatusProvisioning,
			expectError: false,
		},
		{
			name:        "planning to failed",
			from:        StatusPlanning,
			to:          StatusFailed,
			expectError: false,
		},
		{
			name:        "provisioning to ready",
			from:        StatusProvisioning,
			to:          StatusReady,
			expectError: false,
		},
		{
			name:        "provisioning to failed",
			from:        StatusProvisioning,
			to:          StatusFailed,
			expectError: false,
		},
		{
			name:        "ready to updating",
			from:        StatusReady,
			to:          StatusUpdating,
			expectError: false,
		},
		{
			name:        "ready to deleting",
			from:        StatusReady,
			to:          StatusDeleting,
			expectError: false,
		},
		{
			name:        "ready to archiving",
			from:        StatusReady,
			to:          StatusArchiving,
			expectError: false,
		},
		{
			name:        "updating to ready",
			from:        StatusUpdating,
			to:          StatusReady,
			expectError: false,
		},
		{
			name:        "updating to failed",
			from:        StatusUpdating,
			to:          StatusFailed,
			expectError: false,
		},
		{
			name:        "deleting to failed",
			from:        StatusDeleting,
			to:          StatusFailed,
			expectError: false,
		},
		{
			name:        "archiving to archived",
			from:        StatusArchiving,
			to:          StatusArchived,
			expectError: false,
		},
		{
			name:        "failed to archiving",
			from:        StatusFailed,
			to:          StatusArchiving,
			expectError: false,
		},
		{
			name:        "archiving to failed",
			from:        StatusArchiving,
			to:          StatusFailed,
			expectError: false,
		},
		// Invalid transitions
		{
			name:        "requested to provisioning",
			from:        StatusRequested,
			to:          StatusProvisioning,
			expectError: false,
		},
		{
			name:        "requested to planning (no longer valid)",
			from:        StatusRequested,
			to:          StatusPlanning,
			expectError: true,
		},
		{
			name:        "ready to planning (cannot go back)",
			from:        StatusReady,
			to:          StatusPlanning,
			expectError: true,
		},
		{
			name:        "failed to ready (terminal state)",
			from:        StatusFailed,
			to:          StatusReady,
			expectError: true,
		},
		{
			name:        "failed to planning (terminal state)",
			from:        StatusFailed,
			to:          StatusPlanning,
			expectError: true,
		},
		{
			name:        "provisioning to updating (must go to ready first)",
			from:        StatusProvisioning,
			to:          StatusUpdating,
			expectError: true,
		},
		{
			name:        "unknown source status",
			from:        Status("unknown"),
			to:          StatusReady,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTransition(tt.from, tt.to)
			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateTransition() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ValidateTransition() unexpected error: %v", err)
				}
			}
		})
	}
}
