package config

import (
	"fmt"
	"time"
)

// DatabaseConfig holds database connection configuration
// Supports both PostgreSQL and SQLite providers
type DatabaseConfig struct {
	// Provider type: "postgres" (default) or "sqlite"
	Provider string `mapstructure:"provider" env:"DB_PROVIDER" default:"postgres"`

	// PostgreSQL-specific configuration
	Host            string        `mapstructure:"host" env:"DB_HOST" default:"localhost"`
	Port            int           `mapstructure:"port" env:"DB_PORT" default:"5432"`
	User            string        `mapstructure:"user" env:"DB_USER"`
	Password        string        `mapstructure:"password" env:"DB_PASSWORD"`
	Database        string        `mapstructure:"database" env:"DB_DATABASE"`
	SSLMode         string        `mapstructure:"ssl_mode" env:"DB_SSLMODE" default:"prefer"`
	MaxConnections  int32         `mapstructure:"max_connections" env:"DB_MAX_CONNECTIONS" default:"25"`
	MinConnections  int32         `mapstructure:"min_connections" env:"DB_MIN_CONNECTIONS" default:"5"`
	ConnectTimeout  time.Duration `mapstructure:"connect_timeout" env:"DB_CONNECT_TIMEOUT" default:"10s"`
	MaxConnLifetime time.Duration `mapstructure:"max_conn_lifetime" env:"DB_MAX_CONN_LIFETIME" default:"1h"`
	MaxConnIdleTime time.Duration `mapstructure:"max_conn_idle_time" env:"DB_MAX_CONN_IDLE_TIME" default:"30m"`

	// SQLite-specific configuration
	SQLite SQLiteConfig `mapstructure:"sqlite"`
}

// SQLiteConfig holds SQLite-specific configuration
type SQLiteConfig struct {
	Path        string        `mapstructure:"path" env:"DB_SQLITE_PATH" default:"landlord.db"`
	BusyTimeout time.Duration `mapstructure:"busy_timeout" env:"DB_SQLITE_BUSY_TIMEOUT" default:"5s"`
	Pragmas     []string      `mapstructure:"pragmas" env:"DB_SQLITE_PRAGMAS"`
}

// Validate validates database configuration
func (d *DatabaseConfig) Validate() error {
	// Validate provider type
	validProviders := map[string]bool{
		"postgres":   true,
		"postgresql": true,
		"sqlite":     true,
	}
	if !validProviders[d.Provider] {
		return fmt.Errorf("invalid provider: %s (supported: postgres, sqlite)", d.Provider)
	}

	// Provider-specific validation
	switch d.Provider {
	case "postgres", "postgresql":
		return d.validatePostgres()
	case "sqlite":
		return d.validateSQLite()
	}
	return nil
}

// validatePostgres validates PostgreSQL-specific configuration
func (d *DatabaseConfig) validatePostgres() error {
	if d.Port < 1 || d.Port > 65535 {
		return fmt.Errorf("invalid port: %d (must be 1-65535)", d.Port)
	}
	if d.MaxConnections < 1 {
		return fmt.Errorf("max connections must be at least 1")
	}
	if d.MinConnections < 0 {
		return fmt.Errorf("min connections must be non-negative")
	}
	if d.MinConnections > d.MaxConnections {
		return fmt.Errorf("min connections (%d) cannot exceed max connections (%d)", d.MinConnections, d.MaxConnections)
	}
	validSSLModes := map[string]bool{
		"disable":     true,
		"allow":       true,
		"prefer":      true,
		"require":     true,
		"verify-ca":   true,
		"verify-full": true,
	}
	if !validSSLModes[d.SSLMode] {
		return fmt.Errorf("invalid SSL mode: %s", d.SSLMode)
	}
	return nil
}

// validateSQLite validates SQLite-specific configuration
func (d *DatabaseConfig) validateSQLite() error {
	return d.SQLite.Validate()
}

// Validate validates SQLite configuration
func (s *SQLiteConfig) Validate() error {
	if s.Path == "" {
		return fmt.Errorf("SQLite path cannot be empty")
	}
	if s.BusyTimeout < 0 {
		return fmt.Errorf("busy timeout must be non-negative")
	}
	return nil
}

// ConnectionString returns a PostgreSQL connection string
func (d *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Database, d.SSLMode,
	)
}

// MigrationConnectionString returns the appropriate connection string for migrations
// based on the provider type
func (d *DatabaseConfig) MigrationConnectionString() string {
	switch d.Provider {
	case "postgres", "postgresql":
		// PostgreSQL uses pgx driver format
		return fmt.Sprintf("pgx5://%s:%s@%s:%d/%s?sslmode=%s",
			d.User, d.Password, d.Host, d.Port, d.Database, d.SSLMode)
	case "sqlite":
		// SQLite uses sqlite3 driver format
		return fmt.Sprintf("sqlite3://%s", d.SQLite.Path)
	default:
		return ""
	}
}
