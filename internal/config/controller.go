package config

import (
	"fmt"
	"time"
)

// ControllerConfig holds configuration for the reconciliation controller
type ControllerConfig struct {
	// Enabled controls whether the controller is started
	Enabled bool `mapstructure:"enabled"`

	// WorkflowProvider overrides the workflow provider used by the controller
	// If empty, workflow.default_provider is used.
	WorkflowProvider string `mapstructure:"workflow_provider"`

	// ReconciliationInterval is how often to poll for tenants requiring reconciliation
	ReconciliationInterval time.Duration `mapstructure:"reconciliation_interval"`

	// StatusPollInterval is how often to poll for in-flight workflow status
	StatusPollInterval time.Duration `mapstructure:"status_poll_interval"`

	// Workers is the number of concurrent worker goroutines
	Workers int `mapstructure:"workers"`

	// WorkflowTriggerTimeout is the timeout for triggering workflows
	WorkflowTriggerTimeout time.Duration `mapstructure:"workflow_trigger_timeout"`

	// ShutdownTimeout is the maximum time to wait for graceful shutdown
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`

	// MaxRetries is the maximum number of retry attempts before marking a tenant as failed
	MaxRetries int `mapstructure:"max_retries"`
}

// Validate checks the controller configuration
func (c *ControllerConfig) Validate() error {
	if c.Enabled {
		if c.WorkflowProvider != "" {
			validProviders := map[string]bool{"mock": true, "step-functions": true, "restate": true}
			if !validProviders[c.WorkflowProvider] {
				return fmt.Errorf("workflow_provider must be mock, step-functions, or restate")
			}
		}
		if c.ReconciliationInterval <= 0 {
			return fmt.Errorf("reconciliation_interval must be positive")
		}
		if c.StatusPollInterval <= 0 {
			return fmt.Errorf("status_poll_interval must be positive")
		}
		if c.Workers <= 0 {
			return fmt.Errorf("workers must be positive")
		}
		if c.WorkflowTriggerTimeout <= 0 {
			return fmt.Errorf("workflow_trigger_timeout must be positive")
		}
		if c.ShutdownTimeout <= 0 {
			return fmt.Errorf("shutdown_timeout must be positive")
		}
		if c.MaxRetries < 0 {
			return fmt.Errorf("max_retries must be non-negative")
		}
	}
	return nil
}

// SetDefaults sets default values for controller configuration
func (c *ControllerConfig) SetDefaults() {
	if c.ReconciliationInterval == 0 {
		c.ReconciliationInterval = 10 * time.Second
	}
	if c.StatusPollInterval == 0 {
		c.StatusPollInterval = c.ReconciliationInterval
	}
	if c.Workers == 0 {
		c.Workers = 3
	}
	if c.WorkflowTriggerTimeout == 0 {
		c.WorkflowTriggerTimeout = 30 * time.Second
	}
	if c.ShutdownTimeout == 0 {
		c.ShutdownTimeout = 30 * time.Second
	}
	if c.MaxRetries == 0 {
		c.MaxRetries = 5
	}
}
