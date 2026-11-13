package kubernetes

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRestConfigFromConfigUsesDirectAPIServer(t *testing.T) {
	config, err := restConfigFromConfig(Config{
		APIServer:             "https://cluster.example",
		BearerToken:           "token",
		InsecureSkipTLSVerify: true,
	})
	require.NoError(t, err)
	require.Equal(t, "https://cluster.example", config.Host)
	require.Equal(t, "token", config.BearerToken)
	require.True(t, config.TLSClientConfig.Insecure)
	require.Equal(t, clientUserAgent, config.UserAgent)
}

func TestRestConfigFromConfigUsesConfiguredKubeconfigContext(t *testing.T) {
	kubeconfig := filepath.Join(t.TempDir(), "config")
	err := os.WriteFile(kubeconfig, []byte(`apiVersion: v1
clusters:
- cluster:
    server: https://default.example
  name: default
- cluster:
    server: https://selected.example
  name: selected
contexts:
- context:
    cluster: default
    user: default
  name: default
- context:
    cluster: selected
    user: selected
  name: selected
current-context: default
users:
- name: default
  user:
    token: default-token
- name: selected
  user:
    token: selected-token
`), 0o600)
	require.NoError(t, err)

	config, err := restConfigFromConfig(Config{Kubeconfig: kubeconfig, Context: "selected"})
	require.NoError(t, err)
	require.Equal(t, "https://selected.example", config.Host)
	require.Equal(t, "selected-token", config.BearerToken)
	require.Equal(t, clientUserAgent, config.UserAgent)
}

func TestNewConstructsKubernetesClient(t *testing.T) {
	provider, err := New(Config{APIServer: "https://cluster.example", BearerToken: "token"}, nil)
	require.NoError(t, err)
	require.NotNil(t, provider.client)
}
