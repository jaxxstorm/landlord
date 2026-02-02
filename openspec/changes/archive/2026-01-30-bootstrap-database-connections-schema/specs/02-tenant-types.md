# Specification: Tenant Domain Types

## Overview

This specification defines the core domain types for tenant lifecycle management: the Tenant entity, Status enumeration, and StateTransition audit records.

## Tenant Entity

### Location
`internal/tenant/tenant.go`

### Tenant Struct

```go
package tenant

import (
    "encoding/json"
    "time"
)

// Tenant represents a logical tenant instance managed by the control plane
// A tenant has desired state (what should exist) and observed state (what actually exists)
type Tenant struct {
    // Identity
    // ID is the internal database identifier (UUID)
    ID string `json:"id"`
    
    // TenantID is the user-facing stable identifier
    // Must be unique, lowercase alphanumeric with hyphens, max 255 chars
    // Example: "acme-corp", "customer-123"
    TenantID string `json:"tenant_id"`
    
    // Current Lifecycle State
    // Status represents where the tenant is in its lifecycle
    Status Status `json:"status"`
    
    // StatusMessage provides human-readable context about current status
    // Examples: "Provisioning ECS task", "Waiting for health check", "Ready to serve traffic"
    StatusMessage string `json:"status_message,omitempty"`
    
    // Desired State (Declarative)
    // DesiredImage is the container image the tenant should run
    // Example: "docker.io/myapp:v1.2.3"
    DesiredImage string `json:"desired_image"`
    
    // DesiredConfig is tenant-specific configuration as JSON
    // Schema is flexible and provider-specific
    // Example: {"replicas": 2, "cpu": "512", "memory": "1024"}
    DesiredConfig json.RawMessage `json:"desired_config"`
    
    // Observed State (Actual)
    // ObservedImage is the container image currently running (may differ from desired during updates)
    ObservedImage string `json:"observed_image,omitempty"`
    
    // ObservedConfig is the actual configuration applied to running resources
    ObservedConfig json.RawMessage `json:"observed_config,omitempty"`
    
    // ObservedResourceIDs contains provider-specific resource identifiers
    // Example: {"task_arn": "arn:aws:ecs:...", "target_group_arn": "arn:aws:elasticloadbalancing:..."}
    ObservedResourceIDs json.RawMessage `json:"observed_resource_ids,omitempty"`
    
    // Metadata
    // CreatedAt is when the tenant was first created
    CreatedAt time.Time `json:"created_at"`
    
    // UpdatedAt is when the tenant was last modified
    UpdatedAt time.Time `json:"updated_at"`
    
    // DeletedAt is when the tenant was soft-deleted (nil if not deleted)
    DeletedAt *time.Time `json:"deleted_at,omitempty"`
    
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
```

## Status Enumeration

### Status Type

```go
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
    // Next states: StatusDeleted, StatusFailed
    StatusDeleting Status = "deleting"
    
    // StatusDeleted: Tenant successfully deleted, all resources cleaned up
    // Terminal state, tenant is soft-deleted (row remains in database)
    // No next states (terminal)
    StatusDeleted Status = "deleted"
    
    // StatusFailed: Operation failed, manual intervention may be required
    // StatusMessage contains error details
    // Next states: Retry to previous state, or StatusDeleting
    StatusFailed Status = "failed"
)
```

### Valid Status Transitions

```
┌─────────────┐
│  requested  │
└──────┬──────┘
       │
       ↓
┌─────────────┐
│  planning   │
└──────┬──────┘
       │
       ↓
┌──────────────┐     ┌──────────────┐
│ provisioning │────→│    failed    │
└──────┬───────┘     └──────────────┘
       │                     ↑
       ↓                     │
┌──────────────┐            │
│    ready     │            │
└──────┬───────┘            │
       │                     │
       ↓                     │
┌──────────────┐            │
│   updating   │────────────┘
└──────┬───────┘
       │
       ↓
┌──────────────┐
│   deleting   │
└──────┬───────┘
       │
       ↓
┌──────────────┐
│   deleted    │
└──────────────┘
```

### Transition Validation

```go
// ValidTransitions defines allowed state transitions
var ValidTransitions = map[Status][]Status{
    StatusRequested:    {StatusPlanning, StatusFailed},
    StatusPlanning:     {StatusProvisioning, StatusFailed},
    StatusProvisioning: {StatusReady, StatusFailed},
    StatusReady:        {StatusUpdating, StatusDeleting},
    StatusUpdating:     {StatusReady, StatusFailed},
    StatusDeleting:     {StatusDeleted, StatusFailed},
    StatusDeleted:      {}, // Terminal state
    StatusFailed:       {StatusDeleting}, // Can only delete failed tenants
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
```

## StateTransition Audit Record

### StateTransition Struct

