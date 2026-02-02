package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/jaxxstorm/landlord/internal/config"
	"github.com/jaxxstorm/landlord/internal/tenant"
	"github.com/jaxxstorm/landlord/internal/workflow"
)

// workflowClientInterface defines methods used by reconciler
type workflowClientInterface interface {
	TriggerWorkflow(ctx context.Context, t *tenant.Tenant, action string) (string, error)
	GetExecutionStatus(ctx context.Context, executionID string) (*workflow.ExecutionStatus, error)
	DetermineAction(status tenant.Status) (string, error)
}

// Reconciler manages tenant reconciliation
type Reconciler struct {
	tenantRepo     tenant.Repository
	workflowClient workflowClientInterface
	queue          *Queue
	config         config.ControllerConfig
	logger         *zap.Logger

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Retry tracking per tenant
	retryCount map[string]int
	retryMu    sync.RWMutex
}

// NewReconciler creates a new reconciler instance
func NewReconciler(
	tenantRepo tenant.Repository,
	workflowClient *WorkflowClient,
	cfg config.ControllerConfig,
	logger *zap.Logger,
) *Reconciler {
	ctx, cancel := context.WithCancel(context.Background())

	return &Reconciler{
		tenantRepo:     tenantRepo,
		workflowClient: workflowClient,
		queue:          NewRateLimitingQueue(),
		config:         cfg,
		logger:         logger.With(zap.String("component", "reconciler")),
		ctx:            ctx,
		cancel:         cancel,
		retryCount:     make(map[string]int),
	}
}

// Start begins the reconciliation loop and workers
func (r *Reconciler) Start() error {
	if !r.config.Enabled {
		r.logger.Info("controller disabled, not starting reconciler")
		return nil
	}

	r.logger.Info("starting reconciler",
		zap.Duration("interval", r.config.ReconciliationInterval),
		zap.Duration("status_interval", r.config.StatusPollInterval),
		zap.Int("workers", r.config.Workers))

	// Start polling loops
	r.wg.Add(1)
	go r.pollInvocationLoop()

	r.wg.Add(1)
	go r.pollStatusLoop()

	// Start worker goroutines
	for i := 0; i < r.config.Workers; i++ {
		r.wg.Add(1)
		go r.runWorker(i)
	}

	return nil
}

// Stop gracefully shuts down the reconciler
func (r *Reconciler) Stop() error {
	r.logger.Info("stopping reconciler",
		zap.Int("queue_depth", r.queue.Len()))

	// Signal shutdown
	r.cancel()
	r.queue.ShutDown()

	// Wait for workers with timeout
	done := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		r.logger.Info("reconciler stopped gracefully")
		return nil
	case <-time.After(r.config.ShutdownTimeout):
		r.logger.Warn("reconciler shutdown timeout exceeded, forcing exit")
		return fmt.Errorf("shutdown timeout exceeded")
	}
}

// pollInvocationLoop continuously polls for tenants needing workflow invocation
func (r *Reconciler) pollInvocationLoop() {
	defer r.wg.Done()

	ticker := time.NewTicker(r.config.ReconciliationInterval)
	defer ticker.Stop()

	r.logger.Info("invocation poll loop started")

	for {
		select {
		case <-r.ctx.Done():
			r.logger.Info("invocation poll loop stopped")
			return
		case <-ticker.C:
			r.pollTenantsByStatus([]tenant.Status{tenant.StatusRequested, tenant.StatusPlanning})
		}
	}
}

// pollStatusLoop continuously polls for tenants with in-flight workflows
func (r *Reconciler) pollStatusLoop() {
	defer r.wg.Done()

	ticker := time.NewTicker(r.config.StatusPollInterval)
	defer ticker.Stop()

	r.logger.Info("status poll loop started")

	for {
		select {
		case <-r.ctx.Done():
			r.logger.Info("status poll loop stopped")
			return
		case <-ticker.C:
			r.pollTenantsByStatus([]tenant.Status{tenant.StatusProvisioning, tenant.StatusUpdating, tenant.StatusDeleting, tenant.StatusArchiving})
		}
	}
}

// pollTenantsByStatus queries database and enqueues tenants for reconciliation
func (r *Reconciler) pollTenantsByStatus(statuses []tenant.Status) {
	ctx, cancel := context.WithTimeout(r.ctx, 10*time.Second)
	defer cancel()

	tenants, err := r.tenantRepo.ListTenants(ctx, tenant.ListFilters{Statuses: statuses})
	if err != nil {
		r.logger.Error("failed to list tenants for reconciliation", zap.Error(err))
		return
	}

	r.logger.Debug("polled tenants", zap.Int("count", len(tenants)))

	for _, t := range tenants {
		r.queue.Add(t.ID.String())
	}
}

// runWorker processes items from the queue
func (r *Reconciler) runWorker(id int) {
	defer r.wg.Done()

	r.logger.Info("worker started", zap.Int("worker_id", id))

	for {
		item, shutdown := r.queue.Get()
		if shutdown {
			r.logger.Info("worker stopped", zap.Int("worker_id", id))
			return
		}

		r.processItem(item)
	}
}

