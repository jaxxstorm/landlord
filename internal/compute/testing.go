package compute

import (
	"context"
	"encoding/json"
)

// testProvider is a minimal mock for registry tests
type testProvider struct {
	name string
}

func (t *testProvider) Name() string {
	return t.name
}

func (t *testProvider) Provision(ctx context.Context, spec *TenantComputeSpec) (*ProvisionResult, error) {
	return nil, nil
}

func (t *testProvider) Update(ctx context.Context, tenantID string, spec *TenantComputeSpec) (*UpdateResult, error) {
	return nil, nil
}

func (t *testProvider) Destroy(ctx context.Context, tenantID string) error {
	return nil
}

func (t *testProvider) GetStatus(ctx context.Context, tenantID string) (*ComputeStatus, error) {
	return nil, nil
}

func (t *testProvider) Validate(ctx context.Context, spec *TenantComputeSpec) error {
	return nil
}

func (t *testProvider) ValidateConfig(config json.RawMessage) error {
	return nil
}

func (t *testProvider) ConfigSchema() json.RawMessage {
	return json.RawMessage(`{}`)
}

func (t *testProvider) ConfigDefaults() json.RawMessage {
	return nil
}
