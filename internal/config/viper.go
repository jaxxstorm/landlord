package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// NewViperInstance creates and configures a new viper instance with defaults
func NewViperInstance() *viper.Viper {
	v := viper.New()

	// Set default values matching the Config struct defaults
	v.SetDefault("database.provider", "postgres")
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.ssl_mode", "prefer")
	v.SetDefault("database.max_connections", 25)
	v.SetDefault("database.min_connections", 5)
	v.SetDefault("database.connect_timeout", "10s")
	v.SetDefault("database.max_conn_lifetime", "1h")
	v.SetDefault("database.max_conn_idle_time", "30m")

	v.SetDefault("http.host", "0.0.0.0")
	v.SetDefault("http.port", 8080)
	v.SetDefault("http.read_timeout", "10s")
	v.SetDefault("http.write_timeout", "10s")
	v.SetDefault("http.idle_timeout", "120s")
	v.SetDefault("http.shutdown_timeout", "30s")

	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "development")

	v.SetDefault("compute.default_provider", "mock")
	v.SetDefault("workflow.default_provider", "mock")
	v.SetDefault("workflow.step_functions.region", "us-west-2")
	v.SetDefault("workflow.restate.worker_register_on_startup", true)
	v.SetDefault("workflow.restate.worker_compute_cache_ttl", "5m")

	return v
}

