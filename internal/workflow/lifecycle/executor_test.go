package lifecycle

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jaxxstorm/landlord/internal/compute"
	computemock "github.com/jaxxstorm/landlord/internal/compute/providers/mock"
	"github.com/jaxxstorm/landlord/internal/workflow"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

type trackingComputeManager struct {
	provisionWorkflowID string
	updateWorkflowID    string
	deleteWorkflowID    string
	execution           *compute.ComputeExecution
}

type cancelledMutationGuard struct{}

func (cancelledMutationGuard) BeforeMutation(context.Context, *workflow.ProvisionRequest) error {
	return context.Canceled
}

func (m *trackingComputeManager) ProvisionTenantWithTracking(_ context.Context, _ *compute.TenantComputeSpec, workflowExecutionID string) (*compute.ComputeExecution, error) {
	m.provisionWorkflowID = workflowExecutionID
	return m.execution, nil
}

func (m *trackingComputeManager) UpdateTenantWithTracking(_ context.Context, _ string, _ *compute.TenantComputeSpec, workflowExecutionID string) (*compute.ComputeExecution, error) {
	m.updateWorkflowID = workflowExecutionID
	return m.execution, nil
}

func (m *trackingComputeManager) DeleteTenantWithTracking(_ context.Context, _ string, _ string, workflowExecutionID string) (*compute.ComputeExecution, error) {
	m.deleteWorkflowID = workflowExecutionID
	return m.execution, nil
}

func (m *trackingComputeManager) GetComputeExecution(context.Context, string) (*compute.ComputeExecution, error) {
	return nil, nil
}

func (m *trackingComputeManager) ListProviders() []string { return nil }

func (m *trackingComputeManager) Health(context.Context) error { return nil }

func (m *trackingComputeManager) MapProviderErrorToComputeError(error) *compute.ComputeError {
	return nil
}

func newTestExecutor(t *testing.T) *Executor {
	t.Helper()
	logger := zaptest.NewLogger(t)
	registry := compute.NewRegistry(logger)
	require.NoError(t, registry.Register(computemock.New()))
	executor, err := NewExecutor(registry, nil, "test", logger)
	require.NoError(t, err)
	return executor
}

func TestExecutorProvisionUpdateDelete(t *testing.T) {
	executor := newTestExecutor(t)

	for _, operation := range []string{"provision", "update", "delete"} {
		t.Run(operation, func(t *testing.T) {
			status, err := executor.Execute(context.Background(), &workflow.ProvisionRequest{
				TenantID:        "acme",
				Operation:       operation,
				ComputeProvider: "mock",
				DesiredConfig:   map[string]interface{}{"image": "nginx:latest"},
			})
			require.NoError(t, err)
			require.Equal(t, workflow.StateSucceeded, status.State)
			require.Equal(t, "test", status.ProviderType)
		})
	}
}

func TestExecutorRejectsInvalidRequests(t *testing.T) {
	executor := newTestExecutor(t)

	_, err := executor.Execute(context.Background(), nil)
	require.ErrorContains(t, err, "request is required")

	_, err = executor.Execute(context.Background(), &workflow.ProvisionRequest{Operation: "provision"})
	require.ErrorContains(t, err, "tenant identifier is required")

	_, err = executor.Execute(context.Background(), &workflow.ProvisionRequest{TenantID: "acme", Operation: "unknown"})
	require.ErrorContains(t, err, "unknown operation")
}

func TestExecutorRejectsUnknownComputeProvider(t *testing.T) {
	executor := newTestExecutor(t)

	_, err := executor.Execute(context.Background(), &workflow.ProvisionRequest{
		TenantID:        "acme",
		Operation:       "provision",
		ComputeProvider: "does-not-exist",
	})
	require.ErrorContains(t, err, "compute provider lookup failed")
}

func TestExecutorUsesTrackedComputeManager(t *testing.T) {
	executor := newTestExecutor(t)
	manager := &trackingComputeManager{execution: &compute.ComputeExecution{
		ExecutionID: "compute-123",
		Status:      compute.ExecutionStatusSucceeded,
		ResourceIDs: json.RawMessage(`{"service":"acme"}`),
	}}
	executor.WithComputeManager(manager)

	for _, operation := range []string{"provision", "update", "delete"} {
		t.Run(operation, func(t *testing.T) {
			status, err := executor.Execute(context.Background(), &workflow.ProvisionRequest{
				TenantID:        "acme",
				Operation:       operation,
				ComputeProvider: "mock",
				Metadata:        map[string]string{"workflow_execution_id": "arn:aws:states:region:account:execution:machine:acme"},
			})
			require.NoError(t, err)
			require.JSONEq(t, `{"compute_execution_id":"compute-123","compute_status":"succeeded","tenant_id":"acme","resource_ids":{"service":"acme"}}`, string(status.Output))
		})
	}
	require.NotEmpty(t, manager.provisionWorkflowID)
	require.NotEmpty(t, manager.updateWorkflowID)
	require.NotEmpty(t, manager.deleteWorkflowID)
}

func TestExecutorChecksCancellationBeforeTrackedMutation(t *testing.T) {
	executor := newTestExecutor(t)
	manager := &trackingComputeManager{execution: &compute.ComputeExecution{ExecutionID: "compute-123"}}
	executor.WithComputeManager(manager).WithMutationGuard(cancelledMutationGuard{})

	_, err := executor.Execute(context.Background(), &workflow.ProvisionRequest{
		TenantID:        "acme",
		ComputeProvider: "mock",
		Metadata:        map[string]string{"workflow_execution_id": "arn:aws:states:region:account:execution:machine:acme"},
	})
	require.ErrorIs(t, err, context.Canceled)
	require.Empty(t, manager.provisionWorkflowID)
}
