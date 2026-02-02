package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"go.uber.org/zap"

	"github.com/jaxxstorm/landlord/internal/compute"
)

// Provider implements the compute.Provider interface using Docker
type Provider struct {
	mu     sync.RWMutex
	client *client.Client
	logger *zap.Logger
	// tenantContainers maps tenant IDs to container IDs
	tenantContainers map[string]string
	// tenantSpecs stores the specs for provisioned tenants
	tenantSpecs map[string]*compute.TenantComputeSpec
}

// Config represents Docker provider configuration
type Config struct {
	// Host is the Docker API endpoint (e.g., "unix:///var/run/docker.sock", "tcp://localhost:2375")
	// Defaults to Docker socket if empty
	Host string `json:"host,omitempty"`

	// NetworkName is the Docker network to use for tenant containers
	// Defaults to "bridge" if empty
	NetworkName string `json:"network_name,omitempty"`

	// NetworkDriver is the driver for the Docker network
	// Common values: "bridge", "overlay"
	NetworkDriver string `json:"network_driver,omitempty"`

	// LabelPrefix is used to label containers for identification
	// Defaults to "landlord"
	LabelPrefix string `json:"label_prefix,omitempty"`
}

const (
	defaultNetworkName   = "bridge"
	defaultNetworkDriver = "bridge"
	defaultLabelPrefix   = "landlord"
	defaultHost          = ""
)

// New creates a new Docker provider
func New(cfg *Config, logger *zap.Logger) (*Provider, error) {
	if cfg == nil {
		cfg = &Config{}
	}

	logger = logger.With(zap.String("component", "docker-provider"))

	// Set defaults
	if cfg.Host == "" {
		cfg.Host = defaultHost
	}
	if cfg.NetworkName == "" {
		cfg.NetworkName = defaultNetworkName
	}
	if cfg.NetworkDriver == "" {
		cfg.NetworkDriver = defaultNetworkDriver
	}
	if cfg.LabelPrefix == "" {
		cfg.LabelPrefix = defaultLabelPrefix
	}

	// Allow overriding host via environment variable for in-container scenarios
	if env := os.Getenv("DOCKER_HOST"); env != "" {
		cfg.Host = env
	}

	// Create Docker client
	// If Host is empty, client.NewClientWithOpts will use the standard Docker socket
	opts := []client.Opt{}
	if cfg.Host != "" {
		opts = append(opts, client.WithHost(cfg.Host))
	}

	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		logger.Error("failed to create docker client", zap.Error(err))
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	// Test the connection
	_, err = cli.Ping(context.Background())
	if err != nil {
		logger.Error("failed to connect to docker daemon", zap.Error(err))
		cli.Close()
		return nil, fmt.Errorf("failed to connect to docker daemon: %w", err)
	}

	p := &Provider{
		client:           cli,
		logger:           logger,
		tenantContainers: make(map[string]string),
		tenantSpecs:      make(map[string]*compute.TenantComputeSpec),
	}

	logger.Info("docker provider initialized", zap.String("host", cfg.Host), zap.String("network", cfg.NetworkName))
	return p, nil
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "docker"
}

