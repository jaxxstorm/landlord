package database

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/jaxxstorm/landlord/internal/config"
	"github.com/jaxxstorm/landlord/internal/database/providers/postgres"
	"github.com/jaxxstorm/landlord/internal/database/providers/sqlite"
)

// NewProvider creates a database provider based on configuration
func NewProvider(ctx context.Context, cfg *config.DatabaseConfig, logger *zap.Logger) (Provider, error) {
	logger = logger.With(zap.String("component", "database-factory"))

	switch cfg.Provider {
	case "postgres", "postgresql":
		logger.Info("initializing PostgreSQL provider")
		return postgres.New(ctx, cfg, logger)
	case "sqlite":
		logger.Info("initializing SQLite provider")
		return sqlite.New(ctx, cfg, logger)
	default:
		return nil, fmt.Errorf("unknown database provider: %s (supported: postgres, sqlite)", cfg.Provider)
	}
}
