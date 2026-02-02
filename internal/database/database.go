package database

import (
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"go.uber.org/zap"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// RunMigrations applies all pending database migrations
// Supports both PostgreSQL and SQLite connection strings
func RunMigrations(connString string, logger *zap.Logger) error {
	logger = logger.With(zap.String("component", "migrations"))
	logger.Info("applying database migrations")

	// Create migration source from embedded files
	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	// Create migrate instance
	// The connection string format determines which driver is used:
	// - postgres://... uses pgx/v5 driver
	// - sqlite:... or file:... uses sqlite3 driver
	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, connString)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	// Get current version
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get current migration version: %w", err)
	}
	if dirty {
		return fmt.Errorf("database is in dirty state at version %d", version)
	}

	logger.Info("current migration version", zap.Uint("version", version))

	// Run migrations
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			logger.Info("no pending migrations")
			return nil
		}
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	// Get new version
	newVersion, _, err := m.Version()
	if err != nil {
		return fmt.Errorf("failed to get new migration version: %w", err)
	}

	logger.Info("migrations applied successfully", zap.Uint("new_version", newVersion))
	return nil
}
