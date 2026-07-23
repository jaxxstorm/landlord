package stepfunctions

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	sfnTypes "github.com/aws/aws-sdk-go-v2/service/sfn/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/jaxxstorm/landlord/internal/workflow"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

type testSFNClient struct{}

func (testSFNClient) StartExecution(context.Context, *sfn.StartExecutionInput, ...func(*sfn.Options)) (*sfn.StartExecutionOutput, error) {
	return &sfn.StartExecutionOutput{ExecutionArn: aws.String("arn:aws:states:us-east-1:123456789012:execution:landlord:test"), StartDate: aws.Time(time.Now())}, nil
}

func TestCreateWorkflowUsesConfiguredStateMachineARN(t *testing.T) {
	const stateMachineARN = "arn:aws:states:us-east-1:123456789012:stateMachine:landlord"
	provider, err := NewWithClients(Config{
		Region:          "us-east-1",
		StateMachineARN: stateMachineARN,
	}, zaptest.NewLogger(t), testSFNClient{}, testSTSClient{})
	require.NoError(t, err)

	result, err := provider.CreateWorkflow(context.Background(), &workflow.WorkflowSpec{
		WorkflowID: "tenant-provisioning",
		Definition: json.RawMessage(`{"StartAt":"Complete","States":{"Complete":{"Type":"Succeed"}}}`),
	})
	require.NoError(t, err)
	require.Equal(t, stateMachineARN, result.ResourceIDs["arn"])
}

func TestDeleteWorkflowRejectsDeploymentManagedStateMachine(t *testing.T) {
	const stateMachineARN = "arn:aws:states:us-east-1:123456789012:stateMachine:landlord"
	provider, err := NewWithClients(Config{
		Region:          "us-east-1",
		StateMachineARN: stateMachineARN,
	}, zaptest.NewLogger(t), testSFNClient{}, testSTSClient{})
	require.NoError(t, err)

	err = provider.DeleteWorkflow(context.Background(), stateMachineARN)
	require.Error(t, err)
	require.Contains(t, err.Error(), "deployment-managed")
}

func (testSFNClient) DescribeExecution(context.Context, *sfn.DescribeExecutionInput, ...func(*sfn.Options)) (*sfn.DescribeExecutionOutput, error) {
	return nil, nil
}

func (testSFNClient) GetExecutionHistory(context.Context, *sfn.GetExecutionHistoryInput, ...func(*sfn.Options)) (*sfn.GetExecutionHistoryOutput, error) {
	return nil, nil
}

func (testSFNClient) ListExecutions(context.Context, *sfn.ListExecutionsInput, ...func(*sfn.Options)) (*sfn.ListExecutionsOutput, error) {
	return &sfn.ListExecutionsOutput{}, nil
}

func (testSFNClient) StopExecution(context.Context, *sfn.StopExecutionInput, ...func(*sfn.Options)) (*sfn.StopExecutionOutput, error) {
	return nil, nil
}

type testSTSClient struct{}

func (testSTSClient) GetCallerIdentity(context.Context, *sts.GetCallerIdentityInput, ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
	return nil, nil
}

func TestNewWithClients(t *testing.T) {
	tests := []struct {
		name      string
		cfg       Config
		shouldErr bool
	}{
		{
			name: "valid injected clients",
			cfg: Config{
				Region:          "us-east-1",
				StateMachineARN: "arn:aws:states:us-east-1:123456789012:stateMachine:landlord",
			},
		},
		{
			name: "missing region",
			cfg: Config{
				StateMachineARN: "arn:aws:states:us-east-1:123456789012:stateMachine:landlord",
			},
			shouldErr: true,
		},
		{
			name: "missing state machine ARN",
			cfg: Config{
				Region: "us-east-1",
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewWithClients(tt.cfg, zaptest.NewLogger(t), testSFNClient{}, testSTSClient{})
			if tt.shouldErr {
				require.Error(t, err)
				require.Nil(t, provider)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, provider)
		})
	}
}

