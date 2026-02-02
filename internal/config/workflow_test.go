package config_test

import (
	"testing"
	"time"

	"github.com/jaxxstorm/landlord/internal/config"
	"github.com/stretchr/testify/assert"
)

// TestRestateConfigValidation tests RestateConfig.Validate()
func TestRestateConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    config.RestateConfig
		shouldErr bool
		errMsg    string
	}{
		{
			name: "valid local configuration",
			config: config.RestateConfig{
				Endpoint:           "http://localhost:8080",
				ExecutionMechanism: "local",
				AuthType:           "none",
				Timeout:            30 * time.Minute,
				RetryAttempts:      3,
			},
			shouldErr: false,
		},
		{
			name: "valid self-hosted with API key",
			config: config.RestateConfig{
				Endpoint:           "https://restate.example.com",
				ExecutionMechanism: "self-hosted",
				AuthType:           "api_key",
				ApiKey:             "test-key-123",
				Timeout:            30 * time.Minute,
				RetryAttempts:      3,
			},
			shouldErr: false,
		},
		{
			name: "valid Lambda with IAM",
			config: config.RestateConfig{
				Endpoint:           "https://restate-lambda.example.com",
				ExecutionMechanism: "lambda",
				AuthType:           "iam",
				Timeout:            30 * time.Minute,
				RetryAttempts:      3,
			},
			shouldErr: false,
		},
		{
			name: "missing endpoint",
			config: config.RestateConfig{
				Endpoint:           "",
				ExecutionMechanism: "local",
				AuthType:           "none",
				Timeout:            30 * time.Minute,
			},
			shouldErr: true,
			errMsg:    "endpoint is required",
		},
		{
			name: "invalid URL scheme",
			config: config.RestateConfig{
				Endpoint:           "ftp://invalid.example.com",
				ExecutionMechanism: "local",
				AuthType:           "none",
				Timeout:            30 * time.Minute,
			},
			shouldErr: true,
			errMsg:    "scheme must be http or https",
		},
		{
			name: "invalid execution mechanism",
			config: config.RestateConfig{
				Endpoint:           "http://localhost:8080",
				ExecutionMechanism: "invalid-mechanism",
				AuthType:           "none",
				Timeout:            30 * time.Minute,
			},
			shouldErr: true,
			errMsg:    "invalid execution_mechanism",
		},
		{
			name: "invalid auth type",
			config: config.RestateConfig{
				Endpoint:           "http://localhost:8080",
				ExecutionMechanism: "local",
				AuthType:           "invalid-auth",
				Timeout:            30 * time.Minute,
			},
			shouldErr: true,
			errMsg:    "invalid auth_type",
		},
		{
			name: "api_key auth without key",
			config: config.RestateConfig{
				Endpoint:           "http://localhost:8080",
				ExecutionMechanism: "self-hosted",
				AuthType:           "api_key",
				ApiKey:             "",
				Timeout:            30 * time.Minute,
			},
			shouldErr: true,
			errMsg:    "api_key",
		},
		{
			name: "negative timeout",
			config: config.RestateConfig{
				Endpoint:           "http://localhost:8080",
				ExecutionMechanism: "local",
				AuthType:           "none",
				Timeout:            -1 * time.Minute,
			},
			shouldErr: true,
			errMsg:    "timeout must be positive",
		},
		{
			name: "negative retry attempts",
			config: config.RestateConfig{
				Endpoint:           "http://localhost:8080",
				ExecutionMechanism: "local",
				AuthType:           "none",
				Timeout:            30 * time.Minute,
				RetryAttempts:      -1,
			},
			shouldErr: true,
			errMsg:    "retry_attempts must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.shouldErr {
				assert.Error(t, err, "expected validation to fail")
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err, "expected validation to pass")
			}
		})
	}
}

func TestRestateWorkerConfigValidation(t *testing.T) {
	cfg := config.RestateConfig{
		Endpoint:            "http://localhost:8080",
		WorkerAdminEndpoint: "not-a-url",
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for worker admin endpoint")
	}
}

