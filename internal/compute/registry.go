package compute

import (
	"encoding/json"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// Registry manages registered compute providers
type Registry struct {
	providers map[string]Provider
	mu        sync.RWMutex
	logger    *zap.Logger
}

// NewRegistry creates a new provider registry
func NewRegistry(logger *zap.Logger) *Registry {
	return &Registry{
		providers: make(map[string]Provider),
		logger:    logger.With(zap.String("component", "compute-registry")),
	}
}

// Register adds a provider to the registry
// Returns ErrProviderConflict if provider name already registered
func (r *Registry) Register(provider Provider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := provider.Name()
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("%w: %s", ErrProviderConflict, name)
	}

	r.providers[name] = provider
	r.logger.Info("registered compute provider", zap.String("provider", name))
	return nil
}

// Get retrieves a provider by name
// Returns ErrProviderNotFound if provider not registered
func (r *Registry) Get(providerType string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[providerType]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrProviderNotFound, providerType)
	}

	return provider, nil
}

// List returns names of all registered providers
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// Has checks if a provider is registered
func (r *Registry) Has(providerType string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.providers[providerType]
	return exists
}

// GetProviderSchema returns the config schema and defaults for a provider.
func (r *Registry) GetProviderSchema(providerType string) (schema json.RawMessage, defaults json.RawMessage, err error) {
	provider, err := r.Get(providerType)
	if err != nil {
		return nil, nil, err
	}
	schema = provider.ConfigSchema()
	defaults = provider.ConfigDefaults()
	if len(schema) == 0 {
		return nil, nil, fmt.Errorf("compute provider schema not available: %s", providerType)
	}
	return schema, defaults, nil
}
