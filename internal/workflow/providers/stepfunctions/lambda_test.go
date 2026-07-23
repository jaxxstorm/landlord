package stepfunctions

import (
	"context"
	"errors"
	"testing"

	"github.com/jaxxstorm/landlord/internal/workflow"
	"github.com/stretchr/testify/require"
)

type testLifecycleExecutor struct {
	request *workflow.ProvisionRequest
	status  *workflow.ExecutionStatus
	err     error
}

type testExecutionStatusReader struct {
	status *workflow.ExecutionStatus
	err    error
}

func (r *testExecutionStatusReader) GetExecutionStatus(context.Context, string) (*workflow.ExecutionStatus, error) {
	return r.status, r.err
}

func (e *testLifecycleExecutor) Execute(_ context.Context, request *workflow.ProvisionRequest) (*workflow.ExecutionStatus, error) {
	e.request = request
	return e.status, e.err
}

func TestLifecycleHandler(t *testing.T) {
	executor := &testLifecycleExecutor{status: &workflow.ExecutionStatus{ExecutionID: "execution", State: workflow.StateSucceeded}}
	handler, err := NewLifecycleHandler(executor)
	require.NoError(t, err)

	request := &workflow.ProvisionRequest{TenantID: "acme", Operation: "provision"}
	status, err := handler.Handle(context.Background(), &LifecycleRequest{Request: request, WorkflowExecutionID: "arn:aws:states:region:account:execution:machine:acme"})
	require.NoError(t, err)
	require.Same(t, request, executor.request)
	require.Equal(t, "arn:aws:states:region:account:execution:machine:acme", executor.request.Metadata["workflow_execution_id"])
	require.Equal(t, executor.status, status)
}

func TestLifecycleHandlerRejectsInvalidDependenciesAndRequests(t *testing.T) {
	_, err := NewLifecycleHandler(nil)
	require.ErrorContains(t, err, "lifecycle executor is required")

	handler, err := NewLifecycleHandler(&testLifecycleExecutor{})
	require.NoError(t, err)
	_, err = handler.Handle(context.Background(), nil)
	require.ErrorContains(t, err, "Step Functions request is required")
}

func TestLifecycleHandlerWrapsExecutionFailure(t *testing.T) {
	handler, err := NewLifecycleHandler(&testLifecycleExecutor{err: errors.New("provider unavailable")})
	require.NoError(t, err)

	_, err = handler.Handle(context.Background(), &LifecycleRequest{Request: &workflow.ProvisionRequest{TenantID: "acme"}})
	require.ErrorContains(t, err, "execute lifecycle request")
	require.ErrorContains(t, err, "provider unavailable")
}

func TestCancellationCheckerRejectsStoppedExecution(t *testing.T) {
	checker, err := NewCancellationChecker(&testExecutionStatusReader{status: &workflow.ExecutionStatus{State: workflow.StateCancelled}})
	require.NoError(t, err)

	err = checker.BeforeMutation(context.Background(), &workflow.ProvisionRequest{Metadata: map[string]string{"workflow_execution_id": "arn:aws:states:region:account:execution:machine:acme"}})
	require.ErrorIs(t, err, context.Canceled)
}

func TestCancellationCheckerAllowsRequestWithoutExecutionARN(t *testing.T) {
	checker, err := NewCancellationChecker(&testExecutionStatusReader{})
	require.NoError(t, err)
	require.NoError(t, checker.BeforeMutation(context.Background(), &workflow.ProvisionRequest{}))
}