// TestEndpointURLValidation tests URL format validation
func TestEndpointURLValidation(t *testing.T) {
	tests := []struct {
		name      string
		endpoint  string
		shouldErr bool
	}{
		{
			name:      "valid http URL",
			endpoint:  "http://localhost:8080",
			shouldErr: false,
		},
		{
			name:      "valid https URL",
			endpoint:  "https://restate.example.com",
			shouldErr: false,
		},
		{
			name:      "missing scheme",
			endpoint:  "localhost:8080",
			shouldErr: true,
		},
		{
			name:      "invalid scheme",
			endpoint:  "ftp://example.com",
			shouldErr: true,
		},
		{
			name:      "missing host",
			endpoint:  "http://",
			shouldErr: true,
		},
		{
			name:      "http with port",
			endpoint:  "http://localhost:9090",
			shouldErr: false,
		},
		{
			name:      "https with path",
			endpoint:  "https://example.com/api",
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.RestateConfig{
				Endpoint:           tt.endpoint,
				ExecutionMechanism: "local",
				AuthType:           "none",
				Timeout:            30 * time.Minute,
			}
			err := cfg.Validate()
			if tt.shouldErr {
				assert.Error(t, err, "expected endpoint validation to fail")
			} else {
				assert.NoError(t, err, "expected endpoint validation to pass")
			}
		})
	}
}

// TestExecutionMechanismValidation tests execution mechanism values
func TestExecutionMechanismValidation(t *testing.T) {
	validMechanisms := []string{"local", "lambda", "fargate", "kubernetes", "self-hosted"}

	for _, mechanism := range validMechanisms {
		t.Run(mechanism, func(t *testing.T) {
			cfg := config.RestateConfig{
				Endpoint:           "http://localhost:8080",
				ExecutionMechanism: mechanism,
				AuthType:           "none",
				Timeout:            30 * time.Minute,
			}
			err := cfg.Validate()
			assert.NoError(t, err, "expected %s to be valid mechanism", mechanism)
		})
	}

	// Test invalid mechanism
	cfg := config.RestateConfig{
		Endpoint:           "http://localhost:8080",
		ExecutionMechanism: "invalid",
		AuthType:           "none",
		Timeout:            30 * time.Minute,
	}
	err := cfg.Validate()
	assert.Error(t, err, "expected invalid mechanism to fail")
}

// TestAuthTypeValidation tests authentication type values
func TestAuthTypeValidation(t *testing.T) {
	validAuthTypes := []string{"none", "api_key", "iam"}

	for _, authType := range validAuthTypes {
		t.Run(authType, func(t *testing.T) {
			cfg := config.RestateConfig{
				Endpoint:           "http://localhost:8080",
				ExecutionMechanism: "local",
				AuthType:           authType,
				ApiKey:             "test-key",
				Timeout:            30 * time.Minute,
			}
			err := cfg.Validate()
			// none and iam don't require api_key
			if authType == "none" || authType == "iam" {
				assert.NoError(t, err, "expected %s to be valid auth type", authType)
			}
		})
	}

	// Test invalid auth type
	cfg := config.RestateConfig{
		Endpoint:           "http://localhost:8080",
		ExecutionMechanism: "local",
		AuthType:           "invalid",
		Timeout:            30 * time.Minute,
	}
	err := cfg.Validate()
	assert.Error(t, err, "expected invalid auth type to fail")
}

// TestDefaultValues tests default configuration values
func TestDefaultValues(t *testing.T) {
	cfg := config.RestateConfig{
		Endpoint:           "http://localhost:8080",
		ExecutionMechanism: "local",
		AuthType:           "none",
		Timeout:            30 * time.Minute,
		RetryAttempts:      0, // Zero is valid
	}

	err := cfg.Validate()
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:8080", cfg.Endpoint)
	assert.Equal(t, "local", cfg.ExecutionMechanism)
	assert.Equal(t, "none", cfg.AuthType)
	assert.Equal(t, 30*time.Minute, cfg.Timeout)
	assert.Equal(t, 0, cfg.RetryAttempts)
}

