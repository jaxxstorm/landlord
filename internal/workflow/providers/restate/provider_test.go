package restate_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/jaxxstorm/landlord/internal/config"
	"github.com/jaxxstorm/landlord/internal/workflow"
	"github.com/jaxxstorm/landlord/internal/workflow/providers/restate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestProviderName(t *testing.T) {
	logger := zaptest.NewLogger(t)
	server := newFakeRestateServer(t)
	cfg := config.RestateConfig{
		Endpoint:           server.URL(),
		ExecutionMechanism: "local",
		AuthType:           "none",
		Timeout:            30 * time.Minute,
	}

	provider, err := restate.New(cfg, logger)
	require.NoError(t, err)
	require.NotNil(t, provider)

	assert.Equal(t, "restate", provider.Name())
}

func TestValidateWorkflowSpec(t *testing.T) {
	logger := zaptest.NewLogger(t)
	server := newFakeRestateServer(t)
	cfg := config.RestateConfig{
		Endpoint:           server.URL(),
		ExecutionMechanism: "local",
		AuthType:           "none",
		Timeout:            30 * time.Minute,
	}

	provider, err := restate.New(cfg, logger)
	require.NoError(t, err)

	ctx := context.Background()

	// Test with valid spec
	validSpec := &workflow.WorkflowSpec{
		WorkflowID: "test-workflow",
		Definition: json.RawMessage(`{"type":"test"}`),
	}
	err = provider.Validate(ctx, validSpec)
	assert.NoError(t, err)

	// Test with nil spec
	err = provider.Validate(ctx, nil)
	assert.Error(t, err)

	// Test with empty workflow ID
	invalidSpec := &workflow.WorkflowSpec{
		WorkflowID: "",
		Definition: json.RawMessage(`{}`),
	}
	err = provider.Validate(ctx, invalidSpec)
	assert.Error(t, err)
}

func TestCreateWorkflow(t *testing.T) {
	logger := zaptest.NewLogger(t)
	server := newFakeRestateServer(t)
	cfg := config.RestateConfig{
		Endpoint:           server.URL(),
		ExecutionMechanism: "local",
		AuthType:           "none",
		Timeout:            30 * time.Minute,
	}

	provider, err := restate.New(cfg, logger)
	require.NoError(t, err)

	ctx := context.Background()
	spec := &workflow.WorkflowSpec{
		WorkflowID: "test-workflow",
		Definition: json.RawMessage(`{"type":"test"}`),
	}

	result, err := provider.CreateWorkflow(ctx, spec)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-workflow", result.WorkflowID)
	assert.Equal(t, "restate", result.ProviderType)
}

