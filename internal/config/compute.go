package config

import (
	"encoding/json"
	"fmt"
)

// ComputeConfig holds compute provisioning configuration
type ComputeConfig struct {
	DefaultProvider string        `mapstructure:"default_provider" env:"COMPUTE_DEFAULT_PROVIDER" default:"mock"`
	Docker          *DockerConfig `mapstructure:"docker"`
}

// DockerConfig holds Docker provider configuration
type DockerConfig struct {
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
}

// Validate validates compute configuration
func (c *ComputeConfig) Validate() error {
	if c.DefaultProvider == "" {
		return fmt.Errorf("default compute provider must be specified")
	}

	if c.Docker != nil {
		if err := c.Docker.Validate(); err != nil {
			return fmt.Errorf("docker config: %w", err)
		}
	}

	return nil
}

// Validate validates Docker configuration
func (d *DockerConfig) Validate() error {
	// Docker config is valid even if sparse - defaults will be applied
	return nil
}

// ToProviderConfig converts DockerConfig to a JSON-encoded provider config
func (d *DockerConfig) ToProviderConfig() (json.RawMessage, error) {
	if d == nil {
		return json.RawMessage("{}"), nil
	}
	return json.Marshal(d)
}
