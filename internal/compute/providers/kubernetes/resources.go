package kubernetes

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jaxxstorm/landlord/internal/compute"
)

const deploymentName = "runner"

// tenantNamespaceName returns a DNS-label-safe namespace name. Long or
// non-DNS tenant identifiers retain a hash suffix to avoid collisions.
func (p *Provider) tenantNamespaceName(tenantID string) string {
	return p.config.NamespacePrefix + kubernetesName(tenantID, 63-len(p.config.NamespacePrefix))
}

func kubernetesName(value string, maxLength int) string {
	var name strings.Builder
	for _, character := range strings.ToLower(value) {
		switch {
		case character >= 'a' && character <= 'z', character >= '0' && character <= '9':
			name.WriteRune(character)
		default:
			if name.Len() == 0 || !strings.HasSuffix(name.String(), "-") {
				name.WriteByte('-')
			}
		}
	}

	normalized := strings.Trim(name.String(), "-")
	if normalized == "" {
		normalized = "tenant"
	}
	if len(normalized) <= maxLength {
		return normalized
	}
	suffix := fmt.Sprintf("-%x", sha256.Sum256([]byte(value)))[:9]
	return strings.TrimRight(normalized[:maxLength-len(suffix)], "-") + suffix
}

func (p *Provider) tenantLabels(spec *compute.TenantComputeSpec) map[string]string {
	metadataSpec := *spec
	metadataSpec.ProviderType = p.Name()
	return compute.MergeLabels(spec.Labels, compute.DefaultMetadata(&metadataSpec))
}

func (p *Provider) workloadSelectorLabels(spec *compute.TenantComputeSpec) map[string]string {
	metadataSpec := *spec
	metadataSpec.ProviderType = p.Name()
	return compute.DefaultMetadata(&metadataSpec)
}

// namespaceFor creates the namespace resource that owns a tenant workload.
func (p *Provider) namespaceFor(spec *compute.TenantComputeSpec) *corev1.Namespace {
	return &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
		Name:   p.tenantNamespaceName(spec.TenantID),
		Labels: p.tenantLabels(spec),
	}}
}

// deploymentFor creates the single tenant runner Deployment. The namespace
// owner reference is populated after namespace creation, when its UID is known.
func (p *Provider) deploymentFor(spec *compute.TenantComputeSpec, namespace *corev1.Namespace) *appsv1.Deployment {
	labels := p.tenantLabels(spec)
	selectorLabels := p.workloadSelectorLabels(spec)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: deploymentName, Namespace: namespace.Name, Labels: copyLabels(labels)},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Pointer(1),
			Selector: &metav1.LabelSelector{MatchLabels: selectorLabels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: copyLabels(labels)},
				Spec:       corev1.PodSpec{Containers: containersFor(spec)},
			},
		},
	}
	if namespace.UID != "" {
		controller := true
		deployment.OwnerReferences = []metav1.OwnerReference{{
			APIVersion: "v1",
			Kind:       "Namespace",
			Name:       namespace.Name,
			UID:        namespace.UID,
			Controller: &controller,
		}}
	}
	return deployment
}

func copyLabels(labels map[string]string) map[string]string {
	copy := make(map[string]string, len(labels))
	for key, value := range labels {
		copy[key] = value
	}
	return copy
}

func containersFor(spec *compute.TenantComputeSpec) []corev1.Container {
	containers := make([]corev1.Container, 0, len(spec.Containers))
	for _, container := range spec.Containers {
		containers = append(containers, corev1.Container{
			Name:      container.Name,
			Image:     container.Image,
			Command:   container.Command,
			Args:      container.Args,
			Env:       environmentVariables(container.Env),
			Ports:     portsFor(container.Ports),
			Resources: resourcesFor(container.Resources, spec.Resources),
		})
	}
	return containers
}

func environmentVariables(values map[string]string) []corev1.EnvVar {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	variables := make([]corev1.EnvVar, 0, len(keys))
	for _, key := range keys {
		variables = append(variables, corev1.EnvVar{Name: key, Value: values[key]})
	}
	return variables
}

func portsFor(mappings []compute.PortMapping) []corev1.ContainerPort {
	ports := make([]corev1.ContainerPort, 0, len(mappings))
	for _, mapping := range mappings {
		protocol := corev1.ProtocolTCP
		if strings.EqualFold(mapping.Protocol, string(corev1.ProtocolUDP)) {
			protocol = corev1.ProtocolUDP
		}
		ports = append(ports, corev1.ContainerPort{Name: mapping.Name, ContainerPort: int32(mapping.ContainerPort), HostPort: int32(mapping.HostPort), Protocol: protocol})
	}
	return ports
}

func resourcesFor(containerResources *compute.ResourceRequirements, defaultResources compute.ResourceRequirements) corev1.ResourceRequirements {
	resources := defaultResources
	if containerResources != nil {
		resources = *containerResources
	}
	limits := corev1.ResourceList{}
	if resources.CPU > 0 {
		limits[corev1.ResourceCPU] = *resource.NewMilliQuantity(int64(resources.CPU), resource.DecimalSI)
	}
	if resources.Memory > 0 {
		limits[corev1.ResourceMemory] = *resource.NewQuantity(int64(resources.Memory)*1024*1024, resource.BinarySI)
	}
	if resources.Storage > 0 {
		limits[corev1.ResourceEphemeralStorage] = *resource.NewQuantity(int64(resources.Storage)*1024*1024, resource.BinarySI)
	}
	if len(limits) == 0 {
		return corev1.ResourceRequirements{}
	}
	return corev1.ResourceRequirements{Limits: limits}
}

func int32Pointer(value int32) *int32 { return &value }
