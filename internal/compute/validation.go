package compute

import (
	"errors"
	"fmt"
	"regexp"
)

var (
	// tenantIDPattern validates tenant IDs
	tenantIDPattern = regexp.MustCompile(`^[a-z0-9-]+$`)

	// containerNamePattern validates container names
	containerNamePattern = regexp.MustCompile(`^[a-z0-9-]+$`)
)

// ValidateComputeSpec performs structural validation on a compute specification
func ValidateComputeSpec(spec *TenantComputeSpec) error {
	if spec == nil {
		return errors.New("spec is nil")
	}

	// Validate TenantID
	if spec.TenantID == "" {
		return errors.New("tenant_id required")
	}
	if !tenantIDPattern.MatchString(spec.TenantID) {
		return fmt.Errorf("tenant_id must match pattern ^[a-z0-9-]+$")
	}

	// Validate ProviderType
	if spec.ProviderType == "" {
		return errors.New("provider_type required")
	}

	// Validate Containers
	if len(spec.Containers) == 0 {
		return errors.New("at least one container required")
	}

	// Validate container names are unique
	names := make(map[string]bool)
	for i, c := range spec.Containers {
		if c.Name == "" {
			return fmt.Errorf("container[%d]: name required", i)
		}
		if !containerNamePattern.MatchString(c.Name) {
			return fmt.Errorf("container[%d]: name must match pattern ^[a-z0-9-]+$", i)
		}
		if names[c.Name] {
			return fmt.Errorf("duplicate container name: %s", c.Name)
		}
		names[c.Name] = true

		// Validate image
		if c.Image == "" {
			return fmt.Errorf("container[%d]: image required", i)
		}

		// Validate ports
		for j, p := range c.Ports {
			if p.ContainerPort < 1 || p.ContainerPort > 65535 {
				return fmt.Errorf("container[%d].ports[%d]: container_port must be 1-65535", i, j)
			}
			if p.Protocol != "" && p.Protocol != "tcp" && p.Protocol != "udp" {
				return fmt.Errorf("container[%d].ports[%d]: protocol must be 'tcp' or 'udp'", i, j)
			}
		}

		// Validate health check
		if c.HealthCheck != nil {
			if err := validateHealthCheck(c.HealthCheck, i); err != nil {
				return err
			}
		}
	}

	// Validate resources
	if spec.Resources.CPU < 128 {
		return errors.New("cpu must be at least 128 millicores")
	}
	if spec.Resources.Memory < 128 {
		return errors.New("memory must be at least 128 MB")
	}

	return nil
}

// validateHealthCheck validates health check configuration
func validateHealthCheck(hc *HealthCheckConfig, containerIndex int) error {
	if hc.Type != "http" && hc.Type != "tcp" && hc.Type != "exec" {
		return fmt.Errorf("container[%d].health_check: type must be 'http', 'tcp', or 'exec'", containerIndex)
	}

	if hc.IntervalSeconds < 5 {
		return fmt.Errorf("container[%d].health_check: interval_seconds must be >= 5", containerIndex)
	}

	if hc.TimeoutSeconds < 1 {
		return fmt.Errorf("container[%d].health_check: timeout_seconds must be >= 1", containerIndex)
	}

	if hc.TimeoutSeconds >= hc.IntervalSeconds {
		return fmt.Errorf("container[%d].health_check: timeout_seconds must be < interval_seconds", containerIndex)
	}

	return nil
}
