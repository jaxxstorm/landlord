//go:build integration

package ec2runner

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/jaxxstorm/landlord/internal/cloud/awsconfig"
	"github.com/jaxxstorm/landlord/internal/compute"
	"github.com/jaxxstorm/landlord/internal/testutil/floci"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestEC2RunnerProviderIntegration(t *testing.T) {
	ctx := context.Background()
	fixture := floci.Start(t)
	fixture.SetEnvironment(t)

	awsCfg, err := awsconfig.Load(ctx, awsconfig.Options{Region: fixture.Region})
	require.NoError(t, err)
	client := ec2.NewFromConfig(awsCfg)
	provider, err := New(ctx, Config{
		Region:           fixture.Region,
		InstanceType:     "t3.micro",
		SubnetID:         "subnet-default-a",
		SecurityGroupIDs: []string{"sg-default"},
	}, zaptest.NewLogger(t))
	require.NoError(t, err)

	tenantID := fixture.ResourcePrefix + "-tenant"
	spec := integrationSpec(t, tenantID, "#!/bin/sh\necho initial")
	provisioned, err := provider.Provision(ctx, spec)
	require.NoError(t, err)
	instanceID := provisioned.ResourceIDs["instance"]
	require.NotEmpty(t, instanceID)

	t.Cleanup(func() {
		if err := provider.Destroy(ctx, tenantID); err != nil {
			t.Errorf("destroy Floci EC2 runner: %s", err)
		}
	})

	description, err := client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{InstanceIds: []string{instanceID}})
	require.NoError(t, err)
	require.Len(t, description.Reservations, 1)
	require.Len(t, description.Reservations[0].Instances, 1)
	require.Equal(t, compute.MetadataOwnerValue, instanceTag(description.Reservations[0].Instances[0], compute.MetadataOwnerKey))
	require.Equal(t, tenantID, instanceTag(description.Reservations[0].Instances[0], compute.MetadataTenantIDKey))

	require.Eventually(t, func() bool {
		status, err := provider.GetStatus(ctx, tenantID)
		return err == nil && status.State == compute.ComputeStateRunning
	}, 30*time.Second, time.Second)

	replacement := integrationSpec(t, tenantID, "#!/bin/sh\necho replacement")
	updated, err := provider.Update(ctx, tenantID, replacement)
	require.NoError(t, err)
	require.Equal(t, compute.UpdateStatusSuccess, updated.Status)

	var replacementID string
	require.Eventually(t, func() bool {
		status, err := provider.GetStatus(ctx, tenantID)
		if err != nil || status.State != compute.ComputeStateRunning {
			return false
		}
		replacementID = status.Metadata["instance"]
		return replacementID != "" && replacementID != instanceID
	}, 30*time.Second, time.Second)

	require.NoError(t, provider.Destroy(ctx, tenantID))
	require.Eventually(t, func() bool {
		_, err := provider.GetStatus(ctx, tenantID)
		return errors.Is(err, compute.ErrTenantNotFound)
	}, 30*time.Second, time.Second)
	require.NoError(t, provider.Destroy(ctx, tenantID))
}

func integrationSpec(t *testing.T, tenantID, userData string) *compute.TenantComputeSpec {
	t.Helper()
	config, err := json.Marshal(LaunchConfig{AMI: "ami-alpine", UserData: userData})
	require.NoError(t, err)
	return &compute.TenantComputeSpec{
		TenantID:       tenantID,
		ProviderType:   "ec2-runner",
		ProviderConfig: config,
	}
}

func instanceTag(instance ec2types.Instance, key string) string {
	for _, tag := range instance.Tags {
		if tag.Key != nil && *tag.Key == key && tag.Value != nil {
			return *tag.Value
		}
	}
	return ""
}
