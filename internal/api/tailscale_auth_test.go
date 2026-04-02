package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/jaxxstorm/landlord/internal/config"
)

type stubTailscaleAuthorizer struct {
	authorize func(ctx context.Context, remoteAddr, capability string) (*TailscaleIdentity, error)
}

func (s stubTailscaleAuthorizer) Authorize(ctx context.Context, remoteAddr, capability string) (*TailscaleIdentity, error) {
	return s.authorize(ctx, remoteAddr, capability)
}

func newTestHTTPConfig() *config.HTTPConfig {
	return &config.HTTPConfig{
		Host:            "127.0.0.1",
		Port:            8080,
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     120 * time.Second,
		ShutdownTimeout: 30 * time.Second,
	}
}

func TestTailscaleAuthMiddlewareAllowsRequestAndSetsIdentity(t *testing.T) {
	srv := &Server{
		logger: zap.NewNop(),
		httpConfig: &config.HTTPConfig{
			TailscaleAuth: config.TailscaleAuthConfig{
				Enabled: true,
				ProtectedEndpoints: []config.TailscaleProtectedEndpoint{{
					Method:     http.MethodPut,
					Path:       "/v1/tenants/{id}",
					Capability: config.TailscaleCapabilityPrefix,
				}},
			},
		},
		tailscaleAuthorizer: stubTailscaleAuthorizer{authorize: func(ctx context.Context, remoteAddr, capability string) (*TailscaleIdentity, error) {
			return &TailscaleIdentity{LoginName: "user@example.com", NodeName: "test-node", RemoteAddr: remoteAddr}, nil
		}},
	}

	handler := srv.tailscaleAuthMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := TailscaleIdentityFromContext(r.Context())
		if !ok {
			t.Fatal("expected tailscale identity in context")
		}
		if identity.LoginName != "user@example.com" {
			t.Fatalf("unexpected login name: %s", identity.LoginName)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPut, "/v1/tenants/123", nil)
	req.RemoteAddr = "100.64.0.10:1234"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestTailscaleAuthMiddlewareDeniesMissingCapability(t *testing.T) {
	srv := &Server{
		logger: zap.NewNop(),
		httpConfig: &config.HTTPConfig{
			TailscaleAuth: config.TailscaleAuthConfig{
				Enabled: true,
				ProtectedEndpoints: []config.TailscaleProtectedEndpoint{{
					Method:     http.MethodGet,
					Path:       "/v1/docs",
					Capability: config.TailscaleCapabilityPrefix,
				}},
			},
		},
		tailscaleAuthorizer: stubTailscaleAuthorizer{authorize: func(ctx context.Context, remoteAddr, capability string) (*TailscaleIdentity, error) {
			return &TailscaleIdentity{LoginName: "user@example.com", NodeName: "test-node", RemoteAddr: remoteAddr}, ErrTailscaleCapabilityDenied
		}},
	}

	called := false
	handler := srv.tailscaleAuthMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/docs", nil)
	req.RemoteAddr = "100.64.0.10:1234"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if called {
		t.Fatal("handler should not have been called")
	}
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestTailscaleAuthMiddlewareDeniesWhenIdentityUnavailable(t *testing.T) {
	srv := &Server{
		logger: zap.NewNop(),
		httpConfig: &config.HTTPConfig{
			TailscaleAuth: config.TailscaleAuthConfig{
				Enabled: true,
				ProtectedEndpoints: []config.TailscaleProtectedEndpoint{{
					Method:     http.MethodGet,
					Path:       "/v1/docs",
					Capability: config.TailscaleCapabilityPrefix,
				}},
			},
		},
		tailscaleAuthorizer: stubTailscaleAuthorizer{authorize: func(ctx context.Context, remoteAddr, capability string) (*TailscaleIdentity, error) {
			return nil, ErrTailscaleIdentityMissing
		}},
	}

	handler := srv.tailscaleAuthMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/docs", nil)
	req.RemoteAddr = "100.64.0.10:1234"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestTailscaleAuthMiddlewareFailsClosedOnInternalError(t *testing.T) {
	srv := &Server{
		logger: zap.NewNop(),
		httpConfig: &config.HTTPConfig{
			TailscaleAuth: config.TailscaleAuthConfig{
				Enabled: true,
				ProtectedEndpoints: []config.TailscaleProtectedEndpoint{{
					Method:     http.MethodGet,
					Path:       "/v1/docs",
					Capability: config.TailscaleCapabilityPrefix,
				}},
			},
		},
		tailscaleAuthorizer: stubTailscaleAuthorizer{authorize: func(ctx context.Context, remoteAddr, capability string) (*TailscaleIdentity, error) {
			return nil, errors.New("boom")
		}},
	}

	handler := srv.tailscaleAuthMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/docs", nil)
	req.RemoteAddr = "100.64.0.10:1234"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}
}

func TestServerWiringProtectsConfiguredRoute(t *testing.T) {
	cfg := newTestHTTPConfig()
	cfg.TailscaleAuth = config.TailscaleAuthConfig{
		Enabled:  true,
		Hostname: "landlord-test",
		ProtectedEndpoints: []config.TailscaleProtectedEndpoint{{
			Method:     http.MethodGet,
			Path:       "/v1/docs",
			Capability: config.TailscaleCapabilityPrefix,
		}},
	}

	srv := New(cfg, nil, nil, "", nil, nil, zap.NewNop())
	srv.tailscaleAuthorizer = stubTailscaleAuthorizer{authorize: func(ctx context.Context, remoteAddr, capability string) (*TailscaleIdentity, error) {
		return nil, ErrTailscaleCapabilityDenied
	}}

	req := httptest.NewRequest(http.MethodGet, "/v1/docs", nil)
	req.RemoteAddr = "100.64.0.10:1234"
	w := httptest.NewRecorder()

	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestServerWiringLeavesRoutesUnchangedWhenDisabled(t *testing.T) {
	cfg := newTestHTTPConfig()
	srv := New(cfg, nil, nil, "", nil, nil, zap.NewNop())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
