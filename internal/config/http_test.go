package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTailscaleAuthConfigValidate(t *testing.T) {
	tests := []struct {
		name      string
		cfg       TailscaleAuthConfig
		shouldErr bool
		errMsg    string
	}{
		{
			name: "disabled auth is valid",
			cfg:  TailscaleAuthConfig{},
		},
		{
			name: "enabled auth with valid endpoint rules",
			cfg: TailscaleAuthConfig{
				Enabled:  true,
				Hostname: "landlord-test",
				ProtectedEndpoints: []TailscaleProtectedEndpoint{{
					Method:     "get",
					Path:       "/v1/tenants/{id}",
					Capability: TailscaleCapabilityPrefix,
				}},
			},
		},
		{
			name: "missing hostname when enabled",
			cfg: TailscaleAuthConfig{
				Enabled: true,
			},
			shouldErr: true,
			errMsg:    "hostname is required",
		},
		{
			name: "invalid http method",
			cfg: TailscaleAuthConfig{
				Enabled:  true,
				Hostname: "landlord-test",
				ProtectedEndpoints: []TailscaleProtectedEndpoint{{
					Method:     "TRACE",
					Path:       "/v1/tenants",
					Capability: TailscaleCapabilityPrefix,
				}},
			},
			shouldErr: true,
			errMsg:    "unsupported method",
		},
		{
			name: "missing capability",
			cfg: TailscaleAuthConfig{
				Enabled:  true,
				Hostname: "landlord-test",
				ProtectedEndpoints: []TailscaleProtectedEndpoint{{
					Method: "GET",
					Path:   "/v1/tenants",
				}},
			},
			shouldErr: true,
			errMsg:    "capability is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate(8080)
			if tt.shouldErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestHTTPConfigValidateWithTailscaleAuth(t *testing.T) {
	cfg := HTTPConfig{
		Host:            "0.0.0.0",
		Port:            8080,
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     120 * time.Second,
		ShutdownTimeout: 30 * time.Second,
		TailscaleAuth: TailscaleAuthConfig{
			Enabled:  true,
			Hostname: "landlord-test",
			ProtectedEndpoints: []TailscaleProtectedEndpoint{{
				Method:     "GET",
				Path:       "/v1/docs",
				Capability: TailscaleCapabilityPrefix,
			}},
		},
	}

	require.NoError(t, cfg.Validate())
	assert.Equal(t, "GET", cfg.TailscaleAuth.ProtectedEndpoints[0].Method)
	assert.Equal(t, "/v1/docs", cfg.TailscaleAuth.ProtectedEndpoints[0].Path)
}
