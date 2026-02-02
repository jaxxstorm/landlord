package config

import (
	"fmt"
	"net/url"
	"time"
)

// WorkflowConfig holds workflow orchestration configuration
type WorkflowConfig struct {
	DefaultProvider string              `mapstructure:"default_provider" env:"WORKFLOW_DEFAULT_PROVIDER" default:"mock"`
	StepFunctions   StepFunctionsConfig `mapstructure:"step_functions"`
	Restate         RestateConfig       `mapstructure:"restate"`
}

// StepFunctionsConfig holds AWS Step Functions provider configuration
type StepFunctionsConfig struct {
	Region  string `mapstructure:"region" env:"WORKFLOW_SFN_REGION" default:"us-west-2"`
	RoleARN string `mapstructure:"role_arn" env:"WORKFLOW_SFN_ROLE_ARN"`
}

// RestateConfig holds Restate.dev workflow provider configuration
type RestateConfig struct {
	Endpoint           string        `mapstructure:"endpoint" env:"WORKFLOW_RESTATE_ENDPOINT" default:"http://localhost:8080"`
	AdminEndpoint      string        `mapstructure:"admin_endpoint" env:"WORKFLOW_RESTATE_ADMIN_ENDPOINT" default:"http://localhost:9070"`
	ExecutionMechanism string        `mapstructure:"execution_mechanism" env:"WORKFLOW_RESTATE_EXECUTION_MECHANISM" default:"local"`
	ServiceName        string        `mapstructure:"service_name" env:"WORKFLOW_RESTATE_SERVICE_NAME"`
	AuthType           string        `mapstructure:"auth_type" env:"WORKFLOW_RESTATE_AUTH_TYPE" default:"none"`
	ApiKey             string        `mapstructure:"api_key" env:"WORKFLOW_RESTATE_API_KEY"`
	Timeout            time.Duration `mapstructure:"timeout" env:"WORKFLOW_RESTATE_TIMEOUT" default:"30m"`
	RetryAttempts      int           `mapstructure:"retry_attempts" env:"WORKFLOW_RESTATE_RETRY_ATTEMPTS" default:"3"`

	WorkerRegisterOnStartup bool          `mapstructure:"worker_register_on_startup" env:"WORKFLOW_RESTATE_WORKER_REGISTER_ON_STARTUP" default:"true"`
	WorkerAdminEndpoint     string        `mapstructure:"worker_admin_endpoint" env:"WORKFLOW_RESTATE_WORKER_ADMIN_ENDPOINT"`
	WorkerNamespace         string        `mapstructure:"worker_namespace" env:"WORKFLOW_RESTATE_WORKER_NAMESPACE"`
	WorkerServicePrefix     string        `mapstructure:"worker_service_prefix" env:"WORKFLOW_RESTATE_WORKER_SERVICE_PREFIX"`
	WorkerLandlordAPIURL    string        `mapstructure:"worker_landlord_api_url" env:"WORKFLOW_RESTATE_WORKER_LANDLORD_API_URL"`
	WorkerComputeProvider   string        `mapstructure:"worker_compute_provider" env:"WORKFLOW_RESTATE_WORKER_COMPUTE_PROVIDER"`
	WorkerComputeCacheTTL   time.Duration `mapstructure:"worker_compute_cache_ttl" env:"WORKFLOW_RESTATE_WORKER_COMPUTE_CACHE_TTL" default:"5m"`
	WorkerAdvertisedURL     string        `mapstructure:"worker_advertised_url" env:"WORKFLOW_RESTATE_WORKER_ADVERTISED_URL"`
}

// Validate validates workflow configuration
func (w *WorkflowConfig) Validate() error {
	// Validate default provider value
	validProviders := map[string]bool{"mock": true, "step-functions": true, "restate": true}
	if !validProviders[w.DefaultProvider] {
		return fmt.Errorf("invalid default_provider: %s (must be mock, step-functions, or restate)", w.DefaultProvider)
	}

	// If Step Functions is the default provider, ensure RoleARN is configured
	if w.DefaultProvider == "step-functions" {
		if err := w.StepFunctions.Validate(); err != nil {
			return fmt.Errorf("step functions config: %w", err)
		}
	}

	// If Restate is the default provider, ensure configuration is valid
	if w.DefaultProvider == "restate" {
		if err := w.Restate.Validate(); err != nil {
			return fmt.Errorf("restate config: %w", err)
		}
	}

	// Always validate Restate config if it's provided, even if not default
	if w.Restate.Endpoint != "" {
		if err := w.Restate.Validate(); err != nil {
			return fmt.Errorf("restate config: %w", err)
		}
	}

	return nil
}