// Provision creates a Docker container for a tenant
func (p *Provider) Provision(ctx context.Context, spec *compute.TenantComputeSpec) (*compute.ProvisionResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.tenantContainers[spec.TenantID]; exists {
		return nil, fmt.Errorf("tenant %s already provisioned", spec.TenantID)
	}

	// Each tenant gets exactly one container
	if len(spec.Containers) != 1 {
		return nil, fmt.Errorf("docker provider expects exactly 1 container, got %d", len(spec.Containers))
	}

	parsedConfig, err := applyProviderConfig(spec)
	if err != nil {
		return nil, err
	}

	containerSpec := spec.Containers[0]

	// Create container config
	containerConfig := &container.Config{
		Image: containerSpec.Image,
		Env:   convertEnv(containerSpec.Env),
	}
	labels := buildContainerLabels(spec, parsedConfig)
	if len(labels) > 0 {
		containerConfig.Labels = labels
	}

	if len(containerSpec.Command) > 0 {
		containerConfig.Cmd = containerSpec.Command
	}
	if len(containerSpec.Args) > 0 {
		if containerConfig.Cmd == nil {
			containerConfig.Cmd = containerSpec.Args
		} else {
			containerConfig.Cmd = append(containerConfig.Cmd, containerSpec.Args...)
		}
	}

	// Create port bindings
	portMap := nat.PortMap{}
	for _, port := range containerSpec.Ports {
		natPort, err := nat.NewPort(port.Protocol, fmt.Sprintf("%d", port.ContainerPort))
		if err != nil {
			p.logger.Error("invalid port specification", zap.Int("port", port.ContainerPort), zap.Error(err))
			continue
		}
		portMap[natPort] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: fmt.Sprintf("%d", port.HostPort),
			},
		}
	}

	// Create host config with resource limits
	hostConfig := &container.HostConfig{
		PortBindings: portMap,
		RestartPolicy: container.RestartPolicy{
			Name:              "unless-stopped",
			MaximumRetryCount: 0,
		},
	}
	if parsedConfig != nil {
		if len(parsedConfig.Volumes) > 0 {
			hostConfig.Binds = parsedConfig.Volumes
		}
		if parsedConfig.NetworkMode != "" {
			hostConfig.NetworkMode = container.NetworkMode(parsedConfig.NetworkMode)
		}
		if parsedConfig.RestartPolicy != "" {
			hostConfig.RestartPolicy = container.RestartPolicy{
				Name:              container.RestartPolicyMode(parsedConfig.RestartPolicy),
				MaximumRetryCount: 0,
			}
		}
	}

	// Set resource limits
	if spec.Resources.CPU > 0 {
		hostConfig.CPUQuota = int64(spec.Resources.CPU)
	}
	if spec.Resources.Memory > 0 {
		hostConfig.Memory = int64(spec.Resources.Memory * 1024 * 1024) // Convert MB to bytes
	}

	// Create container
	containerName := fmt.Sprintf("%s-tenant-%s", defaultLabelPrefix, spec.TenantID)
	resp, err := p.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, containerName)
	if err != nil {
		p.logger.Error("failed to create container", zap.String("tenant_id", spec.TenantID), zap.Error(err))
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	containerID := resp.ID

	// Start container
	if err := p.client.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		p.logger.Error("failed to start container", zap.String("container_id", containerID), zap.Error(err))
		// Clean up on start failure
		p.client.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Store references
	p.tenantContainers[spec.TenantID] = containerID
	p.tenantSpecs[spec.TenantID] = spec

	// Get container inspect to build endpoints
	inspectResp, err := p.client.ContainerInspect(ctx, containerID)
	if err != nil {
		p.logger.Error("failed to inspect container", zap.String("container_id", containerID), zap.Error(err))
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	endpoints := buildEndpoints(&containerSpec, &inspectResp)

	p.logger.Info("container provisioned", zap.String("tenant_id", spec.TenantID), zap.String("container_id", containerID))

	return &compute.ProvisionResult{
		TenantID:      spec.TenantID,
		ProviderType:  "docker",
		Status:        compute.ProvisionStatusSuccess,
		ResourceIDs:   map[string]string{"container_id": containerID},
		Endpoints:     endpoints,
		Message:       "Container provisioned successfully",
		ProvisionedAt: time.Now(),
	}, nil
}

// Update modifies an existing tenant's container
func (p *Provider) Update(ctx context.Context, tenantID string, spec *compute.TenantComputeSpec) (*compute.UpdateResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	containerID, exists := p.tenantContainers[tenantID]
	if !exists {
		return nil, fmt.Errorf("%w: %s", compute.ErrTenantNotFound, tenantID)
	}

	oldSpec := p.tenantSpecs[tenantID]
	changes := []string{}

	if _, err := applyProviderConfig(spec); err != nil {
		return nil, err
	}

	// Check for changes
	if len(spec.Containers) != len(oldSpec.Containers) {
		return nil, fmt.Errorf("cannot change number of containers")
	}

	oldContainer := oldSpec.Containers[0]
	newContainer := spec.Containers[0]

	// If image changed, we need to recreate the container
	if oldContainer.Image != newContainer.Image {
		changes = append(changes, "container image changed")
	}

	// If ports changed, we need to recreate the container
	if !portMappingsEqual(oldContainer.Ports, newContainer.Ports) {
		changes = append(changes, "port mappings changed")
	}
	// If environment changed, we need to recreate the container
	if !mapsEqual(oldContainer.Env, newContainer.Env) {
		changes = append(changes, "environment variables changed")
	}
	// Provider-specific config changes should trigger recreation
	if !bytes.Equal(oldSpec.ProviderConfig, spec.ProviderConfig) {
		changes = append(changes, "provider config changed")
	}

	// If resources changed, we can update some but may need restart
	resourcesChanged := oldSpec.Resources.CPU != spec.Resources.CPU ||
		oldSpec.Resources.Memory != spec.Resources.Memory
	if resourcesChanged {
		changes = append(changes, "resource limits changed")
	}

	if len(changes) == 0 {
		return &compute.UpdateResult{
			TenantID:     tenantID,
			ProviderType: "docker",
			Status:       compute.UpdateStatusNoChanges,
			Changes:      changes,
			Message:      "No changes detected",
			UpdatedAt:    time.Now(),
		}, nil
	}

	// For changes that require container recreation, stop and remove the old one
	if len(changes) > 0 {
		timeout := 10 // seconds
		if err := p.client.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout}); err != nil {
			p.logger.Warn("failed to stop container during update", zap.String("container_id", containerID), zap.Error(err))
		}

		if err := p.client.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true}); err != nil {
			p.logger.Error("failed to remove container during update", zap.String("container_id", containerID), zap.Error(err))
			return nil, fmt.Errorf("failed to remove container: %w", err)
		}

		// Re-provision with new spec
		_, err := p.provisionInternal(ctx, spec)
		if err != nil {
			return nil, err
		}

		return &compute.UpdateResult{
			TenantID:     tenantID,
			ProviderType: "docker",
			Status:       compute.UpdateStatusSuccess,
			Changes:      changes,
			Message:      "Container updated successfully",
			UpdatedAt:    time.Now(),
		}, nil
	}

	return &compute.UpdateResult{
		TenantID:     tenantID,
		ProviderType: "docker",
		Status:       compute.UpdateStatusNoChanges,
		Changes:      changes,
		Message:      "Update completed",
		UpdatedAt:    time.Now(),
	}, nil
}

