package postgres

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/jaxxstorm/landlord/internal/config"
)

func TestPostgresProvider_InvalidConfig(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	ctx := context.Background()

	cfg := &config.DatabaseConfig{
		Provider:        "postgres",
		Host:            "localhost",
		Port:            9999,
		User:            "test",
		Password:        "test",
		Database:        "test",
		SSLMode:         "disable",
		MaxConnections:  2,
		MinConnections:  1,
		ConnectTimeout:  1 * time.Second,
		MaxConnLifetime: time.Hour,
		MaxConnIdleTime: 30 * time.Minute,
	}

	_, err := New(ctx, cfg, logger)
	if err == nil {
		t.Error("expected error for invalid configuration")
	}
}
