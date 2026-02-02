package workflow

import (
	"fmt"
	"sort"
	"sync"

	"go.uber.org/zap"
)

// Registry stores registered workflow providers
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
	logger    *zap.Logger
}

// NewRegistry creates a new provider registry
func NewRegistry(logger *zap.Logger) *Registry {
	return &Registry{
		providers: make(map[string]Provider),
		logger:    logger.With(zap.String("component", "workflow-registry")),
	}
}

// Register registers a workflow provider
func (r *Registry) Register(provider Provider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := provider.Name()
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	if _, exists := r.providers[name]; exists {
		r.logger.Error("provider already registered",
			zap.String("provider", name),
		)
		return fmt.Errorf("%w: %s", ErrProviderConflict, name)
	}

	r.providers[name] = provider
	r.logger.Info("registered workflow provider",
		zap.String("provider", name),
	)

	return nil
}

// Get retrieves a provider by name
func (r *Registry) Get(providerType string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[providerType]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrProviderNotFound, providerType)
	}

	return provider, nil
}

// List returns all registered provider names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	sort.Strings(names)

	return names
}

// Has checks if a provider is registered
func (r *Registry) Has(providerType string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.providers[providerType]
	return exists
}
