package compute

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"go.uber.org/zap"
)

type mockProvider struct {
	name               string
	provisionFunc      func(ctx context.Context, spec *TenantComputeSpec) (*ProvisionResult, error)
	updateFunc         func(ctx context.Context, tenantID string, spec *TenantComputeSpec) (*UpdateResult, error)
	destroyFunc        func(ctx context.Context, tenantID string) error
	statusFunc         func(ctx context.Context, tenantID string) (*ComputeStatus, error)
	validateFunc       func(ctx context.Context, spec *TenantComputeSpec) error
	validateConfigFunc func(config json.RawMessage) error
	schema             json.RawMessage
	defaults           json.RawMessage
}

func (m *mockProvider) Name() string { return m.name }

func (m *mockProvider) Provision(ctx context.Context, spec *TenantComputeSpec) (*ProvisionResult, error) {
	if m.provisionFunc != nil {
		return m.provisionFunc(ctx, spec)
	}
	return &ProvisionResult{
		Status: ProvisionStatusSuccess,
	}, nil
}

func (m *mockProvider) Update(ctx context.Context, tenantID string, spec *TenantComputeSpec) (*UpdateResult, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, tenantID, spec)
	}
	return &UpdateResult{
		Status: UpdateStatusSuccess,
	}, nil
}

func (m *mockProvider) Destroy(ctx context.Context, tenantID string) error {
	if m.destroyFunc != nil {
		return m.destroyFunc(ctx, tenantID)
	}
	return nil
}

func (m *mockProvider) GetStatus(ctx context.Context, tenantID string) (*ComputeStatus, error) {
	if m.statusFunc != nil {
		return m.statusFunc(ctx, tenantID)
	}
	return &ComputeStatus{
		State: ComputeStateRunning,
	}, nil
}

func (m *mockProvider) Validate(ctx context.Context, spec *TenantComputeSpec) error {
	if m.validateFunc != nil {
		return m.validateFunc(ctx, spec)
	}
	return nil
}

func (m *mockProvider) ValidateConfig(config json.RawMessage) error {
	if m.validateConfigFunc != nil {
		return m.validateConfigFunc(config)
	}
	return nil
}

func (m *mockProvider) ConfigSchema() json.RawMessage {
	if m.schema != nil {
		return m.schema
	}
	return json.RawMessage(`{}`)
}

func (m *mockProvider) ConfigDefaults() json.RawMessage {
	return m.defaults
}

func TestManagerProvisionTenant(t *testing.T) {
	registry := NewRegistry(zap.NewNop())
	provider := &mockProvider{name: "test"}
	registry.Register(provider)

	manager := New(registry, zap.NewNop())

	spec := &TenantComputeSpec{
		TenantID:     "tenant-123",
		ProviderType: "test",
		Containers: []ContainerSpec{
			{
				Name:  "app",
				Image: "nginx:latest",
			},
		},
		Resources: ResourceRequirements{
			CPU:    256,
			Memory: 512,
		},
	}

	result, err := manager.ProvisionTenant(context.Background(), spec)
	if err != nil {
		t.Fatalf("ProvisionTenant failed: %v", err)
	}

	if result.Status != ProvisionStatusSuccess {
		t.Errorf("expected status %s, got %s", ProvisionStatusSuccess, result.Status)
	}
}

func TestManagerProvisionTenantAppliesDefaultMetadata(t *testing.T) {
	registry := NewRegistry(zap.NewNop())
	var captured *TenantComputeSpec
	provider := &mockProvider{
		name: "test",
		provisionFunc: func(ctx context.Context, spec *TenantComputeSpec) (*ProvisionResult, error) {
			captured = spec
			return &ProvisionResult{Status: ProvisionStatusSuccess}, nil
		},
	}
	registry.Register(provider)

	manager := New(registry, zap.NewNop())

	spec := &TenantComputeSpec{
		TenantID:     "tenant-456",
		ProviderType: "test",
		Containers: []ContainerSpec{
			{
				Name:  "app",
				Image: "nginx:latest",
			},
		},
		Resources: ResourceRequirements{
			CPU:    256,
			Memory: 512,
		},
	}

	_, err := manager.ProvisionTenant(context.Background(), spec)
	if err != nil {
		t.Fatalf("ProvisionTenant failed: %v", err)
	}

	if captured == nil || captured.Labels == nil {
		t.Fatalf("expected labels to be applied")
	}
	if captured.Labels[MetadataOwnerKey] != MetadataOwnerValue {
		t.Fatalf("expected owner label to be %q", MetadataOwnerValue)
	}
	if captured.Labels[MetadataTenantIDKey] != "tenant-456" {
		t.Fatalf("expected tenant id label to be tenant-456")
	}
	if captured.Labels[MetadataProviderKey] != "test" {
		t.Fatalf("expected provider label to be test")
	}
}

