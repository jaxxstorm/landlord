package restate_test

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jaxxstorm/landlord/internal/compute"
	computemock "github.com/jaxxstorm/landlord/internal/compute/providers/mock"
	"github.com/jaxxstorm/landlord/internal/config"
	"github.com/jaxxstorm/landlord/internal/workflow"
	"github.com/jaxxstorm/landlord/internal/workflow/providers/restate"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestWorkerEngineRegistrationRetries(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	var deployAttempts int32
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("skipping test; cannot open local listener: %v", err)
	}
	listener.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			w.WriteHeader(http.StatusOK)
		case "/deployments":
			attempt := atomic.AddInt32(&deployAttempts, 1)
			if attempt == 1 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	cfg := config.RestateConfig{
		Endpoint:                server.URL,
		AdminEndpoint:           server.URL,
		AuthType:                "none",
		WorkerRegisterOnStartup: true,
		WorkerAdvertisedURL:     "http://127.0.0.1:9999",
		RetryAttempts:           2,
		Timeout:                 2 * time.Second,
	}

	registry := compute.NewRegistry(logger)
	mockProvider := computemock.New()
	require.NoError(t, registry.Register(mockProvider))

	resolver := workflow.NewCachedComputeProviderResolver(nil, nil, "mock", time.Minute, logger)
	worker, err := restate.NewWorkerEngine(cfg, registry, resolver, logger)
	require.NoError(t, err)

	require.NoError(t, worker.Register(ctx))
	require.Equal(t, int32(2), atomic.LoadInt32(&deployAttempts))
}