```go
// StateTransition represents a single state change in tenant lifecycle
// Immutable audit log entry
type StateTransition struct {
    // Identity
    // ID is the unique identifier for this transition record
    ID string `json:"id"`
    
    // TenantID links this transition to a tenant (database UUID, not tenant_id)
    TenantID string `json:"tenant_id"`
    
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
    DesiredStateSnapshot json.RawMessage `json:"desired_state_snapshot,omitempty"`
    
    // ObservedStateSnapshot captures observed state at transition time
    // Useful for debugging drift and understanding actual system state
    ObservedStateSnapshot json.RawMessage `json:"observed_state_snapshot,omitempty"`
    
    // Metadata
    // CreatedAt is when this transition was recorded
    CreatedAt time.Time `json:"created_at"`
}
```

### StateTransition Creation

```go
// NewStateTransition creates a new state transition record
func NewStateTransition(tenant *Tenant, toStatus Status, reason, triggeredBy string) *StateTransition {
    transition := &StateTransition{
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
```

## Validation Rules

### Tenant Validation

```go
// Validate checks if a tenant is valid
func (t *Tenant) Validate() error {
    if t.TenantID == "" {
        return fmt.Errorf("tenant_id is required")
    }
    if len(t.TenantID) > 255 {
        return fmt.Errorf("tenant_id must be <= 255 characters")
    }
    if !tenantIDPattern.MatchString(t.TenantID) {
        return fmt.Errorf("tenant_id must be lowercase alphanumeric with hyphens")
    }
    if t.DesiredImage == "" {
        return fmt.Errorf("desired_image is required")
    }
    if t.Status == "" {
        return fmt.Errorf("status is required")
    }
    if !t.Status.IsValid() {
        return fmt.Errorf("invalid status: %s", t.Status)
    }
    return nil
}

var tenantIDPattern = regexp.MustCompile(`^[a-z0-9-]+$`)
```

### Status Validation

```go
// IsValid checks if a status is a known valid status
func (s Status) IsValid() bool {
    switch s {
    case StatusRequested, StatusPlanning, StatusProvisioning,
         StatusReady, StatusUpdating, StatusDeleting,
         StatusDeleted, StatusFailed:
        return true
    default:
        return false
    }
}

// IsTerminal returns true if this status is terminal (no further transitions)
func (s Status) IsTerminal() bool {
    return s == StatusDeleted
}

// IsHealthy returns true if tenant is in a healthy operational state
func (s Status) IsHealthy() bool {
    return s == StatusReady
}
```

### StateTransition Validation

```go
// Validate checks if a state transition is valid
func (st *StateTransition) Validate() error {
    if st.TenantID == "" {
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
```

## Helper Methods

### Tenant Helpers

```go
// IsDeleted returns true if tenant has been soft-deleted
func (t *Tenant) IsDeleted() bool {
    return t.DeletedAt != nil
}

// IsDrifted returns true if desired state doesn't match observed state
func (t *Tenant) IsDrifted() bool {
    if t.Status != StatusReady {
        return false // Only check drift for ready tenants
    }
    return t.DesiredImage != t.ObservedImage
}

// Clone creates a deep copy of the tenant
func (t *Tenant) Clone() *Tenant {
    clone := *t
    if t.DeletedAt != nil {
        deletedAt := *t.DeletedAt
        clone.DeletedAt = &deletedAt
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
```

## JSON Serialization

### Example Tenant JSON

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "tenant_id": "acme-corp",
  "status": "ready",
  "status_message": "Tenant is healthy and serving traffic",
  "desired_image": "docker.io/myapp:v1.2.3",
  "desired_config": {
    "replicas": 2,
    "cpu": "512",
    "memory": "1024",
    "env": {
      "DATABASE_URL": "postgres://..."
    }
  },
  "observed_image": "docker.io/myapp:v1.2.3",
  "observed_config": {
    "replicas": 2,
    "cpu": "512",
    "memory": "1024"
  },
  "observed_resource_ids": {
    "task_arn": "arn:aws:ecs:us-west-2:123456789012:task/cluster/abc123",
    "target_group_arn": "arn:aws:elasticloadbalancing:..."
  },
  "created_at": "2026-01-30T10:00:00Z",
  "updated_at": "2026-01-30T10:05:00Z",
  "version": 3,
  "labels": {
    "environment": "production",
    "team": "platform"
  },
  "annotations": {
    "oncall": "team-platform@example.com"
  }
}
```

### Example StateTransition JSON

```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "from_status": "provisioning",
  "to_status": "ready",
  "reason": "ECS task reached healthy state, all health checks passing",
  "triggered_by": "workflow:provision-tenant-acme-corp",
  "desired_state_snapshot": {
    "replicas": 2,
    "cpu": "512"
  },
  "observed_state_snapshot": {
    "replicas": 2,
    "cpu": "512",
    "task_arn": "arn:aws:ecs:..."
  },
  "created_at": "2026-01-30T10:05:00Z"
}
```

## Size Limits

- `TenantID`: max 255 characters
- `DesiredImage` / `ObservedImage`: max 500 characters
- `StatusMessage`: max 1024 characters
- `DesiredConfig` / `ObservedConfig` / `ObservedResourceIDs`: max 64 KB
- `Reason`: max 1024 characters
- Label/Annotation keys: max 128 characters
- Label/Annotation values: max 256 characters
- Labels/Annotations count: max 50 per tenant
