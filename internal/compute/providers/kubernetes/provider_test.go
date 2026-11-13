package kubernetes

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/jaxxstorm/landlord/internal/compute"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewValidatesOperatorConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		errText string
	}{
		{name: "missing connection", errText: "kubeconfig or api_server is required"},
		{name: "both connection types", config: Config{Kubeconfig: "/tmp/config", APIServer: "https://cluster.example", BearerToken: "token"}, errText: "only one"},
		{name: "invalid api server", config: Config{APIServer: "cluster.example", BearerToken: "token"}, errText: "valid HTTP(S) URL"},
		{name: "api server without token", config: Config{APIServer: "https://cluster.example"}, errText: "bearer_token is required"},
		{name: "invalid namespace prefix", config: Config{Kubeconfig: "/tmp/config", NamespacePrefix: "landlord"}, errText: "must end with a hyphen"},
		{name: "namespace prefix is too long", config: Config{Kubeconfig: "/tmp/config", NamespacePrefix: strings.Repeat("a", 54) + "-"}, errText: "at most 54 characters"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := New(test.config, nil)
			require.ErrorContains(t, err, test.errText)
		})
	}
}

func TestNewAppliesDefaultsAndExposesConfigDefaults(t *testing.T) {
	provider, err := New(Config{
		APIServer:   "https://cluster.example",
		BearerToken: "token",
		Defaults:    map[string]interface{}{"image_pull_policy": "IfNotPresent"},
	}, nil)
	require.NoError(t, err)
	require.Equal(t, defaultNamespacePrefix, provider.config.NamespacePrefix)
	require.JSONEq(t, `{"image_pull_policy":"IfNotPresent"}`, string(provider.ConfigDefaults()))
}

func TestProviderValidatesTenantConfiguration(t *testing.T) {
	provider, err := New(Config{APIServer: "https://cluster.example", BearerToken: "token"}, nil)
	require.NoError(t, err)
	require.NoError(t, provider.ValidateConfig(nil))
	require.NoError(t, provider.ValidateConfig(json.RawMessage(`{"image_pull_policy":"Always"}`)))
	require.ErrorContains(t, provider.ValidateConfig(json.RawMessage(`[]`)), "must be an object")
	require.ErrorContains(t, provider.ValidateConfig(json.RawMessage(`not-json`)), "invalid Kubernetes config")

	require.NoError(t, provider.Validate(context.Background(), &compute.TenantComputeSpec{TenantID: "tenant-1", ProviderType: "kubernetes"}))
	require.ErrorContains(t, provider.Validate(context.Background(), &compute.TenantComputeSpec{TenantID: "tenant-1", ProviderType: "docker"}), "provider type must be kubernetes")
}

func TestProviderRegisters(t *testing.T) {
	provider, err := New(Config{APIServer: "https://cluster.example", BearerToken: "token"}, nil)
	require.NoError(t, err)

	registry := compute.NewRegistry(zap.NewNop())
	require.NoError(t, registry.Register(provider))
	require.True(t, registry.Has("kubernetes"))
}
