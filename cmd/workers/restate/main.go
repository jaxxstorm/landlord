package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jaxxstorm/landlord/internal/compute"
	computedocker "github.com/jaxxstorm/landlord/internal/compute/providers/docker"
	computemock "github.com/jaxxstorm/landlord/internal/compute/providers/mock"
	"github.com/jaxxstorm/landlord/internal/config"
	"github.com/jaxxstorm/landlord/internal/logger"
	"github.com/jaxxstorm/landlord/internal/workflow"
	"github.com/jaxxstorm/landlord/internal/workflow/providers/restate"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	v := config.NewViperInstance()
	if err := config.BindEnvironmentVariables(v); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to bind environment variables: %v\n", err)
		os.Exit(1)
	}

	// Find and load config file
	configFile, err := config.FindConfigFile("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to find config file: %v\n", err)
		os.Exit(1)
	}

	if configFile != "" {
		if err := config.LoadConfigFile(v, configFile); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load config file: %v\n", err)
			os.Exit(1)
		}
	}

	cfg, err := config.LoadFromViper(v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.New(cfg.Log.Format, cfg.Log.Level)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("starting landlord workflow worker")

	ctx := context.Background()

	// Initialize compute registry and register providers
	computeRegistry := compute.NewRegistry(log)
	computeRegistry.Register(computemock.New())

	// Register Docker provider if configured
	if cfg.Compute.Docker != nil {
		log.Info("registering Docker compute provider")
		dockerProvider, err := computedocker.New(
			&computedocker.Config{
				Host:          cfg.Compute.Docker.Host,
				NetworkName:   cfg.Compute.Docker.NetworkName,
				NetworkDriver: cfg.Compute.Docker.NetworkDriver,
				LabelPrefix:   cfg.Compute.Docker.LabelPrefix,
			},
			log,
		)
		if err != nil {
			log.Fatal("Failed to initialize Docker provider", zap.Error(err))
		}
		computeRegistry.Register(dockerProvider)
	}

	if cfg.Workflow.Restate.WorkerComputeProvider == "" {
		cfg.Workflow.Restate.WorkerComputeProvider = cfg.Compute.DefaultProvider
	}

	var landlordClient workflow.LandlordClient
	if cfg.Workflow.Restate.WorkerLandlordAPIURL != "" {
		landlordClient = workflow.NewHTTPLandlordClient(cfg.Workflow.Restate.WorkerLandlordAPIURL, log)
	}

	var computeResolver workflow.ComputeProviderResolver
	if landlordClient != nil || cfg.Workflow.Restate.WorkerComputeProvider != "" {
		computeResolver = workflow.NewCachedComputeProviderResolver(
			landlordClient,
			nil,
			cfg.Workflow.Restate.WorkerComputeProvider,
			cfg.Workflow.Restate.WorkerComputeCacheTTL,
			log,
		)
	}

	restateWorker, err := restate.NewWorkerEngine(cfg.Workflow.Restate, computeRegistry, computeResolver, log)
	if err != nil {
		log.Fatal("Failed to initialize restate worker engine", zap.Error(err))
	}

	workerRegistry := workflow.NewWorkerRegistry(log)
	if err := workerRegistry.Register(restateWorker); err != nil {
		log.Fatal("Failed to register restate worker engine", zap.Error(err))
	}

	workerName := cfg.Workflow.DefaultProvider
	if workerName == "" {
		workerName = restateWorker.Name()
	}

	selectedWorker, err := workerRegistry.Get(workerName)
	if err != nil {
		log.Fatal("No worker engine registered for configured workflow provider",
			zap.String("configured_provider", workerName),
			zap.Error(err),
		)
	}

	// Start the worker
	workerCtx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	workerAddr := getWorkerAddress()
	log.Info("starting worker server",
		zap.String("address", workerAddr),
		zap.String("worker_engine", selectedWorker.Name()),
	)

	startErr := make(chan error, 1)
	go func() {
		startErr <- selectedWorker.Start(workerCtx, workerAddr)
	}()

	// Give the worker server a moment to start before registering with Restate.
	time.Sleep(500 * time.Millisecond)

	if err := selectedWorker.Register(ctx); err != nil {
		log.Fatal("Failed to register worker engine", zap.Error(err))
	}

	log.Info("worker started, waiting for workflows",
		zap.String("address", workerAddr),
		zap.String("worker_engine", selectedWorker.Name()),
	)

	if err := <-startErr; err != nil {
		log.Fatal("Worker failed", zap.Error(err))
	}

	log.Info("worker stopped")
}

func getWorkerAddress() string {
	if addr := os.Getenv("LANDLORD_RESTATE_WORKER_ADDRESS"); addr != "" {
		return addr
	}
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}
	return ":9080"
}
