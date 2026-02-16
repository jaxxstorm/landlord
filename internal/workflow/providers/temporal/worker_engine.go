package temporal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jaxxstorm/landlord/internal/compute"
	"github.com/jaxxstorm/landlord/internal/config"
	"github.com/jaxxstorm/landlord/internal/workflow"
	"go.temporal.io/sdk/activity"
	temporalclient "go.temporal.io/sdk/client"
	temporal "go.temporal.io/sdk/temporal"
	temporalworker "go.temporal.io/sdk/worker"
	temporalworkflow "go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

// WorkerEngine executes tenant lifecycle operations for Temporal workflows.
type WorkerEngine struct {
	config          config.TemporalConfig
	logger          *zap.Logger
	computeRegistry *compute.Registry
	computeResolver workflow.ComputeProviderResolver
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
	return &WorkerEngine{
		config:          cfg,
		logger:          logger.With(zap.String("component", "temporal-worker-engine")),
		computeRegistry: computeRegistry,
		computeResolver: computeResolver,
	}, nil
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
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	tenantID := req.TenantUUID
	if tenantID == "" {
		tenantID = req.TenantID
	}
	if tenantID == "" {
		return nil, fmt.Errorf("tenant identifier is required")
	}

	op := req.Operation
	if op == "" {
		op = "provision"
	}

	switch op {
	case "plan":
		return succeededStatus("plan", tenantID, "temporal", map[string]string{"status": "planned", "tenant_id": tenantID})
	case "create", "apply", "provision":
		return w.provision(ctx, tenantID, req)
	case "update":
		return w.update(ctx, tenantID, req)
	case "delete", "destroy":
		return w.destroy(ctx, tenantID, req)
	default:
		return nil, fmt.Errorf("unknown operation: %s", op)
	}
}

func (w *WorkerEngine) provision(ctx context.Context, tenantID string, req *workflow.ProvisionRequest) (*workflow.ExecutionStatus, error) {
	provider, providerType, err := w.resolveComputeProvider(ctx, req)
	if err != nil {
		return nil, err
	}

	spec := buildComputeSpec(tenantID, providerType, req.DesiredConfig)
	result, err := provider.Provision(ctx, spec)
	if err != nil {
		if status, statusErr := provider.GetStatus(ctx, tenantID); statusErr == nil {
			return succeededStatus("provision", tenantID, "temporal", status)
		}
		return nil, fmt.Errorf("compute provisioning failed: %w", err)
	}

	return succeededStatus("provision", tenantID, "temporal", result)
}

func (w *WorkerEngine) update(ctx context.Context, tenantID string, req *workflow.ProvisionRequest) (*workflow.ExecutionStatus, error) {
	provider, providerType, err := w.resolveComputeProvider(ctx, req)
	if err != nil {
		return nil, err
	}

	spec := buildComputeSpec(tenantID, providerType, req.DesiredConfig)
	result, err := provider.Update(ctx, tenantID, spec)
	if err != nil {
		return nil, fmt.Errorf("compute update failed: %w", err)
	}

	return succeededStatus("update", tenantID, "temporal", result)
}

func (w *WorkerEngine) destroy(ctx context.Context, tenantID string, req *workflow.ProvisionRequest) (*workflow.ExecutionStatus, error) {
	provider, _, err := w.resolveComputeProvider(ctx, req)
	if err != nil {
		return nil, err
	}

	if err := provider.Destroy(ctx, tenantID); err != nil {
		if !errors.Is(err, compute.ErrTenantNotFound) {
			return nil, fmt.Errorf("compute deprovisioning failed: %w", err)
		}
	}

	return succeededStatus("delete", tenantID, "temporal", map[string]string{"status": "archived", "tenant_id": tenantID})
}

func (w *WorkerEngine) resolveComputeProvider(ctx context.Context, req *workflow.ProvisionRequest) (compute.Provider, string, error) {
	providerType := req.ComputeProvider
	if providerType == "" && w.computeResolver != nil {
		resolved, err := w.computeResolver.ResolveProvider(ctx, req.TenantID, req.TenantUUID)
		if err != nil {
			w.logger.Warn("failed to resolve compute provider", zap.Error(err))
		} else if resolved != "" {
			providerType = resolved
		}
	}

	if providerType == "" {
		providers := w.computeRegistry.List()
		if len(providers) == 1 {
			providerType = providers[0]
		}
	}
	if providerType == "" {
		return nil, "", fmt.Errorf("compute provider not specified")
	}

	provider, err := w.computeRegistry.Get(providerType)
	if err != nil {
		return nil, "", fmt.Errorf("compute provider lookup failed: %w", err)
	}
	return provider, providerType, nil
}

func buildComputeSpec(tenantID, providerType string, desiredConfig map[string]interface{}) *compute.TenantComputeSpec {
	spec := &compute.TenantComputeSpec{TenantID: tenantID, ProviderType: providerType}

	if providerType == "docker" {
		spec.Containers = []compute.ContainerSpec{{Name: "app"}}
	}

	if len(desiredConfig) > 0 {
		if raw, err := json.Marshal(desiredConfig); err == nil {
			spec.ProviderConfig = raw
		}
	}

	return spec
}

func succeededStatus(operation, tenantID, provider string, payload interface{}) (*workflow.ExecutionStatus, error) {
	output, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal output: %w", err)
	}

	return &workflow.ExecutionStatus{
		ExecutionID:  fmt.Sprintf("%s-%s", operation, tenantID),
		ProviderType: provider,
		State:        workflow.StateSucceeded,
		Output:       output,
	}, nil
}
