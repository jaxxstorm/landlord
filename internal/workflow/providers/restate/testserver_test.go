package restate_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

type fakeRestateServer struct {
	t        *testing.T
	server   *httptest.Server
	mu       sync.Mutex
	services map[string]struct{}
	invokes  [][]byte
}

func newFakeRestateServer(t *testing.T) *fakeRestateServer {
	t.Helper()

	defer func() {
		if r := recover(); r != nil {
			t.Skipf("skipping restate test server: %v", r)
		}
	}()

	frs := &fakeRestateServer{
		t:        t,
		services: make(map[string]struct{}),
	}

	frs.server = httptest.NewServer(http.HandlerFunc(frs.handle))
	t.Cleanup(frs.server.Close)
	return frs
}

func (f *fakeRestateServer) URL() string {
	return f.server.URL
}

func (f *fakeRestateServer) InvokeCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.invokes)
}

func (f *fakeRestateServer) LastInvokePayload() []byte {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.invokes) == 0 {
		return nil
	}
	payload := make([]byte, len(f.invokes[len(f.invokes)-1]))
	copy(payload, f.invokes[len(f.invokes)-1])
	return payload
}

func (f *fakeRestateServer) handle(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/health":
		w.WriteHeader(http.StatusOK)
		return
	case r.Method == http.MethodGet && r.URL.Path == "/services":
		f.mu.Lock()
		names := make([]string, 0, len(f.services))
		for name := range f.services {
			names = append(names, name)
		}
		f.mu.Unlock()
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"services": names})
		return
	case strings.HasPrefix(r.URL.Path, "/services/"):
		serviceName := strings.TrimPrefix(r.URL.Path, "/services/")
		f.handleService(w, r, serviceName)
		return
	case strings.HasSuffix(r.URL.Path, "/execute/send") && r.Method == http.MethodPost:
		f.handleInvoke(w, r)
		return
	case r.Method == http.MethodPost && r.URL.Path == "/query":
		f.handleQuery(w, r)
		return
	case strings.HasPrefix(r.URL.Path, "/invocations/") && r.Method == http.MethodGet:
		f.handleInvocationStatus(w, r)
		return
	case strings.HasPrefix(r.URL.Path, "/invocations/") && strings.HasSuffix(r.URL.Path, "/kill") && r.Method == http.MethodPatch:
		f.handleKillInvocation(w, r)
		return
	case r.Method == http.MethodPost && r.URL.Path == "/services":
		f.handleRegister(w, r)
		return
	case r.Method == http.MethodPost && r.URL.Path == "/deployments":
		f.handleRegisterDeployment(w, r)
		return
	default:
		http.NotFound(w, r)
		return
	}
}

func (f *fakeRestateServer) handleService(w http.ResponseWriter, r *http.Request, serviceName string) {
	f.mu.Lock()
	_, exists := f.services[serviceName]
	f.mu.Unlock()

	switch r.Method {
	case http.MethodGet:
		if !exists {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"name": serviceName})
	case http.MethodDelete:
		if !exists {
			http.NotFound(w, r)
			return
		}
		f.mu.Lock()
		delete(f.services, serviceName)
		f.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (f *fakeRestateServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	if _, exists := f.services[payload.Name]; exists {
		w.WriteHeader(http.StatusConflict)
		return
	}
	f.services[payload.Name] = struct{}{}
	w.WriteHeader(http.StatusCreated)
}

func (f *fakeRestateServer) handleRegisterDeployment(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		URI string `json:"uri"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.URI == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (f *fakeRestateServer) handleInvoke(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err == nil && len(body) > 0 {
		f.mu.Lock()
		f.invokes = append(f.invokes, body)
		f.mu.Unlock()
	}
	w.WriteHeader(http.StatusAccepted)
}

func (f *fakeRestateServer) handleInvocationStatus(w http.ResponseWriter, r *http.Request) {
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "completed",
	})
}

func (f *fakeRestateServer) handleKillInvocation(w http.ResponseWriter, r *http.Request) {
	// Accept kill requests with 202 Accepted
	w.WriteHeader(http.StatusAccepted)
}

func (f *fakeRestateServer) handleQuery(w http.ResponseWriter, r *http.Request) {
	var payload map[string]string
	_ = json.NewDecoder(r.Body).Decode(&payload)
	query := payload["query"]
	if query == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"columns": []interface{}{"id", "status"},
		"rows": []interface{}{
			[]interface{}{"inv_test-execution", "completed"},
		},
	})
}
