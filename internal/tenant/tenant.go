package tenant

import (
	"fmt"
	"reflect"
	"regexp"
	"time"

	"github.com/google/uuid"
)

// tenantNamePattern validates that tenant name is lowercase alphanumeric with hyphens
var tenantNamePattern = regexp.MustCompile(`^[a-z0-9-]+$`)

// Status represents a tenant's position in its lifecycle
type Status string

const (
	// StatusRequested: Tenant creation requested, not yet validated
	// Initial state when tenant is first created via API
	// Next states: StatusPlanning, StatusFailed
	StatusRequested Status = "requested"

	// StatusPlanning: System is computing required actions (plan generation)
	// Comparing desired vs observed state, determining what to provision
	// Next states: StatusProvisioning, StatusFailed
	StatusPlanning Status = "planning"

	// StatusProvisioning: Resources are being created
	// Workflow is executing, compute/networking being provisioned
	// Next states: StatusReady, StatusFailed
	StatusProvisioning Status = "provisioning"

	// StatusReady: Tenant is fully operational and serving traffic
	// Desired state matches observed state
	// Next states: StatusUpdating, StatusDeleting
	StatusReady Status = "ready"

	// StatusUpdating: Tenant is being modified (image update, config change)
	// Temporary state during reconciliation
	// Next states: StatusReady, StatusFailed
	StatusUpdating Status = "updating"

	// StatusDeleting: Tenant deletion in progress
	// Resources are being torn down
	// Next states: StatusArchived, StatusFailed
	StatusDeleting Status = "deleting"

	// StatusArchiving: Tenant resources are being archived (compute removed)
	// Next states: StatusArchived, StatusFailed
	StatusArchiving Status = "archiving"

	// StatusArchived: Tenant resources cleaned up, record retained
	// No next states (terminal)
	StatusArchived Status = "archived"

	// StatusFailed: Operation failed, manual intervention may be required
	// StatusMessage contains error details
	// Next states: Retry to previous state, or StatusDeleting
	StatusFailed Status = "failed"
)

// ValidTransitions defines allowed state transitions
var ValidTransitions = map[Status][]Status{
	StatusRequested:    {StatusProvisioning, StatusFailed},
	StatusPlanning:     {StatusProvisioning, StatusFailed},
	StatusProvisioning: {StatusReady, StatusFailed},
	StatusReady:        {StatusUpdating, StatusDeleting, StatusArchiving},
	StatusUpdating:     {StatusReady, StatusFailed},
	StatusDeleting:     {StatusArchived, StatusFailed},
	StatusArchiving:    {StatusArchived, StatusFailed},
	StatusArchived:     {},                                // Terminal state
	StatusFailed:       {StatusDeleting, StatusArchiving}, // Can archive or delete failed tenants
}

// IsValid checks if a status is a known valid status
func (s Status) IsValid() bool {
	switch s {
	case StatusRequested, StatusPlanning, StatusProvisioning,
		StatusReady, StatusUpdating, StatusDeleting, StatusArchiving,
		StatusArchived, StatusFailed:
		return true
	default:
		return false
	}
}

// IsTerminal returns true if this status is terminal (no further transitions)
func (s Status) IsTerminal() bool {
	return s == StatusArchived
}

// IsHealthy returns true if tenant is in a healthy operational state
func (s Status) IsHealthy() bool {
	return s == StatusReady
}

// CanTransition checks if a transition is valid
func (s Status) CanTransition(to Status) bool {
	allowed, exists := ValidTransitions[s]
	if !exists {
		return false
	}
	for _, valid := range allowed {
		if valid == to {
			return true
		}
	}
	return false
}

// Tenant represents a logical tenant instance managed by the control plane
// A tenant has desired state (what should exist) and observed state (what actually exists)
type Tenant struct {
	// Identity
	// ID is the internal database identifier (UUID)
	ID uuid.UUID `json:"id"`

	// Name is the user-facing stable identifier
	// Must be unique, lowercase alphanumeric with hyphens, max 255 chars
	// Example: "acme-corp", "customer-123"
	Name string `json:"name"`

	// Current Lifecycle State
	// Status represents where the tenant is in its lifecycle
	Status Status `json:"status"`

	// StatusMessage provides human-readable context about current status
	// Examples: "Provisioning ECS task", "Waiting for health check", "Ready to serve traffic"
	StatusMessage string `json:"status_message,omitempty"`

	// WorkflowExecutionID stores the ID of the current or last workflow execution
	// Used to track and monitor workflow progress
	WorkflowExecutionID *string `json:"workflow_execution_id,omitempty"`

	// WorkflowSubState provides provider-agnostic execution sub-state (running, backing-off, error)
	WorkflowSubState *string `json:"workflow_sub_state,omitempty"`

	// WorkflowRetryCount tracks workflow retry attempts
	WorkflowRetryCount *int `json:"workflow_retry_count,omitempty"`

	// WorkflowErrorMessage captures latest workflow error message
	WorkflowErrorMessage *string `json:"workflow_error_message,omitempty"`

	// Desired State (Declarative)
	// DesiredConfig is tenant-specific configuration as map
	// Schema is flexible and provider-specific
	// Example: {"replicas": "2", "cpu": "512", "memory": "1024"}
	DesiredConfig map[string]interface{} `json:"desired_config,omitempty"`

	// Observed State (Actual)
	// ObservedConfig is the actual configuration applied to running resources
	ObservedConfig map[string]interface{} `json:"observed_config,omitempty"`

	// ObservedResourceIDs contains provider-specific resource identifiers
	// Example: {"task_arn": "arn:aws:ecs:...", "target_group_arn": "arn:aws:elasticloadbalancing:..."}
	ObservedResourceIDs map[string]string `json:"observed_resource_ids,omitempty"`

	// Metadata
	// CreatedAt is when the tenant was first created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the tenant was last modified
	UpdatedAt time.Time `json:"updated_at"`

	// Concurrency Control
	// Version is incremented on every update for optimistic locking
	// Prevents lost updates from concurrent modifications
	Version int `json:"version"`

	// Labels and Annotations
	// Labels are key-value pairs for filtering and grouping
	// Example: {"environment": "production", "team": "platform"}
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations are key-value pairs for metadata not used in queries
	// Example: {"oncall": "team-platform@example.com", "cost-center": "engineering"}
	Annotations map[string]string `json:"annotations,omitempty"`
}

