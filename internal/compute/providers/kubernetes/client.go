package kubernetes

import (
	"fmt"
	"strings"

	appsv1client "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const clientUserAgent = "landlord-kubernetes-provider"

// KubernetesClient is the Kubernetes API surface required by this provider.
// Lifecycle operations use only namespaced workload and namespace APIs.
type KubernetesClient interface {
	AppsV1() appsv1client.AppsV1Interface
	CoreV1() corev1client.CoreV1Interface
}

func restConfigFromConfig(cfg Config) (*rest.Config, error) {
	if strings.TrimSpace(cfg.Kubeconfig) != "" {
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		loadingRules.ExplicitPath = cfg.Kubeconfig
		overrides := &clientcmd.ConfigOverrides{}
		if strings.TrimSpace(cfg.Context) != "" {
			overrides.CurrentContext = cfg.Context
		}
		restConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides).ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("load Kubernetes kubeconfig: %w", err)
		}
		if cfg.InsecureSkipTLSVerify {
			restConfig.TLSClientConfig.Insecure = true
		}
		restConfig.UserAgent = clientUserAgent
		return restConfig, nil
	}

	return &rest.Config{
		Host:        strings.TrimSpace(cfg.APIServer),
		BearerToken: strings.TrimSpace(cfg.BearerToken),
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: cfg.InsecureSkipTLSVerify,
		},
		UserAgent: clientUserAgent,
	}, nil
}
