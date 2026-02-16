package temporal

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/jaxxstorm/landlord/internal/config"
	"github.com/jaxxstorm/landlord/internal/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	workflowservice "go.temporal.io/api/workflowservice/v1"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	workflowpb "go.temporal.io/api/workflow/v1"
	temporalclient "go.temporal.io/sdk/client"
	"go.uber.org/zap/zaptest"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type fakeWorkflowRun struct {
	id    string
	runID string
}

func (f *fakeWorkflowRun) GetID() string {
	return f.id
}

func (f *fakeWorkflowRun) GetRunID() string {
	return f.runID
}

func (f *fakeWorkflowRun) Get(ctx context.Context, valuePtr interface{}) error {
	_ = ctx
	_ = valuePtr
	return nil
}

func (f *fakeWorkflowRun) GetWithOptions(ctx context.Context, valuePtr interface{}, options temporalclient.WorkflowRunGetOptions) error {
	_ = ctx
	_ = valuePtr
	_ = options
	return nil
}

type fakeTemporalClient struct {
	executeCalls   int
	cancelCalls    int
	terminateCalls int
	lastArgs       []interface{}

	startErr     error
	cancelErr    error
	terminateErr error
	describeErr  error

	statusByWorkflow map[string]enumspb.WorkflowExecutionStatus
	runByWorkflow    map[string]string
}

func newFakeTemporalClient() *fakeTemporalClient {
	return &fakeTemporalClient{
		statusByWorkflow: make(map[string]enumspb.WorkflowExecutionStatus),
		runByWorkflow:    make(map[string]string),
	}
}

func (f *fakeTemporalClient) ExecuteWorkflow(ctx context.Context, options temporalclient.StartWorkflowOptions, wf interface{}, args ...interface{}) (temporalclient.WorkflowRun, error) {
	_ = ctx
	_ = wf
	f.lastArgs = append([]interface{}(nil), args...)
	if f.startErr != nil {
		return nil, f.startErr
	}
	f.executeCalls++
	runID := "run-" + options.ID
	f.runByWorkflow[options.ID] = runID
	if _, exists := f.statusByWorkflow[options.ID]; !exists {
		f.statusByWorkflow[options.ID] = enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING
	}
	return &fakeWorkflowRun{id: options.ID, runID: runID}, nil
}

func (f *fakeTemporalClient) GetWorkflow(ctx context.Context, workflowID string, runID string) temporalclient.WorkflowRun {
	_ = ctx
	if runID == "" {
		runID = f.runByWorkflow[workflowID]
	}
	if runID == "" {
		runID = "run-" + workflowID
	}
	return &fakeWorkflowRun{id: workflowID, runID: runID}
}

func (f *fakeTemporalClient) CancelWorkflow(ctx context.Context, workflowID string, runID string) error {
	_ = ctx
	_ = runID
	f.cancelCalls++
	if f.cancelErr != nil {
		return f.cancelErr
	}
	f.statusByWorkflow[workflowID] = enumspb.WORKFLOW_EXECUTION_STATUS_CANCELED
	return nil
}

func (f *fakeTemporalClient) TerminateWorkflow(ctx context.Context, workflowID string, runID string, reason string, details ...interface{}) error {
	_ = ctx
	_ = runID
	_ = reason
	_ = details
	f.terminateCalls++
	if f.terminateErr != nil {
		return f.terminateErr
	}
	f.statusByWorkflow[workflowID] = enumspb.WORKFLOW_EXECUTION_STATUS_TERMINATED
	return nil
}

func (f *fakeTemporalClient) DescribeWorkflowExecution(ctx context.Context, workflowID string, runID string) (*workflowservice.DescribeWorkflowExecutionResponse, error) {
	_ = ctx
	_ = runID
	if f.describeErr != nil {
		return nil, f.describeErr
	}
	status := f.statusByWorkflow[workflowID]
	if status == enumspb.WORKFLOW_EXECUTION_STATUS_UNSPECIFIED {
		status = enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING
	}
	start := timestamppb.New(time.Now().Add(-2 * time.Minute))
	resp := &workflowservice.DescribeWorkflowExecutionResponse{
		WorkflowExecutionInfo: &workflowpb.WorkflowExecutionInfo{
			Status:    status,
			StartTime: start,
		},
	}
	if status == enumspb.WORKFLOW_EXECUTION_STATUS_COMPLETED ||
		status == enumspb.WORKFLOW_EXECUTION_STATUS_FAILED ||
		status == enumspb.WORKFLOW_EXECUTION_STATUS_CANCELED ||
		status == enumspb.WORKFLOW_EXECUTION_STATUS_TIMED_OUT ||
		status == enumspb.WORKFLOW_EXECUTION_STATUS_TERMINATED {
		resp.WorkflowExecutionInfo.CloseTime = timestamppb.New(time.Now())
	}
	return resp, nil
}

