// Package kubernetes provides Kubernetes-backed tenant compute workloads.
package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"

	"github.com/jaxxstorm/landlord/internal/compute"
)

const defaultNamespacePrefix = "landlord-tenant-"

// Config contains operator-owned Kubernetes connection and naming settings.
type Config struct {
	Kubeconfig            string
	Context               string
	APIServer             string
	BearerToken           string
	NamespacePrefix       string
	InsecureSkipTLSVerify bool
	Defaults              map[string]interface{}
}

// Provider is the Kubernetes compute provider configuration boundary.
//
// Kubernetes client construction and workload lifecycle operations are added
// separately. The methods below satisfy compute.Provider so the provider can be
// registered and its configuration discovered before that work is complete.
type Provider struct {
	config   Config
	defaults json.RawMessage
	logger   *zap.Logger
	client   KubernetesClient
}

var _ compute.Provider = (*Provider)(nil)

// New creates a Kubernetes provider with a client built from operator configuration.
func New(cfg Config, logger *zap.Logger) (*Provider, error) {
	if err := validateOperatorConfig(cfg); err != nil {
		return nil, err
	}
	client, err := newKubernetesClient(cfg)
	if err != nil {
		return nil, err
	}
	return NewWithClient(cfg, logger, client)
}

// NewWithClient creates a Kubernetes provider with an injected client. It is
// used by lifecycle tests so they do not require Kubernetes credentials.
func NewWithClient(cfg Config, logger *zap.Logger, client KubernetesClient) (*Provider, error) {
	if err := validateOperatorConfig(cfg); err != nil {
		return nil, err
	}
	if client == nil {
		return nil, fmt.Errorf("Kubernetes client is required")
	}
	if cfg.NamespacePrefix == "" {
		cfg.NamespacePrefix = defaultNamespacePrefix
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	defaults, err := marshalDefaults(cfg.Defaults)
	if err != nil {
		return nil, fmt.Errorf("marshal Kubernetes config defaults: %w", err)
	}
	return &Provider{
		config:   cfg,
		defaults: defaults,
		logger:   logger.With(zap.String("component", "kubernetes-provider")),
		client:   client,
	}, nil
}

func newKubernetesClient(cfg Config) (*kubernetes.Clientset, error) {
	restConfig, err := restConfigFromConfig(cfg)
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("create Kubernetes client: %w", err)
	}
	return client, nil
}

func validateOperatorConfig(cfg Config) error {
	hasKubeconfig := strings.TrimSpace(cfg.Kubeconfig) != ""
	hasAPIServer := strings.TrimSpace(cfg.APIServer) != ""
	switch {
	case !hasKubeconfig && !hasAPIServer:
		return fmt.Errorf("kubeconfig or api_server is required")
	case hasKubeconfig && hasAPIServer:
		return fmt.Errorf("only one of kubeconfig or api_server may be configured")
	}
	if hasAPIServer {
		endpoint, err := url.ParseRequestURI(cfg.APIServer)
		if err != nil || endpoint.Host == "" || (endpoint.Scheme != "https" && endpoint.Scheme != "http") {
			return fmt.Errorf("api_server must be a valid HTTP(S) URL")
		}
		if strings.TrimSpace(cfg.BearerToken) == "" {
			return fmt.Errorf("bearer_token is required when api_server is configured")
		}
	}
	if strings.TrimSpace(cfg.NamespacePrefix) != "" && !strings.HasSuffix(cfg.NamespacePrefix, "-") {
		return fmt.Errorf("namespace_prefix must end with a hyphen")
	}
	if len(cfg.NamespacePrefix) > 54 {
		return fmt.Errorf("namespace_prefix must be at most 54 characters")
	}
	return nil
}

func marshalDefaults(defaults map[string]interface{}) (json.RawMessage, error) {
	if len(defaults) == 0 {
		return nil, nil
	}
	return json.Marshal(defaults)
}

// Name returns the compute provider identifier.
func (p *Provider) Name() string { return "kubernetes" }

// ValidateConfig validates that tenant-specific Kubernetes configuration is a
// JSON object. Workload-specific fields will be introduced with lifecycle work.
func (p *Provider) ValidateConfig(raw json.RawMessage) error {
	if len(raw) == 0 {
		return nil
	}
	var value interface{}
	if err := json.Unmarshal(raw, &value); err != nil {
		return fmt.Errorf("invalid Kubernetes config: %w", err)
	}
	if _, ok := value.(map[string]interface{}); !ok {
		return fmt.Errorf("Kubernetes config must be an object")
	}
	return nil
}

// ConfigSchema returns the tenant-facing Kubernetes configuration schema.
func (p *Provider) ConfigSchema() json.RawMessage { return configSchema }

