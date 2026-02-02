package sqlite

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/jaxxstorm/landlord/internal/config"
)

func TestSQLiteProvider_InMemory(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ctx := context.Background()

	cfg := &config.DatabaseConfig{
		Provider: "sqlite",
		SQLite: config.SQLiteConfig{
			Path:        ":memory:",
			BusyTimeout: 5 * time.Second,
		},
	}

	provider, err := New(ctx, cfg, logger)
	if err != nil {
		t.Fatalf("failed to create in-memory provider: %v", err)
	}
	defer provider.Close()

	if provider.Pool() == nil {
		t.Error("Pool() returned nil for in-memory database")
	}

	if err := provider.Health(ctx); err != nil {
		t.Errorf("health check failed: %v", err)
	}
}

func TestSQLiteProvider_Pragmas(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ctx := context.Background()

	cfg := &config.DatabaseConfig{
		Provider: "sqlite",
		SQLite: config.SQLiteConfig{
			Path:        ":memory:",
			BusyTimeout: 5 * time.Second,
			Pragmas: []string{
				"PRAGMA cache_size=-65000",
			},
		},
	}

	provider, err := New(ctx, cfg, logger)
	if err != nil {
		t.Fatalf("failed to create provider with custom pragmas: %v", err)
	}
	defer provider.Close()

	if err := provider.Health(ctx); err != nil {
		t.Errorf("health check with custom pragmas failed: %v", err)
	}
}

func TestSQLiteProvider_Close(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ctx := context.Background()

	cfg := &config.DatabaseConfig{
		Provider: "sqlite",
		SQLite: config.SQLiteConfig{
			Path:        ":memory:",
			BusyTimeout: 5 * time.Second,
		},
	}

	provider, err := New(ctx, cfg, logger)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	provider.Close()
}