type startSFNClient struct {
	testSFNClient
	startInput *sfn.StartExecutionInput
	startErr   error
	listOutput *sfn.ListExecutionsOutput
	describe   *sfn.DescribeExecutionOutput
	stopInput  *sfn.StopExecutionInput
	stopErr    error
}

type statusSFNClient struct {
	testSFNClient
	describe     *sfn.DescribeExecutionOutput
	describeErr  error
	historyPages []*sfn.GetExecutionHistoryOutput
	historyCalls int
}

func (c *statusSFNClient) DescribeExecution(context.Context, *sfn.DescribeExecutionInput, ...func(*sfn.Options)) (*sfn.DescribeExecutionOutput, error) {
	return c.describe, c.describeErr
}

func (c *statusSFNClient) GetExecutionHistory(context.Context, *sfn.GetExecutionHistoryInput, ...func(*sfn.Options)) (*sfn.GetExecutionHistoryOutput, error) {
	result := c.historyPages[c.historyCalls]
	c.historyCalls++
	return result, nil
}

func (c *startSFNClient) StartExecution(_ context.Context, input *sfn.StartExecutionInput, _ ...func(*sfn.Options)) (*sfn.StartExecutionOutput, error) {
	c.startInput = input
	if c.startErr != nil {
		return nil, c.startErr
	}
	return &sfn.StartExecutionOutput{ExecutionArn: aws.String("arn:aws:states:us-east-1:123456789012:execution:landlord:new"), StartDate: aws.Time(time.Now())}, nil
}

func (c *startSFNClient) ListExecutions(context.Context, *sfn.ListExecutionsInput, ...func(*sfn.Options)) (*sfn.ListExecutionsOutput, error) {
	if c.listOutput == nil {
		return &sfn.ListExecutionsOutput{}, nil
	}
	return c.listOutput, nil
}

func (c *startSFNClient) DescribeExecution(context.Context, *sfn.DescribeExecutionInput, ...func(*sfn.Options)) (*sfn.DescribeExecutionOutput, error) {
	if c.describe == nil {
		return &sfn.DescribeExecutionOutput{Input: aws.String(`{}`)}, nil
	}
	return c.describe, nil
}

func (c *startSFNClient) StopExecution(_ context.Context, input *sfn.StopExecutionInput, _ ...func(*sfn.Options)) (*sfn.StopExecutionOutput, error) {
	c.stopInput = input
	if c.stopErr != nil {
		return nil, c.stopErr
	}
	return &sfn.StopExecutionOutput{}, nil
}

func TestStartExecutionUsesConfiguredStateMachine(t *testing.T) {
	client := &startSFNClient{}
	provider, err := NewWithClients(Config{Region: "us-east-1", StateMachineARN: "arn:aws:states:us-east-1:123456789012:stateMachine:landlord"}, zaptest.NewLogger(t), client, testSTSClient{})
	require.NoError(t, err)

	result, err := provider.StartExecution(context.Background(), "tenant-provisioning", &workflow.ExecutionInput{ExecutionName: "tenant-1-provision-abc", Input: json.RawMessage(`{"tenant_id":"tenant-1"}`)})
	require.NoError(t, err)
	require.Equal(t, "arn:aws:states:us-east-1:123456789012:execution:landlord:new", result.ExecutionID)
	require.Equal(t, "arn:aws:states:us-east-1:123456789012:stateMachine:landlord", aws.ToString(client.startInput.StateMachineArn))
	require.Equal(t, "tenant-1-provision-abc", aws.ToString(client.startInput.Name))
}