// Validate checks if a tenant is valid
func (t *Tenant) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(t.Name) > 255 {
		return fmt.Errorf("name must be <= 255 characters")
	}
	if !tenantNamePattern.MatchString(t.Name) {
		return fmt.Errorf("name must be lowercase alphanumeric with hyphens")
	}
	if t.Status == "" {
		return fmt.Errorf("status is required")
	}
	if !t.Status.IsValid() {
		return fmt.Errorf("invalid status: %s", t.Status)
	}
	return nil
}

// IsArchived returns true if tenant resources have been archived
func (t *Tenant) IsArchived() bool {
	return t.Status == StatusArchived
}

// IsDrifted returns true if desired state doesn't match observed state
func (t *Tenant) IsDrifted() bool {
	if t.Status != StatusReady {
		return false // Only check drift for ready tenants
	}
	return !reflect.DeepEqual(t.DesiredConfig, t.ObservedConfig)
}

// Clone creates a deep copy of the tenant
func (t *Tenant) Clone() *Tenant {
	clone := *t
	if t.WorkflowExecutionID != nil {
		id := *t.WorkflowExecutionID
		clone.WorkflowExecutionID = &id
	}
	if t.WorkflowSubState != nil {
		state := *t.WorkflowSubState
		clone.WorkflowSubState = &state
	}
	if t.WorkflowRetryCount != nil {
		count := *t.WorkflowRetryCount
		clone.WorkflowRetryCount = &count
	}
	if t.WorkflowErrorMessage != nil {
		msg := *t.WorkflowErrorMessage
		clone.WorkflowErrorMessage = &msg
	}
	if t.Labels != nil {
		clone.Labels = make(map[string]string, len(t.Labels))
		for k, v := range t.Labels {
			clone.Labels[k] = v
		}
	}
	if t.Annotations != nil {
		clone.Annotations = make(map[string]string, len(t.Annotations))
		for k, v := range t.Annotations {
			clone.Annotations[k] = v
		}
	}
	return &clone
}

// StateTransition represents a single state change in tenant lifecycle
// Immutable audit log entry
type StateTransition struct {
	// Identity
	// ID is the unique identifier for this transition record
	ID uuid.UUID `json:"id"`

	// TenantID links this transition to a tenant (database UUID)
	TenantID uuid.UUID `json:"tenant_id"`

	// Transition Details
	// FromStatus is the previous state (nil for initial creation)
	FromStatus *Status `json:"from_status,omitempty"`

	// ToStatus is the new state after transition
	ToStatus Status `json:"to_status"`

	// Context
	// Reason explains why the transition occurred
	// Required field, must be human-readable
	// Examples: "User requested tenant creation", "Health check failed", "Workflow completed successfully"
	Reason string `json:"reason"`

	// TriggeredBy identifies who/what initiated the transition
	// Examples: "user@example.com", "reconciliation-loop", "workflow:provision-tenant"
	TriggeredBy string `json:"triggered_by,omitempty"`

	// State Snapshots (Optional)
	// DesiredStateSnapshot captures desired state at transition time
	// Useful for understanding what was intended when this transition occurred
	DesiredStateSnapshot map[string]interface{} `json:"desired_state_snapshot,omitempty"`

	// ObservedStateSnapshot captures observed state at transition time
	// Useful for debugging drift and understanding actual system state
	ObservedStateSnapshot map[string]interface{} `json:"observed_state_snapshot,omitempty"`

	// Metadata
	// CreatedAt is when this transition was recorded
	CreatedAt time.Time `json:"created_at"`
}

// NewStateTransition creates a new state transition record
func NewStateTransition(tenant *Tenant, toStatus Status, reason, triggeredBy string) *StateTransition {
	transition := &StateTransition{
		ID:          uuid.New(),
		TenantID:    tenant.ID,
		ToStatus:    toStatus,
		Reason:      reason,
		TriggeredBy: triggeredBy,
		CreatedAt:   time.Now(),
	}

	// Set FromStatus if not initial state
	if tenant.Status != "" {
		fromStatus := tenant.Status
		transition.FromStatus = &fromStatus
	}

	// Optionally capture snapshots
	if tenant.DesiredConfig != nil {
		transition.DesiredStateSnapshot = tenant.DesiredConfig
	}
	if tenant.ObservedConfig != nil {
		transition.ObservedStateSnapshot = tenant.ObservedConfig
	}

	return transition
}

// Validate checks if a state transition is valid
func (st *StateTransition) Validate() error {
	if st.TenantID == uuid.Nil {
		return fmt.Errorf("tenant_id is required")
	}
	if st.ToStatus == "" {
		return fmt.Errorf("to_status is required")
	}
	if !st.ToStatus.IsValid() {
		return fmt.Errorf("invalid to_status: %s", st.ToStatus)
	}
	if st.Reason == "" {
		return fmt.Errorf("reason is required")
	}

	// Validate transition is allowed
	if st.FromStatus != nil {
		if !st.FromStatus.CanTransition(st.ToStatus) {
			return fmt.Errorf("invalid transition from %s to %s", *st.FromStatus, st.ToStatus)
		}
	}

	return nil
}