// processItem reconciles a single tenant
func (r *Reconciler) processItem(item interface{}) {
	defer r.queue.Done(item)

	tenantID, ok := item.(string)
	if !ok {
		r.logger.Error("invalid item type in queue", zap.Any("item", item))
		return
	}

	err := r.reconcile(tenantID)
	if err != nil {
		r.handleReconcileError(tenantID, err)
	} else {
		// Success - forget backoff
		r.queue.Forget(item)
		r.resetRetryCount(tenantID)
	}
}

// reconcile performs reconciliation for a single tenant
func (r *Reconciler) reconcile(tenantID string) error {
	ctx, cancel := context.WithTimeout(r.ctx, 30*time.Second)
	defer cancel()

	startTime := time.Now()

	r.logger.Info("reconciling tenant", zap.String("tenant_id", tenantID))

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return fmt.Errorf("invalid tenant id %q: %w", tenantID, err)
	}

	// Fetch tenant
	t, err := r.tenantRepo.GetTenantByID(ctx, tenantUUID)
	if err != nil {
		if err == tenant.ErrTenantNotFound {
			r.logger.Info("tenant not found, skipping", zap.String("tenant_id", tenantID))
			return nil // Not an error - tenant was deleted
		}
		return fmt.Errorf("fetch tenant: %w", err)
	}

	// Check if still needs reconciliation
	if !shouldReconcile(t.Status) {
		r.logger.Debug("tenant no longer needs reconciliation",
			zap.String("tenant_id", tenantID),
			zap.String("tenant_name", t.Name),
			zap.String("status", string(t.Status)))
		return nil
	}

	// If a workflow execution is in-flight, poll status and update tenant
	if isInFlightStatus(t.Status) && t.WorkflowExecutionID != nil && *t.WorkflowExecutionID != "" {
		execStatus, err := r.workflowClient.GetExecutionStatus(ctx, *t.WorkflowExecutionID)
		if err != nil {
			r.logger.Warn("failed to check workflow status, will retry later",
				zap.String("tenant_id", tenantID),
				zap.String("tenant_name", t.Name),
				zap.String("execution_id", *t.WorkflowExecutionID),
				zap.Error(err))
			return nil
		} else {
			if execStatus.State == workflow.StatePending || execStatus.State == workflow.StateRunning {
				r.logger.Info("workflow still active, skipping trigger",
					zap.String("tenant_id", tenantID),
					zap.String("tenant_name", t.Name),
					zap.String("execution_id", *t.WorkflowExecutionID),
					zap.String("state", string(execStatus.State)))
				return nil
			}

			if execStatus.State == workflow.StateSucceeded {
				if err := r.handleWorkflowSuccess(ctx, t, execStatus); err != nil {
					return err
				}
				duration := time.Since(startTime)
				r.logger.Info("tenant reconciled successfully",
					zap.String("tenant_id", tenantID),
					zap.String("tenant_name", t.Name),
					zap.String("status", string(t.Status)),
					zap.String("execution_id", *t.WorkflowExecutionID),
					zap.Duration("duration", duration))
				return nil
			}

			if err := r.handleWorkflowFailure(ctx, t, execStatus); err != nil {
				return err
			}
			return nil
		}
	}

	// Determine action for new or retried workflow invocation
	action, err := r.workflowClient.DetermineAction(t.Status)
	if err != nil {
		return fmt.Errorf("determine action: %w", err)
	}

	// Trigger workflow
	executionID, err := r.workflowClient.TriggerWorkflow(ctx, t, action)
	if err != nil {
		return fmt.Errorf("trigger workflow: %w", err)
	}

	r.logger.Info("workflow triggered with new execution ID",
		zap.String("tenant_id", tenantID),
		zap.String("tenant_name", t.Name),
		zap.String("new_execution_id", executionID),
		zap.String("action", action))

	// Update tenant with execution ID and move into provisioning where appropriate
	previousStatus := t.Status
	if t.Status == tenant.StatusRequested || t.Status == tenant.StatusPlanning {
		t.Status = tenant.StatusProvisioning
	}
	t.StatusMessage = fmt.Sprintf("Workflow execution started: %s", executionID)
	t.WorkflowExecutionID = &executionID

	if err := r.tenantRepo.UpdateTenant(ctx, t); err != nil {
		return fmt.Errorf("update tenant: %w", err)
	}

	duration := time.Since(startTime)
	r.logger.Info("tenant reconciled successfully",
		zap.String("tenant_id", tenantID),
		zap.String("tenant_name", t.Name),
		zap.String("status", string(t.Status)),
		zap.String("previous_status", string(previousStatus)),
		zap.String("execution_id", executionID),
		zap.Duration("duration", duration))

	return nil
}

func isInFlightStatus(status tenant.Status) bool {
	return status == tenant.StatusProvisioning ||
		status == tenant.StatusUpdating ||
		status == tenant.StatusDeleting ||
		status == tenant.StatusArchiving
}

