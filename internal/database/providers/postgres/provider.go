package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/jaxxstorm/landlord/internal/config"
)

// Provider implements the database Provider interface for PostgreSQL
type Provider struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

// New creates a new PostgreSQL database provider with retry logic
func New(ctx context.Context, cfg *config.DatabaseConfig, logger *zap.Logger) (*Provider, error) {
	logger = logger.With(zap.String("component", "postgres-provider"))

	// Build connection config
	poolConfig, err := pgxpool.ParseConfig(cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Configure pool settings
	poolConfig.MaxConns = cfg.MaxConnections
	poolConfig.MinConns = cfg.MinConnections
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	// Retry connection with exponential backoff
	var pool *pgxpool.Pool
	maxRetries := 5
	backoff := 1 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		logger.Info("attempting database connection",
			zap.Int("attempt", attempt),
			zap.Int("max_retries", maxRetries),
		)

		connectCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
		pool, err = pgxpool.NewWithConfig(connectCtx, poolConfig)
		cancel()

		if err == nil {
			// Test the connection
			pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
			err = pool.Ping(pingCtx)
			pingCancel()

			if err == nil {
				logger.Info("database connection established",
					zap.String("host", cfg.Host),
					zap.Int("port", cfg.Port),
					zap.String("database", cfg.Database),
				)
				return &Provider{pool: pool, logger: logger}, nil
			}
		}

		logger.Warn("database connection failed",
			zap.Error(err),
			zap.Int("attempt", attempt),
			zap.Duration("retry_in", backoff),
		)

		if attempt < maxRetries {
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during connection retry: %w", ctx.Err())
			case <-time.After(backoff):
			}
			backoff *= 2 // Exponential backoff
		}
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
}

// Pool returns the underlying *pgxpool.Pool
func (p *Provider) Pool() interface{} {
	return p.pool
}

// Health checks if the database connection is healthy
func (p *Provider) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := p.pool.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	return nil
}

// Close gracefully closes the database connection pool
func (p *Provider) Close() {
	p.logger.Info("closing PostgreSQL connections")
	p.pool.Close()
	p.logger.Info("PostgreSQL connections closed")
}