func TestStartExecutionReturnsExistingExecution(t *testing.T) {
	client := &startSFNClient{
		startErr: &sfnTypes.ExecutionAlreadyExists{},
		listOutput: &sfn.ListExecutionsOutput{Executions: []sfnTypes.ExecutionListItem{{
			Name:         aws.String("tenant-1-provision-abc"),
			ExecutionArn: aws.String("arn:aws:states:us-east-1:123456789012:execution:landlord:existing"),
			StartDate:    aws.Time(time.Now()),
			Status:       sfnTypes.ExecutionStatusRunning,
		}}},
		describe: &sfn.DescribeExecutionOutput{Input: aws.String(`{}`)},
	}
	provider, err := NewWithClients(Config{Region: "us-east-1", StateMachineARN: "arn:aws:states:us-east-1:123456789012:stateMachine:landlord"}, zaptest.NewLogger(t), client, testSTSClient{})
	require.NoError(t, err)

	result, err := provider.StartExecution(context.Background(), "tenant-provisioning", &workflow.ExecutionInput{ExecutionName: "tenant-1-provision-abc", Input: json.RawMessage(`{}`)})
	require.NoError(t, err)
	require.Equal(t, "arn:aws:states:us-east-1:123456789012:execution:landlord:existing", result.ExecutionID)
	require.Equal(t, "already started", result.Message)
}

func TestStartExecutionRejectsDuplicateWithDifferentInput(t *testing.T) {
	client := &startSFNClient{
		startErr: &sfnTypes.ExecutionAlreadyExists{},
		listOutput: &sfn.ListExecutionsOutput{Executions: []sfnTypes.ExecutionListItem{{
			Name:         aws.String("tenant-1-provision-abc"),
			ExecutionArn: aws.String("arn:aws:states:us-east-1:123456789012:execution:landlord:existing"),
			StartDate:    aws.Time(time.Now()),
			Status:       sfnTypes.ExecutionStatusRunning,
		}}},
		describe: &sfn.DescribeExecutionOutput{Input: aws.String(`{"tenant_id":"other"}`)},
	}
	provider, err := NewWithClients(Config{Region: "us-east-1", StateMachineARN: "arn:aws:states:us-east-1:123456789012:stateMachine:landlord"}, zaptest.NewLogger(t), client, testSTSClient{})
	require.NoError(t, err)

	_, err = provider.StartExecution(context.Background(), "tenant-provisioning", &workflow.ExecutionInput{ExecutionName: "tenant-1-provision-abc", Input: json.RawMessage(`{}`)})
	require.Error(t, err)
	require.Contains(t, err.Error(), "different input")
}

func TestStartExecutionPreservesContextCancellation(t *testing.T) {
	client := &startSFNClient{startErr: context.Canceled}
	provider, err := NewWithClients(Config{Region: "us-east-1", StateMachineARN: "arn:aws:states:us-east-1:123456789012:stateMachine:landlord"}, zaptest.NewLogger(t), client, testSTSClient{})
	require.NoError(t, err)

	_, err = provider.StartExecution(context.Background(), "tenant-provisioning", &workflow.ExecutionInput{ExecutionName: "tenant-1-provision-abc", Input: json.RawMessage(`{}`)})
	require.Error(t, err)
	require.True(t, errors.Is(err, context.Canceled))
}

func TestStopExecutionUsesReasonAndIsIdempotent(t *testing.T) {
	client := &startSFNClient{}
	provider, err := NewWithClients(Config{Region: "us-east-1", StateMachineARN: "arn:aws:states:us-east-1:123456789012:stateMachine:landlord"}, zaptest.NewLogger(t), client, testSTSClient{})
	require.NoError(t, err)

	err = provider.StopExecution(context.Background(), "arn:aws:states:us-east-1:123456789012:execution:landlord:tenant", "tenant deleted")
	require.NoError(t, err)
	require.Equal(t, "LandlordExecutionStopped", aws.ToString(client.stopInput.Error))
	require.Equal(t, "tenant deleted", aws.ToString(client.stopInput.Cause))

	client.stopErr = &sfnTypes.ExecutionDoesNotExist{}
	err = provider.StopExecution(context.Background(), "arn:aws:states:us-east-1:123456789012:execution:landlord:tenant", "tenant deleted")
	require.NoError(t, err)
}

