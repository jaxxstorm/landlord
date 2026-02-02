package restate_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"go.uber.org/zap/zaptest"

	"github.com/jaxxstorm/landlord/internal/compute"
	"github.com/jaxxstorm/landlord/internal/compute/providers/docker"
	computemock "github.com/jaxxstorm/landlord/internal/compute/providers/mock"
	"github.com/jaxxstorm/landlord/internal/config"
	"github.com/jaxxstorm/landlord/internal/workflow"
	"github.com/jaxxstorm/landlord/internal/workflow/providers/restate"
)

func TestRestateIntegrationTenantProvisioningWithDocker(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("skipping integration test: INTEGRATION_TEST not set")
	}

	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	server := newFakeRestateServer(t)
	cfg := config.RestateConfig{
		Endpoint:           server.URL(),
		ExecutionMechanism: "local",
		AuthType:           "none",
		Timeout:            2 * time.Second,
	}

	provider, err := restate.New(cfg, logger)
	requireNoError(t, err, "restate provider init")

	execResult, err := provider.StartExecution(ctx, "tenant-provisioning", &workflow.ExecutionInput{
		ExecutionName: "integration-exec",
		Input:         []byte(`{}`),
	})
	requireNoError(t, err, "start execution")
	if execResult == nil {
		t.Fatal("expected execution result")
	}

	dockerProvider, err := docker.New(&docker.Config{}, logger)
	if err != nil {
		t.Skipf("skipping docker integration: %v", err)
	}

	registry := compute.NewRegistry(logger)
	if err := registry.Register(dockerProvider); err != nil {
		t.Fatalf("failed to register docker provider: %v", err)
	}

	imageRef := "alpine:3.18"
	if err := ensureImageAvailable(ctx, t, imageRef); err != nil {
		t.Skipf("skipping docker integration (image pull failed): %v", err)
	}

	tenantID := fmt.Sprintf("integration-%d", time.Now().UnixNano())
	service := restate.NewTenantProvisioningService(registry, "docker", nil, logger)

	_, err = service.Execute(ctx, &restate.ProvisioningRequest{
		TenantID:     tenantID,
		Operation:    "plan",
		DesiredImage: imageRef,
	})
	requireNoError(t, err, "plan execution")

	_, err = service.Execute(ctx, &restate.ProvisioningRequest{
		TenantID:     tenantID,
		Operation:    "apply",
		DesiredImage: imageRef,
	})
	requireNoError(t, err, "apply execution")

	containerName := fmt.Sprintf("landlord-tenant-%s", tenantID)
	if !containerExists(ctx, t, containerName) {
		t.Fatalf("expected container %s to exist", containerName)
	}

	_, err = service.Execute(ctx, &restate.ProvisioningRequest{
		TenantID:  tenantID,
		Operation: "destroy",
	})
	requireNoError(t, err, "destroy execution")

	waitForContainerRemoval(ctx, t, containerName, 10*time.Second)
}

func TestRestateWorkerLifecycleWithRegistration(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	server := newFakeRestateServer(t)
	cfg := config.RestateConfig{
		Endpoint:                server.URL(),
		AdminEndpoint:           server.URL(),
		AuthType:                "none",
		WorkerRegisterOnStartup: true,
		WorkerAdvertisedURL:     "http://127.0.0.1:9999",
		Timeout:                 30 * time.Second,
	}

	registry := compute.NewRegistry(logger)
	mockProvider := computemock.New()
	if err := registry.Register(mockProvider); err != nil {
		t.Fatalf("failed to register compute provider: %v", err)
	}

	resolver := workflow.NewCachedComputeProviderResolver(nil, nil, "mock", 1*time.Minute, logger)
	worker, err := restate.NewWorkerEngine(cfg, registry, resolver, logger)
	if err != nil {
		t.Fatalf("failed to create worker engine: %v", err)
	}

	if err := worker.Register(ctx); err != nil {
		t.Fatalf("worker registration failed: %v", err)
	}

	service := restate.NewTenantProvisioningService(registry, "mock", resolver, logger)

	_, err = service.Execute(ctx, &restate.ProvisioningRequest{
		TenantID:     "worker-test",
		Operation:    "plan",
		DesiredImage: "example:v1",
	})
	requireNoError(t, err, "plan execution")

	_, err = service.Execute(ctx, &restate.ProvisioningRequest{
		TenantID:     "worker-test",
		Operation:    "apply",
		DesiredImage: "example:v1",
	})
	requireNoError(t, err, "apply execution")

	_, err = service.Execute(ctx, &restate.ProvisioningRequest{
		TenantID:     "worker-test",
		Operation:    "update",
		DesiredImage: "example:v1",
	})
	requireNoError(t, err, "update execution")

	_, err = service.Execute(ctx, &restate.ProvisioningRequest{
		TenantID:  "worker-test",
		Operation: "delete",
	})
	requireNoError(t, err, "delete execution")
}

