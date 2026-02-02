package restate

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jaxxstorm/landlord/internal/compute"
	"github.com/jaxxstorm/landlord/internal/config"
	"github.com/jaxxstorm/landlord/internal/workflow"
	"github.com/restatedev/sdk-go/server"
	"go.uber.org/zap"
)

// WorkerEngine implements a Restate workflow worker.
type WorkerEngine struct {
	config          config.RestateConfig
	logger          *zap.Logger
	computeRegistry *compute.Registry
	computeResolver workflow.ComputeProviderResolver
}

// NewWorkerEngine creates a new Restate worker engine.
func NewWorkerEngine(cfg config.RestateConfig, computeRegistry *compute.Registry, computeResolver workflow.ComputeProviderResolver, logger *zap.Logger) (*WorkerEngine, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid restate configuration: %w", err)
	}
	if computeRegistry == nil {
		return nil, fmt.Errorf("compute registry is required")
	}
	return &WorkerEngine{
		config:          cfg,
		logger:          logger.With(zap.String("component", "restate-worker-engine")),
		computeRegistry: computeRegistry,
		computeResolver: computeResolver,
	}, nil
}

// Name returns the worker engine identifier.
func (w *WorkerEngine) Name() string {
	return "restate"
}

// Register registers worker services with the Restate admin API.
func (w *WorkerEngine) Register(ctx context.Context) error {
	if !w.config.WorkerRegisterOnStartup {
		w.logger.Info("worker registration disabled")
		return nil
	}

	clientCfg := w.config
	if w.config.WorkerAdminEndpoint != "" {
		clientCfg.AdminEndpoint = w.config.WorkerAdminEndpoint
	}

	client, err := NewClient(ctx, clientCfg, w.logger)
	if err != nil {
		return fmt.Errorf("init restate client: %w", err)
	}

	_ = WorkerServiceName(w.config)

	attempts := w.config.RetryAttempts
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	backoff := 500 * time.Millisecond
	for i := 0; i < attempts; i++ {
		if w.config.WorkerAdvertisedURL == "" {
			return fmt.Errorf("worker_advertised_url is required for registration")
		}

		if err := client.RegisterDeployment(ctx, w.config.WorkerAdvertisedURL); err != nil {
			if errors.Is(err, errAdminAPINotSupported) {
				w.logger.Warn("worker registration not supported by restate admin api", zap.Error(err))
				return nil
			}
			lastErr = err
			w.logger.Warn("worker registration failed", zap.Error(err))
			time.Sleep(backoff)
			backoff = backoff * 2
			continue
		}
		w.logger.Info("worker registered",
			zap.String("service_name", WorkerServiceName(w.config)),
			zap.String("uri", w.config.WorkerAdvertisedURL),
		)
		return nil
	}

	return fmt.Errorf("worker registration failed after %d attempt(s): %w", attempts, lastErr)
}

// Start starts the Restate worker server.
func (w *WorkerEngine) Start(ctx context.Context, addr string) error {
	if addr == "" {
		return fmt.Errorf("worker address is required")
	}

	restateServer := server.NewRestate()
	service := NewTenantProvisioningService(w.computeRegistry, w.config.WorkerComputeProvider, w.computeResolver, w.logger)
	service.Bind(restateServer, WorkerServiceName(w.config))

	w.logger.Info("starting restate worker",
		zap.String("address", addr),
		zap.String("service_name", WorkerServiceName(w.config)),
	)

	return restateServer.Start(ctx, addr)
}
