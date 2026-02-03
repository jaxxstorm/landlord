package docker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/jaxxstorm/landlord/internal/compute"
)

// TestNewProvider tests creating a new Docker provider
func TestNewProvider(t *testing.T) {
	logger := zap.NewNop()
	defaults := map[string]interface{}{"image": "nginx:latest"}

	t.Run("creates provider successfully", func(t *testing.T) {
		provider, err := New(&Config{}, defaults, logger)
		if err != nil {
			// Skip test if Docker daemon is not available
			t.Skip("Docker daemon not available:", err)
		}
		require.NoError(t, err)
		require.NotNil(t, provider)
		assert.Equal(t, "docker", provider.Name())
		provider.Close()
	})

	t.Run("sets default configuration values", func(t *testing.T) {
		cfg := &Config{
			Host:          "",
			NetworkName:   "",
			NetworkDriver: "",
			LabelPrefix:   "",
		}
		provider, err := New(cfg, defaults, logger)
		if err != nil {
			t.Skip("Docker daemon not available:", err)
		}
		require.NoError(t, err)
		require.NotNil(t, provider)
		provider.Close()
	})

	t.Run("handles nil config", func(t *testing.T) {
		provider, err := New(nil, defaults, logger)
		if err != nil {
			t.Skip("Docker daemon not available:", err)
		}
		require.NoError(t, err)
		require.NotNil(t, provider)
		provider.Close()
	})
}

// TestValidate tests spec validation
func TestValidate(t *testing.T) {
	logger := zap.NewNop()
	provider, err := New(&Config{}, nil, logger)
	if err != nil {
		t.Skip("Docker daemon not available:", err)
	}
	defer provider.Close()

	t.Run("rejects empty image", func(t *testing.T) {
		spec := &compute.TenantComputeSpec{
			TenantID:     "test-tenant",
			ProviderType: "docker",
			Containers: []compute.ContainerSpec{
				{
					Name:  "app",
					Image: "",
				},
			},
		}
		err := provider.Validate(context.Background(), spec)
		assert.Error(t, err)
	})

	t.Run("rejects multiple containers", func(t *testing.T) {
		spec := &compute.TenantComputeSpec{
			TenantID:     "test-tenant",
			ProviderType: "docker",
			Containers: []compute.ContainerSpec{
				{
					Name:  "app1",
					Image: "nginx:latest",
				},
				{
					Name:  "app2",
					Image: "redis:latest",
				},
			},
		}
		err := provider.Validate(context.Background(), spec)
		assert.Error(t, err)
	})

	t.Run("rejects no containers", func(t *testing.T) {
		spec := &compute.TenantComputeSpec{
			TenantID:     "test-tenant",
			ProviderType: "docker",
			Containers:   []compute.ContainerSpec{},
		}
		err := provider.Validate(context.Background(), spec)
		assert.Error(t, err)
	})

	t.Run("accepts valid spec", func(t *testing.T) {
		spec := &compute.TenantComputeSpec{
			TenantID:     "test-tenant",
			ProviderType: "docker",
			Containers: []compute.ContainerSpec{
				{
					Name:  "app",
					Image: "alpine:latest",
				},
			},
		}
		err := provider.Validate(context.Background(), spec)
		assert.NoError(t, err)
	})
}

// TestName tests the provider name
func TestName(t *testing.T) {
	logger := zap.NewNop()
	provider, err := New(&Config{}, map[string]interface{}{"image": "nginx:latest"}, logger)
	if err != nil {
		t.Skip("Docker daemon not available:", err)
	}
	defer provider.Close()

	assert.Equal(t, "docker", provider.Name())
}