func (f *fakeTemporalClient) Close() {}

func temporalTestConfig() config.TemporalConfig {
	return config.TemporalConfig{
		HostPort:      "localhost:7233",
		Namespace:     "default",
		TaskQueue:     "landlord",
		Timeout:       30 * time.Minute,
		RetryAttempts: 3,
	}
}

func TestProviderName(t *testing.T) {
	provider, err := New(temporalTestConfig(), zaptest.NewLogger(t), WithClient(newFakeTemporalClient()))
	require.NoError(t, err)
	assert.Equal(t, "temporal", provider.Name())
}

func TestProviderRegistersDefaultWorkflowOnStartup(t *testing.T) {
	fakeClient := newFakeTemporalClient()
	provider, err := New(temporalTestConfig(), zaptest.NewLogger(t), WithClient(fakeClient))
	require.NoError(t, err)

	result, err := provider.StartExecution(context.Background(), tenantProvisioningWorkflowID, &workflow.ExecutionInput{
		ExecutionName: "exec-default-reg",
		Input:         json.RawMessage(`{"tenant_id":"acme","operation":"provision"}`),
	})
	require.NoError(t, err)
	assert.Equal(t, "exec-default-reg", result.ExecutionID)
	assert.Equal(t, 1, fakeClient.executeCalls)
}

func TestCreateWorkflowIdempotent(t *testing.T) {
	provider, err := New(temporalTestConfig(), zaptest.NewLogger(t), WithClient(newFakeTemporalClient()))
	require.NoError(t, err)

	spec := &workflow.WorkflowSpec{WorkflowID: "tenant-provisioning", Definition: json.RawMessage(`{"workflow":"temporal"}`)}
	result1, err := provider.CreateWorkflow(context.Background(), spec)
	require.NoError(t, err)
	result2, err := provider.CreateWorkflow(context.Background(), spec)
	require.NoError(t, err)

	assert.Equal(t, result1.WorkflowID, result2.WorkflowID)
	assert.Equal(t, result1.CreatedAt, result2.CreatedAt)
}

func TestStartExecutionIdempotentWithDeterministicName(t *testing.T) {
	fakeClient := newFakeTemporalClient()
	provider, err := New(temporalTestConfig(), zaptest.NewLogger(t), WithClient(fakeClient))
	require.NoError(t, err)

	_, err = provider.CreateWorkflow(context.Background(), &workflow.WorkflowSpec{WorkflowID: "tenant-provisioning", Definition: json.RawMessage(`{"ok":true}`)})
	require.NoError(t, err)

	input := &workflow.ExecutionInput{Input: json.RawMessage(`{"tenant_id":"acme"}`), Metadata: map[string]string{"config_hash": "abc123"}}
	first, err := provider.StartExecution(context.Background(), "tenant-provisioning", input)
	require.NoError(t, err)
	second, err := provider.StartExecution(context.Background(), "tenant-provisioning", input)
	require.NoError(t, err)

	assert.Equal(t, first.ExecutionID, second.ExecutionID)
	assert.Equal(t, 1, fakeClient.executeCalls)
}

func TestStartExecutionTransientFailure(t *testing.T) {
	fakeClient := newFakeTemporalClient()
	fakeClient.startErr = serviceerror.NewUnavailable("temporary backend issue")

	provider, err := New(temporalTestConfig(), zaptest.NewLogger(t), WithClient(fakeClient))
	require.NoError(t, err)
	_, err = provider.CreateWorkflow(context.Background(), &workflow.WorkflowSpec{WorkflowID: "tenant-provisioning", Definition: json.RawMessage(`{"ok":true}`)})
	require.NoError(t, err)

	_, err = provider.StartExecution(context.Background(), "tenant-provisioning", &workflow.ExecutionInput{ExecutionName: "exec-1", Input: json.RawMessage(`{"tenant_id":"acme"}`)})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "start temporal execution")
}

