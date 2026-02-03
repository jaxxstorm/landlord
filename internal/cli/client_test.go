package cli

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaxxstorm/landlord/internal/api/models"
)

func newTestServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("skipping test server: %v", err)
	}

	server := httptest.NewUnstartedServer(handler)
	server.Listener = ln
	server.Start()
	t.Cleanup(server.Close)
	return server
}

func TestClientCreateListDelete(t *testing.T) {
	t.Parallel()

	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/tenants":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"123","name":"demo","status":"planning","desired_config":{"image":"nginx:alpine"},"compute_config":{"image":"nginx:alpine"}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v1/tenants/123/archive":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"123","name":"demo","status":"archiving","desired_config":{"image":"nginx:alpine"},"compute_config":{"image":"nginx:alpine"}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/v1/tenants":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"tenants":[{"id":"123","name":"demo","status":"archived","desired_config":{"image":"nginx:alpine"},"compute_config":{"image":"nginx:alpine"}}],"total":1,"limit":50,"offset":0}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/v1/tenants/123":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"123","name":"demo","status":"deleting","desired_config":{"image":"nginx:alpine"},"compute_config":{"image":"nginx:alpine"}}`))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	client := NewClient(server.URL)

	if _, err := client.CreateTenant(context.Background(), models.CreateTenantRequest{
		Name:          "demo",
		ComputeConfig: map[string]interface{}{"image": "nginx:alpine"},
	}); err != nil {
		t.Fatalf("create tenant failed: %v", err)
	}

	if _, err := client.ListTenants(context.Background(), false); err != nil {
		t.Fatalf("list tenants failed: %v", err)
	}

	if _, err := client.ArchiveTenant(context.Background(), "demo"); err != nil {
		t.Fatalf("archive tenant failed: %v", err)
	}

	if _, err := client.DeleteTenant(context.Background(), "demo"); err != nil {
		t.Fatalf("delete tenant failed: %v", err)
	}
}

func TestClientHandlesErrors(t *testing.T) {
	t.Parallel()

	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	}))

	client := NewClient(server.URL)
	_, err := client.ListTenants(context.Background(), false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestClientGetUpdateTenant(t *testing.T) {
	t.Parallel()

	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/tenants":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"tenants":[{"id":"123","name":"demo","status":"ready","desired_config":{"image":"nginx:alpine"},"compute_config":{"image":"nginx:alpine"}}],"total":1,"limit":50,"offset":0}`))
		case r.Method == http.MethodGet && r.URL.Path == "/v1/tenants/123":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"123","name":"demo","status":"ready","desired_config":{"image":"nginx:alpine"},"compute_config":{"image":"nginx:alpine"}}`))
		case r.Method == http.MethodPut && r.URL.Path == "/v1/tenants/123":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"123","name":"demo","status":"planning","desired_config":{"image":"nginx:1.25"},"compute_config":{"image":"nginx:1.25"}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	client := NewClient(server.URL)

	if _, err := client.GetTenant(context.Background(), "demo"); err != nil {
		t.Fatalf("get tenant failed: %v", err)
	}

	if _, err := client.UpdateTenant(context.Background(), "demo", http.MethodPut, models.UpdateTenantRequest{
		ComputeConfig: map[string]interface{}{"image": "nginx:1.25"},
	}); err != nil {
		t.Fatalf("update tenant failed: %v", err)
	}
}

func TestClientComputeConfigDiscovery(t *testing.T) {
	t.Parallel()

	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/compute/config" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.URL.Query().Get("provider") != "docker" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"provider":"docker","schema":{"type":"object"},"defaults":{"env":{"FOO":"bar"}}}`))
	}))

	client := NewClient(server.URL)
	resp, err := client.GetComputeConfigDiscovery(context.Background(), "docker")
	if err != nil {
		t.Fatalf("compute config discovery failed: %v", err)
	}
	if resp.Provider != "docker" {
		t.Fatalf("expected provider docker, got %s", resp.Provider)
	}
	if len(resp.Schema) == 0 {
		t.Fatalf("expected schema to be set")
	}
	if len(resp.Defaults) == 0 {
		t.Fatalf("expected defaults to be set")
	}
}