// TestProvisionValidation tests that provision validates input
func TestProvisionValidation(t *testing.T) {
	logger := zap.NewNop()
	provider, err := New(&Config{}, map[string]interface{}{"image": "nginx:latest"}, logger)
	if err != nil {
		t.Skip("Docker daemon not available:", err)
	}
	defer provider.Close()

	t.Run("rejects multiple containers", func(t *testing.T) {
		spec := &compute.TenantComputeSpec{
			TenantID:     "test-tenant",
			ProviderType: "docker",
			Containers: []compute.ContainerSpec{
				{
					Name:  "app1",
					Image: "nginx:latest",
				},
				{
					Name:  "app2",
					Image: "redis:latest",
				},
			},
		}
		_, err := provider.Provision(context.Background(), spec)
		assert.Error(t, err)
	})

	t.Run("rejects no containers", func(t *testing.T) {
		spec := &compute.TenantComputeSpec{
			TenantID:     "test-tenant",
			ProviderType: "docker",
			Containers:   []compute.ContainerSpec{},
		}
		_, err := provider.Provision(context.Background(), spec)
		assert.Error(t, err)
	})
}

// TestHelperFunctions tests utility functions
func TestHelperFunctions(t *testing.T) {
	t.Run("convertEnv", func(t *testing.T) {
		envMap := map[string]string{
			"FOO": "bar",
			"BAZ": "qux",
		}
		env := convertEnv(envMap)
		assert.Len(t, env, 2)
		// Check both values are present (order might vary)
		envStr := ""
		for _, e := range env {
			envStr += e
		}
		assert.Contains(t, envStr, "FOO=bar")
		assert.Contains(t, envStr, "BAZ=qux")
	})

	t.Run("mapsEqual", func(t *testing.T) {
		m1 := map[string]string{"a": "1", "b": "2"}
		m2 := map[string]string{"a": "1", "b": "2"}
		m3 := map[string]string{"a": "1", "b": "3"}
		m4 := map[string]string{"a": "1"}

		assert.True(t, mapsEqual(m1, m2))
		assert.False(t, mapsEqual(m1, m3))
		assert.False(t, mapsEqual(m1, m4))
	})

	t.Run("portMappingsEqual", func(t *testing.T) {
		p1 := []compute.PortMapping{
			{ContainerPort: 80, HostPort: 8080, Protocol: "tcp"},
		}
		p2 := []compute.PortMapping{
			{ContainerPort: 80, HostPort: 8080, Protocol: "tcp"},
		}
		p3 := []compute.PortMapping{
			{ContainerPort: 80, HostPort: 8081, Protocol: "tcp"},
		}

		assert.True(t, portMappingsEqual(p1, p2))
		assert.False(t, portMappingsEqual(p1, p3))
	})

	t.Run("isValidImageRef", func(t *testing.T) {
		assert.True(t, isValidImageRef("alpine"))
		assert.True(t, isValidImageRef("alpine:3.18"))
		assert.True(t, isValidImageRef("docker.io/library/alpine"))
		assert.True(t, isValidImageRef("docker.io/library/alpine:3.18"))
		assert.False(t, isValidImageRef(""))
		assert.False(t, isValidImageRef("-invalid"))
		assert.False(t, isValidImageRef(":invalid"))
	})

	t.Run("buildContainerLabels", func(t *testing.T) {
		spec := &compute.TenantComputeSpec{
			TenantID:     "tenant-123",
			ProviderType: "docker",
			Labels: map[string]string{
				"custom":                    "value",
				compute.MetadataOwnerKey:    "override",
				compute.MetadataTenantIDKey: "override",
			},
		}
		parsedConfig := &DockerComputeConfig{
			Labels: map[string]string{
				"provider_label": "from-config",
			},
		}

		labels := buildContainerLabels(spec, parsedConfig)
		assert.Equal(t, compute.MetadataOwnerValue, labels[compute.MetadataOwnerKey])
		assert.Equal(t, "tenant-123", labels[compute.MetadataTenantIDKey])
		assert.Equal(t, "docker", labels[compute.MetadataProviderKey])
		assert.Equal(t, "value", labels["custom"])
		assert.Equal(t, "from-config", labels["provider_label"])
	})
}
