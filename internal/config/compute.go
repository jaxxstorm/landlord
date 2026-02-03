package config

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ComputeConfig holds compute provisioning configuration
type ComputeConfig struct {
	Docker  *DockerProviderConfig `mapstructure:"docker"`
	ECS     *ECSProviderConfig    `mapstructure:"ecs"`
	Mock    *MockProviderConfig   `mapstructure:"mock"`
	Unknown map[string]interface{} `mapstructure:",remain"`
}

// DockerProviderConfig holds Docker provider configuration
type DockerProviderConfig struct {
	// Host is the Docker API endpoint (e.g., "unix:///var/run/docker.sock", "tcp://localhost:2375")
	// Defaults to Docker socket if empty. Can be overridden with DOCKER_HOST env var.
	Host string `mapstructure:"host" env:"DOCKER_HOST" default:""`

	// NetworkName is the Docker network to use for tenant containers
	// Defaults to "bridge" if empty
	NetworkName string `mapstructure:"network_name" default:"bridge"`

	// NetworkDriver is the driver for the Docker network
	// Common values: "bridge", "overlay"
	NetworkDriver string `mapstructure:"network_driver" default:"bridge"`

	// LabelPrefix is used to label containers for identification
	// Defaults to "landlord"
	LabelPrefix string `mapstructure:"label_prefix" default:"landlord"`

	// Defaults holds provider-specific compute_config defaults (e.g., image).
	Defaults map[string]interface{} `mapstructure:",remain"`
}

// ECSProviderConfig holds ECS provider configuration defaults.
type ECSProviderConfig struct {
	Defaults map[string]interface{} `mapstructure:",remain"`
}

// MockProviderConfig holds mock provider configuration defaults.
type MockProviderConfig struct {
	Defaults map[string]interface{} `mapstructure:",remain"`
}

// Validate validates compute configuration.
func (c *ComputeConfig) Validate() error {
	if len(c.Unknown) > 0 {
		legacy := []string{}
		if _, ok := c.Unknown["default_provider"]; ok {
			legacy = append(legacy, "compute.default_provider")
		}
		if _, ok := c.Unknown["defaults"]; ok {
			legacy = append(legacy, "compute.defaults")
		}
		if len(legacy) > 0 {
			return fmt.Errorf("legacy compute config keys are no longer supported: %s", strings.Join(legacy, ", "))
		}

		unknown := make([]string, 0, len(c.Unknown))
		for key := range c.Unknown {
			unknown = append(unknown, key)
		}
		return fmt.Errorf("unknown compute provider(s): %s", strings.Join(unknown, ", "))
	}

	if len(c.EnabledProviders()) == 0 {
		return fmt.Errorf("at least one compute provider must be configured")
	}

	if c.Docker != nil {
		if err := c.Docker.Validate(); err != nil {
			return fmt.Errorf("docker config: %w", err)
		}
	}
	if c.ECS != nil {
		if err := c.ECS.Validate(); err != nil {
			return fmt.Errorf("ecs config: %w", err)
		}
	}
	if c.Mock != nil {
		if err := c.Mock.Validate(); err != nil {
			return fmt.Errorf("mock config: %w", err)
		}
	}

	return nil
}

// EnabledProviders lists configured compute providers.
func (c *ComputeConfig) EnabledProviders() []string {
	providers := []string{}
	if c.Docker != nil {
		providers = append(providers, "docker")
	}
	if c.ECS != nil {
		providers = append(providers, "ecs")
	}
	if c.Mock != nil {
		providers = append(providers, "mock")
	}
	return providers
}

// DefaultProvider returns the only enabled provider when exactly one is configured.
func (c *ComputeConfig) DefaultProvider() string {
	providers := c.EnabledProviders()
	if len(providers) == 1 {
		return providers[0]
	}
	return ""
}

// Validate validates Docker configuration defaults.
func (d *DockerProviderConfig) Validate() error {
	if d == nil {
		return nil
	}
	if len(d.Defaults) == 0 {
		return fmt.Errorf("compute.docker must include default compute_config values (e.g., image)")
	}
	image, ok := d.Defaults["image"]
	if !ok {
		return fmt.Errorf("compute.docker.image is required")
	}
	imageStr, ok := image.(string)
	if !ok || strings.TrimSpace(imageStr) == "" {
		return fmt.Errorf("compute.docker.image must be a non-empty string")
	}
	return nil
}

// Validate validates ECS configuration defaults.
func (e *ECSProviderConfig) Validate() error {
	if e == nil {
		return nil
	}
	if len(e.Defaults) == 0 {
		return fmt.Errorf("compute.ecs must include default compute_config values (e.g., cluster_arn, task_definition_arn, service_name_prefix)")
	}
	cluster, ok := e.Defaults["cluster_arn"]
	if !ok {
		return fmt.Errorf("compute.ecs.cluster_arn is required")
	}
	clusterStr, ok := cluster.(string)
	if !ok || strings.TrimSpace(clusterStr) == "" {
		return fmt.Errorf("compute.ecs.cluster_arn must be a non-empty string")
	}
	taskDef, ok := e.Defaults["task_definition_arn"]
	if !ok {
		return fmt.Errorf("compute.ecs.task_definition_arn is required")
	}
	taskDefStr, ok := taskDef.(string)
	if !ok || strings.TrimSpace(taskDefStr) == "" {
		return fmt.Errorf("compute.ecs.task_definition_arn must be a non-empty string")
	}
	serviceName, hasServiceName := e.Defaults["service_name"]
	serviceNamePrefix, hasServiceNamePrefix := e.Defaults["service_name_prefix"]
	if !hasServiceName && !hasServiceNamePrefix {
		return fmt.Errorf("compute.ecs.service_name or compute.ecs.service_name_prefix is required")
	}
	if hasServiceName {
		serviceNameStr, ok := serviceName.(string)
		if !ok || strings.TrimSpace(serviceNameStr) == "" {
			return fmt.Errorf("compute.ecs.service_name must be a non-empty string")
		}
	}
	if hasServiceNamePrefix {
		serviceNamePrefixStr, ok := serviceNamePrefix.(string)
		if !ok || strings.TrimSpace(serviceNamePrefixStr) == "" {
			return fmt.Errorf("compute.ecs.service_name_prefix must be a non-empty string")
		}
	}
	return nil
}

// Validate validates mock configuration defaults.
func (m *MockProviderConfig) Validate() error {
	return nil
}

// ToProviderConfig converts DockerProviderConfig to a JSON-encoded provider config
func (d *DockerProviderConfig) ToProviderConfig() (json.RawMessage, error) {
	if d == nil {
		return json.RawMessage("{}"), nil
	}
	return json.Marshal(d)
}
