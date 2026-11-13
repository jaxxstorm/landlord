//go:build integration

// Package floci starts isolated Floci instances for AWS-compatible integration tests.
package floci

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultRegion    = "us-east-1"
	defaultSecretKey = "test"
	flociPort        = "4566/tcp"
)

// Fixture provides the connection details for an isolated Floci instance.
type Fixture struct {
	Endpoint        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	ResourcePrefix  string
}

// Start launches Floci and registers container cleanup with the test.
func Start(t *testing.T) *Fixture {
	t.Helper()
	testcontainers.SkipIfProviderIsNotHealthy(t)

	ctx := context.Background()
	instance, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "floci/floci:latest",
			ExposedPorts: []string{flociPort},
			Env:          map[string]string{"FLOCI_SERVICES_EKS_ECR_REGISTRY_MIRROR": "false"},
			HostConfigModifier: func(hostConfig *container.HostConfig) {
				hostConfig.Binds = append(hostConfig.Binds, "/var/run/docker.sock:/var/run/docker.sock")
			},
			ConfigModifier: func(config *container.Config) {
				config.User = "root"
			},
			WaitingFor: wait.ForHTTP("/_localstack/health").
				WithPort(nat.Port(flociPort)).
				WithStartupTimeout(30 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		t.Skipf("skipping Floci integration test (container start failed): %s", err)
	}

	t.Cleanup(func() {
		if err := instance.Terminate(ctx); err != nil {
			t.Errorf("terminate Floci container: %s", err)
		}
	})

	host, err := instance.Host(ctx)
	if err != nil {
		t.Fatalf("get Floci container host: %s", err)
	}
	port, err := instance.MappedPort(ctx, nat.Port(flociPort))
	if err != nil {
		t.Fatalf("get Floci container port: %s", err)
	}

	return &Fixture{
		Endpoint:        fmt.Sprintf("http://%s:%s", host, port.Port()),
		Region:          defaultRegion,
		AccessKeyID:     accountID(),
		SecretAccessKey: defaultSecretKey,
		ResourcePrefix:  "landlord-" + strings.ReplaceAll(uuid.NewString(), "-", ""),
	}
}

// SetEnvironment configures the AWS SDK default chain to target this fixture.
func (f *Fixture) SetEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("AWS_ACCESS_KEY_ID", f.AccessKeyID)
	t.Setenv("AWS_SECRET_ACCESS_KEY", f.SecretAccessKey)
	t.Setenv("AWS_DEFAULT_REGION", f.Region)
	t.Setenv("AWS_REGION", f.Region)
	t.Setenv("AWS_ENDPOINT_URL", f.Endpoint)
}

func accountID() string {
	return fmt.Sprintf("%012d", time.Now().UnixNano()%900_000_000_000+100_000_000_000)
}
