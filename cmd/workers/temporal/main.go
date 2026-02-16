package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jaxxstorm/landlord/internal/compute"
	computedocker "github.com/jaxxstorm/landlord/internal/compute/providers/docker"
	computedecs "github.com/jaxxstorm/landlord/internal/compute/providers/ecs"
	computemock "github.com/jaxxstorm/landlord/internal/compute/providers/mock"
	"github.com/jaxxstorm/landlord/internal/config"
	"github.com/jaxxstorm/landlord/internal/logger"
	"github.com/jaxxstorm/landlord/internal/workflow"
	"github.com/jaxxstorm/landlord/internal/workflow/providers/temporal"
	"go.uber.org/zap"
)

func main() {
	v := config.NewViperInstance()
	if err := config.BindEnvironmentVariables(v); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to bind environment variables: %v\n", err)
		os.Exit(1)
	}

	configFile, err := config.FindConfigFile("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to find config file: %v\n", err)
		os.Exit(1)
	}
	if configFile == "" {
		fmt.Fprintln(os.Stderr, "Config file is required for startup")
		os.Exit(1)
	}

	if err := config.LoadConfigFile(v, configFile); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config file: %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.LoadFromViper(v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	log, err := logger.New(cfg.Log.Format, cfg.Log.Level)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	ctx := context.Background()

	computeRegistry := compute.NewRegistry(log)
	if cfg.Compute.Mock != nil {
		computeRegistry.Register(computemock.New())
	}
	if cfg.Compute.ECS != nil {
		ecsProvider := computedecs.New(log, cfg.Compute.ECS.Defaults)
		if err := validateProviderDefaults("ecs", ecsProvider, cfg.Compute.ECS.Defaults); err != nil {
			log.Fatal("Invalid ECS compute defaults", zap.Error(err))
		}
		computeRegistry.Register(ecsProvider)
	}
	if cfg.Compute.Docker != nil {
		dockerProvider, err := computedocker.New(
			&computedocker.Config{
				Host:          cfg.Compute.Docker.Host,
				NetworkName:   cfg.Compute.Docker.NetworkName,
				NetworkDriver: cfg.Compute.Docker.NetworkDriver,
				LabelPrefix:   cfg.Compute.Docker.LabelPrefix,
			},
			cfg.Compute.Docker.Defaults,
			log,
		)
		if err != nil {
			log.Fatal("Failed to initialize Docker provider", zap.Error(err))
		}
		if err := validateProviderDefaults("docker", dockerProvider, cfg.Compute.Docker.Defaults); err != nil {
			log.Fatal("Invalid Docker compute defaults", zap.Error(err))
		}
		computeRegistry.Register(dockerProvider)
	}

	workerRegistry := workflow.NewWorkerRegistry(log)
	temporalWorker, err := temporal.NewWorkerEngine(cfg.Workflow.Temporal, computeRegistry, nil, log)
	if err != nil {
		log.Fatal("Failed to initialize temporal worker engine", zap.Error(err))
	}
	if err := workerRegistry.Register(temporalWorker); err != nil {
		log.Fatal("Failed to register temporal worker engine", zap.Error(err))
	}

	workerEngine, err := workerRegistry.Get("temporal")
	if err != nil {
		log.Fatal("Failed to select temporal worker engine", zap.Error(err))
	}

	if err := workerEngine.Register(ctx); err != nil {
		log.Fatal("Failed to register worker engine", zap.Error(err))
	}

	workerCtx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	workerAddr := getWorkerAddress()
	if err := workerEngine.Start(workerCtx, workerAddr); err != nil {
		log.Fatal("Worker failed", zap.Error(err))
	}
}

func getWorkerAddress() string {
	if addr := os.Getenv("LANDLORD_WORKER_ADDRESS"); addr != "" {
		return addr
	}
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}
	return ":9081"
}

func validateProviderDefaults(providerName string, provider compute.Provider, defaults map[string]interface{}) error {
	if provider == nil {
		return nil
	}
	if len(defaults) == 0 {
		return fmt.Errorf("compute.%s must include default compute_config values", providerName)
	}
	raw, err := json.Marshal(defaults)
	if err != nil {
		return fmt.Errorf("marshal %s defaults: %w", providerName, err)
	}
	if err := provider.ValidateConfig(raw); err != nil {
		return fmt.Errorf("invalid %s compute defaults: %w", providerName, err)
	}
	return nil
}
