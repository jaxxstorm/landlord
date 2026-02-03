package restate_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jaxxstorm/landlord/internal/compute"
	computemock "github.com/jaxxstorm/landlord/internal/compute/providers/mock"
	"github.com/jaxxstorm/landlord/internal/workflow"
	"github.com/jaxxstorm/landlord/internal/workflow/providers/restate"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

type trackingProvider struct {
	name           string
	provisionCalls int
}

func (p *trackingProvider) Name() string { return p.name }

func (p *trackingProvider) Provision(ctx context.Context, spec *compute.TenantComputeSpec) (*compute.ProvisionResult, error) {
	p.provisionCalls++
	return &compute.ProvisionResult{
		TenantID:      spec.TenantID,
		ProviderType:  p.name,
		Status:        compute.ProvisionStatusSuccess,
		ResourceIDs:   map[string]string{"service": "mock"},
		ProvisionedAt: time.Now(),
	}, nil
}

func (p *trackingProvider) Update(ctx context.Context, tenantID string, spec *compute.TenantComputeSpec) (*compute.UpdateResult, error) {
	return &compute.UpdateResult{TenantID: tenantID, ProviderType: p.name, Status: compute.UpdateStatusSuccess, UpdatedAt: time.Now()}, nil
}

func (p *trackingProvider) Destroy(ctx context.Context, tenantID string) error { return nil }

func (p *trackingProvider) GetStatus(ctx context.Context, tenantID string) (*compute.ComputeStatus, error) {
	return &compute.ComputeStatus{TenantID: tenantID, ProviderType: p.name, State: compute.ComputeStateRunning, Health: compute.HealthStatusHealthy, LastUpdated: time.Now()}, nil
}

func (p *trackingProvider) Validate(ctx context.Context, spec *compute.TenantComputeSpec) error {
	return nil
}

func (p *trackingProvider) ValidateConfig(config json.RawMessage) error { return nil }

func (p *trackingProvider) ConfigSchema() json.RawMessage { return json.RawMessage(`{}`) }

func (p *trackingProvider) ConfigDefaults() json.RawMessage { return nil }

func TestTenantProvisioningUsesRequestedProvider(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	registry := compute.NewRegistry(logger)
	require.NoError(t, registry.Register(computemock.New()))

	ecsProvider := &trackingProvider{name: "ecs"}
	require.NoError(t, registry.Register(ecsProvider))

	resolver := workflow.NewCachedComputeProviderResolver(nil, nil, "", time.Minute, logger)
	service := restate.NewTenantProvisioningService(registry, "mock", resolver, logger)

	_, err := service.Execute(ctx, &restate.ProvisioningRequest{
		TenantID:        "tenant-ecs",
		Operation:       "apply",
		DesiredConfig:   map[string]interface{}{"image": "example:v1"},
		ComputeProvider: "ecs",
	})
	require.NoError(t, err)
	require.Equal(t, 1, ecsProvider.provisionCalls)
}
