package compute

const (
	// MetadataNamespace is the prefix for all Landlord-managed metadata.
	MetadataNamespace = "landlord"

	MetadataOwnerKey    = "landlord.owner"
	MetadataTenantIDKey = "landlord.tenant_id"
	MetadataProviderKey = "landlord.provider"

	MetadataOwnerValue = "landlord"
)

// DefaultMetadata returns the standard metadata applied to compute resources.
func DefaultMetadata(spec *TenantComputeSpec) map[string]string {
	if spec == nil {
		return nil
	}

	metadata := map[string]string{
		MetadataOwnerKey:    MetadataOwnerValue,
		MetadataTenantIDKey: spec.TenantID,
	}

	if spec.ProviderType != "" {
		metadata[MetadataProviderKey] = spec.ProviderType
	}

	return metadata
}

// MergeLabels merges label maps in order, where later maps override earlier ones.
func MergeLabels(labelSets ...map[string]string) map[string]string {
	merged := map[string]string{}

	for _, labels := range labelSets {
		if len(labels) == 0 {
			continue
		}
		for key, value := range labels {
			merged[key] = value
		}
	}

	if len(merged) == 0 {
		return nil
	}

	return merged
}

// ApplyDefaultMetadata ensures the default metadata is present on the spec.
func ApplyDefaultMetadata(spec *TenantComputeSpec) {
	if spec == nil {
		return
	}

	spec.Labels = MergeLabels(spec.Labels, DefaultMetadata(spec))
}
