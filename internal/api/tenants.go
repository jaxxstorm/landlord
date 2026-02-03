package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
	"github.com/jaxxstorm/landlord/internal/api/models"
	"github.com/jaxxstorm/landlord/internal/compute"
	"github.com/jaxxstorm/landlord/internal/tenant"
)

// handleCreateTenant creates a new tenant
// @Summary Create a new tenant
// @Description Creates a new tenant with provided configuration
// @Tags tenants
// @Accept json
// @Produce json
// @Param body body models.CreateTenantRequest true "Tenant creation request"
// @Success 201 {object} models.TenantResponse "Tenant created successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request or validation error"
// @Failure 409 {object} models.ErrorResponse "Tenant name already exists"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /v1/tenants [post]
func (s *Server) handleCreateTenant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := r.Header.Get("X-Request-ID")

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, "Failed to read request body", nil, requestID)
		return
	}
	defer r.Body.Close()

	var req models.CreateTenantRequest
	if err := json.Unmarshal(body, &req); err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", []string{err.Error()}, requestID)
		return
	}

	// Validate required fields
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		s.writeErrorResponse(w, http.StatusBadRequest, "name is required", nil, requestID)
		return
	}

	if len(req.Name) > 255 {
		s.writeErrorResponse(w, http.StatusBadRequest, "name must be <= 255 characters", nil, requestID)
		return
	}

	if req.ComputeConfig == nil {
		s.writeErrorResponse(w, http.StatusBadRequest, "compute_config is required", nil, requestID)
		return
	}

	// Validate compute configuration if provided
	if req.ComputeConfig != nil {
		provider, _, err := s.resolveComputeProvider(req.ComputeConfig, req.Labels, req.Annotations, nil)
		if err != nil {
			status := http.StatusBadRequest
			message := "Compute provider not available"
			if errors.Is(err, compute.ErrProviderNotFound) {
				message = "Compute provider not available"
			} else if strings.Contains(err.Error(), "registry not configured") {
				status = http.StatusInternalServerError
				message = "Compute provider registry not configured"
			} else if strings.Contains(err.Error(), "multiple providers") {
				message = "compute_provider is required when multiple compute providers are configured"
			}
			s.writeErrorResponse(w, status, message, []string{err.Error()}, requestID)
			return
		}
		// Convert map to JSON for validation
		configJSON, err := json.Marshal(req.ComputeConfig)
		if err != nil {
			s.writeErrorResponse(w, http.StatusBadRequest, "Invalid compute configuration format", []string{err.Error()}, requestID)
			return
		}
		if err := compute.ValidateConfigAgainstSchema(provider, configJSON); err != nil {
			s.writeErrorResponse(w, http.StatusBadRequest, "Invalid compute configuration", computeSchemaErrorDetails(err), requestID)
			return
		}
		if err := provider.ValidateConfig(configJSON); err != nil {
			s.writeErrorResponse(w, http.StatusBadRequest, "Invalid compute configuration", []string{err.Error()}, requestID)
			return
		}
	}

	// Convert request to domain model
	t, err := models.FromCreateRequest(&req)
	if err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, "Failed to process request", []string{err.Error()}, requestID)
		return
	}

	// Set ID and timestamps
	t.ID = uuid.New()
	now := time.Now()
	t.CreatedAt = now
	t.UpdatedAt = now
	t.Version = 1

	// Create tenant in database
	if err := s.tenantRepo.CreateTenant(ctx, t); err != nil {
		// Check if it's a duplicate key error
		if errors.Is(err, tenant.ErrTenantExists) {
			s.writeErrorResponse(w, http.StatusConflict, "Tenant name already exists", nil, requestID)
			return
		}
		s.logger.Error("failed to create tenant", zap.Error(err), zap.String("request_id", requestID))
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create tenant", nil, requestID)
		return
	}

	s.logger.Info("tenant created, awaiting reconciliation",
		zap.String("tenant_name", t.Name),
		zap.String("request_id", requestID))

	// Return created tenant with HTTP 201 Created
	resp := models.ToTenantResponse(t)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// handleGetTenant retrieves a single tenant by ID
