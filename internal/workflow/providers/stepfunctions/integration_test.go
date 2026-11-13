//go:build integration

package stepfunctions

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/jaxxstorm/landlord/internal/cloud/awsconfig"
	"github.com/jaxxstorm/landlord/internal/testutil/floci"
	"github.com/jaxxstorm/landlord/internal/workflow"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestStepFunctionsProviderIntegration(t *testing.T) {
	ctx := context.Background()
	fixture := floci.Start(t)
	fixture.SetEnvironment(t)

	awsCfg, err := awsconfig.Load(ctx, awsconfig.Options{Region: fixture.Region})
	require.NoError(t, err)
	client := sfn.NewFromConfig(awsCfg)
	stateMachine, err := client.CreateStateMachine(ctx, &sfn.CreateStateMachineInput{
		Name:       aws.String(fixture.ResourcePrefix + "-lifecycle"),
		Definition: aws.String(`{"StartAt":"Wait","States":{"Wait":{"Type":"Wait","Seconds":60,"End":true}}}`),
		RoleArn:    aws.String(fmt.Sprintf("arn:aws:iam::%s:role/floci-integration", fixture.AccessKeyID)),
	})
	require.NoError(t, err)
	require.NotNil(t, stateMachine.StateMachineArn)
	t.Cleanup(func() {
		_, err := client.DeleteStateMachine(ctx, &sfn.DeleteStateMachineInput{StateMachineArn: stateMachine.StateMachineArn})
		if err != nil {
			t.Errorf("delete Floci state machine: %s", err)
		}
	})

	provider, err := New(ctx, Config{Region: fixture.Region, StateMachineARN: *stateMachine.StateMachineArn}, zaptest.NewLogger(t))
	require.NoError(t, err)
	var executionID string
	executionStopped := false
	t.Cleanup(func() {
		if executionID != "" && !executionStopped {
			if err := provider.StopExecution(ctx, executionID, "test cleanup"); err != nil {
				t.Errorf("stop Floci Step Functions execution: %s", err)
			}
		}
	})

	input, err := json.Marshal(&workflow.ProvisionRequest{
		TenantID:        "integration-tenant",
		Operation:       "plan",
		ComputeProvider: "mock",
	})
	require.NoError(t, err)
	executionName := fixture.ResourcePrefix + "-execution"
	execution, err := provider.StartExecution(ctx, "lifecycle", &workflow.ExecutionInput{ExecutionName: executionName, Input: input})
	require.NoError(t, err)
	executionID = execution.ExecutionID

	status, err := provider.GetExecutionStatus(ctx, execution.ExecutionID)
	require.NoError(t, err)
	require.Equal(t, workflow.StateRunning, status.State)
	require.NotEmpty(t, status.History)

	duplicate, err := provider.StartExecution(ctx, "lifecycle", &workflow.ExecutionInput{ExecutionName: executionName, Input: input})
	require.NoError(t, err)
	require.Equal(t, execution.ExecutionID, duplicate.ExecutionID)

	require.NoError(t, provider.StopExecution(ctx, execution.ExecutionID, "integration cleanup"))
	executionStopped = true

	require.Eventually(t, func() bool {
		status, err := provider.GetExecutionStatus(ctx, execution.ExecutionID)
		return err == nil && status.State == workflow.StateCancelled
	}, 30*time.Second, time.Second)
}