// BindEnvironmentVariables binds environment variables to viper keys
func BindEnvironmentVariables(v *viper.Viper) error {
	// Database configuration
	if err := v.BindEnv("database.provider", "DB_PROVIDER"); err != nil {
		return fmt.Errorf("failed to bind DB_PROVIDER: %w", err)
	}
	if err := v.BindEnv("database.host", "DB_HOST"); err != nil {
		return fmt.Errorf("failed to bind DB_HOST: %w", err)
	}
	if err := v.BindEnv("database.port", "DB_PORT"); err != nil {
		return fmt.Errorf("failed to bind DB_PORT: %w", err)
	}
	if err := v.BindEnv("database.user", "DB_USER"); err != nil {
		return fmt.Errorf("failed to bind DB_USER: %w", err)
	}
	if err := v.BindEnv("database.password", "DB_PASSWORD"); err != nil {
		return fmt.Errorf("failed to bind DB_PASSWORD: %w", err)
	}
	if err := v.BindEnv("database.database", "DB_DATABASE"); err != nil {
		return fmt.Errorf("failed to bind DB_DATABASE: %w", err)
	}
	if err := v.BindEnv("database.ssl_mode", "DB_SSLMODE"); err != nil {
		return fmt.Errorf("failed to bind DB_SSLMODE: %w", err)
	}
	if err := v.BindEnv("database.max_connections", "DB_MAX_CONNECTIONS"); err != nil {
		return fmt.Errorf("failed to bind DB_MAX_CONNECTIONS: %w", err)
	}
	if err := v.BindEnv("database.min_connections", "DB_MIN_CONNECTIONS"); err != nil {
		return fmt.Errorf("failed to bind DB_MIN_CONNECTIONS: %w", err)
	}
	if err := v.BindEnv("database.connect_timeout", "DB_CONNECT_TIMEOUT"); err != nil {
		return fmt.Errorf("failed to bind DB_CONNECT_TIMEOUT: %w", err)
	}
	if err := v.BindEnv("database.max_conn_lifetime", "DB_MAX_CONN_LIFETIME"); err != nil {
		return fmt.Errorf("failed to bind DB_MAX_CONN_LIFETIME: %w", err)
	}
	if err := v.BindEnv("database.max_conn_idle_time", "DB_MAX_CONN_IDLE_TIME"); err != nil {
		return fmt.Errorf("failed to bind DB_MAX_CONN_IDLE_TIME: %w", err)
	}

	// HTTP configuration
	if err := v.BindEnv("http.host", "HTTP_HOST"); err != nil {
		return fmt.Errorf("failed to bind HTTP_HOST: %w", err)
	}
	if err := v.BindEnv("http.port", "HTTP_PORT"); err != nil {
		return fmt.Errorf("failed to bind HTTP_PORT: %w", err)
	}
	if err := v.BindEnv("http.read_timeout", "HTTP_READ_TIMEOUT"); err != nil {
		return fmt.Errorf("failed to bind HTTP_READ_TIMEOUT: %w", err)
	}
	if err := v.BindEnv("http.write_timeout", "HTTP_WRITE_TIMEOUT"); err != nil {
		return fmt.Errorf("failed to bind HTTP_WRITE_TIMEOUT: %w", err)
	}
	if err := v.BindEnv("http.idle_timeout", "HTTP_IDLE_TIMEOUT"); err != nil {
		return fmt.Errorf("failed to bind HTTP_IDLE_TIMEOUT: %w", err)
	}
	if err := v.BindEnv("http.shutdown_timeout", "HTTP_SHUTDOWN_TIMEOUT"); err != nil {
		return fmt.Errorf("failed to bind HTTP_SHUTDOWN_TIMEOUT: %w", err)
	}

	// Logging configuration
	if err := v.BindEnv("log.level", "LOG_LEVEL"); err != nil {
		return fmt.Errorf("failed to bind LOG_LEVEL: %w", err)
	}
	if err := v.BindEnv("log.format", "LOG_FORMAT"); err != nil {
		return fmt.Errorf("failed to bind LOG_FORMAT: %w", err)
	}

	// Compute configuration
	if err := v.BindEnv("compute.default_provider", "COMPUTE_DEFAULT_PROVIDER"); err != nil {
		return fmt.Errorf("failed to bind COMPUTE_DEFAULT_PROVIDER: %w", err)
	}

	// Workflow configuration
	if err := v.BindEnv("workflow.default_provider", "WORKFLOW_DEFAULT_PROVIDER"); err != nil {
		return fmt.Errorf("failed to bind WORKFLOW_DEFAULT_PROVIDER: %w", err)
	}
	if err := v.BindEnv("workflow.step_functions.region", "WORKFLOW_SFN_REGION"); err != nil {
		return fmt.Errorf("failed to bind WORKFLOW_SFN_REGION: %w", err)
	}
	if err := v.BindEnv("workflow.step_functions.role_arn", "WORKFLOW_SFN_ROLE_ARN"); err != nil {
		return fmt.Errorf("failed to bind WORKFLOW_SFN_ROLE_ARN: %w", err)
	}
	if err := v.BindEnv("workflow.restate.endpoint", "WORKFLOW_RESTATE_ENDPOINT"); err != nil {
		return fmt.Errorf("failed to bind WORKFLOW_RESTATE_ENDPOINT: %w", err)
	}
	if err := v.BindEnv("workflow.restate.admin_endpoint", "WORKFLOW_RESTATE_ADMIN_ENDPOINT"); err != nil {
		return fmt.Errorf("failed to bind WORKFLOW_RESTATE_ADMIN_ENDPOINT: %w", err)
	}
	if err := v.BindEnv("workflow.restate.execution_mechanism", "WORKFLOW_RESTATE_EXECUTION_MECHANISM"); err != nil {
		return fmt.Errorf("failed to bind WORKFLOW_RESTATE_EXECUTION_MECHANISM: %w", err)
	}
	if err := v.BindEnv("workflow.restate.service_name", "WORKFLOW_RESTATE_SERVICE_NAME"); err != nil {
		return fmt.Errorf("failed to bind WORKFLOW_RESTATE_SERVICE_NAME: %w", err)
	}
	if err := v.BindEnv("workflow.restate.auth_type", "WORKFLOW_RESTATE_AUTH_TYPE"); err != nil {
		return fmt.Errorf("failed to bind WORKFLOW_RESTATE_AUTH_TYPE: %w", err)
	}
	if err := v.BindEnv("workflow.restate.api_key", "WORKFLOW_RESTATE_API_KEY"); err != nil {
		return fmt.Errorf("failed to bind WORKFLOW_RESTATE_API_KEY: %w", err)
	}
	if err := v.BindEnv("workflow.restate.timeout", "WORKFLOW_RESTATE_TIMEOUT"); err != nil {
		return fmt.Errorf("failed to bind WORKFLOW_RESTATE_TIMEOUT: %w", err)
	}
	if err := v.BindEnv("workflow.restate.retry_attempts", "WORKFLOW_RESTATE_RETRY_ATTEMPTS"); err != nil {
		return fmt.Errorf("failed to bind WORKFLOW_RESTATE_RETRY_ATTEMPTS: %w", err)
	}
	if err := v.BindEnv("workflow.restate.worker_register_on_startup", "WORKFLOW_RESTATE_WORKER_REGISTER_ON_STARTUP"); err != nil {
		return fmt.Errorf("failed to bind WORKFLOW_RESTATE_WORKER_REGISTER_ON_STARTUP: %w", err)
	}
	if err := v.BindEnv("workflow.restate.worker_admin_endpoint", "WORKFLOW_RESTATE_WORKER_ADMIN_ENDPOINT"); err != nil {
		return fmt.Errorf("failed to bind WORKFLOW_RESTATE_WORKER_ADMIN_ENDPOINT: %w", err)
	}
	if err := v.BindEnv("workflow.restate.worker_namespace", "WORKFLOW_RESTATE_WORKER_NAMESPACE"); err != nil {
		return fmt.Errorf("failed to bind WORKFLOW_RESTATE_WORKER_NAMESPACE: %w", err)
	}
	if err := v.BindEnv("workflow.restate.worker_service_prefix", "WORKFLOW_RESTATE_WORKER_SERVICE_PREFIX"); err != nil {
		return fmt.Errorf("failed to bind WORKFLOW_RESTATE_WORKER_SERVICE_PREFIX: %w", err)
	}
	if err := v.BindEnv("workflow.restate.worker_landlord_api_url", "WORKFLOW_RESTATE_WORKER_LANDLORD_API_URL"); err != nil {
		return fmt.Errorf("failed to bind WORKFLOW_RESTATE_WORKER_LANDLORD_API_URL: %w", err)
	}
	if err := v.BindEnv("workflow.restate.worker_compute_provider", "WORKFLOW_RESTATE_WORKER_COMPUTE_PROVIDER"); err != nil {
		return fmt.Errorf("failed to bind WORKFLOW_RESTATE_WORKER_COMPUTE_PROVIDER: %w", err)
	}
	if err := v.BindEnv("workflow.restate.worker_compute_cache_ttl", "WORKFLOW_RESTATE_WORKER_COMPUTE_CACHE_TTL"); err != nil {
		return fmt.Errorf("failed to bind WORKFLOW_RESTATE_WORKER_COMPUTE_CACHE_TTL: %w", err)
	}
	if err := v.BindEnv("workflow.restate.worker_advertised_url", "WORKFLOW_RESTATE_WORKER_ADVERTISED_URL"); err != nil {
		return fmt.Errorf("failed to bind WORKFLOW_RESTATE_WORKER_ADVERTISED_URL: %w", err)
	}

	return nil
}

