package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComputeConfigValidate_SingleProvider(t *testing.T) {
	cfg := ComputeConfig{
		Docker: &DockerProviderConfig{
			Defaults: map[string]interface{}{
				"image": "nginx:latest",
			},
		},
	}

	require.NoError(t, cfg.Validate())
}

func TestComputeConfigValidate_MultipleProviders(t *testing.T) {
	cfg := ComputeConfig{
		Docker: &DockerProviderConfig{
			Defaults: map[string]interface{}{
				"image": "nginx:latest",
			},
		},
		ECS: &ECSProviderConfig{
			Defaults: map[string]interface{}{
				"cluster_arn":         "arn:aws:ecs:us-east-1:123456789012:cluster/test",
				"task_definition_arn": "arn:aws:ecs:us-east-1:123456789012:task-definition/test:1",
				"service_name_prefix": "landlord-tenant-",
			},
		},
	}

	require.NoError(t, cfg.Validate())
}

func TestComputeConfigValidate_UnknownProvider(t *testing.T) {
	cfg := ComputeConfig{
		Unknown: map[string]interface{}{
			"kubernetes": map[string]interface{}{},
		},
	}

	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown compute provider")
}

func TestComputeConfigValidate_LegacyKeys(t *testing.T) {
	cfg := ComputeConfig{
		Unknown: map[string]interface{}{
			"default_provider": "mock",
			"defaults":         map[string]interface{}{},
		},
	}

	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "legacy compute config keys")
}

func TestComputeConfigValidate_DockerDefaultsRequired(t *testing.T) {
	cfg := ComputeConfig{
		Docker: &DockerProviderConfig{
			Defaults: map[string]interface{}{},
		},
	}

	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "compute.docker")
}

func TestComputeConfigValidate_ECSDefaultsRequired(t *testing.T) {
	cfg := ComputeConfig{
		ECS: &ECSProviderConfig{
			Defaults: map[string]interface{}{},
		},
	}

	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "compute.ecs")
}

func TestComputeConfigValidate_EC2RunnerDefaults(t *testing.T) {
	cfg := ComputeConfig{EC2Runner: &EC2RunnerProviderConfig{
		Region:           "us-west-2",
		InstanceType:     "t3.micro",
		SubnetID:         "subnet-123",
		SecurityGroupIDs: []string{"sg-123"},
	}}

	require.NoError(t, cfg.Validate())
	require.Equal(t, []string{"ec2-runner"}, cfg.EnabledProviders())
	require.Equal(t, "ec2-runner", cfg.DefaultProvider())
}

func TestComputeConfigValidate_RejectsIncompleteEC2RunnerDefaults(t *testing.T) {
	cfg := ComputeConfig{EC2Runner: &EC2RunnerProviderConfig{
		Region:       "us-west-2",
		InstanceType: "t3.micro",
	}}

	err := cfg.Validate()
	require.ErrorContains(t, err, "compute.ec2_runner.subnet_id is required")
}

func TestComputeConfigValidate_Kubernetes(t *testing.T) {
	cfg := ComputeConfig{Kubernetes: &KubernetesProviderConfig{
		Kubeconfig:      "/etc/landlord/kubeconfig",
		NamespacePrefix: "landlord-tenant-",
		Defaults:        map[string]interface{}{"image_pull_policy": "IfNotPresent"},
	}}

	require.NoError(t, cfg.Validate())
	require.Equal(t, []string{"kubernetes"}, cfg.EnabledProviders())
	require.Equal(t, "kubernetes", cfg.DefaultProvider())
}

func TestComputeConfigValidate_RejectsKubernetesAPIServerWithoutToken(t *testing.T) {
	cfg := ComputeConfig{Kubernetes: &KubernetesProviderConfig{APIServer: "https://cluster.example"}}

	err := cfg.Validate()
	require.ErrorContains(t, err, "compute.kubernetes.bearer_token is required")
}

func TestComputeConfigValidate_RejectsInvalidKubernetesAPIServer(t *testing.T) {
	cfg := ComputeConfig{Kubernetes: &KubernetesProviderConfig{APIServer: "cluster.example", BearerToken: "token"}}

	err := cfg.Validate()
	require.ErrorContains(t, err, "compute.kubernetes.api_server must be a valid HTTP(S) URL")
}