func TestManagerProvisionTenant_InvalidSpec(t *testing.T) {
	registry := NewRegistry(zap.NewNop())
	provider := &mockProvider{name: "test"}
	registry.Register(provider)

	manager := New(registry, zap.NewNop())

	spec := &TenantComputeSpec{
		TenantID:     "invalid tenant",
		ProviderType: "test",
	}

	_, err := manager.ProvisionTenant(context.Background(), spec)
	if err == nil {
		t.Fatal("expected error for invalid spec")
	}

	if !errors.Is(err, ErrInvalidSpec) {
		t.Errorf("expected ErrInvalidSpec, got %v", err)
	}
}

func TestManagerProvisionTenant_ProviderNotFound(t *testing.T) {
	registry := NewRegistry(zap.NewNop())
	manager := New(registry, zap.NewNop())

	spec := &TenantComputeSpec{
		TenantID:     "tenant-123",
		ProviderType: "nonexistent",
		Containers: []ContainerSpec{
			{
				Name:  "app",
				Image: "nginx:latest",
			},
		},
		Resources: ResourceRequirements{
			CPU:    256,
			Memory: 512,
		},
	}

	_, err := manager.ProvisionTenant(context.Background(), spec)
	if err == nil {
		t.Fatal("expected error for missing provider")
	}

	if !errors.Is(err, ErrProviderNotFound) {
		t.Errorf("expected ErrProviderNotFound, got %v", err)
	}
}

func TestManagerUpdateTenant(t *testing.T) {
	registry := NewRegistry(zap.NewNop())
	provider := &mockProvider{name: "test"}
	registry.Register(provider)

	manager := New(registry, zap.NewNop())

	spec := &TenantComputeSpec{
		TenantID:     "tenant-123",
		ProviderType: "test",
		Containers: []ContainerSpec{
			{
				Name:  "app",
				Image: "nginx:2.0",
			},
		},
		Resources: ResourceRequirements{
			CPU:    256,
			Memory: 512,
		},
	}

	result, err := manager.UpdateTenant(context.Background(), "tenant-123", spec)
	if err != nil {
		t.Fatalf("UpdateTenant failed: %v", err)
	}

	if result.Status != UpdateStatusSuccess {
		t.Errorf("expected status %s, got %s", UpdateStatusSuccess, result.Status)
	}
}

func TestManagerDestroyTenant(t *testing.T) {
	registry := NewRegistry(zap.NewNop())
	provider := &mockProvider{name: "test"}
	registry.Register(provider)

	manager := New(registry, zap.NewNop())

	err := manager.DestroyTenant(context.Background(), "tenant-123", "test")
	if err != nil {
		t.Fatalf("DestroyTenant failed: %v", err)
	}
}

func TestManagerGetTenantStatus(t *testing.T) {
	registry := NewRegistry(zap.NewNop())
	provider := &mockProvider{name: "test"}
	registry.Register(provider)

	manager := New(registry, zap.NewNop())

	status, err := manager.GetTenantStatus(context.Background(), "tenant-123", "test")
	if err != nil {
		t.Fatalf("GetTenantStatus failed: %v", err)
	}

	if status.State != ComputeStateRunning {
		t.Errorf("expected state %s, got %s", ComputeStateRunning, status.State)
	}
}

func TestManagerValidateTenantSpec(t *testing.T) {
	registry := NewRegistry(zap.NewNop())
	provider := &mockProvider{name: "test"}
	registry.Register(provider)

	manager := New(registry, zap.NewNop())

	spec := &TenantComputeSpec{
		TenantID:     "tenant-123",
		ProviderType: "test",
		Containers: []ContainerSpec{
			{
				Name:  "app",
				Image: "nginx:latest",
			},
		},
		Resources: ResourceRequirements{
			CPU:    256,
			Memory: 512,
		},
	}

	err := manager.ValidateTenantSpec(context.Background(), spec)
	if err != nil {
		t.Fatalf("ValidateTenantSpec failed: %v", err)
	}
}

func TestManagerListProviders(t *testing.T) {
	registry := NewRegistry(zap.NewNop())
	registry.Register(&mockProvider{name: "ecs"})
	registry.Register(&mockProvider{name: "kubernetes"})

	manager := New(registry, zap.NewNop())

	providers := manager.ListProviders()
	if len(providers) != 2 {
		t.Errorf("expected 2 providers, got %d", len(providers))
	}
}
