package database

import (
	"context"
	"os"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/jaxxstorm/landlord/internal/config"
)

func TestDatabaseIntegration(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Check if database is available via environment variable
	if os.Getenv("INTEGRATION_TEST") == "" {
		t.Skip("skipping integration test: INTEGRATION_TEST not set")
	}

	cfg := &config.DatabaseConfig{
		Provider:        "postgres",
		Host:            getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:            5432,
		User:            getEnvOrDefault("TEST_DB_USER", "postgres"),
		Password:        getEnvOrDefault("TEST_DB_PASSWORD", "postgres"),
		Database:        getEnvOrDefault("TEST_DB_DATABASE", "landlord_test"),
		SSLMode:         "disable",
		MaxConnections:  10,
		MinConnections:  2,
		ConnectTimeout:  10 * time.Second,
		MaxConnLifetime: 1 * time.Hour,
		MaxConnIdleTime: 30 * time.Minute,
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	ctx := context.Background()

	t.Run("Connection", func(t *testing.T) {
		db, err := NewProvider(ctx, cfg, logger)
		if err != nil {
			t.Fatalf("failed to connect to database: %v", err)
		}
		defer db.Close()

		if db.Pool() == nil {
			t.Error("expected connection pool but got nil")
		}
	})

	t.Run("HealthCheck", func(t *testing.T) {
		db, err := NewProvider(ctx, cfg, logger)
		if err != nil {
			t.Fatalf("failed to connect to database: %v", err)
		}
		defer db.Close()

		if err := db.Health(ctx); err != nil {
			t.Errorf("health check failed: %v", err)
		}
	})

	t.Run("ConnectionRetry", func(t *testing.T) {
		// Use invalid config to test retry logic
		badCfg := *cfg
		badCfg.Port = 9999 // Invalid port

		shortCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		_, err := NewProvider(shortCtx, &badCfg, logger)
		if err == nil {
			t.Error("expected connection to fail with invalid config")
		}
	})
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