// Destroy removes a tenant's container
func (p *Provider) Destroy(ctx context.Context, tenantID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	containerID, exists := p.tenantContainers[tenantID]
	if !exists {
		// Idempotent - don't error if already gone
		return nil
	}

	// Stop the container
	stopCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	timeout := 10 // seconds

	if err := p.client.ContainerStop(stopCtx, containerID, container.StopOptions{Timeout: &timeout}); err != nil {
		// Log but continue with removal
		p.logger.Warn("failed to stop container", zap.String("container_id", containerID), zap.Error(err))
	}

	// Remove the container
	if err := p.client.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true}); err != nil {
		p.logger.Error("failed to remove container", zap.String("container_id", containerID), zap.Error(err))
		return fmt.Errorf("failed to remove container: %w", err)
	}

	delete(p.tenantContainers, tenantID)
	delete(p.tenantSpecs, tenantID)

	p.logger.Info("container destroyed", zap.String("tenant_id", tenantID), zap.String("container_id", containerID))
	return nil
}

// GetStatus returns the current status of a tenant's container
func (p *Provider) GetStatus(ctx context.Context, tenantID string) (*compute.ComputeStatus, error) {
	p.mu.RLock()
	containerID, exists := p.tenantContainers[tenantID]
	p.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("%w: %s", compute.ErrTenantNotFound, tenantID)
	}

	inspectResp, err := p.client.ContainerInspect(ctx, containerID)
	if err != nil {
		p.logger.Error("failed to inspect container", zap.String("container_id", containerID), zap.Error(err))
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	return buildComputeStatus(tenantID, &inspectResp), nil
}