// Validate validates Step Functions configuration
func (s *StepFunctionsConfig) Validate() error {
	if s.RoleARN == "" {
		return fmt.Errorf("role ARN is required for Step Functions provider")
	}
	if s.Region == "" {
		return fmt.Errorf("region is required for Step Functions provider")
	}
	return nil
}

// Validate validates Restate configuration
func (r *RestateConfig) Validate() error {
	if r.Endpoint == "" {
		return fmt.Errorf("endpoint is required for Restate provider")
	}

	// Validate endpoint URL format
	if err := validateEndpointURL(r.Endpoint); err != nil {
		return fmt.Errorf("invalid endpoint URL: %w", err)
	}

	if r.AdminEndpoint != "" {
		if err := validateEndpointURL(r.AdminEndpoint); err != nil {
			return fmt.Errorf("invalid admin endpoint URL: %w", err)
		}
	}

	// Validate execution mechanism
	validMechanisms := map[string]bool{
		"local":       true,
		"lambda":      true,
		"fargate":     true,
		"kubernetes":  true,
		"self-hosted": true,
	}
	if r.ExecutionMechanism != "" && !validMechanisms[r.ExecutionMechanism] {
		return fmt.Errorf("invalid execution_mechanism: %s (must be local, lambda, fargate, kubernetes, or self-hosted)", r.ExecutionMechanism)
	}

	// Validate auth type
	validAuthTypes := map[string]bool{"none": true, "api_key": true, "iam": true}
	if r.AuthType != "" && !validAuthTypes[r.AuthType] {
		return fmt.Errorf("invalid auth_type: %s (must be none, api_key, or iam)", r.AuthType)
	}

	// Validate auth type matches execution mechanism
	if err := validateAuthMechanism(r.ExecutionMechanism, r.AuthType, r.ApiKey); err != nil {
		return err
	}

	// Validate timeout is positive (must be > 0)
	if r.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	// Validate retry attempts is non-negative
	if r.RetryAttempts < 0 {
		return fmt.Errorf("retry_attempts must be non-negative")
	}

	if r.WorkerAdminEndpoint != "" {
		if err := validateEndpointURL(r.WorkerAdminEndpoint); err != nil {
			return fmt.Errorf("invalid worker admin endpoint URL: %w", err)
		}
	}

	if r.WorkerLandlordAPIURL != "" {
		if err := validateEndpointURL(r.WorkerLandlordAPIURL); err != nil {
			return fmt.Errorf("invalid worker landlord api url: %w", err)
		}
	}

	if r.WorkerAdvertisedURL != "" {
		if err := validateEndpointURL(r.WorkerAdvertisedURL); err != nil {
			return fmt.Errorf("invalid worker advertised url: %w", err)
		}
	}

	if r.WorkerComputeCacheTTL < 0 {
		return fmt.Errorf("worker_compute_cache_ttl must be non-negative")
	}

	return nil
}

// validateEndpointURL validates the endpoint is a valid HTTP/HTTPS URL
func validateEndpointURL(endpoint string) error {
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("failed to parse endpoint: %w", err)
	}

	if parsedURL.Scheme == "" {
		return fmt.Errorf("endpoint must include scheme (http or https)")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("endpoint scheme must be http or https, got %s", parsedURL.Scheme)
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("endpoint must include host")
	}

	return nil
}

// validateAuthMechanism validates that auth type is appropriate for the execution mechanism
func validateAuthMechanism(mechanism, authType, apiKey string) error {
	// If no mechanism specified or localhost endpoint, none auth is fine
	if mechanism == "local" || mechanism == "" {
		return nil
	}

	// Lambda and Fargate typically use IAM
	if mechanism == "lambda" || mechanism == "fargate" {
		if authType == "api_key" && apiKey == "" {
			return fmt.Errorf("api_key auth_type requires api_key to be configured")
		}
		return nil
	}

	// Kubernetes can use either api_key or iam
	if mechanism == "kubernetes" {
		if authType == "api_key" && apiKey == "" {
			return fmt.Errorf("api_key auth_type requires api_key to be configured")
		}
		return nil
	}

	// Self-hosted typically uses API key
	if mechanism == "self-hosted" {
		if authType == "api_key" && apiKey == "" {
			return fmt.Errorf("api_key auth_type requires api_key to be configured")
		}
		return nil
	}

	return nil
}