func TestGetExecutionStatusReturnsOutputAndHistory(t *testing.T) {
	startedAt := time.Now().Add(-time.Minute).UTC()
	stoppedAt := time.Now().UTC()
	client := &statusSFNClient{
		describe: &sfn.DescribeExecutionOutput{
			StateMachineArn: aws.String("arn:aws:states:us-east-1:123456789012:stateMachine:landlord"),
			Status:          sfnTypes.ExecutionStatusSucceeded,
			StartDate:       &startedAt,
			StopDate:        &stoppedAt,
			Input:           aws.String(`{"tenant_id":"tenant-1"}`),
			Output:          aws.String(`{"status":"ready"}`),
		},
		historyPages: []*sfn.GetExecutionHistoryOutput{
			{Events: []sfnTypes.HistoryEvent{{Timestamp: &startedAt, Type: sfnTypes.HistoryEventType("ExecutionStarted")}}, NextToken: aws.String("next")},
			{Events: []sfnTypes.HistoryEvent{{Timestamp: &stoppedAt, Type: sfnTypes.HistoryEventType("ExecutionSucceeded")}}},
		},
	}
	provider, err := NewWithClients(Config{Region: "us-east-1", StateMachineARN: "arn:aws:states:us-east-1:123456789012:stateMachine:landlord"}, zaptest.NewLogger(t), client, testSTSClient{})
	require.NoError(t, err)

	status, err := provider.GetExecutionStatus(context.Background(), "arn:aws:states:us-east-1:123456789012:execution:landlord:tenant")
	require.NoError(t, err)
	require.Equal(t, workflow.StateSucceeded, status.State)
	require.JSONEq(t, `{"status":"ready"}`, string(status.Output))
	require.Len(t, status.History, 2)
	require.Equal(t, 2, client.historyCalls)
}

func TestGetExecutionStatusMapsTimedOutFailure(t *testing.T) {
	stoppedAt := time.Now().UTC()
	client := &statusSFNClient{
		describe: &sfn.DescribeExecutionOutput{
			Status:   sfnTypes.ExecutionStatusTimedOut,
			StopDate: &stoppedAt,
			Error:    aws.String("States.Timeout"),
			Cause:    aws.String("execution exceeded timeout"),
		},
		historyPages: []*sfn.GetExecutionHistoryOutput{{}},
	}
	provider, err := NewWithClients(Config{Region: "us-east-1", StateMachineARN: "arn:aws:states:us-east-1:123456789012:stateMachine:landlord"}, zaptest.NewLogger(t), client, testSTSClient{})
	require.NoError(t, err)

	status, err := provider.GetExecutionStatus(context.Background(), "arn:aws:states:us-east-1:123456789012:execution:landlord:tenant")
	require.NoError(t, err)
	require.Equal(t, workflow.StateTimedOut, status.State)
	require.NotNil(t, status.Error)
	require.Equal(t, "States.Timeout", status.Error.Code)
	require.Equal(t, "execution exceeded timeout", status.Error.Message)
}

func TestGetExecutionStatusMapsMissingExecution(t *testing.T) {
	client := &statusSFNClient{describeErr: &sfnTypes.ExecutionDoesNotExist{}}
	provider, err := NewWithClients(Config{Region: "us-east-1", StateMachineARN: "arn:aws:states:us-east-1:123456789012:stateMachine:landlord"}, zaptest.NewLogger(t), client, testSTSClient{})
	require.NoError(t, err)

	_, err = provider.GetExecutionStatus(context.Background(), "arn:aws:states:us-east-1:123456789012:execution:landlord:missing")
	require.Error(t, err)
	require.True(t, errors.Is(err, workflow.ErrExecutionNotFound))
}
