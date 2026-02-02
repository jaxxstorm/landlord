package database

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/jaxxstorm/landlord/internal/config"
)

func TestPostgresProvider(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	logger, _ := zap.NewDevelopment()
	ctx := context.Background()

	cfg := &config.DatabaseConfig{
		Provider:        "postgres",
		Host:            "localhost",
		Port:            5432,
		User:            "test",
		Password:        "test",
		Database:        "test",
		SSLMode:         "disable",
		MaxConnections:  10,
		MinConnections:  2,
		ConnectTimeout:  10 * time.Second,
		MaxConnLifetime: time.Hour,
		MaxConnIdleTime: 30 * time.Minute,
	}

	provider, err := NewProvider(ctx, cfg, logger)
	if err != nil {
		t.Skip("PostgreSQL not available:", err)
		return
	}
	defer provider.Close()

	if err := provider.Health(ctx); err != nil {
		t.Errorf("Health check failed: %v", err)
	}

	if provider.Pool() == nil {
		t.Error("Pool() returned nil")
	}
}

func TestSQLiteProvider(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ctx := context.Background()

	cfg := &config.DatabaseConfig{
		Provider: "sqlite",
		SQLite: config.SQLiteConfig{
			Path:        ":memory:",
			BusyTimeout: 5 * time.Second,
		},
	}

	provider, err := NewProvider(ctx, cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create SQLite provider: %v", err)
	}
	defer provider.Close()

	if err := provider.Health(ctx); err != nil {
		t.Errorf("Health check failed: %v", err)
	}

	if provider.Pool() == nil {
		t.Error("Pool() returned nil")
	}
}

func TestUnknownProvider(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ctx := context.Background()

	cfg := &config.DatabaseConfig{
		Provider: "mysql",
	}

	_, err := NewProvider(ctx, cfg, logger)
	if err == nil {
		t.Error("Expected error for unknown provider, got nil")
	}
}
