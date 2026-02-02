package compute

import "testing"

func TestDefaultMetadata(t *testing.T) {
	spec := &TenantComputeSpec{
		TenantID:     "tenant-123",
		ProviderType: "docker",
	}

	metadata := DefaultMetadata(spec)
	if metadata[MetadataOwnerKey] != MetadataOwnerValue {
		t.Fatalf("expected owner %q, got %q", MetadataOwnerValue, metadata[MetadataOwnerKey])
	}
	if metadata[MetadataTenantIDKey] != "tenant-123" {
		t.Fatalf("expected tenant id tenant-123, got %q", metadata[MetadataTenantIDKey])
	}
	if metadata[MetadataProviderKey] != "docker" {
		t.Fatalf("expected provider docker, got %q", metadata[MetadataProviderKey])
	}
}

func TestMergeLabelsOverrides(t *testing.T) {
	base := map[string]string{"a": "1", "b": "2"}
	override := map[string]string{"b": "3", "c": "4"}

	merged := MergeLabels(base, override)
	if merged["a"] != "1" {
		t.Fatalf("expected a=1, got %q", merged["a"])
	}
	if merged["b"] != "3" {
		t.Fatalf("expected b=3, got %q", merged["b"])
	}
	if merged["c"] != "4" {
		t.Fatalf("expected c=4, got %q", merged["c"])
	}
}

func TestApplyDefaultMetadata(t *testing.T) {
	spec := &TenantComputeSpec{
		TenantID:     "tenant-abc",
		ProviderType: "docker",
		Labels: map[string]string{
			"custom": "value",
		},
	}

	ApplyDefaultMetadata(spec)
	if spec.Labels[MetadataOwnerKey] != MetadataOwnerValue {
		t.Fatalf("expected owner label")
	}
	if spec.Labels[MetadataTenantIDKey] != "tenant-abc" {
		t.Fatalf("expected tenant id label")
	}
	if spec.Labels["custom"] != "value" {
		t.Fatalf("expected custom label to remain")
	}
}
