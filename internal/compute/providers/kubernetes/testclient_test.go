package kubernetes

import (
	appsv1client "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
)

type testKubernetesClient struct{}

func (testKubernetesClient) AppsV1() appsv1client.AppsV1Interface { return nil }

func (testKubernetesClient) CoreV1() corev1client.CoreV1Interface { return nil }
