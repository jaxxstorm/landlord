package models

import (
	"time"

	"github.com/jaxxstorm/landlord/internal/tenant"
)

// CreateTenantRequest represents the request body for creating a new tenant
type CreateTenantRequest struct {
	// Name is the unique human-friendly name for the tenant
	Name string `json:"name" validate:"required,min=1,max=255"`

	// Image is the container image the tenant should run
	Image string `json:"image" validate:"required"`

	// ComputeConfig is provider-specific configuration (Docker, ECS, K8s, etc.)
	// Validated by the active compute provider
	ComputeConfig map[string]interface{} `json:"compute_config,omitempty"`

	// Labels are key-value pairs for organizing tenants
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations are key-value pairs for metadata
	Annotations map[string]string `json:"annotations,omitempty"`
}

// UpdateTenantRequest represents the request body for updating a tenant
type UpdateTenantRequest struct {
	// Name is the updated tenant name (optional for updates)
	Name *string `json:"name,omitempty"`

	// Image is the container image the tenant should run (optional for updates)
	Image *string `json:"image,omitempty"`

	// ComputeConfig is provider-specific configuration (optional for updates)
	// If provided, will be validated by the active compute provider
	ComputeConfig map[string]interface{} `json:"compute_config,omitempty"`

	// Labels are key-value pairs for organizing tenants (optional for updates)
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations are key-value pairs for metadata (optional for updates)
	Annotations map[string]string `json:"annotations,omitempty"`
}

// TenantResponse represents a tenant in API responses
type TenantResponse struct {
	// ID is the internal database identifier (UUID)
	ID string `json:"id"`

	// Name is the user-facing stable identifier
	Name string `json:"name"`

	// Status represents where the tenant is in its lifecycle
	Status string `json:"status"`

	// StatusMessage provides human-readable context about current status
	StatusMessage string `json:"status_message,omitempty"`

	// DesiredImage is the container image the tenant should run
	DesiredImage string `json:"desired_image"`

	// DesiredConfig is tenant-specific configuration
	DesiredConfig map[string]interface{} `json:"desired_config,omitempty"`

	// ComputeConfig is the provider-specific compute configuration
	ComputeConfig map[string]interface{} `json:"compute_config,omitempty"`

	// ObservedImage is the container image currently running
	ObservedImage string `json:"observed_image,omitempty"`

	// ObservedConfig is the actual configuration applied to running resources
	ObservedConfig map[string]interface{} `json:"observed_config,omitempty"`

	// ObservedResourceIDs contains provider-specific resource identifiers
	ObservedResourceIDs map[string]string `json:"observed_resource_ids,omitempty"`

	// WorkflowExecutionID is the ID of the current or last workflow execution
	WorkflowExecutionID *string `json:"workflow_execution_id,omitempty"`

	// CreatedAt is when the tenant was first created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the tenant was last modified
	UpdatedAt time.Time `json:"updated_at"`

	// Version is the concurrency control version
	Version int `json:"version"`

	// Labels are key-value pairs for filtering and grouping
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations are key-value pairs for metadata
	Annotations map[string]string `json:"annotations,omitempty"`
}

// ListTenantsResponse represents a paginated list of tenants
type ListTenantsResponse struct {
	// Tenants is the array of tenant resources
	Tenants []TenantResponse `json:"tenants"`

	// Pagination metadata
	Total  int `json:"total"`  // Total number of tenants
	Limit  int `json:"limit"`  // Number of items per page
	Offset int `json:"offset"` // Starting position
}

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	// Error is the error message
	Error string `json:"error"`

	// Details provides additional context about the error
	Details []string `json:"details,omitempty"`

	// RequestID is the correlation ID from the request context
	RequestID string `json:"request_id,omitempty"`
}

// ToTenantResponse converts a domain tenant to an API response
func ToTenantResponse(t *tenant.Tenant) TenantResponse {
	resp := TenantResponse{
		ID:                  t.ID.String(),
		Name:                t.Name,
		Status:              string(t.Status),
		StatusMessage:       t.StatusMessage,
		DesiredImage:        t.DesiredImage,
		DesiredConfig:       t.DesiredConfig,
		ObservedImage:       t.ObservedImage,
		ObservedConfig:      t.ObservedConfig,
		ObservedResourceIDs: t.ObservedResourceIDs,
		WorkflowExecutionID: t.WorkflowExecutionID,
		CreatedAt:           t.CreatedAt,
		UpdatedAt:           t.UpdatedAt,
		Version:             t.Version,
		Labels:              t.Labels,
		Annotations:         t.Annotations,
	}

	// Convert DesiredConfig map to ComputeConfig map for API response
	if len(t.DesiredConfig) > 0 {
		resp.ComputeConfig = copyInterfaceMap(t.DesiredConfig)
	}

	return resp
}

// FromCreateRequest converts a create request to a domain tenant
func FromCreateRequest(req *CreateTenantRequest) (*tenant.Tenant, error) {
	t := &tenant.Tenant{
		Name:         req.Name,
		DesiredImage: req.Image,
		Labels:       req.Labels,
		Annotations:  req.Annotations,
		Status:       tenant.StatusRequested,
	}

	// Convert ComputeConfig map to DesiredConfig map if present
	if req.ComputeConfig != nil {
		t.DesiredConfig = copyInterfaceMap(req.ComputeConfig)
	}

	return t, nil
}

// ApplyUpdateRequest applies an update request to an existing tenant
func ApplyUpdateRequest(t *tenant.Tenant, req *UpdateTenantRequest) error {
	if req.Name != nil {
		t.Name = *req.Name
	}
	if req.Image != nil {
		t.DesiredImage = *req.Image
	}

	if req.ComputeConfig != nil {
		t.DesiredConfig = copyInterfaceMap(req.ComputeConfig)
	}

	if req.Labels != nil {
		t.Labels = req.Labels
	}

	if req.Annotations != nil {
		t.Annotations = req.Annotations
	}

	return nil
}

func copyInterfaceMap(input map[string]interface{}) map[string]interface{} {
	if input == nil {
		return nil
	}
	output := make(map[string]interface{}, len(input))
	for k, v := range input {
		output[k] = v
	}
	return output
}
