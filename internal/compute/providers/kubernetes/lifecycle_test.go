package kubernetes

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"

	"github.com/jaxxstorm/landlord/internal/compute"
	"github.com/stretchr/testify/require"
)

func TestLifecycleProvisionStatusUpdateAndDestroy(t *testing.T) {
	ctx := context.Background()
	client := fake.NewSimpleClientset()
	client.PrependReactor("create", "namespaces", func(action k8stesting.Action) (bool, runtime.Object, error) {
		namespace := action.(k8stesting.CreateAction).GetObject().(*corev1.Namespace)
		namespace.UID = types.UID("namespace-uid")
		return false, nil, nil
	})
	provider := lifecycleProvider(t, client)
	spec := lifecycleSpec()

	provisioned, err := provider.Provision(ctx, spec)
	require.NoError(t, err)
	require.Equal(t, compute.ProvisionStatusSuccess, provisioned.Status)
	require.Equal(t, "landlord-tenant-tenant-1", provisioned.ResourceIDs["namespace"])
	require.Equal(t, deploymentName, provisioned.ResourceIDs["deployment"])

	namespace, err := client.CoreV1().Namespaces().Get(ctx, provisioned.ResourceIDs["namespace"], metav1.GetOptions{})
	require.NoError(t, err)
	deployment, err := client.AppsV1().Deployments(namespace.Name).Get(ctx, deploymentName, metav1.GetOptions{})
	require.NoError(t, err)
	require.Len(t, deployment.OwnerReferences, 1)
	require.Equal(t, namespace.UID, deployment.OwnerReferences[0].UID)

	_, err = provider.Provision(ctx, spec)
	require.NoError(t, err)

	deployment.Status.ReadyReplicas = 1
	deployment.Status.AvailableReplicas = 1
	_, err = client.AppsV1().Deployments(namespace.Name).UpdateStatus(ctx, deployment, metav1.UpdateOptions{})
	require.NoError(t, err)
	_, err = client.CoreV1().Pods(namespace.Name).Create(ctx, &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "runner-pod", Namespace: namespace.Name, Labels: deployment.Spec.Selector.MatchLabels},
		Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{
			Name: "worker", Ready: true, State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}},
		}}},
	}, metav1.CreateOptions{})
	require.NoError(t, err)

	status, err := provider.GetStatus(ctx, spec.TenantID)
	require.NoError(t, err)
	require.Equal(t, compute.ComputeStateRunning, status.State)
	require.Equal(t, compute.HealthStatusHealthy, status.Health)
	require.Equal(t, []compute.ContainerStatus{{Name: "worker", State: "running", Ready: true}}, status.Containers)

	updated := lifecycleSpec()
	updated.Containers[0].Image = "example/worker:v2"
	result, err := provider.Update(ctx, spec.TenantID, updated)
	require.NoError(t, err)
	require.Equal(t, compute.UpdateStatusSuccess, result.Status)
	require.Equal(t, []string{"deployment updated"}, result.Changes)

	deployment, err = client.AppsV1().Deployments(namespace.Name).Get(ctx, deploymentName, metav1.GetOptions{})
	require.NoError(t, err)
	require.Equal(t, "example/worker:v2", deployment.Spec.Template.Spec.Containers[0].Image)
	require.Equal(t, namespace.UID, deployment.OwnerReferences[0].UID)

	result, err = provider.Update(ctx, spec.TenantID, updated)
	require.NoError(t, err)
	require.Equal(t, compute.UpdateStatusNoChanges, result.Status)

	require.NoError(t, provider.Destroy(ctx, spec.TenantID))
	require.NoError(t, provider.Destroy(ctx, spec.TenantID))
	_, err = client.CoreV1().Namespaces().Get(ctx, namespace.Name, metav1.GetOptions{})
	require.True(t, apierrors.IsNotFound(err))
	_, err = client.AppsV1().Deployments(namespace.Name).Get(ctx, deploymentName, metav1.GetOptions{})
	require.True(t, apierrors.IsNotFound(err))
}

func TestGetStatusMapsFailedPod(t *testing.T) {
	ctx := context.Background()
	provider := lifecycleProvider(t, fake.NewSimpleClientset())
	spec := lifecycleSpec()
	namespace := provider.namespaceFor(spec)
	namespace.UID = types.UID("namespace-uid")
	_, err := provider.client.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	require.NoError(t, err)
	deployment := provider.deploymentFor(spec, namespace)
	deployment.Status.Conditions = []appsv1.DeploymentCondition{{Type: appsv1.DeploymentProgressing, Status: corev1.ConditionFalse}}
	_, err = provider.client.AppsV1().Deployments(namespace.Name).Create(ctx, deployment, metav1.CreateOptions{})
	require.NoError(t, err)
	_, err = provider.client.CoreV1().Pods(namespace.Name).Create(ctx, &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "failed-runner", Namespace: namespace.Name, Labels: deployment.Spec.Selector.MatchLabels},
		Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{{
			Name: "worker", State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"}},
		}}},
	}, metav1.CreateOptions{})
	require.NoError(t, err)

	status, err := provider.GetStatus(ctx, spec.TenantID)
	require.NoError(t, err)
	require.Equal(t, compute.ComputeStateFailed, status.State)
	require.Equal(t, compute.HealthStatusUnhealthy, status.Health)
	require.Equal(t, "failed", status.Containers[0].State)
}

func lifecycleProvider(t *testing.T, client KubernetesClient) *Provider {
	t.Helper()
	provider, err := NewWithClient(Config{APIServer: "https://cluster.example", BearerToken: "token"}, nil, client)
	require.NoError(t, err)
	return provider
}

func lifecycleSpec() *compute.TenantComputeSpec {
	return &compute.TenantComputeSpec{
		TenantID: "tenant-1", ProviderType: "kubernetes",
		Containers: []compute.ContainerSpec{{Name: "worker", Image: "example/worker:v1", Env: map[string]string{"LOG_LEVEL": "info"}}},
	}
}
