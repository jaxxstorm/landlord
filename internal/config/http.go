package config

import (
	"fmt"
	"time"
)

// HTTPConfig holds HTTP server configuration
type HTTPConfig struct {
	Host            string        `mapstructure:"host" env:"HTTP_HOST" default:"0.0.0.0"`
	Port            int           `mapstructure:"port" env:"HTTP_PORT" default:"8080"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout" env:"HTTP_READ_TIMEOUT" default:"10s"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout" env:"HTTP_WRITE_TIMEOUT" default:"10s"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout" env:"HTTP_IDLE_TIMEOUT" default:"120s"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" env:"HTTP_SHUTDOWN_TIMEOUT" default:"30s"`
}

// Validate validates HTTP configuration
func (h *HTTPConfig) Validate() error {
	if h.Port < 1 || h.Port > 65535 {
		return fmt.Errorf("invalid port: %d (must be 1-65535)", h.Port)
	}
	if h.ReadTimeout < 0 {
		return fmt.Errorf("read timeout must be non-negative")
	}
	if h.WriteTimeout < 0 {
		return fmt.Errorf("write timeout must be non-negative")
	}
	if h.ShutdownTimeout < 0 {
		return fmt.Errorf("shutdown timeout must be non-negative")
	}
	return nil
}

// Address returns the HTTP server address
func (h *HTTPConfig) Address() string {
	return fmt.Sprintf("%s:%d", h.Host, h.Port)
}