func TestStartExecutionPassesTypedProvisionRequest(t *testing.T) {
	fakeClient := newFakeTemporalClient()
	provider, err := New(temporalTestConfig(), zaptest.NewLogger(t), WithClient(fakeClient))
	require.NoError(t, err)

	input := &workflow.ExecutionInput{
		ExecutionName: "exec-typed-arg",
		Input: json.RawMessage(`{
			"tenant_id":"acme",
			"tenant_uuid":"acme-uuid",
			"operation":"provision",
			"compute_provider":"docker"
		}`),
	}

	_, err = provider.StartExecution(context.Background(), tenantProvisioningWorkflowID, input)
	require.NoError(t, err)
	require.Len(t, fakeClient.lastArgs, 1)

	arg, ok := fakeClient.lastArgs[0].(workflow.ProvisionRequest)
	require.True(t, ok)
	assert.Equal(t, "acme", arg.TenantID)
	assert.Equal(t, "acme-uuid", arg.TenantUUID)
	assert.Equal(t, "provision", arg.Operation)
	assert.Equal(t, "docker", arg.ComputeProvider)
}

func TestGetExecutionStatusMapping(t *testing.T) {
	fakeClient := newFakeTemporalClient()
	provider, err := New(temporalTestConfig(), zaptest.NewLogger(t), WithClient(fakeClient))
	require.NoError(t, err)

	_, err = provider.CreateWorkflow(context.Background(), &workflow.WorkflowSpec{WorkflowID: "tenant-provisioning", Definition: json.RawMessage(`{"ok":true}`)})
	require.NoError(t, err)

	res, err := provider.StartExecution(context.Background(), "tenant-provisioning", &workflow.ExecutionInput{ExecutionName: "exec-status", Input: json.RawMessage(`{"tenant_id":"acme"}`)})
	require.NoError(t, err)

	fakeClient.statusByWorkflow[res.ExecutionID] = enumspb.WORKFLOW_EXECUTION_STATUS_COMPLETED
	status, err := provider.GetExecutionStatus(context.Background(), res.ExecutionID)
	require.NoError(t, err)
	assert.Equal(t, workflow.StateSucceeded, status.State)
	assert.Equal(t, "succeeded", status.Metadata["workflow_sub_state"])
	assert.Equal(t, "Completed", status.Metadata["temporal_status"])
}

func TestStopExecutionIdempotent(t *testing.T) {
	fakeClient := newFakeTemporalClient()
	fakeClient.cancelErr = serviceerror.NewCancellationAlreadyRequested("already cancelled")
	provider, err := New(temporalTestConfig(), zaptest.NewLogger(t), WithClient(fakeClient))
	require.NoError(t, err)

	_, err = provider.CreateWorkflow(context.Background(), &workflow.WorkflowSpec{WorkflowID: "tenant-provisioning", Definition: json.RawMessage(`{"ok":true}`)})
	require.NoError(t, err)

	_, err = provider.StartExecution(context.Background(), "tenant-provisioning", &workflow.ExecutionInput{ExecutionName: "exec-stop", Input: json.RawMessage(`{"tenant_id":"acme"}`)})
	require.NoError(t, err)

	require.NoError(t, provider.StopExecution(context.Background(), "exec-stop", "cancel one"))
	require.NoError(t, provider.StopExecution(context.Background(), "exec-stop", "cancel two"))
	assert.Equal(t, 2, fakeClient.cancelCalls)
}

func TestDeleteWorkflowIdempotent(t *testing.T) {
	provider, err := New(temporalTestConfig(), zaptest.NewLogger(t), WithClient(newFakeTemporalClient()))
	require.NoError(t, err)

	require.NoError(t, provider.DeleteWorkflow(context.Background(), "tenant-provisioning"))
	require.NoError(t, provider.DeleteWorkflow(context.Background(), "tenant-provisioning"))
}

func TestPostComputeCallbackRequiresExecution(t *testing.T) {
	provider, err := New(temporalTestConfig(), zaptest.NewLogger(t), WithClient(newFakeTemporalClient()))
	require.NoError(t, err)

	err = provider.PostComputeCallback(context.Background(), "missing", nil, nil)
	require.Error(t, err)
	assert.False(t, errors.Is(err, workflow.ErrExecutionNotFound))
}