// ConfigDefaults returns configured tenant-facing Kubernetes defaults.
func (p *Provider) ConfigDefaults() json.RawMessage { return p.defaults }

// Validate validates a Kubernetes compute specification without provisioning.
func (p *Provider) Validate(_ context.Context, spec *compute.TenantComputeSpec) error {
	if spec == nil || strings.TrimSpace(spec.TenantID) == "" {
		return fmt.Errorf("tenant ID is required")
	}
	if spec.ProviderType != "" && spec.ProviderType != p.Name() {
		return fmt.Errorf("provider type must be %s", p.Name())
	}
	return p.ValidateConfig(spec.ProviderConfig)
}

// Provision creates the tenant namespace and its runner Deployment. Existing
// managed resources are returned unchanged so retries do not create duplicates.
func (p *Provider) Provision(ctx context.Context, spec *compute.TenantComputeSpec) (*compute.ProvisionResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := p.Validate(ctx, spec); err != nil {
		return nil, fmt.Errorf("%w: %v", compute.ErrInvalidSpec, err)
	}

	namespace, err := p.client.CoreV1().Namespaces().Get(ctx, p.tenantNamespaceName(spec.TenantID), metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		namespace, err = p.client.CoreV1().Namespaces().Create(ctx, p.namespaceFor(spec), metav1.CreateOptions{})
	}
	if err != nil {
		return nil, provisioningError(ctx, "create Kubernetes namespace", err)
	}
	if !p.ownsNamespace(namespace, spec.TenantID) {
		return nil, fmt.Errorf("%w: namespace %q is not managed for tenant %q", compute.ErrProvisionFailed, namespace.Name, spec.TenantID)
	}

	deployment, err := p.client.AppsV1().Deployments(namespace.Name).Get(ctx, deploymentName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		deployment, err = p.client.AppsV1().Deployments(namespace.Name).Create(ctx, p.deploymentFor(spec, namespace), metav1.CreateOptions{})
	}
	if err != nil {
		return nil, provisioningError(ctx, "create Kubernetes Deployment", err)
	}
	if !p.ownsDeployment(deployment, namespace, spec.TenantID) {
		return nil, fmt.Errorf("%w: Deployment %q in namespace %q is not managed for tenant %q", compute.ErrProvisionFailed, deployment.Name, namespace.Name, spec.TenantID)
	}

	return &compute.ProvisionResult{
		TenantID:      spec.TenantID,
		ProviderType:  p.Name(),
		Status:        compute.ProvisionStatusSuccess,
		ResourceIDs:   map[string]string{"namespace": namespace.Name, "deployment": deployment.Name},
		Message:       "Kubernetes runner provisioned",
		ProvisionedAt: time.Now(),
	}, nil
}

// Update applies a desired runner Deployment without replacing the tenant
// namespace. A semantically unchanged Deployment is left untouched.
func (p *Provider) Update(ctx context.Context, tenantID string, spec *compute.TenantComputeSpec) (*compute.UpdateResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if spec == nil {
		return nil, fmt.Errorf("%w: spec is required", compute.ErrInvalidSpec)
	}
	desiredSpec := *spec
	desiredSpec.TenantID = tenantID
	if err := p.Validate(ctx, &desiredSpec); err != nil {
		return nil, fmt.Errorf("%w: %v", compute.ErrInvalidSpec, err)
	}

	namespace, err := p.client.CoreV1().Namespaces().Get(ctx, p.tenantNamespaceName(tenantID), metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return nil, fmt.Errorf("%w: %s", compute.ErrTenantNotFound, tenantID)
	}
	if err != nil {
		return nil, lifecycleError(ctx, "get Kubernetes namespace", err)
	}
	if !p.ownsNamespace(namespace, tenantID) {
		return nil, fmt.Errorf("%w: %s", compute.ErrTenantNotFound, tenantID)
	}

	existing, err := p.client.AppsV1().Deployments(namespace.Name).Get(ctx, deploymentName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return nil, fmt.Errorf("%w: %s", compute.ErrTenantNotFound, tenantID)
	}
	if err != nil {
		return nil, lifecycleError(ctx, "get Kubernetes Deployment", err)
	}
	if !p.ownsDeployment(existing, namespace, tenantID) {
		return nil, fmt.Errorf("%w: %s", compute.ErrTenantNotFound, tenantID)
	}

	desired := p.deploymentFor(&desiredSpec, namespace)
	if deploymentMatchesDesired(existing, desired) {
		return updateResult(tenantID, compute.UpdateStatusNoChanges, nil), nil
	}
	desired.ResourceVersion = existing.ResourceVersion
	if _, err := p.client.AppsV1().Deployments(namespace.Name).Update(ctx, desired, metav1.UpdateOptions{}); err != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("%w: update Kubernetes Deployment: %v", compute.ErrUpdateFailed, err)
	}
	return updateResult(tenantID, compute.UpdateStatusSuccess, []string{"deployment updated"}), nil
}

