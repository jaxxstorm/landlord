package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jaxxstorm/landlord/internal/api/models"
)

// handleComputeConfigDiscovery returns the requested compute provider config schema.
// @Summary Get compute config discovery
// @Description Returns the requested compute provider and its compute_config schema (and defaults if available)
// @Tags compute
// @Produce json
// @Param provider query string true "Compute provider identifier"
// @Success 200 {object} models.ComputeConfigDiscoveryResponse "Compute config discovery"
// @Failure 400 {object} models.ErrorResponse "Compute provider not available"
// @Failure 500 {object} models.ErrorResponse "Compute provider registry not configured"
// @Router /v1/compute/config [get]
func (s *Server) handleComputeConfigDiscovery(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get("X-Request-ID")
	if s.computeRegistry == nil {
		s.writeErrorResponse(w, http.StatusInternalServerError, "Compute provider registry not configured", nil, requestID)
		return
	}

	provider := strings.TrimSpace(r.URL.Query().Get("provider"))
	if provider == "" {
		s.writeErrorResponse(w, http.StatusBadRequest, "provider is required", nil, requestID)
		return
	}

	schema, defaults, err := s.computeRegistry.GetProviderSchema(provider)
	if err != nil {
		s.writeErrorResponse(w, http.StatusBadRequest, "Compute provider not available", []string{err.Error()}, requestID)
		return
	}

	resp := models.ComputeConfigDiscoveryResponse{
		Provider: provider,
		Schema:   schema,
	}

	if len(defaults) > 0 {
		resp.Defaults = defaults
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
