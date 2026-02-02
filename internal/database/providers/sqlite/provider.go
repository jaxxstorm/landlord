package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
	"go.uber.org/zap"

	"github.com/jaxxstorm/landlord/internal/config"
)

// Provider implements the database Provider interface for SQLite
type Provider struct {
	db     *sqlx.DB
	logger *zap.Logger
	path   string
}

// New creates a new SQLite database provider
func New(ctx context.Context, cfg *config.DatabaseConfig, logger *zap.Logger) (*Provider, error) {
	logger = logger.With(zap.String("component", "sqlite-provider"))

	// Get SQLite configuration
	sqliteCfg := cfg.SQLite
	path := sqliteCfg.Path

	// Handle special cases for in-memory databases
	if strings.HasPrefix(path, ":memory:") || strings.HasPrefix(path, "file::memory:") {
		logger.Info("initializing in-memory SQLite database")
	} else {
		// Ensure parent directory exists for file-based databases
		if !strings.HasPrefix(path, "file:") {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
			}
			path = absPath
		}
		logger.Info("initializing file-based SQLite database", zap.String("path", path))
	}

	// Open database connection
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Wrap with sqlx for enhanced functionality
	dbx := sqlx.NewDb(db, "sqlite")

	// Configure connection pool
	// SQLite benefits from limited connections due to single-writer model
	dbx.SetMaxOpenConns(10)          // Allow multiple readers
	dbx.SetMaxIdleConns(5)            // Maintain some idle connections
	dbx.SetConnMaxLifetime(time.Hour) // Recycle connections hourly

	// Test the connection
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := dbx.PingContext(pingCtx); err != nil {
		dbx.Close()
		return nil, fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	provider := &Provider{
		db:     dbx,
		logger: logger,
		path:   path,
	}

	// Apply pragmas
	if err := provider.applyPragmas(ctx, sqliteCfg); err != nil {
		dbx.Close()
		return nil, fmt.Errorf("failed to apply pragmas: %w", err)
	}

	logger.Info("SQLite database initialized successfully")
	return provider, nil
}

// applyPragmas configures SQLite with appropriate pragmas
func (p *Provider) applyPragmas(ctx context.Context, cfg config.SQLiteConfig) error {
	// Default pragmas for optimal performance and correctness
	defaultPragmas := []string{
		"PRAGMA journal_mode=WAL",            // Enable Write-Ahead Logging for concurrency
		"PRAGMA synchronous=NORMAL",          // Balance durability and performance
		"PRAGMA foreign_keys=ON",             // Enable foreign key constraints
		"PRAGMA temp_store=MEMORY",           // Use memory for temporary tables
		fmt.Sprintf("PRAGMA busy_timeout=%d", int(cfg.BusyTimeout.Milliseconds())), // Set busy timeout
	}

	// Apply default pragmas
	for _, pragma := range defaultPragmas {
		p.logger.Debug("applying pragma", zap.String("pragma", pragma))
		if _, err := p.db.ExecContext(ctx, pragma); err != nil {
			return fmt.Errorf("failed to apply pragma %s: %w", pragma, err)
		}
	}

	// Apply user-configured pragmas (these override defaults if duplicated)
	for _, pragma := range cfg.Pragmas {
		p.logger.Debug("applying custom pragma", zap.String("pragma", pragma))
		if _, err := p.db.ExecContext(ctx, pragma); err != nil {
			return fmt.Errorf("failed to apply custom pragma %s: %w", pragma, err)
		}
	}

	// Log effective journal mode for verification
	var journalMode string
	if err := p.db.GetContext(ctx, &journalMode, "PRAGMA journal_mode"); err == nil {
		p.logger.Info("SQLite journal mode", zap.String("mode", journalMode))
	}

	return nil
}

// Pool returns the underlying *sqlx.DB
func (p *Provider) Pool() interface{} {
	return p.db
}

// Health checks if the database connection is healthy
func (p *Provider) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Simple SELECT query to verify database accessibility
	var result int
	if err := p.db.GetContext(ctx, &result, "SELECT 1"); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// Close gracefully closes the database connections
func (p *Provider) Close() {
	p.logger.Info("closing SQLite connections")
	if err := p.db.Close(); err != nil {
		p.logger.Error("error closing SQLite database", zap.Error(err))
	} else {
		p.logger.Info("SQLite connections closed")
	}
}