// @Summary Get a tenant by ID
// @Description Retrieves a specific tenant resource
// @Tags tenants
// @Produce json
// @Param id path string true "Tenant identifier (UUID or name)"
// @Success 200 {object} models.TenantResponse "Tenant found"
// @Failure 400 {object} models.ErrorResponse "Invalid tenant identifier format"
// @Failure 404 {object} models.ErrorResponse "Tenant not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /v1/tenants/{id} [get]
func (s *Server) handleGetTenant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := r.Header.Get("X-Request-ID")

	identifier := chi.URLParam(r, "id")
	if strings.TrimSpace(identifier) == "" {
		s.writeErrorResponse(w, http.StatusBadRequest, "tenant identifier is required", nil, requestID)
		return
	}
	if isUUIDLike(identifier) {
		if _, err := uuid.Parse(identifier); err != nil {
			s.writeErrorResponse(w, http.StatusBadRequest, "invalid tenant identifier format", []string{err.Error()}, requestID)
			return
		}
	}

	// Get tenant from database
	t, err := s.lookupTenant(ctx, identifier)
	if err != nil {
		if errors.Is(err, tenant.ErrTenantNotFound) {
			s.writeErrorResponse(w, http.StatusNotFound, "Tenant not found", nil, requestID)
			return
		}
		s.logger.Error("failed to get tenant", zap.Error(err), zap.String("request_id", requestID))
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve tenant", nil, requestID)
		return
	}

	// Return tenant
	resp := models.ToTenantResponse(t)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// handleListTenants lists all tenants with pagination