// Destroy deletes the managed runner and namespace. Missing resources are a
// successful no-op, as required by the compute.Provider contract.
func (p *Provider) Destroy(ctx context.Context, tenantID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	namespace, err := p.client.CoreV1().Namespaces().Get(ctx, p.tenantNamespaceName(tenantID), metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return lifecycleError(ctx, "get Kubernetes namespace", err)
	}
	if !p.ownsNamespace(namespace, tenantID) {
		return fmt.Errorf("namespace %q is not managed for tenant %q", namespace.Name, tenantID)
	}

	deployment, err := p.client.AppsV1().Deployments(namespace.Name).Get(ctx, deploymentName, metav1.GetOptions{})
	if err == nil {
		if !p.ownsDeployment(deployment, namespace, tenantID) {
			return fmt.Errorf("Deployment %q in namespace %q is not managed for tenant %q", deployment.Name, namespace.Name, tenantID)
		}
		if err := p.client.AppsV1().Deployments(namespace.Name).Delete(ctx, deploymentName, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
			return lifecycleError(ctx, "delete Kubernetes Deployment", err)
		}
	} else if !apierrors.IsNotFound(err) {
		return lifecycleError(ctx, "get Kubernetes Deployment", err)
	}

	if err := p.client.CoreV1().Namespaces().Delete(ctx, namespace.Name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return lifecycleError(ctx, "delete Kubernetes namespace", err)
	}
	return nil
}

// GetStatus maps Deployment availability and Pod container readiness to the
// provider-neutral compute status model.
func (p *Provider) GetStatus(ctx context.Context, tenantID string) (*compute.ComputeStatus, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	namespace, err := p.client.CoreV1().Namespaces().Get(ctx, p.tenantNamespaceName(tenantID), metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return nil, fmt.Errorf("%w: %s", compute.ErrTenantNotFound, tenantID)
	}
	if err != nil {
		return nil, lifecycleError(ctx, "get Kubernetes namespace", err)
	}
	if !p.ownsNamespace(namespace, tenantID) {
		return nil, fmt.Errorf("%w: %s", compute.ErrTenantNotFound, tenantID)
	}

	deployment, err := p.client.AppsV1().Deployments(namespace.Name).Get(ctx, deploymentName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return nil, fmt.Errorf("%w: %s", compute.ErrTenantNotFound, tenantID)
	}
	if err != nil {
		return nil, lifecycleError(ctx, "get Kubernetes Deployment", err)
	}
	if !p.ownsDeployment(deployment, namespace, tenantID) {
		return nil, fmt.Errorf("%w: %s", compute.ErrTenantNotFound, tenantID)
	}

	pods, err := p.client.CoreV1().Pods(namespace.Name).List(ctx, metav1.ListOptions{LabelSelector: labels.Set(deployment.Spec.Selector.MatchLabels).String()})
	if err != nil {
		return nil, lifecycleError(ctx, "list Kubernetes Pods", err)
	}
	containers, podsReady, podsFailed := podStatuses(pods.Items)
	state := deploymentState(deployment, podsReady, podsFailed)
	health := compute.HealthStatusUnknown
	if state == compute.ComputeStateRunning {
		health = compute.HealthStatusHealthy
	} else if state == compute.ComputeStateFailed {
		health = compute.HealthStatusUnhealthy
	}

	return &compute.ComputeStatus{
		TenantID: tenantID, ProviderType: p.Name(), State: state, Containers: containers, Health: health,
		LastUpdated: time.Now(), Metadata: map[string]string{
			"namespace":      namespace.Name,
			"deployment":     deployment.Name,
			"ready_replicas": strconv.FormatInt(int64(deployment.Status.ReadyReplicas), 10),
		},
	}, nil
}

func provisioningError(ctx context.Context, operation string, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return fmt.Errorf("%w: %s: %v", compute.ErrProvisionFailed, operation, err)
}

