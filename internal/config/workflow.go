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
	Temporal        TemporalConfig      `mapstructure:"temporal"`
}

// StepFunctionsConfig holds AWS Step Functions provider configuration
type StepFunctionsConfig struct {
	Region           string                         `mapstructure:"region" env:"WORKFLOW_SFN_REGION" default:"us-west-2"`
	StateMachineARN  string                         `mapstructure:"state_machine_arn" env:"WORKFLOW_SFN_STATE_MACHINE_ARN"`
	CallerAssumeRole *StepFunctionsAssumeRoleConfig `mapstructure:"caller_assume_role"`
}

// StepFunctionsAssumeRoleConfig configures an optional role used by Landlord
// when it calls the Step Functions control-plane API.
type StepFunctionsAssumeRoleConfig struct {
	RoleARN     string `mapstructure:"role_arn"`
	ExternalID  string `mapstructure:"external_id"`
	SessionName string `mapstructure:"session_name"`
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

// TemporalConfig holds Temporal workflow provider configuration.
type TemporalConfig struct {
	HostPort      string        `mapstructure:"host_port" env:"WORKFLOW_TEMPORAL_HOST_PORT" default:"localhost:7233"`
	Namespace     string        `mapstructure:"namespace" env:"WORKFLOW_TEMPORAL_NAMESPACE" default:"default"`
	TaskQueue     string        `mapstructure:"task_queue" env:"WORKFLOW_TEMPORAL_TASK_QUEUE" default:"landlord"`
	Timeout       time.Duration `mapstructure:"timeout" env:"WORKFLOW_TEMPORAL_TIMEOUT" default:"30m"`
	RetryAttempts int           `mapstructure:"retry_attempts" env:"WORKFLOW_TEMPORAL_RETRY_ATTEMPTS" default:"3"`
}

// Validate validates workflow configuration
func (w *WorkflowConfig) Validate() error {
	// Validate default provider value
	validProviders := map[string]bool{"mock": true, "step-functions": true, "restate": true, "temporal": true}
	if !validProviders[w.DefaultProvider] {
		return fmt.Errorf("invalid default_provider: %s (must be mock, step-functions, restate, or temporal)", w.DefaultProvider)
	}

	// If Step Functions is the default provider, ensure its execution target is configured.
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

	// If Temporal is the default provider, ensure configuration is valid
	if w.DefaultProvider == "temporal" {
		if err := w.Temporal.Validate(); err != nil {
			return fmt.Errorf("temporal config: %w", err)
		}
	}

	// Always validate Restate config if it's provided, even if not default
	if w.Restate.Endpoint != "" {
		if err := w.Restate.Validate(); err != nil {
			return fmt.Errorf("restate config: %w", err)
		}
	}

	// Always validate Temporal config if it's partially provided, even when not default.
	if w.Temporal.HostPort != "" || w.Temporal.Namespace != "" || w.Temporal.TaskQueue != "" {
		if err := w.Temporal.Validate(); err != nil {
			return fmt.Errorf("temporal config: %w", err)
		}
	}

	return nil
}

// Validate validates Step Functions configuration.
func (s *StepFunctionsConfig) Validate() error {
	if s.Region == "" {
		return fmt.Errorf("region is required for Step Functions provider")
	}
	if s.StateMachineARN == "" {
		return fmt.Errorf("state machine ARN is required for Step Functions provider")
	}
	if s.CallerAssumeRole != nil && s.CallerAssumeRole.RoleARN == "" {
		return fmt.Errorf("caller assume-role ARN is required when caller assume-role is configured")
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

// Validate validates Temporal configuration.
func (t *TemporalConfig) Validate() error {
	if t.HostPort == "" {
		return fmt.Errorf("host_port is required for Temporal provider")
	}
	if t.Namespace == "" {
		return fmt.Errorf("namespace is required for Temporal provider")
	}
	if t.TaskQueue == "" {
		return fmt.Errorf("task_queue is required for Temporal provider")
	}
	if t.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	if t.RetryAttempts < 0 {
		return fmt.Errorf("retry_attempts must be non-negative")
	}
	return nil
}