// Validate validates a compute spec for Docker
func (p *Provider) Validate(ctx context.Context, spec *compute.TenantComputeSpec) error {
	// Docker provider expects exactly one container
	if len(spec.Containers) != 1 {
		return fmt.Errorf("docker provider requires exactly 1 container, got %d", len(spec.Containers))
	}

	containerSpec := spec.Containers[0]

	// Validate container image
	if containerSpec.Image == "" {
		return fmt.Errorf("container image cannot be empty")
	}

	// Basic image format validation
	if !isValidImageRef(containerSpec.Image) {
		return fmt.Errorf("invalid image reference: %s", containerSpec.Image)
	}

	return nil
}

// Close closes the Docker client connection
func (p *Provider) Close() error {
	return p.client.Close()
}

// Internal helper methods

func (p *Provider) provisionInternal(ctx context.Context, spec *compute.TenantComputeSpec) (*compute.ProvisionResult, error) {
	if len(spec.Containers) != 1 {
		return nil, fmt.Errorf("docker provider expects exactly 1 container, got %d", len(spec.Containers))
	}

	parsedConfig, err := applyProviderConfig(spec)
	if err != nil {
		return nil, err
	}

	containerSpec := spec.Containers[0]

	containerConfig := &container.Config{
		Image: containerSpec.Image,
		Env:   convertEnv(containerSpec.Env),
	}
	labels := buildContainerLabels(spec, parsedConfig)
	if len(labels) > 0 {
		containerConfig.Labels = labels
	}

	if len(containerSpec.Command) > 0 {
		containerConfig.Cmd = containerSpec.Command
	}
	if len(containerSpec.Args) > 0 {
		if containerConfig.Cmd == nil {
			containerConfig.Cmd = containerSpec.Args
		} else {
			containerConfig.Cmd = append(containerConfig.Cmd, containerSpec.Args...)
		}
	}

	portMap := nat.PortMap{}
	for _, port := range containerSpec.Ports {
		natPort, err := nat.NewPort(port.Protocol, fmt.Sprintf("%d", port.ContainerPort))
		if err != nil {
			p.logger.Error("invalid port specification", zap.Int("port", port.ContainerPort), zap.Error(err))
			continue
		}
		portMap[natPort] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: fmt.Sprintf("%d", port.HostPort),
			},
		}
	}

	hostConfig := &container.HostConfig{
		PortBindings: portMap,
		RestartPolicy: container.RestartPolicy{
			Name:              "unless-stopped",
			MaximumRetryCount: 0,
		},
	}
	if parsedConfig != nil {
		if len(parsedConfig.Volumes) > 0 {
			hostConfig.Binds = parsedConfig.Volumes
		}
		if parsedConfig.NetworkMode != "" {
			hostConfig.NetworkMode = container.NetworkMode(parsedConfig.NetworkMode)
		}
		if parsedConfig.RestartPolicy != "" {
			hostConfig.RestartPolicy = container.RestartPolicy{
				Name:              container.RestartPolicyMode(parsedConfig.RestartPolicy),
				MaximumRetryCount: 0,
			}
		}
	}

	if spec.Resources.CPU > 0 {
		hostConfig.CPUQuota = int64(spec.Resources.CPU)
	}
	if spec.Resources.Memory > 0 {
		hostConfig.Memory = int64(spec.Resources.Memory * 1024 * 1024)
	}

	containerName := fmt.Sprintf("%s-tenant-%s", defaultLabelPrefix, spec.TenantID)
	resp, err := p.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, containerName)
	if err != nil {
		p.logger.Error("failed to create container", zap.String("tenant_id", spec.TenantID), zap.Error(err))
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	containerID := resp.ID

	if err := p.client.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		p.logger.Error("failed to start container", zap.String("container_id", containerID), zap.Error(err))
		p.client.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	p.tenantContainers[spec.TenantID] = containerID
	p.tenantSpecs[spec.TenantID] = spec

	inspectResp, err := p.client.ContainerInspect(ctx, containerID)
	if err != nil {
		p.logger.Error("failed to inspect container", zap.String("container_id", containerID), zap.Error(err))
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	endpoints := buildEndpoints(&containerSpec, &inspectResp)

	p.logger.Info("container provisioned", zap.String("tenant_id", spec.TenantID), zap.String("container_id", containerID))

	return &compute.ProvisionResult{
		TenantID:      spec.TenantID,
		ProviderType:  "docker",
		Status:        compute.ProvisionStatusSuccess,
		ResourceIDs:   map[string]string{"container_id": containerID},
		Endpoints:     endpoints,
		Message:       "Container provisioned successfully",
		ProvisionedAt: time.Now(),
	}, nil
}

func convertEnv(envMap map[string]string) []string {
	env := make([]string, 0, len(envMap))
	for k, v := range envMap {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	return env
}

func buildContainerLabels(spec *compute.TenantComputeSpec, parsedConfig *DockerComputeConfig) map[string]string {
	var providerLabels map[string]string
	if parsedConfig != nil {
		providerLabels = parsedConfig.Labels
	}

	if spec == nil {
		return compute.MergeLabels(providerLabels)
	}

	return compute.MergeLabels(spec.Labels, providerLabels, compute.DefaultMetadata(spec))
}

func buildEndpoints(containerSpec *compute.ContainerSpec, inspect *types.ContainerJSON) []compute.Endpoint {
	endpoints := []compute.Endpoint{}

	// Get the container's IP address (from first network)
	var containerIP string
	if inspect.NetworkSettings != nil && len(inspect.NetworkSettings.Networks) > 0 {
		// Get first network's IP
		for _, netSettings := range inspect.NetworkSettings.Networks {
			if netSettings.IPAddress != "" {
				containerIP = netSettings.IPAddress
				break
			}
		}
	}
	if containerIP == "" {
		containerIP = "127.0.0.1"
	}

	for _, port := range containerSpec.Ports {
		endpoint := compute.Endpoint{
			Type:    "tcp",
			Address: containerIP,
			Port:    port.ContainerPort,
			URL:     fmt.Sprintf("tcp://%s:%d", containerIP, port.ContainerPort),
		}

		// If host port is specified, also provide host endpoint
		if port.HostPort > 0 {
			endpoint.URL = fmt.Sprintf("tcp://localhost:%d", port.HostPort)
		}

		if port.Name != "" {
			endpoint.Type = port.Name
		}

		endpoints = append(endpoints, endpoint)
	}

	return endpoints
}

func buildComputeStatus(tenantID string, inspect *types.ContainerJSON) *compute.ComputeStatus {
	containerStatus := compute.ContainerStatus{
		Name:    inspect.Name,
		Ready:   inspect.State.Running,
		Message: string(inspect.State.Status),
	}

	if inspect.State.Running {
		containerStatus.State = "running"
	} else if inspect.State.Status == "exited" || inspect.State.Status == "dead" {
		containerStatus.State = "stopped"
	} else {
		containerStatus.State = "unknown"
	}

	if !inspect.State.Running {
		containerStatus.RestartCount = inspect.RestartCount
	}

	state := compute.ComputeStateUnknown
	if inspect.State.Running {
		state = compute.ComputeStateRunning
	} else if inspect.State.Status == "exited" || inspect.State.Status == "dead" {
		state = compute.ComputeStateStopped
	}

	health := compute.HealthStatusUnknown
	if inspect.State.Running {
		health = compute.HealthStatusHealthy
	}

	return &compute.ComputeStatus{
		TenantID:     tenantID,
		ProviderType: "docker",
		State:        state,
		Containers:   []compute.ContainerStatus{containerStatus},
		Health:       health,
		LastUpdated:  time.Now(),
		Metadata: map[string]string{
			"container_id": inspect.ID,
			"image":        inspect.Config.Image,
		},
	}
}

func mapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func portMappingsEqual(a, b []compute.PortMapping) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].ContainerPort != b[i].ContainerPort ||
			a[i].HostPort != b[i].HostPort ||
			a[i].Protocol != b[i].Protocol {
			return false
		}
	}
	return true
}