// @Summary List all tenants
// @Description Returns a paginated list of tenants
// @Tags tenants
// @Produce json
// @Param limit query int false "Maximum number of results (default 50)"
// @Param offset query int false "Number of results to skip (default 0)"
// @Param include_deleted query bool false "Include archived tenants in results"
// @Success 200 {object} models.ListTenantsResponse "List of tenants"
// @Failure 400 {object} models.ErrorResponse "Invalid pagination parameters"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /v1/tenants [get]
func (s *Server) handleListTenants(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := r.Header.Get("X-Request-ID")

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	includeDeletedStr := r.URL.Query().Get("include_deleted")

	limit := 50
	offset := 0
	includeDeleted := false

	if limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil || parsed < 1 {
			s.writeErrorResponse(w, http.StatusBadRequest, "Invalid limit parameter", []string{"limit must be a positive integer"}, requestID)
			return
		}
		limit = parsed
	}

	if offsetStr != "" {
		parsed, err := strconv.Atoi(offsetStr)
		if err != nil || parsed < 0 {
			s.writeErrorResponse(w, http.StatusBadRequest, "Invalid offset parameter", []string{"offset must be a non-negative integer"}, requestID)
			return
		}
		offset = parsed
	}
	if includeDeletedStr != "" {
		parsed, err := strconv.ParseBool(includeDeletedStr)
		if err != nil {
			s.writeErrorResponse(w, http.StatusBadRequest, "Invalid include_deleted parameter", []string{"include_deleted must be a boolean"}, requestID)
			return
		}
		includeDeleted = parsed
	}

	// List tenants from database
	filters := tenant.ListFilters{
		Limit:          limit,
		Offset:         offset,
		IncludeDeleted: includeDeleted,
	}
	tenants, err := s.tenantRepo.ListTenants(ctx, filters)
	if err != nil {
		s.logger.Error("failed to list tenants", zap.Error(err), zap.String("request_id", requestID))
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list tenants", nil, requestID)
		return
	}

	// Get total count - need to query without limit/offset
	countFilters := filters
	countFilters.Limit = 0
	countFilters.Offset = 0
	allTenants, err := s.tenantRepo.ListTenants(ctx, countFilters)
	if err != nil {
		s.logger.Error("failed to count tenants", zap.Error(err), zap.String("request_id", requestID))
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list tenants", nil, requestID)
		return
	}
	total := len(allTenants)

	// Convert to response format
	responses := make([]models.TenantResponse, 0, len(tenants))
	for _, t := range tenants {
		responses = append(responses, models.ToTenantResponse(t))
	}

	resp := models.ListTenantsResponse{
		Tenants: responses,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// handleUpdateTenant updates an existing tenant
// @Summary Update a tenant
// @Description Updates properties of an existing tenant
// @Tags tenants
// @Accept json
// @Produce json
// @Param id path string true "Tenant identifier (UUID or name)"
// @Param body body models.UpdateTenantRequest true "Tenant update request"
// @Success 200 {object} models.TenantResponse "Tenant updated successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request or validation error"
// @Failure 404 {object} models.ErrorResponse "Tenant not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /v1/tenants/{id} [put]
func (s *Server) handleUpdateTenant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := r.Header.Get("X-Request-ID")

	identifier := chi.URLParam(r, "id")
	if strings.TrimSpace(identifier) == "" {
		s.writeErrorResponse(w, http.StatusBadRequest, "tenant identifier is required", nil, requestID)
		return
	}
	if isUUIDLike(identifier) {
		if _, err := uuid.Parse(identifier); err != nil {
			s.writeErrorResponse(w, http.StatusBadRequest, "invalid tenant identifier format", []string{err.Error()}, requestID)
			return
		}
	}

	// Parse request body
	var req models.UpdateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", []string{err.Error()}, requestID)
		return
	}
	defer r.Body.Close()

	if req.ComputeConfig == nil {
		s.writeErrorResponse(w, http.StatusBadRequest, "compute_config is required", nil, requestID)
		return
	}

	// Get existing tenant
	t, err := s.lookupTenant(ctx, identifier)
	if err != nil {
		if errors.Is(err, tenant.ErrTenantNotFound) {
			s.writeErrorResponse(w, http.StatusNotFound, "Tenant not found", nil, requestID)
			return
		}
		s.logger.Error("failed to get tenant", zap.Error(err), zap.String("request_id", requestID))
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve tenant", nil, requestID)
		return
	}

	// Check for archived tenant
	if t.Status == tenant.StatusArchived {
		s.writeErrorResponse(w, http.StatusConflict, "Tenant is archived", nil, requestID)
		return
	}

	// Validate compute configuration if provided
	if req.ComputeConfig != nil {
		provider, _, err := s.resolveComputeProvider(req.ComputeConfig, req.Labels, req.Annotations, t)
		if err != nil {
			status := http.StatusBadRequest
			message := "Compute provider not available"
			if errors.Is(err, compute.ErrProviderNotFound) {
				message = "Compute provider not available"
			} else if strings.Contains(err.Error(), "registry not configured") {
				status = http.StatusInternalServerError
				message = "Compute provider registry not configured"
			} else if strings.Contains(err.Error(), "multiple providers") {
				message = "compute_provider is required when multiple compute providers are configured"
			}
			s.writeErrorResponse(w, status, message, []string{err.Error()}, requestID)
			return
		}

		configJSON, err := json.Marshal(req.ComputeConfig)
		if err != nil {
			s.writeErrorResponse(w, http.StatusBadRequest, "Invalid compute configuration format", []string{err.Error()}, requestID)
			return
		}
		if err := compute.ValidateConfigAgainstSchema(provider, configJSON); err != nil {
			s.writeErrorResponse(w, http.StatusBadRequest, "Invalid compute configuration", computeSchemaErrorDetails(err), requestID)
			return
		}
		if err := provider.ValidateConfig(configJSON); err != nil {
			s.writeErrorResponse(w, http.StatusBadRequest, "Invalid compute configuration", []string{err.Error()}, requestID)
			return
		}
	}

	// Validate name update if provided
	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			s.writeErrorResponse(w, http.StatusBadRequest, "name cannot be empty", nil, requestID)
			return
		}
		if len(trimmed) > 255 {
			s.writeErrorResponse(w, http.StatusBadRequest, "name must be <= 255 characters", nil, requestID)
			return
		}
		req.Name = &trimmed
	}

	// Validate state transition - check if tenant is in terminal failed state
	if t.Status == tenant.StatusFailed {
		s.writeErrorResponse(w, http.StatusConflict, "Cannot update tenant in failed state", nil, requestID)
		return
	}

	// Store previous status for validation
	previousStatus := t.Status

	// Apply update
	if err := models.ApplyUpdateRequest(t, &req); err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, "Failed to process update", []string{err.Error()}, requestID)
		return
	}

	// Set status to updating if currently ready, otherwise keep current status
	if t.Status == tenant.StatusReady {
		t.Status = tenant.StatusUpdating
		t.StatusMessage = "Update requested"
		t.WorkflowExecutionID = nil
	}

	// Validate state transition
	if previousStatus != t.Status {
		if err := tenant.ValidateTransition(previousStatus, t.Status); err != nil {
			s.writeErrorResponse(w, http.StatusConflict, "Invalid state transition", []string{err.Error()}, requestID)
			return
		}
	}

	// Update timestamp and version
	t.UpdatedAt = time.Now()

	// Save to database
	if err := s.tenantRepo.UpdateTenant(ctx, t); err != nil {
		if errors.Is(err, tenant.ErrTenantExists) {
			s.writeErrorResponse(w, http.StatusConflict, "Tenant name already exists", nil, requestID)
			return
		}
		s.logger.Error("failed to update tenant", zap.Error(err), zap.String("request_id", requestID))
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to update tenant", nil, requestID)
		return
	}

	// Return updated tenant with HTTP 202 Accepted if workflow triggered
	resp := models.ToTenantResponse(t)
	w.Header().Set("Content-Type", "application/json")
	if t.Status == tenant.StatusUpdating {
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	json.NewEncoder(w).Encode(resp)
}

