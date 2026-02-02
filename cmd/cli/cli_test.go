package main

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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

func TestCLICommands(t *testing.T) {
	server := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/tenants":
			var payload map[string]any
			_ = json.NewDecoder(r.Body).Decode(&payload)
			if payload["name"] != "demo" && payload["name"] != "demo-config" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"name missing"}`))
				return
			}
			if payload["name"] == "demo-config" {
				config, ok := payload["compute_config"].(map[string]any)
				if !ok || config["env"] == nil {
					w.WriteHeader(http.StatusBadRequest)
					_, _ = w.Write([]byte(`{"error":"compute_config missing"}`))
					return
				}
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"123","name":"demo","status":"planning","desired_image":"nginx:alpine"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/v1/compute/config":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"provider":"docker","schema":{"type":"object"},"defaults":{"env":{"FOO":"bar"}}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/v1/tenants":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"tenants":[{"id":"123","name":"demo","status":"ready","desired_image":"nginx:alpine"}],"total":1,"limit":50,"offset":0}`))
		case r.Method == http.MethodGet && r.URL.Path == "/v1/tenants/123":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"123","name":"demo","status":"archived","desired_image":"nginx:alpine"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v1/tenants/123/archive":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"123","name":"demo","status":"archiving","desired_image":"nginx:alpine"}`))
		case (r.Method == http.MethodPut || r.Method == http.MethodPatch) && r.URL.Path == "/v1/tenants/123":
			var payload map[string]any
			_ = json.NewDecoder(r.Body).Decode(&payload)
			if payload["image"] != "nginx:1.25" && payload["compute_config"] == nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"image missing"}`))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"123","name":"demo","status":"planning","desired_image":"nginx:1.25"}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/v1/tenants/123":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"123","name":"demo","status":"deleting","desired_image":"nginx:alpine"}`))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	t.Setenv("LANDLORD_CLI_API_URL", server.URL)

	run := func(args ...string) (string, error) {
		cmd := newRootCommand()
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		cmd.SetArgs(args)
		err := cmd.Execute()
		return out.String(), err
	}

	output, err := run("create", "--tenant-name", "demo", "--image", "nginx:alpine")
	if err != nil {
		t.Fatalf("create command failed: %v", err)
	}
	if !strings.Contains(output, "Tenant created") {
		t.Fatalf("expected create output, got %s", output)
	}

	output, err = run("create", "--tenant-name", "demo-config", "--image", "nginx:alpine", "--config", `{"env":{"FOO":"bar"}}`)
	if err != nil {
		t.Fatalf("create with config failed: %v", err)
	}
	if !strings.Contains(output, "Tenant created") {
		t.Fatalf("expected create output, got %s", output)
	}

	output, err = run("compute")
	if err != nil {
		t.Fatalf("compute command failed: %v", err)
	}
	if !strings.Contains(output, "Compute config discovery") {
		t.Fatalf("expected compute output, got %s", output)
	}

	output, err = run("list")
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}
	if !strings.Contains(output, "demo") {
		t.Fatalf("expected list output to contain tenant, got %s", output)
	}

	output, err = run("get", "--tenant-name", "demo")
	if err != nil {
		t.Fatalf("get command failed: %v", err)
	}
	if !strings.Contains(output, "Tenant details") {
		t.Fatalf("expected get output, got %s", output)
	}

	output, err = run("set", "--tenant-name", "demo", "--image", "nginx:1.25")
	if err != nil {
		t.Fatalf("set command failed: %v", err)
	}
	if !strings.Contains(output, "Tenant updated") {
		t.Fatalf("expected set output, got %s", output)
	}

	output, err = run("set", "--tenant-name", "demo", "--config", `{"env":{"FOO":"bar"}}`)
	if err != nil {
		t.Fatalf("set config command failed: %v", err)
	}
	if !strings.Contains(output, "Tenant updated") {
		t.Fatalf("expected set output, got %s", output)
	}

	output, err = run("archive", "--tenant-name", "demo")
	if err != nil {
		t.Fatalf("archive command failed: %v", err)
	}
	if !strings.Contains(output, "Tenant archival requested") {
		t.Fatalf("expected archive output, got %s", output)
	}

	output, err = run("delete", "--tenant-name", "demo")
	if err != nil {
		t.Fatalf("delete command failed: %v", err)
	}
	if !strings.Contains(output, "Tenant deletion requested") {
		t.Fatalf("expected delete output, got %s", output)
	}
}
