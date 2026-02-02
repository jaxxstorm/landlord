package mock

import (
	"context"
	"testing"

	"github.com/jaxxstorm/landlord/internal/compute"
)

func TestProvisionTenant(t *testing.T) {
	provider := New()

	spec := &compute.TenantComputeSpec{
		TenantID:     "test-tenant",
		ProviderType: "mock",
		Containers: []compute.ContainerSpec{
			{
				Name:  "app",
				Image: "nginx:latest",
				Ports: []compute.PortMapping{
					{ContainerPort: 80, HostPort: 8080},
				},
			},
		},
		Resources: compute.ResourceRequirements{
			CPU:    256,
			Memory: 512,
		},
	}

	result, err := provider.Provision(context.Background(), spec)
	if err != nil {
		t.Fatalf("Provision failed: %v", err)
	}

	if result.Status != compute.ProvisionStatusSuccess {
		t.Errorf("expected status %s, got %s", compute.ProvisionStatusSuccess, result.Status)
	}

	if result.TenantID != "test-tenant" {
		t.Errorf("expected tenant_id test-tenant, got %s", result.TenantID)
	}

	if len(result.Endpoints) == 0 {
		t.Error("expected at least one endpoint")
	}
}

func TestProvisionDuplicate(t *testing.T) {
	provider := New()

	spec := &compute.TenantComputeSpec{
		TenantID:     "test-tenant",
		ProviderType: "mock",
		Containers: []compute.ContainerSpec{
			{
				Name:  "app",
				Image: "nginx:latest",
			},
		},
		Resources: compute.ResourceRequirements{
			CPU:    256,
			Memory: 512,
		},
	}

	_, err := provider.Provision(context.Background(), spec)
	if err != nil {
		t.Fatalf("First provision failed: %v", err)
	}

	_, err = provider.Provision(context.Background(), spec)
	if err == nil {
		t.Fatal("expected error for duplicate tenant")
	}
}

func TestUpdateTenant(t *testing.T) {
	provider := New()

	spec := &compute.TenantComputeSpec{
		TenantID:     "test-tenant",
		ProviderType: "mock",
		Containers: []compute.ContainerSpec{
			{
				Name:  "app",
				Image: "nginx:1.0",
			},
		},
		Resources: compute.ResourceRequirements{
			CPU:    256,
			Memory: 512,
		},
	}

	_, err := provider.Provision(context.Background(), spec)
	if err != nil {
		t.Fatalf("Provision failed: %v", err)
	}

	updatedSpec := &compute.TenantComputeSpec{
		TenantID:     "test-tenant",
		ProviderType: "mock",
		Containers: []compute.ContainerSpec{
			{
				Name:  "app",
				Image: "nginx:1.0",
			},
		},
		Resources: compute.ResourceRequirements{
			CPU:    512,
			Memory: 512,
		},
	}
	result, err := provider.Update(context.Background(), "test-tenant", updatedSpec)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if result.Status != compute.UpdateStatusSuccess {
		t.Errorf("expected status %s, got %s", compute.UpdateStatusSuccess, result.Status)
	}

	if len(result.Changes) == 0 {
		t.Error("expected at least one change reported")
	}
}

func TestUpdateNonexistent(t *testing.T) {
	provider := New()

	spec := &compute.TenantComputeSpec{
		TenantID:     "nonexistent",
		ProviderType: "mock",
		Containers: []compute.ContainerSpec{
			{
				Name:  "app",
				Image: "nginx:latest",
			},
		},
		Resources: compute.ResourceRequirements{
			CPU:    256,
			Memory: 512,
		},
	}

	_, err := provider.Update(context.Background(), "nonexistent", spec)
	if err == nil {
		t.Fatal("expected error for nonexistent tenant")
	}
}

func TestDestroyTenant(t *testing.T) {
	provider := New()

	spec := &compute.TenantComputeSpec{
		TenantID:     "test-tenant",
		ProviderType: "mock",
		Containers: []compute.ContainerSpec{
			{
				Name:  "app",
				Image: "nginx:latest",
			},
		},
		Resources: compute.ResourceRequirements{
			CPU:    256,
			Memory: 512,
		},
	}

	_, err := provider.Provision(context.Background(), spec)
	if err != nil {
		t.Fatalf("Provision failed: %v", err)
	}

	err = provider.Destroy(context.Background(), "test-tenant")
	if err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}

	// Verify tenant is gone
	_, err = provider.GetStatus(context.Background(), "test-tenant")
	if err == nil {
		t.Fatal("expected error for destroyed tenant")
	}
}

func TestGetStatus(t *testing.T) {
	provider := New()

	spec := &compute.TenantComputeSpec{
		TenantID:     "test-tenant",
		ProviderType: "mock",
		Containers: []compute.ContainerSpec{
			{
				Name:  "app",
				Image: "nginx:latest",
			},
		},
		Resources: compute.ResourceRequirements{
			CPU:    256,
			Memory: 512,
		},
	}

	_, err := provider.Provision(context.Background(), spec)
	if err != nil {
		t.Fatalf("Provision failed: %v", err)
	}

	status, err := provider.GetStatus(context.Background(), "test-tenant")
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if status.State != compute.ComputeStateRunning {
		t.Errorf("expected state %s, got %s", compute.ComputeStateRunning, status.State)
	}

	if len(status.Containers) != 1 {
		t.Errorf("expected 1 container, got %d", len(status.Containers))
	}

	if status.Health != compute.HealthStatusHealthy {
		t.Errorf("expected health %s, got %s", compute.HealthStatusHealthy, status.Health)
	}
}

func TestValidate(t *testing.T) {
	provider := New()

	spec := &compute.TenantComputeSpec{
		TenantID:     "test-tenant",
		ProviderType: "mock",
		Containers: []compute.ContainerSpec{
			{
				Name:  "app",
				Image: "nginx:latest",
			},
		},
		Resources: compute.ResourceRequirements{
			CPU:    256,
			Memory: 512,
		},
	}

	err := provider.Validate(context.Background(), spec)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
}
