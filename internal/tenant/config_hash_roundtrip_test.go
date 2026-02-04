package tenant_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jaxxstorm/landlord/internal/tenant"
	tenantpg "github.com/jaxxstorm/landlord/internal/tenant/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap/zaptest"
)

// TestConfigHashStableAfterDBRoundtrip verifies that config hash remains
// stable after storing in DB and reading back (JSON marshal/unmarshal)
func TestConfigHashStableAfterDBRoundtrip(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testcontainers.SkipIfProviderIsNotHealthy(t)

	ctx := context.Background()

	// Start PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Skipf("skipping integration test (container start failed): %s", err)
	}
	defer container.Terminate(ctx)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := "postgres://testuser:testpass@" + host + ":" + port.Port() + "/testdb?sslmode=disable"

	// Run migrations
	migrationPath := "file://../../internal/database/migrations"
	m, err := migrate.New(migrationPath, dsn)
	require.NoError(t, err)
	require.NoError(t, m.Up())

	// Create connection pool
	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	logger := zaptest.NewLogger(t)
	repo, err := tenantpg.New(pool, logger)
	require.NoError(t, err)

	testCases := []struct {
		name   string
		config map[string]interface{}
	}{
		{
			name: "string values",
			config: map[string]interface{}{
				"image":   "myapp:v1",
				"env":     "prod",
				"region":  "us-west-2",
			},
		},
		{
			name: "integer values",
			config: map[string]interface{}{
				"replicas": 3,
				"port":     8080,
				"timeout":  30,
			},
		},
		{
			name: "mixed types",
			config: map[string]interface{}{
				"image":    "myapp:v2",
				"replicas": 5,
				"enabled":  true,
				"cpu":      1.5,
			},
		},
		{
			name: "nested objects",
			config: map[string]interface{}{
				"compute": map[string]interface{}{
					"cpu":    2,
					"memory": "4Gi",
				},
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "test",
						"env": "prod",
					},
				},
			},
		},
		{
			name: "arrays",
			config: map[string]interface{}{
				"ports": []interface{}{8080, 8443, 9090},
				"tags":  []interface{}{"api", "backend", "v1"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Compute hash of original config
			originalHash, err := tenant.ComputeConfigHash(tc.config)
			require.NoError(t, err)
			require.NotEmpty(t, originalHash, "Original hash should not be empty")

			// Create tenant with this config
			tn := &tenant.Tenant{
				ID:                  uuid.New(),
				Name:                "test-tenant-" + uuid.NewString(),
				Status:              tenant.StatusRequested,
				DesiredConfig:       tc.config,
				WorkflowConfigHash:  &originalHash,
			}
			err = repo.CreateTenant(ctx, tn)
			require.NoError(t, err)

			// Read tenant back from DB
			retrieved, err := repo.GetTenantByID(ctx, tn.ID)
			require.NoError(t, err)

			// Compute hash of config after DB roundtrip
			roundtripHash, err := tenant.ComputeConfigHash(retrieved.DesiredConfig)
			require.NoError(t, err)

			// CRITICAL: Hashes must match - if not, config change detection is broken
			assert.Equal(t, originalHash, roundtripHash, 
				"Config hash changed after DB roundtrip - this will cause false workflow restarts!\n"+
				"Original config: %+v\n"+
				"After DB: %+v",
				tc.config, retrieved.DesiredConfig)

			// Also verify the stored hash matches
			assert.Equal(t, originalHash, *retrieved.WorkflowConfigHash,
				"Stored hash should match original")
		})
	}
}
