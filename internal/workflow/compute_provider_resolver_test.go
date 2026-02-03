package workflow

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jaxxstorm/landlord/internal/tenant"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

type fakeTenantRepo struct {
	lookup func(ctx context.Context, name string) (*tenant.Tenant, error)
}

func (f *fakeTenantRepo) CreateTenant(ctx context.Context, t *tenant.Tenant) error {
	return nil
}

func (f *fakeTenantRepo) GetTenantByName(ctx context.Context, name string) (*tenant.Tenant, error) {
	return f.lookup(ctx, name)
}

func (f *fakeTenantRepo) GetTenantByID(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
	return nil, tenant.ErrTenantNotFound
}

func (f *fakeTenantRepo) UpdateTenant(ctx context.Context, t *tenant.Tenant) error {
	return nil
}

func (f *fakeTenantRepo) ListTenants(ctx context.Context, filters tenant.ListFilters) ([]*tenant.Tenant, error) {
	return nil, nil
}

func (f *fakeTenantRepo) ListTenantsForReconciliation(ctx context.Context) ([]*tenant.Tenant, error) {
	return nil, nil
}

func (f *fakeTenantRepo) DeleteTenant(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (f *fakeTenantRepo) RecordStateTransition(ctx context.Context, transition *tenant.StateTransition) error {
	return nil
}

func (f *fakeTenantRepo) GetStateHistory(ctx context.Context, tenantID uuid.UUID) ([]*tenant.StateTransition, error) {
	return nil, nil
}

func TestResolverUsesConfigProvider(t *testing.T) {
	logger := zaptest.NewLogger(t)
	repo := &fakeTenantRepo{
		lookup: func(ctx context.Context, name string) (*tenant.Tenant, error) {
			return &tenant.Tenant{
				Name: "tenant-a",
				DesiredConfig: map[string]interface{}{
					"compute_provider": "ecs",
				},
			}, nil
		},
	}

	resolver := NewCachedComputeProviderResolver(nil, repo, "", time.Minute, logger)
	provider, err := resolver.ResolveProvider(context.Background(), "tenant-a", "")
	require.NoError(t, err)
	require.Equal(t, "ecs", provider)
}

func TestResolverUsesLabelsWhenConfigMissing(t *testing.T) {
	logger := zaptest.NewLogger(t)
	repo := &fakeTenantRepo{
		lookup: func(ctx context.Context, name string) (*tenant.Tenant, error) {
			return &tenant.Tenant{
				Name:   "tenant-b",
				Labels: map[string]string{"compute_provider": "ecs"},
			}, nil
		},
	}

	resolver := NewCachedComputeProviderResolver(nil, repo, "", time.Minute, logger)
	provider, err := resolver.ResolveProvider(context.Background(), "tenant-b", "")
	require.NoError(t, err)
	require.Equal(t, "ecs", provider)
}

func TestResolverUsesOverride(t *testing.T) {
	logger := zaptest.NewLogger(t)
	resolver := NewCachedComputeProviderResolver(nil, nil, "ecs", time.Minute, logger)
	provider, err := resolver.ResolveProvider(context.Background(), "tenant-c", "")
	require.NoError(t, err)
	require.Equal(t, "ecs", provider)
}
