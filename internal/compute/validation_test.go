package compute

import "testing"

func TestValidateComputeSpecAllowsProviderManagedSpec(t *testing.T) {
	spec := &TenantComputeSpec{
		TenantID:       "tenant-123",
		ProviderType:   "ec2-runner",
		ProviderConfig: []byte(`{"ami":"ami-123","user_data":"#!/bin/sh"}`),
	}

	if err := ValidateComputeSpec(spec); err != nil {
		t.Fatalf("ValidateComputeSpec() error = %v", err)
	}
}
