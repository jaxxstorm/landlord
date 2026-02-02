package restate_test

import (
	"testing"
	"time"

	"github.com/jaxxstorm/landlord/internal/config"
	"github.com/jaxxstorm/landlord/internal/workflow/providers/restate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// TestAuthenticationNone tests "none" authentication type
func TestAuthenticationNone(t *testing.T) {
	logger := zaptest.NewLogger(t)

	cfg := config.RestateConfig{
		Endpoint:           "http://localhost:8080",
		ExecutionMechanism: "local",
		AuthType:           "none",
		Timeout:            30 * time.Minute,
	}

	provider, err := restate.New(cfg, logger)
	require.NoError(t, err)
	require.NotNil(t, provider)
	assert.Equal(t, "restate", provider.Name())
}

// TestAuthenticationAPIKey tests "api_key" authentication type
func TestAuthenticationAPIKey(t *testing.T) {
	logger := zaptest.NewLogger(t)

	cfg := config.RestateConfig{
		Endpoint:           "http://localhost:8080",
		ExecutionMechanism: "self-hosted",
		AuthType:           "api_key",
		ApiKey:             "test-api-key-12345",
		Timeout:            30 * time.Minute,
	}

	provider, err := restate.New(cfg, logger)
	require.NoError(t, err)
	require.NotNil(t, provider)
	assert.Equal(t, "restate", provider.Name())
}

// TestAuthenticationIAM tests "iam" authentication type
func TestAuthenticationIAM(t *testing.T) {
	logger := zaptest.NewLogger(t)

	cfg := config.RestateConfig{
		Endpoint:           "https://my-fargate-service.example.com",
		ExecutionMechanism: "lambda",
		AuthType:           "iam",
		Timeout:            30 * time.Minute,
	}

	provider, err := restate.New(cfg, logger)
	require.NoError(t, err)
	require.NotNil(t, provider)
	assert.Equal(t, "restate", provider.Name())
}

// TestAuthenticationMissingAPIKey tests error when api_key auth type is configured but no key provided
func TestAuthenticationMissingAPIKey(t *testing.T) {
	cfg := config.RestateConfig{
		Endpoint:           "http://localhost:8080",
		ExecutionMechanism: "self-hosted",
		AuthType:           "api_key",
		ApiKey:             "",
		Timeout:            30 * time.Minute,
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api_key")
}

// TestAuthenticationInvalidType tests error with invalid authentication type
func TestAuthenticationInvalidType(t *testing.T) {
	cfg := config.RestateConfig{
		Endpoint:           "http://localhost:8080",
		ExecutionMechanism: "local",
		AuthType:           "invalid_auth_type",
		Timeout:            30 * time.Minute,
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid auth_type")
}

// TestAuthenticationMechanismCompatibility tests auth type and execution mechanism compatibility
func TestAuthenticationMechanismCompatibility(t *testing.T) {
	tests := []struct {
		name       string
		mechanism  string
		authType   string
		apiKey     string
		shouldFail bool
	}{
		{
			name:       "local with none",
			mechanism:  "local",
			authType:   "none",
			shouldFail: false,
		},
		{
			name:       "self-hosted with api_key",
			mechanism:  "self-hosted",
			authType:   "api_key",
			apiKey:     "test-key",
			shouldFail: false,
		},
		{
			name:       "lambda with iam",
			mechanism:  "lambda",
			authType:   "iam",
			shouldFail: false,
		},
		{
			name:       "kubernetes with api_key",
			mechanism:  "kubernetes",
			authType:   "api_key",
			apiKey:     "test-key",
			shouldFail: false,
		},
		{
			name:       "api_key without key",
			mechanism:  "self-hosted",
			authType:   "api_key",
			apiKey:     "",
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.RestateConfig{
				Endpoint:           "http://localhost:8080",
				ExecutionMechanism: tt.mechanism,
				AuthType:           tt.authType,
				ApiKey:             tt.apiKey,
				Timeout:            30 * time.Minute,
			}

			err := cfg.Validate()
			if tt.shouldFail {
				assert.Error(t, err, "expected validation to fail")
			} else {
				assert.NoError(t, err, "expected validation to pass")
			}
		})
	}
}
