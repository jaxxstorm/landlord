package temporal

import (
	"context"
	"fmt"

	"github.com/jaxxstorm/landlord/internal/compute"
	"github.com/jaxxstorm/landlord/internal/config"
	"github.com/jaxxstorm/landlord/internal/workflow"
	"github.com/jaxxstorm/landlord/internal/workflow/lifecycle"
	"go.temporal.io/sdk/activity"
	temporalclient "go.temporal.io/sdk/client"
	temporal "go.temporal.io/sdk/temporal"
	temporalworker "go.temporal.io/sdk/worker"
	temporalworkflow "go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

// WorkerEngine executes tenant lifecycle operations for Temporal workflows.
type WorkerEngine struct {
	config   config.TemporalConfig
	logger   *zap.Logger
	executor *lifecycle.Executor
}

var _ workflow.WorkerEngine = (*WorkerEngine)(nil)

// NewWorkerEngine creates a new Temporal worker engine.
func NewWorkerEngine(cfg config.TemporalConfig, computeRegistry *compute.Registry, computeResolver workflow.ComputeProviderResolver, logger *zap.Logger) (*WorkerEngine, error) {
	if err := ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid temporal configuration: %w", err)
	}
	if computeRegistry == nil {
		return nil, fmt.Errorf("compute registry is required")
	}
	executor, err := lifecycle.NewExecutor(computeRegistry, computeResolver, "temporal", logger)
	if err != nil {
		return nil, fmt.Errorf("create lifecycle executor: %w", err)
	}
	return &WorkerEngine{config: cfg, logger: logger.With(zap.String("component", "temporal-worker-engine")), executor: executor}, nil
}

// Name returns the worker engine identifier.
func (w *WorkerEngine) Name() string {
	return "temporal"
}

// Register validates worker readiness.
func (w *WorkerEngine) Register(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	w.logger.Info("temporal worker registration complete",
		zap.String("namespace", w.config.Namespace),
		zap.String("task_queue", w.config.TaskQueue),
		zap.String("workflow_id", tenantProvisioningWorkflowID),
	)
	return nil
}

// Start runs the worker until context cancellation.
func (w *WorkerEngine) Start(ctx context.Context, addr string) error {
	if ctx.Err() != nil {
		return nil
	}

	w.logger.Info("starting temporal worker",
		zap.String("namespace", w.config.Namespace),
		zap.String("task_queue", w.config.TaskQueue),
		zap.String("address", addr),
	)

	client, err := temporalclient.Dial(temporalclient.Options{
		HostPort:  w.config.HostPort,
		Namespace: w.config.Namespace,
	})
	if err != nil {
		return fmt.Errorf("connect temporal client: %w", err)
	}
	defer client.Close()

	worker := temporalworker.New(client, w.config.TaskQueue, temporalworker.Options{})
	worker.RegisterWorkflowWithOptions(w.tenantProvisioningWorkflow, temporalworkflow.RegisterOptions{Name: tenantProvisioningWorkflowID})
	worker.RegisterActivityWithOptions(w.executeLifecycleActivity, activity.RegisterOptions{Name: tenantProvisioningActivityID})

	if err := worker.Start(); err != nil {
		return fmt.Errorf("start temporal worker: %w", err)
	}
	w.logger.Info("temporal worker started",
		zap.String("task_queue", w.config.TaskQueue),
		zap.String("workflow_id", tenantProvisioningWorkflowID),
	)

	<-ctx.Done()
	worker.Stop()
	w.logger.Info("stopping temporal worker", zap.Error(ctx.Err()))
	return nil
}

func (w *WorkerEngine) tenantProvisioningWorkflow(ctx temporalworkflow.Context, req workflow.ProvisionRequest) error {
	activityOpts := temporalworkflow.ActivityOptions{
		StartToCloseTimeout: w.config.Timeout,
	}
	if w.config.RetryAttempts > 0 {
		activityOpts.RetryPolicy = &temporal.RetryPolicy{
			MaximumAttempts: int32(w.config.RetryAttempts),
		}
	}
	ctx = temporalworkflow.WithActivityOptions(ctx, activityOpts)

	return temporalworkflow.ExecuteActivity(ctx, tenantProvisioningActivityID, req).Get(ctx, nil)
}

func (w *WorkerEngine) executeLifecycleActivity(ctx context.Context, req workflow.ProvisionRequest) error {
	_, err := w.Execute(ctx, &req)
	return err
}

// Execute runs tenant lifecycle operations using payload-driven state.
func (w *WorkerEngine) Execute(ctx context.Context, req *workflow.ProvisionRequest) (*workflow.ExecutionStatus, error) {
	return w.executor.Execute(ctx, req)
}
