package temporal

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/jaxxstorm/landlord/internal/config"
	"github.com/jaxxstorm/landlord/internal/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func temporalIntegrationConfig() config.TemporalConfig {
	hostPort := os.Getenv("TEMPORAL_HOST_PORT")
	if hostPort == "" {
		hostPort = "localhost:7233"
	}
	namespace := os.Getenv("TEMPORAL_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}
	taskQueue := os.Getenv("TEMPORAL_TASK_QUEUE")
	if taskQueue == "" {
		taskQueue = "landlord"
	}
	return config.TemporalConfig{
		HostPort:      hostPort,
		Namespace:     namespace,
		TaskQueue:     taskQueue,
		Timeout:       30 * time.Second,
		RetryAttempts: 2,
	}
}

func requireTemporalIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("skipping integration test: INTEGRATION_TEST not set")
	}
}

func TestTemporalProviderIntegrationStartStatusStop(t *testing.T) {
	requireTemporalIntegration(t)

	provider, err := New(temporalIntegrationConfig(), zaptest.NewLogger(t))
	require.NoError(t, err)

	ctx := context.Background()
	_, err = provider.CreateWorkflow(ctx, &workflow.WorkflowSpec{
		WorkflowID:  "tenant-provisioning",
		Definition:  json.RawMessage(`{"type":"temporal"}`),
		ProviderType: "temporal",
		Name:        "Tenant Provisioning",
	})
	require.NoError(t, err)

	execName := "integration-temporal-" + time.Now().Format("20060102150405")
	started, err := provider.StartExecution(ctx, "tenant-provisioning", &workflow.ExecutionInput{
		ExecutionName: execName,
		Input:         json.RawMessage(`{"tenant_id":"acme","operation":"provision"}`),
	})
	require.NoError(t, err)
	require.NotEmpty(t, started.ExecutionID)

	status, err := provider.GetExecutionStatus(ctx, started.ExecutionID)
	require.NoError(t, err)
	assert.NotEmpty(t, status.Metadata["workflow_sub_state"])

	require.NoError(t, provider.StopExecution(ctx, started.ExecutionID, "integration cancellation"))

	cancelled, err := provider.GetExecutionStatus(ctx, started.ExecutionID)
	require.NoError(t, err)
	assert.Contains(t, []workflow.ExecutionState{workflow.StateCancelled, workflow.StateFailed, workflow.StateRunning}, cancelled.State)
}