func isValidImageRef(image string) bool {
	// Basic validation: image should not be empty and should be a reasonable format
	// Docker image refs can be:
	// - name (e.g., "alpine")
	// - name:tag (e.g., "alpine:3.18")
	// - registry/name (e.g., "docker.io/alpine")
	// - registry/name:tag (e.g., "docker.io/alpine:3.18")
	return len(image) > 0 && !strings.HasPrefix(image, "-") && !strings.HasPrefix(image, ":")
}

// DockerComputeConfig represents Docker-specific tenant configuration
type DockerComputeConfig struct {
	// Env is environment variables for the container
	Env map[string]string `json:"env,omitempty"`

	// Volumes defines volume mounts (format: "host_path:container_path" or "host_path:container_path:mode")
	Volumes []string `json:"volumes,omitempty"`

	// NetworkMode specifies the network mode (bridge, host, none, container:<name|id>)
	NetworkMode string `json:"network_mode,omitempty"`

	// Ports defines port mappings
	Ports []PortConfig `json:"ports,omitempty"`

	// RestartPolicy defines restart behavior (no, always, on-failure, unless-stopped)
	RestartPolicy string `json:"restart_policy,omitempty"`

	// Labels are Docker container labels
	Labels map[string]string `json:"labels,omitempty"`
}