func lifecycleError(ctx context.Context, operation string, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func updateResult(tenantID string, status compute.UpdateStatus, changes []string) *compute.UpdateResult {
	return &compute.UpdateResult{TenantID: tenantID, ProviderType: "kubernetes", Status: status, Changes: changes, UpdatedAt: time.Now()}
}

func (p *Provider) ownsNamespace(namespace *corev1.Namespace, tenantID string) bool {
	return namespace != nil && namespace.Labels[compute.MetadataOwnerKey] == compute.MetadataOwnerValue &&
		namespace.Labels[compute.MetadataTenantIDKey] == tenantID && namespace.Labels[compute.MetadataProviderKey] == p.Name()
}

func (p *Provider) ownsDeployment(deployment *appsv1.Deployment, namespace *corev1.Namespace, tenantID string) bool {
	if deployment == nil || namespace == nil || deployment.Namespace != namespace.Name ||
		deployment.Labels[compute.MetadataOwnerKey] != compute.MetadataOwnerValue ||
		deployment.Labels[compute.MetadataTenantIDKey] != tenantID || deployment.Labels[compute.MetadataProviderKey] != p.Name() {
		return false
	}
	if namespace.UID == "" {
		return true
	}
	for _, owner := range deployment.OwnerReferences {
		if owner.APIVersion == "v1" && owner.Kind == "Namespace" && owner.Name == namespace.Name && owner.UID == namespace.UID && owner.Controller != nil && *owner.Controller {
			return true
		}
	}
	return false
}

func deploymentMatchesDesired(existing, desired *appsv1.Deployment) bool {
	return deploymentReplicas(existing) == deploymentReplicas(desired) &&
		reflect.DeepEqual(existing.Labels, desired.Labels) &&
		reflect.DeepEqual(existing.Spec.Selector, desired.Spec.Selector) &&
		reflect.DeepEqual(existing.Spec.Template.Labels, desired.Spec.Template.Labels) &&
		containersMatch(existing.Spec.Template.Spec.Containers, desired.Spec.Template.Spec.Containers)
}

func deploymentReplicas(deployment *appsv1.Deployment) int32 {
	if deployment.Spec.Replicas == nil {
		return 1
	}
	return *deployment.Spec.Replicas
}

func containersMatch(existing, desired []corev1.Container) bool {
	if len(existing) != len(desired) {
		return false
	}
	for index := range desired {
		actual := existing[index]
		expected := desired[index]
		if actual.Name != expected.Name || actual.Image != expected.Image ||
			!reflect.DeepEqual(actual.Command, expected.Command) || !reflect.DeepEqual(actual.Args, expected.Args) ||
			!reflect.DeepEqual(actual.Env, expected.Env) || !reflect.DeepEqual(actual.Ports, expected.Ports) ||
			!reflect.DeepEqual(actual.Resources.Limits, expected.Resources.Limits) {
			return false
		}
	}
	return true
}

func podStatuses(pods []corev1.Pod) ([]compute.ContainerStatus, bool, bool) {
	statuses := make([]compute.ContainerStatus, 0)
	ready := len(pods) > 0
	failed := false
	for _, pod := range pods {
		if len(pod.Status.ContainerStatuses) == 0 {
			ready = false
		}
		for _, status := range pod.Status.ContainerStatuses {
			container := compute.ContainerStatus{Name: status.Name, Ready: status.Ready, RestartCount: int(status.RestartCount)}
			switch {
			case status.State.Running != nil:
				container.State = "running"
			case status.State.Terminated != nil:
				container.Message = status.State.Terminated.Message
				if status.State.Terminated.ExitCode == 0 {
					container.State = "stopped"
				} else {
					container.State = "failed"
					failed = true
				}
			case status.State.Waiting != nil:
				container.Message = status.State.Waiting.Message
				if failedWaitingReason(status.State.Waiting.Reason) {
					container.State = "failed"
					failed = true
				} else {
					container.State = "starting"
				}
			default:
				container.State = "unknown"
			}
			if !container.Ready {
				ready = false
			}
			statuses = append(statuses, container)
		}
	}
	return statuses, ready, failed
}

func failedWaitingReason(reason string) bool {
	switch reason {
	case "CrashLoopBackOff", "CreateContainerConfigError", "ErrImagePull", "ImagePullBackOff", "InvalidImageName", "RunContainerError":
		return true
	default:
		return false
	}
}

func deploymentState(deployment *appsv1.Deployment, podsReady, podsFailed bool) compute.ComputeState {
	if deployment.DeletionTimestamp != nil {
		return compute.ComputeStateStopping
	}
	if deploymentReplicas(deployment) == 0 {
		return compute.ComputeStateStopped
	}
	if podsFailed || deploymentFailed(deployment) {
		return compute.ComputeStateFailed
	}
	desired := deploymentReplicas(deployment)
	if deployment.Status.ReadyReplicas >= desired && deployment.Status.AvailableReplicas >= desired && podsReady {
		return compute.ComputeStateRunning
	}
	return compute.ComputeStateStarting
}

func deploymentFailed(deployment *appsv1.Deployment) bool {
	for _, condition := range deployment.Status.Conditions {
		if (condition.Type == appsv1.DeploymentReplicaFailure && condition.Status == corev1.ConditionTrue) ||
			(condition.Type == appsv1.DeploymentProgressing && condition.Status == corev1.ConditionFalse) {
			return true
		}
	}
	return false
}

var configSchema = json.RawMessage(`{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "additionalProperties": true
}`)
