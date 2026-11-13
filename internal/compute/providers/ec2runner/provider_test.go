package ec2runner

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/jaxxstorm/landlord/internal/compute"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

type mockEC2Client struct {
	describe  func(context.Context, *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
	run       func(context.Context, *ec2.RunInstancesInput) (*ec2.RunInstancesOutput, error)
	terminate func(context.Context, *ec2.TerminateInstancesInput) (*ec2.TerminateInstancesOutput, error)
}

func (m *mockEC2Client) DescribeInstances(ctx context.Context, input *ec2.DescribeInstancesInput, _ ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	if m.describe == nil {
		return &ec2.DescribeInstancesOutput{}, nil
	}
	return m.describe(ctx, input)
}

func (m *mockEC2Client) RunInstances(ctx context.Context, input *ec2.RunInstancesInput, _ ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error) {
	if m.run == nil {
		return &ec2.RunInstancesOutput{}, nil
	}
	return m.run(ctx, input)
}

func (m *mockEC2Client) TerminateInstances(ctx context.Context, input *ec2.TerminateInstancesInput, _ ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error) {
	if m.terminate == nil {
		return &ec2.TerminateInstancesOutput{}, nil
	}
	return m.terminate(ctx, input)
}

func testProvider(t *testing.T, client EC2Client) *Provider {
	t.Helper()
	provider, err := NewWithClient(Config{
		Region: "us-west-2", InstanceType: "t3.micro", SubnetID: "subnet-123", SecurityGroupIDs: []string{"sg-123"},
	}, zaptest.NewLogger(t), client)
	require.NoError(t, err)
	return provider
}

func testSpec() *compute.TenantComputeSpec {
	return &compute.TenantComputeSpec{TenantID: "tenant-1", ProviderType: "ec2-runner", ProviderConfig: []byte(`{"ami":"ami-123","user_data":"#!/bin/bash"}`)}
}

func runnerInstance(id string, state ec2types.InstanceStateName, fingerprint string) ec2types.Instance {
	privateIP := "10.0.0.5"
	return ec2types.Instance{InstanceId: &id, PrivateIpAddress: &privateIP, State: &ec2types.InstanceState{Name: state}, Tags: []ec2types.Tag{{Key: stringPointer(launchFingerprintTag), Value: stringPointer(fingerprint)}}}
}

func TestProvisionLaunchesTaggedRunner(t *testing.T) {
	var runInput *ec2.RunInstancesInput
	client := &mockEC2Client{run: func(_ context.Context, input *ec2.RunInstancesInput) (*ec2.RunInstancesOutput, error) {
		runInput = input
		return &ec2.RunInstancesOutput{Instances: []ec2types.Instance{runnerInstance("i-new", ec2types.InstanceStateNamePending, "")}}, nil
	}}
	result, err := testProvider(t, client).Provision(context.Background(), testSpec())
	require.NoError(t, err)
	require.Equal(t, "i-new", result.ResourceIDs["instance"])
	require.Equal(t, "ami-123", *runInput.ImageId)
	require.Equal(t, base64.StdEncoding.EncodeToString([]byte("#!/bin/bash")), *runInput.UserData)
	require.Equal(t, int32(1), *runInput.MinCount)
	require.NotEmpty(t, runInput.TagSpecifications[0].Tags)
}

func TestProvisionReturnsExistingRunner(t *testing.T) {
	instance := runnerInstance("i-existing", ec2types.InstanceStateNameRunning, "old")
	client := &mockEC2Client{describe: func(context.Context, *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
		return &ec2.DescribeInstancesOutput{Reservations: []ec2types.Reservation{{Instances: []ec2types.Instance{instance}}}}, nil
	}, run: func(context.Context, *ec2.RunInstancesInput) (*ec2.RunInstancesOutput, error) {
		t.Fatal("RunInstances must not be called for an existing runner")
		return nil, nil
	}}
	result, err := testProvider(t, client).Provision(context.Background(), testSpec())
	require.NoError(t, err)
	require.Equal(t, "i-existing", result.ResourceIDs["instance"])
}

func TestGetStatusAndDestroy(t *testing.T) {
	instance := runnerInstance("i-existing", ec2types.InstanceStateNameRunning, "old")
	terminated := false
	client := &mockEC2Client{describe: func(context.Context, *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
		return &ec2.DescribeInstancesOutput{Reservations: []ec2types.Reservation{{Instances: []ec2types.Instance{instance}}}}, nil
	}, terminate: func(_ context.Context, input *ec2.TerminateInstancesInput) (*ec2.TerminateInstancesOutput, error) {
		terminated = true
		require.Equal(t, []string{"i-existing"}, input.InstanceIds)
		return &ec2.TerminateInstancesOutput{}, nil
	}}
	provider := testProvider(t, client)
	status, err := provider.GetStatus(context.Background(), "tenant-1")
	require.NoError(t, err)
	require.Equal(t, compute.ComputeStateRunning, status.State)
	require.Equal(t, "i-existing", status.Metadata["instance"])
	require.NoError(t, provider.Destroy(context.Background(), "tenant-1"))
	require.True(t, terminated)
}

func TestUpdateReplacesChangedLaunchInput(t *testing.T) {
	old := runnerInstance("i-old", ec2types.InstanceStateNameRunning, "different")
	terminated := false
	client := &mockEC2Client{describe: func(context.Context, *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
		return &ec2.DescribeInstancesOutput{Reservations: []ec2types.Reservation{{Instances: []ec2types.Instance{old}}}}, nil
	}, run: func(context.Context, *ec2.RunInstancesInput) (*ec2.RunInstancesOutput, error) {
		return &ec2.RunInstancesOutput{Instances: []ec2types.Instance{runnerInstance("i-new", ec2types.InstanceStateNamePending, "")}}, nil
	}, terminate: func(context.Context, *ec2.TerminateInstancesInput) (*ec2.TerminateInstancesOutput, error) {
		terminated = true
		return &ec2.TerminateInstancesOutput{}, nil
	}}
	result, err := testProvider(t, client).Update(context.Background(), "tenant-1", testSpec())
	require.NoError(t, err)
	require.Equal(t, compute.UpdateStatusSuccess, result.Status)
	require.True(t, terminated)
}

func TestProvisionFailureAndCancellation(t *testing.T) {
	provider := testProvider(t, &mockEC2Client{run: func(context.Context, *ec2.RunInstancesInput) (*ec2.RunInstancesOutput, error) {
		return nil, errors.New("unauthorized")
	}})
	_, err := provider.Provision(context.Background(), testSpec())
	require.ErrorIs(t, err, compute.ErrProvisionFailed)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = provider.Provision(ctx, testSpec())
	require.ErrorIs(t, err, context.Canceled)
}