// PortConfig represents a port mapping configuration
type PortConfig struct {
	ContainerPort int    `json:"container_port"`
	HostPort      int    `json:"host_port,omitempty"`
	Protocol      string `json:"protocol,omitempty"` // tcp or udp
}

func applyProviderConfig(spec *compute.TenantComputeSpec) (*DockerComputeConfig, error) {
	if spec == nil || len(spec.ProviderConfig) == 0 {
		return nil, nil
	}
	if len(spec.Containers) != 1 {
		return nil, fmt.Errorf("docker provider expects exactly 1 container, got %d", len(spec.Containers))
	}

	var dockerConfig DockerComputeConfig
	if err := json.Unmarshal(spec.ProviderConfig, &dockerConfig); err != nil {
		return nil, fmt.Errorf("invalid provider config: %w", err)
	}

	containerSpec := &spec.Containers[0]
	if len(dockerConfig.Env) > 0 {
		containerSpec.Env = dockerConfig.Env
	}
	if len(dockerConfig.Ports) > 0 {
		containerSpec.Ports = toPortMappings(dockerConfig.Ports)
	}

	return &dockerConfig, nil
}

func toPortMappings(ports []PortConfig) []compute.PortMapping {
	if len(ports) == 0 {
		return nil
	}
	mappings := make([]compute.PortMapping, 0, len(ports))
	for _, port := range ports {
		protocol := port.Protocol
		if protocol == "" {
			protocol = "tcp"
		}
		mappings = append(mappings, compute.PortMapping{
			ContainerPort: port.ContainerPort,
			HostPort:      port.HostPort,
			Protocol:      protocol,
		})
	}
	return mappings
}

var dockerConfigSchema = json.RawMessage(`{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "properties": {
    "env": {
      "type": "object",
      "additionalProperties": { "type": "string" }
    },
    "volumes": {
      "type": "array",
      "items": { "type": "string" }
    },
    "network_mode": { "type": "string" },
    "ports": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "container_port": { "type": "integer", "minimum": 1, "maximum": 65535 },
          "host_port": { "type": "integer", "minimum": 1, "maximum": 65535 },
          "protocol": { "type": "string", "enum": ["tcp", "udp"] }
        },
        "required": ["container_port"],
        "additionalProperties": false
      }
    },
    "restart_policy": { "type": "string", "enum": ["no", "always", "on-failure", "unless-stopped"] },
    "labels": {
      "type": "object",
      "additionalProperties": { "type": "string" }
    }
  },
  "additionalProperties": true
}`)

