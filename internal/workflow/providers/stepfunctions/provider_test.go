package stepfunctions

import (
	"context"
	"testing"

	"github.com/jaxxstorm/landlord/internal/workflow"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestInvokeAcceptsProvisioningPayload(t *testing.T) {
	logger := zaptest.NewLogger(t)
	provider, err := New(context.Background(), Config{
		Region:  "us-east-1",
		RoleARN: "arn:aws:iam::123456789012:role/test",
	}, logger)
	require.NoError(t, err)

	request := &workflow.ProvisionRequest{
		TenantID:      "tenant-1",
		Operation:     "plan",
		DesiredConfig: map[string]interface{}{"image": "nginx:latest", "replicas": "2"},
	}

	result, err := provider.Invoke(context.Background(), "tenant-provisioning", request)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, workflow.StateRunning, result.State)

	request.Operation = "delete"
	result, err = provider.Invoke(context.Background(), "tenant-provisioning", request)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, workflow.StateRunning, result.State)
}
