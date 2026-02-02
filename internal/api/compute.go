package api

import (
	"encoding/json"
	"net/http"

	"github.com/jaxxstorm/landlord/internal/api/models"
)

// handleComputeConfigDiscovery returns the active compute provider config schema.
// @Summary Get compute config discovery
// @Description Returns the active compute provider and its compute_config schema (and defaults if available)
// @Tags compute
// @Produce json
// @Success 200 {object} models.ComputeConfigDiscoveryResponse "Compute config discovery"
// @Failure 500 {object} models.ErrorResponse "Compute provider not configured"
// @Router /v1/compute/config [get]
func (s *Server) handleComputeConfigDiscovery(w http.ResponseWriter, r *http.Request) {
	requestID := r.Header.Get("X-Request-ID")
	if s.computeProvider == nil {
		s.writeErrorResponse(w, http.StatusInternalServerError, "Compute provider not configured", nil, requestID)
		return
	}

	schema := s.computeProvider.ConfigSchema()
	if len(schema) == 0 {
		s.writeErrorResponse(w, http.StatusInternalServerError, "Compute provider schema not available", nil, requestID)
		return
	}

	resp := models.ComputeConfigDiscoveryResponse{
		Provider: s.computeProvider.Name(),
		Schema:   schema,
	}

	if defaults := s.computeProvider.ConfigDefaults(); len(defaults) > 0 {
		resp.Defaults = defaults
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
