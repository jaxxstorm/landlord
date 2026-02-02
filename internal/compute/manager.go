package compute

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// WorkflowProvider defines the minimal interface for workflow callback posting
type WorkflowProvider interface {
	PostComputeCallback(ctx context.Context, executionID string, payload *CallbackPayload, opts *CallbackOptions) error
}

// Manager coordinates compute provisioning operations
type Manager struct {
	registry            *Registry
	executionRepository ExecutionRepository
	workflowProvider    WorkflowProvider
	logger              *zap.Logger

	// failedCallbacks stores callbacks that failed delivery for manual retry
	failedCallbacks   map[string]*FailedCallback
	failedCallbacksMu sync.RWMutex
}

// New creates a new compute manager
func New(registry *Registry, logger *zap.Logger) *Manager {
	return &Manager{
		registry:        registry,
		logger:          logger.With(zap.String("component", "compute-manager")),
		failedCallbacks: make(map[string]*FailedCallback),
	}
}

// NewWithTracking creates a compute manager with execution tracking capability
func NewWithTracking(registry *Registry, execRepo ExecutionRepository, logger *zap.Logger) *Manager {
	return &Manager{
		registry:            registry,
		executionRepository: execRepo,
		logger:              logger.With(zap.String("component", "compute-manager")),
		failedCallbacks:     make(map[string]*FailedCallback),
	}
}

// SetWorkflowProvider sets the workflow provider for callback delivery
func (m *Manager) SetWorkflowProvider(wp WorkflowProvider) {
	m.workflowProvider = wp
}

// GenerateComputeExecutionID creates a deterministic execution ID from tenant ID and operation type
// This enables idempotency - the same tenant + operation always produces the same ID
func (m *Manager) GenerateComputeExecutionID(tenantID string, operationType ComputeOperationType) string {
	// Create a deterministic hash from tenant ID and operation type
	// Format: "<tenantID>-<operationType>-<hash of full string>"
	key := fmt.Sprintf("%s:%s", tenantID, operationType)
	hash := md5.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])[:12] // Use first 12 chars of hash

	return fmt.Sprintf("%s-%s-%s", tenantID, operationType, hashStr)
}

// ProvisionTenant provisions compute resources for a tenant
func (m *Manager) ProvisionTenant(ctx context.Context, spec *TenantComputeSpec) (*ProvisionResult, error) {
	m.logger.Info("provisioning tenant",
		zap.String("tenant_id", spec.TenantID),
		zap.String("provider", spec.ProviderType),
	)

	// Validate spec
	if err := ValidateComputeSpec(spec); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidSpec, err)
	}

	ApplyDefaultMetadata(spec)

	// Get provider
	provider, err := m.registry.Get(spec.ProviderType)
	if err != nil {
		return nil, err
	}

	// Delegate to provider
	result, err := provider.Provision(ctx, spec)
	if err != nil {
		m.logger.Error("provisioning failed",
			zap.String("tenant_id", spec.TenantID),
			zap.Error(err),
		)
		return nil, err
	}

	m.logger.Info("provisioning completed",
		zap.String("tenant_id", spec.TenantID),
		zap.String("status", string(result.Status)),
	)

	return result, nil
}

// UpdateTenant updates existing compute resources
func (m *Manager) UpdateTenant(ctx context.Context, tenantID string, spec *TenantComputeSpec) (*UpdateResult, error) {
	m.logger.Info("updating tenant",
		zap.String("tenant_id", tenantID),
		zap.String("provider", spec.ProviderType),
	)

	// Validate spec
	if err := ValidateComputeSpec(spec); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidSpec, err)
	}

	ApplyDefaultMetadata(spec)

	// Get provider
	provider, err := m.registry.Get(spec.ProviderType)
	if err != nil {
		return nil, err
	}

	// Delegate to provider
	result, err := provider.Update(ctx, tenantID, spec)
	if err != nil {
		m.logger.Error("update failed",
			zap.String("tenant_id", tenantID),
			zap.Error(err),
		)
		return nil, err
	}

	m.logger.Info("update completed",
		zap.String("tenant_id", tenantID),
		zap.String("status", string(result.Status)),
	)

	return result, nil
}