func (r *Reconciler) handleWorkflowSuccess(ctx context.Context, t *tenant.Tenant, execStatus *workflow.ExecutionStatus) error {
	if t.Status == tenant.StatusDeleting {
		if err := r.tenantRepo.DeleteTenant(ctx, t.ID); err != nil {
			return fmt.Errorf("delete tenant after workflow: %w", err)
		}
		r.logger.Info("tenant deleted after workflow completion",
			zap.String("tenant_id", t.ID.String()),
			zap.String("tenant_name", t.Name),
		)
		return nil
	}
	if t.Status == tenant.StatusArchiving {
		if t.Annotations != nil && t.Annotations["landlord/delete_after_archive"] == "true" {
			if err := r.tenantRepo.DeleteTenant(ctx, t.ID); err != nil {
				return fmt.Errorf("delete tenant after archive workflow: %w", err)
			}
			r.logger.Info("tenant deleted after archive workflow completion",
				zap.String("tenant_id", t.ID.String()),
				zap.String("tenant_name", t.Name),
			)
			return nil
		}

		t.Status = tenant.StatusArchived
		t.StatusMessage = fmt.Sprintf("Workflow execution completed: %s", execStatus.ExecutionID)
		if err := r.tenantRepo.UpdateTenant(ctx, t); err != nil {
			return fmt.Errorf("update tenant: %w", err)
		}
		return nil
	}

	next, err := nextStatus(t.Status)
	if err != nil {
		return fmt.Errorf("determine next status: %w", err)
	}

	if len(execStatus.Output) > 0 {
		observed := make(map[string]interface{})
		if err := json.Unmarshal(execStatus.Output, &observed); err != nil {
			r.logger.Warn("failed to unmarshal workflow output",
				zap.String("tenant_id", t.ID.String()),
				zap.Error(err))
			observed["compute_result"] = string(execStatus.Output)
			t.ObservedConfig = observed
		} else {
			t.ObservedConfig = observed
		}
	}

	t.Status = next
	t.StatusMessage = fmt.Sprintf("Workflow execution completed: %s", execStatus.ExecutionID)

	if err := r.tenantRepo.UpdateTenant(ctx, t); err != nil {
		return fmt.Errorf("update tenant: %w", err)
	}

	return nil
}

func (r *Reconciler) handleWorkflowFailure(ctx context.Context, t *tenant.Tenant, execStatus *workflow.ExecutionStatus) error {
	message := fmt.Sprintf("Workflow execution failed: %s", execStatus.ExecutionID)
	if execStatus.Error != nil && execStatus.Error.Message != "" {
		message = fmt.Sprintf("%s: %s", message, execStatus.Error.Message)
	}

	t.Status = tenant.StatusFailed
	t.StatusMessage = message

	if err := r.tenantRepo.UpdateTenant(ctx, t); err != nil {
		return fmt.Errorf("update tenant: %w", err)
	}

	return nil
}

// handleReconcileError handles errors during reconciliation
func (r *Reconciler) handleReconcileError(tenantID string, err error) {
	retryCount := r.incrementRetryCount(tenantID)

	r.logger.Error("reconciliation failed",
		zap.String("tenant_id", tenantID),
		zap.Error(err),
		zap.Int("retry_count", retryCount))

	// Check if exceeded max retries
	if retryCount >= r.config.MaxRetries {
		r.logger.Error("max retries exceeded, marking tenant as failed",
			zap.String("tenant_id", tenantID))

		// Mark tenant as failed
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		tenantUUID, parseErr := uuid.Parse(tenantID)
		if parseErr != nil {
			r.logger.Error("failed to parse tenant id for failure update",
				zap.String("tenant_id", tenantID),
				zap.Error(parseErr))
			return
		}

		t, err := r.tenantRepo.GetTenantByID(ctx, tenantUUID)
		if err != nil {
			r.logger.Error("failed to fetch tenant for failure update",
				zap.String("tenant_id", tenantID),
				zap.Error(err))
			return
		}

		t.Status = tenant.StatusFailed
		t.StatusMessage = fmt.Sprintf("Reconciliation failed after %d retries: %v", retryCount, err)

		if err := r.tenantRepo.UpdateTenant(ctx, t); err != nil {
			r.logger.Error("failed to update tenant to failed status",
				zap.String("tenant_id", tenantID),
				zap.Error(err))
		}

		r.resetRetryCount(tenantID)
		return
	}

	// Requeue with rate limiting
	r.queue.AddRateLimited(tenantID)
}

// incrementRetryCount increments the retry counter for a tenant
func (r *Reconciler) incrementRetryCount(tenantID string) int {
	r.retryMu.Lock()
	defer r.retryMu.Unlock()
	r.retryCount[tenantID]++
	return r.retryCount[tenantID]
}

// resetRetryCount resets the retry counter for a tenant
func (r *Reconciler) resetRetryCount(tenantID string) {
	r.retryMu.Lock()
	defer r.retryMu.Unlock()
	delete(r.retryCount, tenantID)
}

// IsReady returns whether the controller is ready
func (r *Reconciler) IsReady() bool {
	return r.queue != nil && !r.queue.ShuttingDown()
}
