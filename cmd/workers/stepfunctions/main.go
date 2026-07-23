package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jaxxstorm/landlord/internal/compute"
	computedocker "github.com/jaxxstorm/landlord/internal/compute/providers/docker"
	computedecs "github.com/jaxxstorm/landlord/internal/compute/providers/ecs"
	computemock "github.com/jaxxstorm/landlord/internal/compute/providers/mock"
	"github.com/jaxxstorm/landlord/internal/config"
	"github.com/jaxxstorm/landlord/internal/database"
	"github.com/jaxxstorm/landlord/internal/logger"
	"github.com/jaxxstorm/landlord/internal/workflow/lifecycle"
	workflowstepfunctions "github.com/jaxxstorm/landlord/internal/workflow/providers/stepfunctions"
	"go.uber.org/zap"
)

type workerDependencies struct {
	computeManager      *compute.Manager
	cancellationChecker lifecycle.MutationGuard
}

func main() {
	log, registry, dependencies, err := initialize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize Step Functions worker: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	executor, err := lifecycle.NewExecutor(registry, nil, "step-functions", log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize lifecycle executor: %v\n", err)
		os.Exit(1)
	}
	handler, err := workflowstepfunctions.NewLifecycleHandler(executor.WithComputeManager(dependencies.computeManager).WithMutationGuard(dependencies.cancellationChecker))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize Lambda handler: %v\n", err)
		os.Exit(1)
	}

	lambda.Start(handler.WithLogger(log).Handle)
}

func initialize() (*zap.Logger, *compute.Registry, *workerDependencies, error) {
	v := config.NewViperInstance()
	if err := config.BindEnvironmentVariables(v); err != nil {
		return nil, nil, nil, fmt.Errorf("bind environment variables: %w", err)
	}
	configFile, err := config.FindConfigFile("")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("find configuration file: %w", err)
	}
	if configFile == "" {
		return nil, nil, nil, fmt.Errorf("config file is required for startup")
	}
	if err := config.LoadConfigFile(v, configFile); err != nil {
		return nil, nil, nil, fmt.Errorf("load configuration file: %w", err)
	}
	cfg, err := config.LoadFromViper(v)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("load configuration: %w", err)
	}
	log, err := logger.New(cfg.Log.Format, cfg.Log.Level)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("initialize logger: %w", err)
	}

	registry := compute.NewRegistry(log)
	if cfg.Compute.Mock != nil {
		if err := registry.Register(computemock.New()); err != nil {
			return nil, nil, nil, err
		}
	}
	if cfg.Compute.ECS != nil {
		provider := computedecs.New(log, cfg.Compute.ECS.Defaults)
		if err := validateProviderDefaults("ecs", provider, cfg.Compute.ECS.Defaults); err != nil {
			return nil, nil, nil, err
		}
		if err := registry.Register(provider); err != nil {
			return nil, nil, nil, err
		}
	}
	if cfg.Compute.Docker != nil {
		provider, err := computedocker.New(&computedocker.Config{
			Host: cfg.Compute.Docker.Host, NetworkName: cfg.Compute.Docker.NetworkName,
			NetworkDriver: cfg.Compute.Docker.NetworkDriver, LabelPrefix: cfg.Compute.Docker.LabelPrefix,
		}, cfg.Compute.Docker.Defaults, log)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("initialize Docker provider: %w", err)
		}
		if err := validateProviderDefaults("docker", provider, cfg.Compute.Docker.Defaults); err != nil {
			return nil, nil, nil, err
		}
		if err := registry.Register(provider); err != nil {
			return nil, nil, nil, err
		}
	}

	databaseProvider, err := database.NewProvider(context.Background(), &cfg.Database, log)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("connect to database: %w", err)
	}
	pgPool, ok := databaseProvider.Pool().(*pgxpool.Pool)
	if !ok {
		return nil, nil, nil, fmt.Errorf("Step Functions worker requires a PostgreSQL database")
	}
	computeManager := compute.NewWithTracking(registry, compute.NewPgExecutionRepository(pgPool, log), log)
	cancellationProvider, err := workflowstepfunctions.New(context.Background(), workflowstepfunctions.Config{
		Region:          cfg.Workflow.StepFunctions.Region,
		StateMachineARN: cfg.Workflow.StepFunctions.StateMachineARN,
	}, log)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("initialize Step Functions client: %w", err)
	}
	cancellationChecker, err := workflowstepfunctions.NewCancellationChecker(cancellationProvider)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("initialize cancellation checker: %w", err)
	}
	return log, registry, &workerDependencies{computeManager: computeManager, cancellationChecker: cancellationChecker}, nil
}

func validateProviderDefaults(providerName string, provider compute.Provider, defaults map[string]interface{}) error {
	if provider == nil || len(defaults) == 0 {
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