// handleArchiveTenant archives a tenant (removes compute but keeps record)
// @Summary Archive a tenant
// @Description Archives a tenant by removing compute resources and retaining the record
// @Tags tenants
// @Param id path string true "Tenant identifier (UUID or name)"
// @Success 200 {object} models.TenantResponse "Tenant already archived"
// @Success 202 {object} models.TenantResponse "Tenant archival initiated"
// @Failure 400 {object} models.ErrorResponse "Invalid tenant identifier format"
// @Failure 404 {object} models.ErrorResponse "Tenant not found"
// @Failure 409 {object} models.ErrorResponse "Invalid state transition"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /v1/tenants/{id}/archive [post]
func (s *Server) handleArchiveTenant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := r.Header.Get("X-Request-ID")

	identifier := chi.URLParam(r, "id")
	if strings.TrimSpace(identifier) == "" {
		s.writeErrorResponse(w, http.StatusBadRequest, "tenant identifier is required", nil, requestID)
		return
	}
	if isUUIDLike(identifier) {
		if _, err := uuid.Parse(identifier); err != nil {
			s.writeErrorResponse(w, http.StatusBadRequest, "invalid tenant identifier format", []string{err.Error()}, requestID)
			return
		}
	}

	t, err := s.lookupTenant(ctx, identifier)
	if err != nil {
		if errors.Is(err, tenant.ErrTenantNotFound) {
			s.writeErrorResponse(w, http.StatusNotFound, "Tenant not found", nil, requestID)
			return
		}
		s.logger.Error("failed to get tenant", zap.Error(err), zap.String("request_id", requestID))
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve tenant", nil, requestID)
		return
	}

	if t.Status == tenant.StatusArchived {
		resp := models.ToTenantResponse(t)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
		return
	}

	if t.Status == tenant.StatusArchiving || t.Status == tenant.StatusDeleting {
		resp := models.ToTenantResponse(t)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(resp)
		return
	}

	previousStatus := t.Status
	t.Status = tenant.StatusArchiving
	t.StatusMessage = "Archival requested"
	t.WorkflowExecutionID = nil
	if err := tenant.ValidateTransition(previousStatus, t.Status); err != nil {
		s.writeInvalidStateError(w, "Invalid state transition", []string{err.Error()}, requestID)
		return
	}

	t.UpdatedAt = time.Now()
	for attempt := 0; attempt < 2; attempt++ {
		if err := s.tenantRepo.UpdateTenant(ctx, t); err != nil {
			if errors.Is(err, tenant.ErrVersionConflict) {
				fresh, fetchErr := s.lookupTenant(ctx, identifier)
				if fetchErr != nil {
					s.logger.Error("failed to refetch tenant after version conflict", zap.Error(fetchErr), zap.String("request_id", requestID))
					s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to initiate archival", nil, requestID)
					return
				}
				t = fresh
				if t.Status == tenant.StatusArchived {
					resp := models.ToTenantResponse(t)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(resp)
					return
				}
				if t.Status == tenant.StatusArchiving || t.Status == tenant.StatusDeleting {
					resp := models.ToTenantResponse(t)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusAccepted)
					json.NewEncoder(w).Encode(resp)
					return
				}
				previousStatus = t.Status
				t.Status = tenant.StatusArchiving
				t.StatusMessage = "Archival requested"
				t.WorkflowExecutionID = nil
				if err := tenant.ValidateTransition(previousStatus, t.Status); err != nil {
					s.writeInvalidStateError(w, "Invalid state transition", []string{err.Error()}, requestID)
					return
				}
				t.UpdatedAt = time.Now()
				continue
			}
			s.logger.Error("failed to update tenant status to archiving", zap.Error(err), zap.String("request_id", requestID))
			s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to initiate archival", nil, requestID)
			return
		}
		break
	}

	resp := models.ToTenantResponse(t)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(resp)
}

