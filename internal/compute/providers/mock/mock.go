package mock

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/jaxxstorm/landlord/internal/compute"
)

// Provider is an in-memory mock provider for testing
type Provider struct {
	mu      sync.RWMutex
	tenants map[string]*tenantState
}

type tenantState struct {
	Spec          *compute.TenantComputeSpec
	ProvisionedAt time.Time
	UpdatedAt     time.Time
}

// New creates a new mock provider
func New() *Provider {
	return &Provider{
		tenants: make(map[string]*tenantState),
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "mock"
}

// Provision creates a new tenant in memory
func (p *Provider) Provision(ctx context.Context, spec *compute.TenantComputeSpec) (*compute.ProvisionResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.tenants[spec.TenantID]; exists {
		return nil, fmt.Errorf("tenant %s already exists", spec.TenantID)
	}

	p.tenants[spec.TenantID] = &tenantState{
		Spec:          spec,
		ProvisionedAt: time.Now(),
	}

	endpoints := make([]compute.Endpoint, 0, len(spec.Containers))
	for _, container := range spec.Containers {
		if len(container.Ports) > 0 {
			endpoints = append(endpoints, compute.Endpoint{
				Type:    "http",
				Address: "mock.local",
				Port:    container.Ports[0].ContainerPort,
				URL:     fmt.Sprintf("http://mock.local:%d", container.Ports[0].ContainerPort),
			})
		}
	}

	return &compute.ProvisionResult{
		TenantID:      spec.TenantID,
		ProviderType:  "mock",
		Status:        compute.ProvisionStatusSuccess,
		ResourceIDs:   map[string]string{"tenant": spec.TenantID},
		Endpoints:     endpoints,
		Message:       "Mock provisioning successful",
		ProvisionedAt: time.Now(),
	}, nil
}

// Update modifies an existing tenant
func (p *Provider) Update(ctx context.Context, tenantID string, spec *compute.TenantComputeSpec) (*compute.UpdateResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	state, exists := p.tenants[tenantID]
	if !exists {
		return nil, fmt.Errorf("%w: %s", compute.ErrTenantNotFound, tenantID)
	}

	changes := []string{}
	if len(state.Spec.Containers) != len(spec.Containers) {
		changes = append(changes, "container count changed")
	}
	if state.Spec.Resources.CPU != spec.Resources.CPU {
		changes = append(changes, "CPU allocation changed")
	}
	if state.Spec.Resources.Memory != spec.Resources.Memory {
		changes = append(changes, "memory allocation changed")
	}

	state.Spec = spec
	state.UpdatedAt = time.Now()

	status := compute.UpdateStatusSuccess
	if len(changes) == 0 {
		status = compute.UpdateStatusNoChanges
	}

	return &compute.UpdateResult{
		TenantID:     tenantID,
		ProviderType: "mock",
		Status:       status,
		Changes:      changes,
		Message:      "Mock update successful",
		UpdatedAt:    time.Now(),
	}, nil
}

// Destroy removes a tenant
func (p *Provider) Destroy(ctx context.Context, tenantID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.tenants[tenantID]; !exists {
		return fmt.Errorf("%w: %s", compute.ErrTenantNotFound, tenantID)
	}

	delete(p.tenants, tenantID)
	return nil
}

// GetStatus returns current status of a tenant
func (p *Provider) GetStatus(ctx context.Context, tenantID string) (*compute.ComputeStatus, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	state, exists := p.tenants[tenantID]
	if !exists {
		return nil, fmt.Errorf("%w: %s", compute.ErrTenantNotFound, tenantID)
	}

	containers := make([]compute.ContainerStatus, 0, len(state.Spec.Containers))
	for _, c := range state.Spec.Containers {
		containers = append(containers, compute.ContainerStatus{
			Name:         c.Name,
			State:        "running",
			Ready:        true,
			RestartCount: 0,
			Message:      "Mock container running",
		})
	}

	return &compute.ComputeStatus{
		TenantID:     tenantID,
		ProviderType: "mock",
		State:        compute.ComputeStateRunning,
		Containers:   containers,
		Health:       compute.HealthStatusHealthy,
		LastUpdated:  time.Now(),
		Metadata:     map[string]string{"mock": "true"},
	}, nil
}

// Validate performs provider-specific validation
func (p *Provider) Validate(ctx context.Context, spec *compute.TenantComputeSpec) error {
	// Mock provider accepts all valid specs
	return nil
}

// ValidateConfig validates provider-specific configuration
// Mock provider accepts any configuration
func (p *Provider) ValidateConfig(config json.RawMessage) error {
	// Mock provider accepts all configurations
	return nil
}

// ConfigSchema returns an empty schema for mock provider.
func (p *Provider) ConfigSchema() json.RawMessage {
	return json.RawMessage(`{}`)
}

// ConfigDefaults returns no defaults for mock provider.
func (p *Provider) ConfigDefaults() json.RawMessage {
	return nil
}
