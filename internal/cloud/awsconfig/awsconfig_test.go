package awsconfig

import (
	"context"
	"testing"
)

func TestLoadUsesRegion(t *testing.T) {
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")

	cfg, err := Load(context.Background(), Options{Region: "us-west-2"})
	if err != nil {
		t.Fatalf("expected load to succeed: %v", err)
	}
	if cfg.Region != "us-west-2" {
		t.Fatalf("expected region to be us-west-2, got %s", cfg.Region)
	}
}

func TestLoadWithAssumeRole(t *testing.T) {
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")

	cfg, err := Load(context.Background(), Options{
		Region: "us-west-2",
		AssumeRole: &AssumeRoleOptions{
			RoleARN:     "arn:aws:iam::123456789012:role/example",
			ExternalID:  "external",
			SessionName: "landlord-test",
		},
	})
	if err != nil {
		t.Fatalf("expected load with assume role to succeed: %v", err)
	}
	if cfg.Credentials == nil {
		t.Fatalf("expected credentials to be configured")
	}
}