// ValidateConfig validates Docker-specific configuration
func (p *Provider) ValidateConfig(config json.RawMessage) error {
	if len(config) == 0 {
		// Empty config is valid - will use defaults
		return nil
	}

	var dockerConfig DockerComputeConfig
	if err := json.Unmarshal(config, &dockerConfig); err != nil {
		return fmt.Errorf("invalid JSON structure: %w", err)
	}

	var errors []string

	// Validate volumes format
	for i, vol := range dockerConfig.Volumes {
		parts := strings.Split(vol, ":")
		if len(parts) < 2 || len(parts) > 3 {
			errors = append(errors, fmt.Sprintf("volumes[%d]: invalid format '%s', expected 'host_path:container_path' or 'host_path:container_path:mode'", i, vol))
		}
		// Check container path is absolute
		if len(parts) >= 2 && !strings.HasPrefix(parts[1], "/") {
			errors = append(errors, fmt.Sprintf("volumes[%d]: container path must be absolute, got '%s'", i, parts[1]))
		}
	}

	// Validate network mode
	if dockerConfig.NetworkMode != "" {
		validModes := []string{"bridge", "host", "none"}
		isValid := false
		for _, mode := range validModes {
			if dockerConfig.NetworkMode == mode {
				isValid = true
				break
			}
		}
		// Also allow container:<name> format
		if !isValid && !strings.HasPrefix(dockerConfig.NetworkMode, "container:") {
			errors = append(errors, fmt.Sprintf("network_mode: invalid value '%s', must be one of: bridge, host, none, or container:<name>", dockerConfig.NetworkMode))
		}
	}

	// Validate ports
	for i, port := range dockerConfig.Ports {
		if port.ContainerPort < 1 || port.ContainerPort > 65535 {
			errors = append(errors, fmt.Sprintf("ports[%d].container_port: must be between 1 and 65535, got %d", i, port.ContainerPort))
		}
		if port.HostPort != 0 && (port.HostPort < 1 || port.HostPort > 65535) {
			errors = append(errors, fmt.Sprintf("ports[%d].host_port: must be between 1 and 65535, got %d", i, port.HostPort))
		}
		if port.Protocol != "" && port.Protocol != "tcp" && port.Protocol != "udp" {
			errors = append(errors, fmt.Sprintf("ports[%d].protocol: must be 'tcp' or 'udp', got '%s'", i, port.Protocol))
		}
	}

	// Validate restart policy
	if dockerConfig.RestartPolicy != "" {
		validPolicies := []string{"no", "always", "on-failure", "unless-stopped"}
		isValid := false
		for _, policy := range validPolicies {
			if dockerConfig.RestartPolicy == policy {
				isValid = true
				break
			}
		}
		if !isValid {
			errors = append(errors, fmt.Sprintf("restart_policy: invalid value '%s', must be one of: no, always, on-failure, unless-stopped", dockerConfig.RestartPolicy))
		}
	}

	// Validate environment variable names
	for key := range dockerConfig.Env {
		if key == "" {
			errors = append(errors, "env: environment variable name cannot be empty")
		}
		if strings.Contains(key, "=") {
			errors = append(errors, fmt.Sprintf("env: environment variable name cannot contain '=', got '%s'", key))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("Docker configuration validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// ConfigSchema returns the JSON Schema for Docker compute_config.
func (p *Provider) ConfigSchema() json.RawMessage {
	return dockerConfigSchema
}

// ConfigDefaults returns defaults for Docker compute_config (none defined).
func (p *Provider) ConfigDefaults() json.RawMessage {
	return nil
}
