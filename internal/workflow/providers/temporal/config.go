package temporal

import (
	"fmt"

	"github.com/jaxxstorm/landlord/internal/config"
)

// ValidateConfig validates temporal-specific provider configuration.
func ValidateConfig(cfg config.TemporalConfig) error {
	if cfg.HostPort == "" {
		return fmt.Errorf("host_port is required")
	}
	if cfg.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}
	if cfg.TaskQueue == "" {
		return fmt.Errorf("task_queue is required")
	}
	if cfg.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	if cfg.RetryAttempts < 0 {
		return fmt.Errorf("retry_attempts must be non-negative")
	}
	return nil
}