// TestStartExecution tests Provider.StartExecution
func TestStartExecution(t *testing.T) {
	logger := zaptest.NewLogger(t)
	server := newFakeRestateServer(t)
	cfg := config.RestateConfig{
		Endpoint:           server.URL(),
		ExecutionMechanism: "local",
		AuthType:           "none",
		Timeout:            30 * time.Minute,
	}

	provider, err := restate.New(cfg, logger)
	require.NoError(t, err)

	ctx := context.Background()

	// Create workflow first
	spec := &workflow.WorkflowSpec{
		WorkflowID: "test-workflow",
		Definition: json.RawMessage(`{"type":"test"}`),
	}
	_, err = provider.CreateWorkflow(ctx, spec)
	require.NoError(t, err)

	// Start execution
	input := &workflow.ExecutionInput{
		ExecutionName: "test-execution",
		Input:         json.RawMessage(`{"key": "value"}`),
	}
	result, err := provider.StartExecution(ctx, "test-workflow", input)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestInvoke(t *testing.T) {
	logger := zaptest.NewLogger(t)
	server := newFakeRestateServer(t)
	cfg := config.RestateConfig{
		Endpoint:           server.URL(),
		ExecutionMechanism: "local",
		AuthType:           "none",
		Timeout:            30 * time.Minute,
	}

	provider, err := restate.New(cfg, logger)
	require.NoError(t, err)

	ctx := context.Background()

	// Create workflow first
	spec := &workflow.WorkflowSpec{
		WorkflowID: "test-workflow",
		Definition: json.RawMessage(`{"type":"test"}`),
	}
	_, err = provider.CreateWorkflow(ctx, spec)
	require.NoError(t, err)

	request := &workflow.ProvisionRequest{
		TenantID:      "tenant-1",
		TenantUUID:    "tenant-uuid-1",
		Operation:     "provision",
		DesiredConfig: map[string]interface{}{"image": "nginx:latest", "replicas": "1"},
	}

	result, err := provider.Invoke(ctx, "test-workflow", request)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "tenant-tenant-uuid-1-test-workflow-provision", result.ExecutionID)
	assert.Equal(t, workflow.StateRunning, result.State)

	payload := server.LastInvokePayload()
	require.NotEmpty(t, payload, "expected invoke payload to be captured")

	var captured map[string]interface{}
	require.NoError(t, json.Unmarshal(payload, &captured))
	assert.Equal(t, "tenant-1", captured["tenant_id"])
	assert.Equal(t, "tenant-uuid-1", captured["tenant_uuid"])
	assert.Equal(t, "provision", captured["operation"])
	if desiredConfig, ok := captured["desired_config"].(map[string]interface{}); ok {
		assert.Equal(t, "nginx:latest", desiredConfig["image"])
		assert.Equal(t, "1", desiredConfig["replicas"])
	} else {
		t.Fatalf("expected desired_config to be an object, got %T", captured["desired_config"])
	}
}

func TestProviderStartupWithUnavailableRestateBackend(t *testing.T) {
	logger := zaptest.NewLogger(t)
	server := newFakeRestateServer(t)
	cfg := config.RestateConfig{
		Endpoint:           server.URL(),
		ExecutionMechanism: "local",
		AuthType:           "none",
		Timeout:            100 * time.Millisecond,
	}

	// Force registration failures by pointing admin endpoint to a path that returns 404.
	cfg.AdminEndpoint = server.URL() + "/missing"

	provider, err := restate.New(cfg, logger)
	require.NoError(t, err, "provider should still initialize when registration fails")
	require.NotNil(t, provider)
}

func TestGetWorkflowStatus(t *testing.T) {
	logger := zaptest.NewLogger(t)
	server := newFakeRestateServer(t)
	cfg := config.RestateConfig{
		Endpoint:           server.URL(),
		ExecutionMechanism: "local",
		AuthType:           "none",
		Timeout:            30 * time.Minute,
	}

	provider, err := restate.New(cfg, logger)
	require.NoError(t, err)

	ctx := context.Background()

	status, err := provider.GetWorkflowStatus(ctx, "inv_test-execution")
	require.NoError(t, err)
	require.NotNil(t, status)
	assert.Equal(t, "inv_test-execution", status.ExecutionID)
	assert.Equal(t, workflow.StateSucceeded, status.State)
}

// TestGetExecutionStatus tests Provider.GetExecutionStatus
func TestGetExecutionStatus(t *testing.T) {
	logger := zaptest.NewLogger(t)
	server := newFakeRestateServer(t)
	cfg := config.RestateConfig{
		Endpoint:           server.URL(),
		ExecutionMechanism: "local",
		AuthType:           "none",
		Timeout:            30 * time.Minute,
	}

	provider, err := restate.New(cfg, logger)
	require.NoError(t, err)

	ctx := context.Background()

	// Get status for a test execution
	status, err := provider.GetExecutionStatus(ctx, "inv_test-execution")
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, "inv_test-execution", status.ExecutionID)
	assert.Equal(t, "restate", status.ProviderType)
}

// TestStopExecution tests Provider.StopExecution
func TestStopExecution(t *testing.T) {
	logger := zaptest.NewLogger(t)
	server := newFakeRestateServer(t)
	cfg := config.RestateConfig{
		Endpoint:           server.URL(),
		ExecutionMechanism: "local",
		AuthType:           "none",
		Timeout:            30 * time.Minute,
	}

	provider, err := restate.New(cfg, logger)
	require.NoError(t, err)

	ctx := context.Background()

	// Stop execution (should be idempotent)
	err = provider.StopExecution(ctx, "test-execution", "user requested cancellation")
	assert.NoError(t, err)

	// Stop again (should still be idempotent)
	err = provider.StopExecution(ctx, "test-execution", "user requested cancellation")
	assert.NoError(t, err)
}