// FindConfigFile finds a configuration file using the precedence order:
// 1. Explicit --config flag (passed via configPath parameter)
// 2. LANDLORD_CONFIG environment variable
// 3. Standard locations: ./config.{yaml,json}, /etc/landlord/config.{yaml,json}, $XDG_CONFIG_HOME/landlord/config.{yaml,json}
func FindConfigFile(configPath string) (string, error) {
	// If explicit config path provided, use it
	if configPath != "" {
		// Verify file exists and is readable
		if _, err := os.Stat(configPath); err != nil {
			if os.IsNotExist(err) {
				return "", fmt.Errorf("config file not found: %s", configPath)
			}
			return "", fmt.Errorf("cannot access config file %s: %w", configPath, err)
		}
		return configPath, nil
	}

	// Check LANDLORD_CONFIG environment variable
	if envPath := os.Getenv("LANDLORD_CONFIG"); envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			return envPath, nil
		}
	}

	// Search standard locations for config files
	locations := []string{
		".",             // Current directory
		"/etc/landlord", // System config directory
	}

	// Add XDG_CONFIG_HOME location if set
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		locations = append(locations, filepath.Join(xdgConfig, "landlord"))
	}

	// Try each location with both YAML and JSON extensions
	for _, loc := range locations {
		for _, ext := range []string{"yaml", "json"} {
			path := filepath.Join(loc, "config."+ext)
			if _, err := os.Stat(path); err == nil {
				return path, nil
			}
		}
	}

	// No config file found - this is not an error, config can come from env vars or defaults
	return "", nil
}

// LoadConfigFile loads a configuration file (YAML or JSON) into viper
func LoadConfigFile(v *viper.Viper, filePath string) error {
	if filePath == "" {
		return nil
	}

	// Determine file type based on extension
	ext := filepath.Ext(filePath)
	switch ext {
	case ".yaml", ".yml":
		v.SetConfigType("yaml")
	case ".json":
		v.SetConfigType("json")
	default:
		return fmt.Errorf("unsupported config file type: %s", ext)
	}

	v.SetConfigFile(filePath)

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	return nil
}

// LoadFromViper unmarshals viper configuration into a Config struct
func LoadFromViper(v *viper.Viper) (*Config, error) {
	var cfg Config

	// Unmarshal all viper values into the config struct
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}
