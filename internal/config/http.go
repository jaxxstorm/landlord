package config

import (
	"fmt"
	"strings"
	"time"
)

// HTTPConfig holds HTTP server configuration
type HTTPConfig struct {
	Host            string              `mapstructure:"host" env:"HTTP_HOST" default:"0.0.0.0"`
	Port            int                 `mapstructure:"port" env:"HTTP_PORT" default:"8080"`
	ReadTimeout     time.Duration       `mapstructure:"read_timeout" env:"HTTP_READ_TIMEOUT" default:"10s"`
	WriteTimeout    time.Duration       `mapstructure:"write_timeout" env:"HTTP_WRITE_TIMEOUT" default:"10s"`
	IdleTimeout     time.Duration       `mapstructure:"idle_timeout" env:"HTTP_IDLE_TIMEOUT" default:"120s"`
	ShutdownTimeout time.Duration       `mapstructure:"shutdown_timeout" env:"HTTP_SHUTDOWN_TIMEOUT" default:"30s"`
	TailscaleAuth   TailscaleAuthConfig `mapstructure:"tailscale_auth"`
}

const TailscaleCapabilityPrefix = "lbrlabs.com/cap/landlord"

// TailscaleAuthConfig holds tsnet-backed API authorization settings.
type TailscaleAuthConfig struct {
	Enabled            bool                         `mapstructure:"enabled" env:"HTTP_TAILSCALE_AUTH_ENABLED"`
	Hostname           string                       `mapstructure:"hostname" env:"HTTP_TAILSCALE_AUTH_HOSTNAME"`
	StateDir           string                       `mapstructure:"state_dir" env:"HTTP_TAILSCALE_AUTH_STATE_DIR"`
	AuthKey            string                       `mapstructure:"auth_key" env:"HTTP_TAILSCALE_AUTH_AUTH_KEY"`
	Ephemeral          bool                         `mapstructure:"ephemeral" env:"HTTP_TAILSCALE_AUTH_EPHEMERAL"`
	ListenAddr         string                       `mapstructure:"listen_addr" env:"HTTP_TAILSCALE_AUTH_LISTEN_ADDR"`
	ProtectedEndpoints []TailscaleProtectedEndpoint `mapstructure:"protected_endpoints"`
}

// TailscaleProtectedEndpoint defines a protected HTTP method/path rule.
type TailscaleProtectedEndpoint struct {
	Method     string `mapstructure:"method"`
	Path       string `mapstructure:"path"`
	Capability string `mapstructure:"capability"`
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
	if err := h.TailscaleAuth.Validate(h.Port); err != nil {
		return fmt.Errorf("tailscale_auth: %w", err)
	}
	return nil
}

// Address returns the HTTP server address
func (h *HTTPConfig) Address() string {
	return fmt.Sprintf("%s:%d", h.Host, h.Port)
}

// ListenAddress returns the address tsnet should bind to when enabled.
func (t *TailscaleAuthConfig) ListenAddress(httpPort int) string {
	if strings.TrimSpace(t.ListenAddr) != "" {
		return strings.TrimSpace(t.ListenAddr)
	}
	return fmt.Sprintf(":%d", httpPort)
}

// Validate validates Tailscale auth configuration.
func (t *TailscaleAuthConfig) Validate(httpPort int) error {
	if !t.Enabled {
		return nil
	}

	if strings.TrimSpace(t.Hostname) == "" {
		return fmt.Errorf("hostname is required when enabled")
	}

	listenAddr := t.ListenAddress(httpPort)
	if !strings.HasPrefix(listenAddr, ":") {
		return fmt.Errorf("listen_addr must be in :port form")
	}

	for i := range t.ProtectedEndpoints {
		if err := t.ProtectedEndpoints[i].Validate(); err != nil {
			return fmt.Errorf("protected_endpoints[%d]: %w", i, err)
		}
	}

	return nil
}

// Validate validates a protected endpoint rule.
func (e *TailscaleProtectedEndpoint) Validate() error {
	method := strings.ToUpper(strings.TrimSpace(e.Method))
	if !isSupportedHTTPMethod(method) {
		return fmt.Errorf("unsupported method %q", e.Method)
	}
	if !strings.HasPrefix(strings.TrimSpace(e.Path), "/") {
		return fmt.Errorf("path must start with /")
	}
	capability := strings.TrimSpace(e.Capability)
	if capability == "" {
		return fmt.Errorf("capability is required")
	}
	if capability != TailscaleCapabilityPrefix && !strings.HasPrefix(capability, TailscaleCapabilityPrefix+"/") {
		return fmt.Errorf("capability must be %q or nested beneath it", TailscaleCapabilityPrefix)
	}

	e.Method = method
	e.Path = strings.TrimSpace(e.Path)
	e.Capability = capability
	return nil
}

func isSupportedHTTPMethod(method string) bool {
	switch method {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
		return true
	default:
		return false
	}
}
