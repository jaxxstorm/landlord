package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func validControllerConfig() ControllerConfig {
	return ControllerConfig{
		Enabled:                true,
		ReconciliationInterval: 10 * time.Second,
		StatusPollInterval:     10 * time.Second,
		Workers:                1,
		WorkflowTriggerTimeout: 30 * time.Second,
		ShutdownTimeout:        30 * time.Second,
		MaxRetries:             3,
	}
}

func TestControllerConfigValidateTemporalWorkflowProvider(t *testing.T) {
	cfg := validControllerConfig()
	cfg.WorkflowProvider = "temporal"

	require.NoError(t, cfg.Validate())
}

func TestControllerConfigValidateUnknownWorkflowProvider(t *testing.T) {
	cfg := validControllerConfig()
	cfg.WorkflowProvider = "unknown"

	require.EqualError(t, cfg.Validate(), "workflow_provider must be mock, step-functions, restate, or temporal")
}
