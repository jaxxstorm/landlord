package api

import (
	"fmt"

	"github.com/jaxxstorm/landlord/internal/compute"
	"github.com/jaxxstorm/landlord/internal/tenant"
)

func providerFromMaps(config map[string]interface{}, labels map[string]string, annotations map[string]string) string {
	if config != nil {
		if provider, ok := config["compute_provider"]; ok {
			if value, ok := provider.(string); ok {
				return value
			}
		}
		if provider, ok := config["compute_provider_type"]; ok {
			if value, ok := provider.(string); ok {
				return value
			}
		}
	}
	if labels != nil {
		if provider, ok := labels["compute_provider"]; ok {
			return provider
		}
	}
	if annotations != nil {
		if provider, ok := annotations["compute_provider"]; ok {
			return provider
		}
	}
	return ""
}

func (s *Server) resolveComputeProvider(config map[string]interface{}, labels map[string]string, annotations map[string]string, fallback *tenant.Tenant) (compute.Provider, string, error) {
	if s.computeRegistry == nil {
		return nil, "", fmt.Errorf("compute provider registry not configured")
	}

	providerName := providerFromMaps(config, labels, annotations)
	if providerName == "" && fallback != nil {
		providerName = providerFromMaps(fallback.DesiredConfig, fallback.Labels, fallback.Annotations)
	}
	if providerName == "" {
		providerName = s.defaultComputeProvider
	}
	if providerName == "" {
		return nil, "", fmt.Errorf("compute provider is required when multiple providers are configured")
	}

	provider, err := s.computeRegistry.Get(providerName)
	if err != nil {
		return nil, "", err
	}
	return provider, providerName, nil
}