// handleDeleteTenant deletes a tenant
// @Summary Delete a tenant
// @Description Deletes a specific tenant resource
// @Tags tenants
// @Param id path string true "Tenant identifier (UUID or name)"
// @Success 202 {object} models.TenantResponse "Tenant deletion initiated"
// @Failure 400 {object} models.ErrorResponse "Invalid tenant identifier format"
// @Failure 404 {object} models.ErrorResponse "Tenant not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /v1/tenants/{id} [delete]
func (s *Server) handleDeleteTenant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := r.Header.Get("X-Request-ID")

	identifier := chi.URLParam(r, "id")
	if strings.TrimSpace(identifier) == "" {
		s.writeErrorResponse(w, http.StatusBadRequest, "tenant identifier is required", nil, requestID)
		return
	}
	if isUUIDLike(identifier) {
		if _, err := uuid.Parse(identifier); err != nil {
			s.writeErrorResponse(w, http.StatusBadRequest, "invalid tenant identifier format", []string{err.Error()}, requestID)
			return
		}
	}

	// Get existing tenant
	t, err := s.lookupTenant(ctx, identifier)
	if err != nil {
		if errors.Is(err, tenant.ErrTenantNotFound) {
			s.writeErrorResponse(w, http.StatusNotFound, "Tenant not found", nil, requestID)
			return
		}
		s.logger.Error("failed to get tenant", zap.Error(err), zap.String("request_id", requestID))
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve tenant", nil, requestID)
		return
	}

	// Hard delete archived tenants
	if t.Status == tenant.StatusArchived {
		t.Status = tenant.StatusDeleting
		t.StatusMessage = "Deletion requested"
		t.WorkflowExecutionID = nil
		t.UpdatedAt = time.Now()

		if err := s.tenantRepo.UpdateTenant(ctx, t); err != nil {
			s.logger.Error("failed to update archived tenant to deleting", zap.Error(err), zap.String("request_id", requestID))
			s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to initiate deletion", nil, requestID)
			return
		}

		resp := models.ToTenantResponse(t)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(resp)
		return
	}
	if t.Status == tenant.StatusArchiving {
		resp := models.ToTenantResponse(t)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(resp)
		return
	}
	if t.Status == tenant.StatusDeleting {
		resp := models.ToTenantResponse(t)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Set status to deleting
	t.Status = tenant.StatusArchiving
	t.StatusMessage = "Archival requested"
	t.WorkflowExecutionID = nil
	t.UpdatedAt = time.Now()
	if t.Annotations == nil {
		t.Annotations = map[string]string{}
	}
	t.Annotations["landlord/delete_after_archive"] = "true"

	// Update tenant status in database
	if err := s.tenantRepo.UpdateTenant(ctx, t); err != nil {
		s.logger.Error("failed to update tenant status to archiving", zap.Error(err), zap.String("request_id", requestID))
		s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to initiate deletion", nil, requestID)
		return
	}

	// Return tenant with HTTP 202 Accepted
	resp := models.ToTenantResponse(t)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(resp)
}

// writeErrorResponse writes a standardized error response
func (s *Server) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, details []string, requestID string) {
	resp := models.ErrorResponse{
		Error:     message,
		Details:   details,
		RequestID: requestID,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}

// writeWorkflowError writes a standardized error response for workflow trigger failures (500)
func (s *Server) writeWorkflowError(w http.ResponseWriter, err error, tenantID string, requestID string) {
	s.logger.Error("failed to trigger workflow",
		zap.Error(err),
		zap.String("tenant_id", tenantID),
		zap.String("request_id", requestID))
	s.writeErrorResponse(w, http.StatusInternalServerError, "Failed to trigger workflow", []string{err.Error()}, requestID)
}

// writeInvalidStateError writes a standardized error response for invalid state transitions (409)
func (s *Server) writeInvalidStateError(w http.ResponseWriter, message string, details []string, requestID string) {
	s.logger.Warn("invalid state transition",
		zap.String("message", message),
		zap.Strings("details", details),
		zap.String("request_id", requestID))
	s.writeErrorResponse(w, http.StatusConflict, message, details, requestID)
}

func (s *Server) lookupTenant(ctx context.Context, identifier string) (*tenant.Tenant, error) {
	if id, err := uuid.Parse(identifier); err == nil {
		return s.tenantRepo.GetTenantByID(ctx, id)
	}
	return s.tenantRepo.GetTenantByName(ctx, identifier)
}

var uuidLikePattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func isUUIDLike(value string) bool {
	return uuidLikePattern.MatchString(value)
}

func computeSchemaErrorDetails(err error) []string {
	if err == nil {
		return nil
	}
	if schemaErr, ok := err.(*compute.SchemaValidationError); ok {
		return schemaErr.Details
	}
	return []string{err.Error()}
}
