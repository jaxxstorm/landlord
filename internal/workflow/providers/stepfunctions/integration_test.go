//go:build integration

package stepfunctions

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/jaxxstorm/landlord/internal/workflow"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestStepFunctionsProviderIntegration(t *testing.T) {
	region := os.Getenv("LANDLORD_SFN_INTEGRATION_REGION")
	stateMachineARN := os.Getenv("LANDLORD_SFN_INTEGRATION_STATE_MACHINE_ARN")
	if region == "" || stateMachineARN == "" {
		t.Skip("set LANDLORD_SFN_INTEGRATION_REGION and LANDLORD_SFN_INTEGRATION_STATE_MACHINE_ARN; configure AWS credentials or AWS_ENDPOINT_URL for LocalStack")
	}

	provider, err := New(context.Background(), Config{Region: region, StateMachineARN: stateMachineARN}, zaptest.NewLogger(t))
	require.NoError(t, err)

	input, err := json.Marshal(&workflow.ProvisionRequest{
		TenantUUID:      "integration-tenant",
		Operation:       "plan",
		ComputeProvider: "mock",
	})
	require.NoError(t, err)
	executionName := "integration-" + time.Now().UTC().Format("20060102150405")
	execution, err := provider.StartExecution(context.Background(), "lifecycle", &workflow.ExecutionInput{ExecutionName: executionName, Input: input})
	require.NoError(t, err)

	duplicate, err := provider.StartExecution(context.Background(), "lifecycle", &workflow.ExecutionInput{ExecutionName: executionName, Input: input})
	require.NoError(t, err)
	require.Equal(t, execution.ExecutionID, duplicate.ExecutionID)

	_, err = provider.GetExecutionStatus(context.Background(), execution.ExecutionID)
	require.NoError(t, err)
	require.NoError(t, provider.StopExecution(context.Background(), execution.ExecutionID, "integration cleanup"))

	require.Eventually(t, func() bool {
		status, err := provider.GetExecutionStatus(context.Background(), execution.ExecutionID)
		return err == nil && status.State == workflow.StateCancelled
	}, 30*time.Second, time.Second)
}
