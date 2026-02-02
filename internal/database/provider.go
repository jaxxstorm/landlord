package database

import "context"

// Provider defines the interface for database providers
// Implementations include PostgreSQL and SQLite providers
type Provider interface {
	// Pool returns the underlying connection pool or database handle
	// Returns *pgxpool.Pool for PostgreSQL or *sqlx.DB for SQLite
	Pool() interface{}

	// Health checks if the database connection is healthy
	Health(ctx context.Context) error

	// Close gracefully closes the database connections
	Close()
}
