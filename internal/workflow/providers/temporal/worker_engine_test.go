package temporal

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jaxxstorm/landlord/internal/compute"
	computemock "github.com/jaxxstorm/landlord/internal/compute/providers/mock"
	"github.com/jaxxstorm/landlord/internal/config"
	"github.com/jaxxstorm/landlord/internal/workflow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

type staticResolver struct {
	provider string
	err      error
}

type simpleProvider struct {
	name string
}

func (s *simpleProvider) Name() string { return s.name }
func (s *simpleProvider) Provision(ctx context.Context, spec *compute.TenantComputeSpec) (*compute.ProvisionResult, error) {
	_ = ctx
	return &compute.ProvisionResult{
		TenantID:      spec.TenantID,
		ProviderType:  s.name,
		Status:        compute.ProvisionStatusSuccess,
		ProvisionedAt: time.Now(),
		ResourceIDs:   map[string]string{"tenant": spec.TenantID},
	}, nil
}
func (s *simpleProvider) Update(ctx context.Context, tenantID string, spec *compute.TenantComputeSpec) (*compute.UpdateResult, error) {
	_ = ctx
	_ = spec
	return &compute.UpdateResult{
		TenantID:     tenantID,
		ProviderType: s.name,
		Status:       compute.UpdateStatusSuccess,
		UpdatedAt:    time.Now(),
	}, nil
}
func (s *simpleProvider) Destroy(ctx context.Context, tenantID string) error {
	_ = ctx
	_ = tenantID
	return nil
}
func (s *simpleProvider) GetStatus(ctx context.Context, tenantID string) (*compute.ComputeStatus, error) {
	_ = ctx
	return &compute.ComputeStatus{TenantID: tenantID, ProviderType: s.name}, nil
}
func (s *simpleProvider) Validate(ctx context.Context, spec *compute.TenantComputeSpec) error {
	_ = ctx
	_ = spec
	return nil
}
func (s *simpleProvider) ValidateConfig(cfg json.RawMessage) error {
	_ = cfg
	return nil
}
func (s *simpleProvider) ConfigSchema() json.RawMessage   { return json.RawMessage(`{}`) }
func (s *simpleProvider) ConfigDefaults() json.RawMessage { return nil }

func (s *staticResolver) ResolveProvider(ctx context.Context, tenantID, tenantUUID string) (string, error) {
	_ = ctx
	_ = tenantID
	_ = tenantUUID
	if s.err != nil {
		return "", s.err
	}
	return s.provider, nil
}

func testWorkerConfig() config.TemporalConfig {
	return config.TemporalConfig{
		HostPort:      "localhost:7233",
		Namespace:     "default",
		TaskQueue:     "landlord",
		Timeout:       time.Minute,
		RetryAttempts: 3,
	}
}

func TestNewWorkerEngine(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := compute.NewRegistry(logger)
	require.NoError(t, registry.Register(computemock.New()))

	worker, err := NewWorkerEngine(testWorkerConfig(), registry, nil, logger)
	require.NoError(t, err)
	assert.Equal(t, "temporal", worker.Name())
}

func TestWorkerRegisterAndStart(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := compute.NewRegistry(logger)
	require.NoError(t, registry.Register(computemock.New()))

	worker, err := NewWorkerEngine(testWorkerConfig(), registry, nil, logger)
	require.NoError(t, err)
	require.NoError(t, worker.Register(context.Background()))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.NoError(t, worker.Start(ctx, ":0"))
}

func TestWorkerExecuteProvisionUpdateDelete(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := compute.NewRegistry(logger)
	require.NoError(t, registry.Register(computemock.New()))

	worker, err := NewWorkerEngine(testWorkerConfig(), registry, &staticResolver{provider: "mock"}, logger)
	require.NoError(t, err)

	ctx := context.Background()

	provisioned, err := worker.Execute(ctx, &workflow.ProvisionRequest{
		TenantID:      "acme",
		Operation:     "provision",
		ComputeProvider: "mock",
		DesiredConfig: map[string]interface{}{"image": "nginx:latest"},
	})
	require.NoError(t, err)
	assert.Equal(t, workflow.StateSucceeded, provisioned.State)

	updated, err := worker.Execute(ctx, &workflow.ProvisionRequest{
		TenantID:      "acme",
		Operation:     "update",
		ComputeProvider: "mock",
		DesiredConfig: map[string]interface{}{"image": "nginx:latest"},
	})
	require.NoError(t, err)
	assert.Equal(t, workflow.StateSucceeded, updated.State)

	deleted, err := worker.Execute(ctx, &workflow.ProvisionRequest{
		TenantID:      "acme",
		Operation:     "delete",
		ComputeProvider: "mock",
	})
	require.NoError(t, err)
	assert.Equal(t, workflow.StateSucceeded, deleted.State)
}

func TestWorkerResolveComputeProviderErrors(t *testing.T) {
	logger := zaptest.NewLogger(t)
	emptyRegistry := compute.NewRegistry(logger)
	worker, err := NewWorkerEngine(testWorkerConfig(), emptyRegistry, nil, logger)
	require.NoError(t, err)

	_, err = worker.Execute(context.Background(), &workflow.ProvisionRequest{
		TenantID:  "acme",
		Operation: "provision",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "compute provider not specified")

	registry := compute.NewRegistry(logger)
	require.NoError(t, registry.Register(computemock.New()))
	worker, err = NewWorkerEngine(testWorkerConfig(), registry, nil, logger)
	require.NoError(t, err)

	_, err = worker.Execute(context.Background(), &workflow.ProvisionRequest{
		TenantID:       "acme",
		Operation:      "provision",
		ComputeProvider: "does-not-exist",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "compute provider lookup failed")
}

func TestWorkerCancellationAware(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := compute.NewRegistry(logger)
	require.NoError(t, registry.Register(computemock.New()))

	worker, err := NewWorkerEngine(testWorkerConfig(), registry, nil, logger)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = worker.Execute(ctx, &workflow.ProvisionRequest{TenantID: "acme", Operation: "provision"})
	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestWorkerExecuteAcrossRegisteredComputeProviders(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := compute.NewRegistry(logger)
	require.NoError(t, registry.Register(&simpleProvider{name: "provider-a"}))
	require.NoError(t, registry.Register(&simpleProvider{name: "provider-b"}))

	worker, err := NewWorkerEngine(testWorkerConfig(), registry, nil, logger)
	require.NoError(t, err)

	for _, provider := range []string{"provider-a", "provider-b"} {
		status, err := worker.Execute(context.Background(), &workflow.ProvisionRequest{
			TenantID:        "acme-" + provider,
			Operation:       "provision",
			ComputeProvider: provider,
		})
		require.NoError(t, err)
		assert.Equal(t, workflow.StateSucceeded, status.State)
	}
}

func TestWorkerRegisterIsStableAcrossRepeats(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := compute.NewRegistry(logger)
	require.NoError(t, registry.Register(computemock.New()))

	worker, err := NewWorkerEngine(testWorkerConfig(), registry, nil, logger)
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		require.NoError(t, worker.Register(context.Background()))
	}
}