// DestroyTenant removes compute resources for a tenant
func (m *Manager) DestroyTenant(ctx context.Context, tenantID, providerType string) error {
	m.logger.Info("destroying tenant",
		zap.String("tenant_id", tenantID),
		zap.String("provider", providerType),
	)

	// Get provider
	provider, err := m.registry.Get(providerType)
	if err != nil {
		return err
	}

	// Delegate to provider
	if err := provider.Destroy(ctx, tenantID); err != nil {
		m.logger.Error("destroy failed",
			zap.String("tenant_id", tenantID),
			zap.Error(err),
		)
		return err
	}

	m.logger.Info("tenant destroyed",
		zap.String("tenant_id", tenantID),
	)

	return nil
}

// GetTenantStatus queries current status of tenant compute
func (m *Manager) GetTenantStatus(ctx context.Context, tenantID, providerType string) (*ComputeStatus, error) {
	// Get provider
	provider, err := m.registry.Get(providerType)
	if err != nil {
		return nil, err
	}

	// Delegate to provider
	status, err := provider.GetStatus(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return status, nil
}

// ValidateTenantSpec validates a spec without provisioning
func (m *Manager) ValidateTenantSpec(ctx context.Context, spec *TenantComputeSpec) error {
	// Validate structure
	if err := ValidateComputeSpec(spec); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidSpec, err)
	}

	ApplyDefaultMetadata(spec)

	// Get provider for provider-specific validation
	provider, err := m.registry.Get(spec.ProviderType)
	if err != nil {
		return err
	}

	// Delegate to provider
	return provider.Validate(ctx, spec)
}

// ListProviders returns available provider types
func (m *Manager) ListProviders() []string {
	return m.registry.List()
}

// Health performs a health check on the compute manager
func (m *Manager) Health(ctx context.Context) error {
	// Basic health check - just verify the manager is initialized
	if m.registry == nil {
		return fmt.Errorf("compute registry not initialized")
	}
	return nil
}

