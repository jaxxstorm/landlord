package lifecycle

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/jaxxstorm/landlord/internal/compute"
	"github.com/jaxxstorm/landlord/internal/compute/providers/ec2runner"
	"github.com/jaxxstorm/landlord/internal/workflow"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

type lifecycleEC2Client struct {
	runInput *ec2.RunInstancesInput
}

func (c *lifecycleEC2Client) DescribeInstances(context.Context, *ec2.DescribeInstancesInput, ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return &ec2.DescribeInstancesOutput{}, nil
}

func (c *lifecycleEC2Client) RunInstances(_ context.Context, input *ec2.RunInstancesInput, _ ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error) {
	c.runInput = input
	id := "i-workflow"
	return &ec2.RunInstancesOutput{Instances: []ec2types.Instance{{InstanceId: &id}}}, nil
}

func (c *lifecycleEC2Client) TerminateInstances(context.Context, *ec2.TerminateInstancesInput, ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error) {
	return &ec2.TerminateInstancesOutput{}, nil
}

func TestExecutorProvisionsEC2RunnerFromWorkflowPayload(t *testing.T) {
	logger := zaptest.NewLogger(t)
	client := &lifecycleEC2Client{}
	provider, err := ec2runner.NewWithClient(ec2runner.Config{
		Region: "us-west-2", InstanceType: "t3.micro", SubnetID: "subnet-123", SecurityGroupIDs: []string{"sg-123"},
	}, logger, client)
	require.NoError(t, err)
	registry := compute.NewRegistry(logger)
	require.NoError(t, registry.Register(provider))
	executor, err := NewExecutor(registry, nil, "step-functions", logger)
	require.NoError(t, err)

	status, err := executor.Execute(context.Background(), &workflow.ProvisionRequest{
		TenantUUID: "tenant-ec2", Operation: "provision", ComputeProvider: "ec2-runner",
		DesiredConfig: map[string]interface{}{"ami": "ami-123", "user_data": "#!/bin/bash"},
	})
	require.NoError(t, err)
	require.Equal(t, workflow.StateSucceeded, status.State)
	require.NotNil(t, client.runInput)
	require.Equal(t, "ami-123", *client.runInput.ImageId)
	var output map[string]interface{}
	require.NoError(t, json.Unmarshal(status.Output, &output))
	require.Equal(t, "tenant-ec2", output["tenant_id"])
	require.Equal(t, "i-workflow", output["resource_ids"].(map[string]interface{})["instance"])
}
