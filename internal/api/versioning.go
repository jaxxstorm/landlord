package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/jaxxstorm/landlord/internal/apiversion"
)

func (s *Server) handleVersionRequired(w http.ResponseWriter, r *http.Request) {
	s.writeVersionError(w, r, "version_required")
}

func (s *Server) handleUnsupportedVersion(w http.ResponseWriter, r *http.Request) {
	version := chi.URLParam(r, "version")
	if apiversion.IsSupported(version) {
		http.NotFound(w, r)
		return
	}
	s.writeVersionError(w, r, "unsupported_version")
}

func (s *Server) writeVersionError(w http.ResponseWriter, r *http.Request, code string) {
	requestID := r.Header.Get("X-Request-ID")
	s.writeErrorResponse(w, http.StatusBadRequest, code, apiversion.SupportedVersions(), requestID)
}