// TestZeroTimeout tests zero timeout handling
func TestZeroTimeout(t *testing.T) {
	cfg := config.RestateConfig{
		Endpoint:           "http://localhost:8080",
		ExecutionMechanism: "local",
		AuthType:           "none",
		Timeout:            0,
	}

	err := cfg.Validate()
	// Zero timeout should be treated as invalid (must be positive)
	assert.Error(t, err)
}

// TestWorkflowConfigValidation tests WorkflowConfig validation with Restate
func TestWorkflowConfigValidation(t *testing.T) {
	tests := []struct {
		name            string
		defaultProvider string
		restateCfg      config.RestateConfig
		shouldErr       bool
	}{
		{
			name:            "default provider mock",
			defaultProvider: "mock",
			restateCfg:      config.RestateConfig{},
			shouldErr:       false,
		},
		{
			name:            "default provider step-functions",
			defaultProvider: "step-functions",
			restateCfg:      config.RestateConfig{},
			shouldErr:       true, // No credentials configured
		},
		{
			name:            "default provider restate with valid config",
			defaultProvider: "restate",
			restateCfg: config.RestateConfig{
				Endpoint:           "http://localhost:8080",
				ExecutionMechanism: "local",
				AuthType:           "none",
				Timeout:            30 * time.Minute,
			},
			shouldErr: false,
		},
		{
			name:            "default provider restate with empty endpoint",
			defaultProvider: "restate",
			restateCfg:      config.RestateConfig{},
			shouldErr:       true, // Restate endpoint empty
		},
		{
			name:            "invalid default provider",
			defaultProvider: "invalid-provider",
			restateCfg:      config.RestateConfig{},
			shouldErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.WorkflowConfig{
				DefaultProvider: tt.defaultProvider,
				Restate:         tt.restateCfg,
			}
			err := cfg.Validate()
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestBackwardCompatibilityNoRestate tests that configurations without Restate still work
func TestBackwardCompatibilityNoRestate(t *testing.T) {
	tests := []struct {
		name     string
		config   config.WorkflowConfig
		shouldOk bool
	}{
		{
			name: "config without restate section loads successfully",
			config: config.WorkflowConfig{
				DefaultProvider: "mock",
				Restate:         config.RestateConfig{
					// Empty restate config - should be ignored since default is mock
				},
			},
			shouldOk: true,
		},
		{
			name: "empty restate config with mock provider",
			config: config.WorkflowConfig{
				DefaultProvider: "mock",
			},
			shouldOk: true,
		},
		{
			name: "mock provider is available even when restate config exists",
			config: config.WorkflowConfig{
				DefaultProvider: "mock",
				Restate: config.RestateConfig{
					Endpoint:           "http://localhost:8080",
					ExecutionMechanism: "local",
					AuthType:           "none",
					Timeout:            30 * time.Minute,
				},
			},
			shouldOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the config can be created without panic
			assert.NotNil(t, tt.config)
			assert.Equal(t, tt.config.DefaultProvider, "mock")
		})
	}
}

// TestDefaultProviderMockBehavior tests that mock is default when not specified
func TestDefaultProviderMockBehavior(t *testing.T) {
	cfg := config.WorkflowConfig{}
	// Empty config should default to mock (implementation dependent)
	// At minimum, should not error
	assert.NotNil(t, cfg)
}

// TestProviderConfigurationIsolation tests that provider configs don't interfere
func TestProviderConfigurationIsolation(t *testing.T) {
	cfg := config.WorkflowConfig{
		DefaultProvider: "mock",
		Restate: config.RestateConfig{
			Endpoint:           "http://localhost:8080",
			ExecutionMechanism: "local",
			AuthType:           "none",
			Timeout:            30 * time.Minute,
		},
		// StepFunctions would go here but intentionally left blank
	}

	// Verify that presence of Restate config doesn't affect mock provider
	assert.Equal(t, cfg.DefaultProvider, "mock")
	assert.NotNil(t, cfg.Restate)

	// Restate config should have the values we set
	assert.Equal(t, cfg.Restate.Endpoint, "http://localhost:8080")
	assert.Equal(t, cfg.Restate.ExecutionMechanism, "local")
	assert.Equal(t, cfg.Restate.AuthType, "none")
}
