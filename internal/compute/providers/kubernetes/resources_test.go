package kubernetes

import (
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/jaxxstorm/landlord/internal/compute"
	"github.com/stretchr/testify/require"
)

func TestTenantNamespaceAndDeploymentAreNamedAndOwnedDeterministically(t *testing.T) {
	provider := testProvider(t)
	spec := &compute.TenantComputeSpec{
		TenantID:     "Tenant / One",
		ProviderType: "docker",
		Labels: map[string]string{
			"team":                      "platform",
			compute.MetadataOwnerKey:    "other",
			compute.MetadataTenantIDKey: "other-tenant",
		},
		Containers: []compute.ContainerSpec{{
			Name:    "worker",
			Image:   "example/worker:v1",
			Command: []string{"worker"},
			Args:    []string{"--serve"},
			Env:     map[string]string{"Z": "last", "A": "first"},
			Ports: []compute.PortMapping{
				{Name: "http", ContainerPort: 8080, Protocol: "tcp"},
				{Name: "metrics", ContainerPort: 9090, Protocol: "udp"},
			},
		}},
		Resources: compute.ResourceRequirements{CPU: 250, Memory: 128, Storage: 64},
	}

	namespace := provider.namespaceFor(spec)
	namespace.UID = types.UID("namespace-uid")
	deployment := provider.deploymentFor(spec, namespace)

	require.Equal(t, "landlord-tenant-tenant-one", namespace.Name)
	require.Equal(t, namespace.Name, deployment.Namespace)
	require.Equal(t, deploymentName, deployment.Name)
	require.Equal(t, compute.MetadataOwnerValue, deployment.Labels[compute.MetadataOwnerKey])
	require.Equal(t, "Tenant / One", deployment.Labels[compute.MetadataTenantIDKey])
	require.Equal(t, provider.Name(), deployment.Labels[compute.MetadataProviderKey])
	require.Equal(t, "platform", deployment.Labels["team"])
	require.NotContains(t, deployment.Spec.Selector.MatchLabels, "team")
	require.Equal(t, compute.MetadataOwnerValue, deployment.Spec.Selector.MatchLabels[compute.MetadataOwnerKey])
	require.Equal(t, "Tenant / One", deployment.Spec.Selector.MatchLabels[compute.MetadataTenantIDKey])
	require.Equal(t, provider.Name(), deployment.Spec.Selector.MatchLabels[compute.MetadataProviderKey])
	require.Equal(t, deployment.Labels, deployment.Spec.Template.Labels)
	require.Equal(t, namespace.Name, deployment.OwnerReferences[0].Name)
	require.Equal(t, namespace.UID, deployment.OwnerReferences[0].UID)
	require.True(t, *deployment.OwnerReferences[0].Controller)

	container := deployment.Spec.Template.Spec.Containers[0]
	require.Equal(t, []string{"worker"}, container.Command)
	require.Equal(t, []string{"--serve"}, container.Args)
	require.Equal(t, []corev1.EnvVar{{Name: "A", Value: "first"}, {Name: "Z", Value: "last"}}, container.Env)
	require.Equal(t, corev1.ProtocolTCP, container.Ports[0].Protocol)
	require.Equal(t, corev1.ProtocolUDP, container.Ports[1].Protocol)
	require.Equal(t, "250m", container.Resources.Limits.Cpu().String())
	require.Equal(t, "128Mi", container.Resources.Limits.Memory().String())
	storage := container.Resources.Limits[corev1.ResourceEphemeralStorage]
	require.Equal(t, "64Mi", storage.String())
}

func TestTenantNamespaceNameIsStableAndWithinKubernetesLimit(t *testing.T) {
	provider := testProvider(t)
	tenantID := "tenant/" + strings.Repeat("a", 100)
	first := provider.tenantNamespaceName(tenantID)

	require.Equal(t, first, provider.tenantNamespaceName(tenantID))
	require.LessOrEqual(t, len(first), 63)
	require.NotEqual(t, first, provider.tenantNamespaceName(tenantID+"other"))
}

func testProvider(t *testing.T) *Provider {
	t.Helper()
	provider, err := NewWithClient(Config{APIServer: "https://cluster.example", BearerToken: "token"}, nil, testKubernetesClient{})
	require.NoError(t, err)
	return provider
}