func TestRestateCreateIdempotency(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	registry := compute.NewRegistry(logger)
	mockProvider := computemock.New()
	if err := registry.Register(mockProvider); err != nil {
		t.Fatalf("failed to register compute provider: %v", err)
	}

	service := restate.NewTenantProvisioningService(registry, "mock", nil, logger)

	_, err := service.Execute(ctx, &restate.ProvisioningRequest{
		TenantID:     "idempotent-create",
		Operation:    "plan",
		DesiredImage: "example:v1",
	})
	requireNoError(t, err, "plan execution")

	_, err = service.Execute(ctx, &restate.ProvisioningRequest{
		TenantID:     "idempotent-create",
		Operation:    "apply",
		DesiredImage: "example:v1",
	})
	requireNoError(t, err, "apply execution")

	// Apply again should be idempotent.
	_, err = service.Execute(ctx, &restate.ProvisioningRequest{
		TenantID:     "idempotent-create",
		Operation:    "apply",
		DesiredImage: "example:v1",
	})
	requireNoError(t, err, "apply idempotency execution")
}

func TestRestateUpdateNoChanges(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	registry := compute.NewRegistry(logger)
	mockProvider := computemock.New()
	if err := registry.Register(mockProvider); err != nil {
		t.Fatalf("failed to register compute provider: %v", err)
	}

	service := restate.NewTenantProvisioningService(registry, "mock", nil, logger)

	_, err := service.Execute(ctx, &restate.ProvisioningRequest{
		TenantID:     "update-noop",
		Operation:    "plan",
		DesiredImage: "example:v1",
	})
	requireNoError(t, err, "plan execution")

	_, err = service.Execute(ctx, &restate.ProvisioningRequest{
		TenantID:     "update-noop",
		Operation:    "apply",
		DesiredImage: "example:v1",
	})
	requireNoError(t, err, "apply execution")

	_, err = service.Execute(ctx, &restate.ProvisioningRequest{
		TenantID:     "update-noop",
		Operation:    "update",
		DesiredImage: "example:v1",
	})
	requireNoError(t, err, "update execution")
}

func containerExists(ctx context.Context, t *testing.T, containerName string) bool {
	t.Helper()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		t.Fatalf("failed to create docker client: %v", err)
	}
	defer cli.Close()

	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		t.Fatalf("failed to list containers: %v", err)
	}

	for _, c := range containers {
		for _, name := range c.Names {
			if strings.TrimPrefix(name, "/") == containerName {
				return true
			}
		}
	}

	return false
}

func waitForContainerRemoval(ctx context.Context, t *testing.T, containerName string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !containerExists(ctx, t, containerName) {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("container %s still exists after %s", containerName, timeout)
}

func ensureImageAvailable(ctx context.Context, t *testing.T, imageRef string) error {
	t.Helper()

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer cli.Close()

	images, err := cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list docker images: %w", err)
	}

	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == imageRef {
				return nil
			}
		}
	}

	reader, err := cli.ImagePull(ctx, imageRef, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", imageRef, err)
	}
	defer reader.Close()

	if _, err := io.Copy(io.Discard, reader); err != nil {
		return fmt.Errorf("failed to read image pull output: %w", err)
	}

	return nil
}

func requireNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}
