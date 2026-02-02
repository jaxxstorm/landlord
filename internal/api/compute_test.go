package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaxxstorm/landlord/internal/compute"
)

type testComputeProvider struct {
	name     string
	schema   json.RawMessage
	defaults json.RawMessage
}

func (t *testComputeProvider) Name() string { return t.name }
func (t *testComputeProvider) Provision(ctx context.Context, spec *compute.TenantComputeSpec) (*compute.ProvisionResult, error) {
	return nil, nil
}
func (t *testComputeProvider) Update(ctx context.Context, tenantID string, spec *compute.TenantComputeSpec) (*compute.UpdateResult, error) {
	return nil, nil
}
func (t *testComputeProvider) Destroy(ctx context.Context, tenantID string) error { return nil }
func (t *testComputeProvider) GetStatus(ctx context.Context, tenantID string) (*compute.ComputeStatus, error) {
	return nil, nil
}
func (t *testComputeProvider) Validate(ctx context.Context, spec *compute.TenantComputeSpec) error {
	return nil
}
func (t *testComputeProvider) ValidateConfig(config json.RawMessage) error { return nil }
func (t *testComputeProvider) ConfigSchema() json.RawMessage               { return t.schema }
func (t *testComputeProvider) ConfigDefaults() json.RawMessage             { return t.defaults }

func TestHandleComputeConfigDiscovery(t *testing.T) {
	srv := &Server{
		computeProvider: &testComputeProvider{
			name:     "docker",
			schema:   json.RawMessage(`{"type":"object"}`),
			defaults: json.RawMessage(`{"env":{"FOO":"bar"}}`),
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/compute/config", nil)
	w := httptest.NewRecorder()

	srv.handleComputeConfigDiscovery(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp["provider"] != "docker" {
		t.Fatalf("expected provider docker, got %v", resp["provider"])
	}
	if _, ok := resp["schema"]; !ok {
		t.Fatalf("expected schema in response")
	}
	if _, ok := resp["defaults"]; !ok {
		t.Fatalf("expected defaults in response")
	}
}

func TestHandleComputeConfigDiscovery_NoProvider(t *testing.T) {
	srv := &Server{}

	req := httptest.NewRequest(http.MethodGet, "/api/compute/config", nil)
	w := httptest.NewRecorder()

	srv.handleComputeConfigDiscovery(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}
