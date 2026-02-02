package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/jaxxstorm/landlord/internal/api/models"
)

func newVersioningTestServer() *Server {
	router := chi.NewRouter()
	srv := &Server{router: router}
	srv.registerRoutes()
	return srv
}

func TestVersionRequiredForUnversionedPaths(t *testing.T) {
	srv := newVersioningTestServer()
	req := httptest.NewRequest(http.MethodGet, "/api/tenants", nil)
	rec := httptest.NewRecorder()

	srv.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	var resp models.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Error != "version_required" {
		t.Fatalf("expected error code version_required, got %q", resp.Error)
	}
	if len(resp.Details) == 0 || resp.Details[0] != "v1" {
		t.Fatalf("expected supported versions list to include v1, got %#v", resp.Details)
	}
}

func TestUnsupportedVersionReturnsError(t *testing.T) {
	srv := newVersioningTestServer()
	req := httptest.NewRequest(http.MethodGet, "/v2/tenants", nil)
	rec := httptest.NewRecorder()

	srv.router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	var resp models.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Error != "unsupported_version" {
		t.Fatalf("expected error code unsupported_version, got %q", resp.Error)
	}
	if len(resp.Details) == 0 || resp.Details[0] != "v1" {
		t.Fatalf("expected supported versions list to include v1, got %#v", resp.Details)
	}
}