// TestDeleteWorkflow tests Provider.DeleteWorkflow
func TestDeleteWorkflow(t *testing.T) {
	logger := zaptest.NewLogger(t)
	server := newFakeRestateServer(t)
	cfg := config.RestateConfig{
		Endpoint:           server.URL(),
		ExecutionMechanism: "local",
		AuthType:           "none",
		Timeout:            30 * time.Minute,
	}

	provider, err := restate.New(cfg, logger)
	require.NoError(t, err)

	ctx := context.Background()

	err = provider.DeleteWorkflow(ctx, "test-workflow")
	assert.NoError(t, err)

	// Delete again (should be consistent behavior)
	err = provider.DeleteWorkflow(ctx, "test-workflow")
	assert.NoError(t, err)
}

// TestCreateWorkflowIdempotency tests that CreateWorkflow is idempotent
func TestCreateWorkflowIdempotency(t *testing.T) {
	logger := zaptest.NewLogger(t)
	server := newFakeRestateServer(t)
	cfg := config.RestateConfig{
		Endpoint:           server.URL(),
		ExecutionMechanism: "local",
		AuthType:           "none",
		Timeout:            30 * time.Minute,
	}

	provider, err := restate.New(cfg, logger)
	require.NoError(t, err)

	ctx := context.Background()
	spec := &workflow.WorkflowSpec{
		WorkflowID: "idempotent-workflow",
		Definition: json.RawMessage(`{"type":"test"}`),
	}

	// Create workflow
	result1, err := provider.CreateWorkflow(ctx, spec)
	require.NoError(t, err)
	assert.NotNil(t, result1)

	// Create again with same spec
	result2, err := provider.CreateWorkflow(ctx, spec)
	require.NoError(t, err)
	assert.NotNil(t, result2)

	// Both should succeed (idempotent)
	assert.Equal(t, result1.WorkflowID, result2.WorkflowID)
}

func TestStartExecutionRequiresRegisteredService(t *testing.T) {
	logger := zaptest.NewLogger(t)
	server := newFakeRestateServer(t)
	cfg := config.RestateConfig{
		Endpoint:           server.URL(),
		ExecutionMechanism: "local",
		AuthType:           "none",
		Timeout:            30 * time.Minute,
	}

	provider, err := restate.New(cfg, logger)
	require.NoError(t, err)

	input := &workflow.ExecutionInput{
		ExecutionName: "missing-exec",
		Input:         json.RawMessage(`{"key": "value"}`),
	}

	_, err = provider.StartExecution(context.Background(), "missing-workflow", input)
	require.Error(t, err)
	assert.True(t, errors.Is(err, workflow.ErrWorkflowNotFound))
}

func TestWorkflowRegistrationAfterRestart(t *testing.T) {
	logger := zaptest.NewLogger(t)
	server := newFakeRestateServer(t)
	cfg := config.RestateConfig{
		Endpoint:           server.URL(),
		ExecutionMechanism: "local",
		AuthType:           "none",
		Timeout:            30 * time.Minute,
	}

	provider, err := restate.New(cfg, logger)
	require.NoError(t, err)

	_, err = provider.StartExecution(context.Background(), "tenant-provisioning", &workflow.ExecutionInput{
		ExecutionName: "restart-exec-1",
		Input:         json.RawMessage(`{}`),
	})
	require.NoError(t, err)

	provider, err = restate.New(cfg, logger)
	require.NoError(t, err)

	_, err = provider.StartExecution(context.Background(), "tenant-provisioning", &workflow.ExecutionInput{
		ExecutionName: "restart-exec-2",
		Input:         json.RawMessage(`{}`),
	})
	require.NoError(t, err)
}