// ProvisionTenantWithTracking provisions compute with execution tracking
func (m *Manager) ProvisionTenantWithTracking(ctx context.Context, spec *TenantComputeSpec, workflowExecutionID string) (*ComputeExecution, error) {
	if m.executionRepository == nil {
		return nil, fmt.Errorf("execution repository not configured")
	}

	ApplyDefaultMetadata(spec)

	// Generate deterministic execution ID
	executionID := m.GenerateComputeExecutionID(spec.TenantID, OperationTypeProvision)

	m.logger.Info("provisioning tenant with tracking",
		zap.String("tenant_id", spec.TenantID),
		zap.String("execution_id", executionID),
		zap.String("provider", spec.ProviderType),
	)

	// Create execution record in pending state
	exec := &ComputeExecution{
		ExecutionID:         executionID,
		TenantID:            spec.TenantID,
		WorkflowExecutionID: workflowExecutionID,
		OperationType:       OperationTypeProvision,
		Status:              ExecutionStatusPending,
	}

	if err := m.executionRepository.CreateComputeExecution(ctx, exec); err != nil {
		m.logger.Error("failed to create execution record",
			zap.String("tenant_id", spec.TenantID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to create execution record: %w", err)
	}

	// Add history entry
	historyDetails := map[string]string{"trigger": "workflow"}
	detailsJSON, _ := json.Marshal(historyDetails)
	history := &ComputeExecutionHistory{
		ComputeExecutionID: executionID,
		Status:             ExecutionStatusPending,
		Details:            detailsJSON,
	}
	_ = m.executionRepository.AddExecutionHistory(ctx, history)

	// Update to running state
	exec.Status = ExecutionStatusRunning
	if err := m.executionRepository.UpdateComputeExecution(ctx, exec); err != nil {
		m.logger.Error("failed to update execution to running",
			zap.String("execution_id", executionID),
			zap.Error(err),
		)
		return nil, err
	}

	// Add history entry for running
	history.Status = ExecutionStatusRunning
	_ = m.executionRepository.AddExecutionHistory(ctx, history)

	// Call provider
	result, err := m.ProvisionTenant(ctx, spec)
	if err != nil {
		// Mark as failed
		errCode := "PROVISIONING_FAILED"
		errMsg := err.Error()
		exec.Status = ExecutionStatusFailed
		exec.ErrorCode = &errCode
		exec.ErrorMessage = &errMsg
		_ = m.executionRepository.UpdateComputeExecution(ctx, exec)

		// Add failure history
		failureDetails := map[string]string{"error": err.Error()}
		detailsJSON, _ := json.Marshal(failureDetails)
		history.Status = ExecutionStatusFailed
		history.Details = detailsJSON
		_ = m.executionRepository.AddExecutionHistory(ctx, history)

		// Post failure callback to workflow provider
		m.postCallbackWithRetry(ctx, executionID, exec, err)

		return exec, err
	}

	// Mark as succeeded with resource IDs
	exec.Status = ExecutionStatusSucceeded
	if result != nil && result.ResourceIDs != nil {
		resourceJSON, _ := json.Marshal(result.ResourceIDs)
		exec.ResourceIDs = resourceJSON
	}
	if err := m.executionRepository.UpdateComputeExecution(ctx, exec); err != nil {
		m.logger.Error("failed to update execution to succeeded",
			zap.String("execution_id", executionID),
			zap.Error(err),
		)
		return nil, err
	}

	// Add success history
	successDetails := map[string]interface{}{"status": "completed", "resources": result.ResourceIDs}
	successJSON, _ := json.Marshal(successDetails)
	history.Status = ExecutionStatusSucceeded
	history.Details = successJSON
	_ = m.executionRepository.AddExecutionHistory(ctx, history)

	m.logger.Info("provisioning with tracking completed",
		zap.String("tenant_id", spec.TenantID),
		zap.String("execution_id", executionID),
	)

	// Post callback to workflow provider about successful completion
	m.postCallbackWithRetry(ctx, executionID, exec, nil)

	return exec, nil
}

// UpdateTenantWithTracking updates compute with execution tracking
func (m *Manager) UpdateTenantWithTracking(ctx context.Context, tenantID string, spec *TenantComputeSpec, workflowExecutionID string) (*ComputeExecution, error) {
	if m.executionRepository == nil {
		return nil, fmt.Errorf("execution repository not configured")
	}

	ApplyDefaultMetadata(spec)

	executionID := m.GenerateComputeExecutionID(tenantID, OperationTypeUpdate)

	m.logger.Info("updating tenant with tracking",
		zap.String("tenant_id", tenantID),
		zap.String("execution_id", executionID),
	)

	// Create execution record
	exec := &ComputeExecution{
		ExecutionID:         executionID,
		TenantID:            tenantID,
		WorkflowExecutionID: workflowExecutionID,
		OperationType:       OperationTypeUpdate,
		Status:              ExecutionStatusPending,
	}

	if err := m.executionRepository.CreateComputeExecution(ctx, exec); err != nil {
		return nil, fmt.Errorf("failed to create execution record: %w", err)
	}

	// Update to running
	exec.Status = ExecutionStatusRunning
	_ = m.executionRepository.UpdateComputeExecution(ctx, exec)

	// Call provider
	result, err := m.UpdateTenant(ctx, tenantID, spec)
	if err != nil {
		errCode := "UPDATE_FAILED"
		errMsg := err.Error()
		exec.Status = ExecutionStatusFailed
		exec.ErrorCode = &errCode
		exec.ErrorMessage = &errMsg
		_ = m.executionRepository.UpdateComputeExecution(ctx, exec)

		// Post failure callback to workflow provider
		m.postCallbackWithRetry(ctx, executionID, exec, err)

		return exec, err
	}

	// Mark as succeeded
	exec.Status = ExecutionStatusSucceeded
	if result != nil && result.Changes != nil {
		changesJSON, _ := json.Marshal(result.Changes)
		exec.ResourceIDs = changesJSON
	}
	_ = m.executionRepository.UpdateComputeExecution(ctx, exec)

	// Post success callback to workflow provider
	m.postCallbackWithRetry(ctx, executionID, exec, nil)

	return exec, nil
}

// DeleteTenantWithTracking deletes compute with execution tracking
func (m *Manager) DeleteTenantWithTracking(ctx context.Context, tenantID, providerType, workflowExecutionID string) (*ComputeExecution, error) {
	if m.executionRepository == nil {
		return nil, fmt.Errorf("execution repository not configured")
	}

	executionID := m.GenerateComputeExecutionID(tenantID, OperationTypeDelete)

	m.logger.Info("deleting tenant with tracking",
		zap.String("tenant_id", tenantID),
		zap.String("execution_id", executionID),
	)

	// Create execution record
	exec := &ComputeExecution{
		ExecutionID:         executionID,
		TenantID:            tenantID,
		WorkflowExecutionID: workflowExecutionID,
		OperationType:       OperationTypeDelete,
		Status:              ExecutionStatusPending,
	}

	if err := m.executionRepository.CreateComputeExecution(ctx, exec); err != nil {
		return nil, fmt.Errorf("failed to create execution record: %w", err)
	}

	// Update to running
	exec.Status = ExecutionStatusRunning
	_ = m.executionRepository.UpdateComputeExecution(ctx, exec)

	// Call provider
	err := m.DestroyTenant(ctx, tenantID, providerType)
	if err != nil {
		errCode := "DELETE_FAILED"
		errMsg := err.Error()
		exec.Status = ExecutionStatusFailed
		exec.ErrorCode = &errCode
		exec.ErrorMessage = &errMsg
		_ = m.executionRepository.UpdateComputeExecution(ctx, exec)

		// Post failure callback to workflow provider
		m.postCallbackWithRetry(ctx, executionID, exec, err)

		return exec, err
	}

	// Mark as succeeded
	exec.Status = ExecutionStatusSucceeded
	_ = m.executionRepository.UpdateComputeExecution(ctx, exec)

	// Post success callback to workflow provider
	m.postCallbackWithRetry(ctx, executionID, exec, nil)

	return exec, nil
}

// GetComputeExecution retrieves an execution by ID (for workflow queries)
func (m *Manager) GetComputeExecution(ctx context.Context, executionID string) (*ComputeExecution, error) {
	if m.executionRepository == nil {
		return nil, fmt.Errorf("execution repository not configured")
	}

	return m.executionRepository.GetComputeExecution(ctx, executionID)
}

// MapProviderErrorToComputeError converts provider errors to standardized ComputeError
func (m *Manager) MapProviderErrorToComputeError(err error) *ComputeError {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Classify common error types
	isRetriable := false
	code := "UNKNOWN_ERROR"

	// Check for timeout/temporary errors
	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "Timeout") || strings.Contains(errStr, "deadline exceeded") || strings.Contains(errStr, "context deadline exceeded") {
		code = "PROVIDER_TIMEOUT"
		isRetriable = true
	} else if strings.Contains(errStr, "unavailable") || strings.Contains(errStr, "Unavailable") {
		code = "PROVIDER_UNAVAILABLE"
		isRetriable = true
	} else if strings.Contains(errStr, "exhausted") || strings.Contains(errStr, "quota") {
		code = "RESOURCE_EXHAUSTED"
		isRetriable = true
	} else if strings.Contains(errStr, "invalid") || strings.Contains(errStr, "Invalid") {
		code = "INVALID_CONFIGURATION"
		isRetriable = false
	} else if strings.Contains(errStr, "not found") || strings.Contains(errStr, "NotFound") {
		code = "RESOURCE_NOT_FOUND"
		isRetriable = false
	}

	return &ComputeError{
		Code:          code,
		Message:       errStr,
		IsRetriable:   isRetriable,
		ProviderError: errStr,
	}
}

// postCallbackWithRetry posts a callback to the workflow provider with retry logic
func (m *Manager) postCallbackWithRetry(ctx context.Context, executionID string, exec *ComputeExecution, opErr error) {
	// If no workflow provider is configured, skip callback
	if m.workflowProvider == nil {
		return
	}

	// Construct callback payload
	payload := &CallbackPayload{
		ExecutionID: executionID,
		TenantID:    exec.TenantID,
		Status:      exec.Status,
	}

	// Add resource IDs if succeeded
	if exec.Status == ExecutionStatusSucceeded && exec.ResourceIDs != nil {
		var resourceIDs map[string]interface{}
		if err := json.Unmarshal(exec.ResourceIDs, &resourceIDs); err == nil {
			payload.ResourceIDs = resourceIDs
		}
	}

	// Add error details if failed
	if exec.Status == ExecutionStatusFailed {
		if exec.ErrorCode != nil {
			payload.ErrorCode = *exec.ErrorCode
		}
		if exec.ErrorMessage != nil {
			payload.ErrorMessage = *exec.ErrorMessage
		}
		// Determine if retriable based on error classification
		if opErr != nil {
			computeErr := m.MapProviderErrorToComputeError(opErr)
			payload.IsRetriable = computeErr.IsRetriable
		}
	}

	// Create callback options with retry settings
	opts := &CallbackOptions{
		MaxRetries:  3,
		RetryDelay:  time.Second * 1,
		BackoffType: "exponential",
	}

	// Retry with exponential backoff (up to 3 retries)
	var lastErr error
	for attempt := 0; attempt <= opts.MaxRetries; attempt++ {
		// Create callback context with timeout
		callbackCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		err := m.workflowProvider.PostComputeCallback(callbackCtx, executionID, payload, opts)
		cancel()

		if err == nil {
			// Success!
			m.logger.Info("compute callback delivered",
				zap.String("execution_id", executionID),
				zap.String("tenant_id", exec.TenantID),
				zap.Int("attempt", attempt+1),
			)
			return
		}

		lastErr = err

		// If this was the last attempt, break
		if attempt == opts.MaxRetries {
			break
		}

		// Calculate backoff delay
		var delay time.Duration
		if opts.BackoffType == "exponential" {
			// Exponential backoff: 1s, 2s, 4s
			delay = opts.RetryDelay * time.Duration(1<<uint(attempt))
		} else {
			// Linear backoff
			delay = opts.RetryDelay * time.Duration(attempt+1)
		}

		m.logger.Warn("callback delivery failed, retrying",
			zap.String("execution_id", executionID),
			zap.String("tenant_id", exec.TenantID),
			zap.Int("attempt", attempt+1),
			zap.Duration("retry_after", delay),
			zap.Error(err),
		)

		// Wait before retry
		time.Sleep(delay)
	}

	// All retries exhausted
	m.logger.Error("failed to deliver compute callback after all retries",
		zap.String("execution_id", executionID),
		zap.String("tenant_id", exec.TenantID),
		zap.Int("attempts", opts.MaxRetries+1),
		zap.Error(lastErr),
	)

	// Store failed callback for manual retry (Task 6.6)
	m.storeFailedCallback(executionID, payload, lastErr)
}

// storeFailedCallback stores a failed callback for manual retry/investigation
func (m *Manager) storeFailedCallback(executionID string, payload *CallbackPayload, err error) {
	m.failedCallbacksMu.Lock()
	defer m.failedCallbacksMu.Unlock()

	now := time.Now()
	if existing, ok := m.failedCallbacks[executionID]; ok {
		// Update existing failed callback
		existing.Attempts++
		existing.LastAttemptAt = now
		existing.Error = err.Error()
	} else {
		// Create new failed callback record
		m.failedCallbacks[executionID] = &FailedCallback{
			ExecutionID:   executionID,
			Payload:       payload,
			Error:         err.Error(),
			Attempts:      1,
			FailedAt:      now,
			LastAttemptAt: now,
		}
	}

	m.logger.Info("stored failed callback for manual retry",
		zap.String("execution_id", executionID),
		zap.String("tenant_id", payload.TenantID),
	)
}

// GetFailedCallbacks returns all failed callbacks for manual inspection/retry
func (m *Manager) GetFailedCallbacks() []*FailedCallback {
	m.failedCallbacksMu.RLock()
	defer m.failedCallbacksMu.RUnlock()

	callbacks := make([]*FailedCallback, 0, len(m.failedCallbacks))
	for _, cb := range m.failedCallbacks {
		callbacks = append(callbacks, cb)
	}
	return callbacks
}

// RetryFailedCallback attempts to re-deliver a failed callback
func (m *Manager) RetryFailedCallback(executionID string) error {
	m.failedCallbacksMu.RLock()
	failed, ok := m.failedCallbacks[executionID]
	m.failedCallbacksMu.RUnlock()

	if !ok {
		return fmt.Errorf("no failed callback found for execution %s", executionID)
	}

	if m.workflowProvider == nil {
		return fmt.Errorf("no workflow provider configured")
	}

	// Try to deliver the callback once
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := &CallbackOptions{
		MaxRetries:  0, // No retries for manual retry
		RetryDelay:  0,
		BackoffType: "exponential",
	}

	err := m.workflowProvider.PostComputeCallback(ctx, executionID, failed.Payload, opts)
	if err != nil {
		// Update failure record
		m.storeFailedCallback(executionID, failed.Payload, err)
		return fmt.Errorf("manual retry failed: %w", err)
	}

	// Success - remove from failed callbacks
	m.failedCallbacksMu.Lock()
	delete(m.failedCallbacks, executionID)
	m.failedCallbacksMu.Unlock()

	m.logger.Info("manually retried callback succeeded",
		zap.String("execution_id", executionID),
	)

	return nil
}
