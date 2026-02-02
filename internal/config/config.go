package config

import "fmt"

// Config holds all application configuration
type Config struct {
	Database   DatabaseConfig   `mapstructure:"database"`
	HTTP       HTTPConfig       `mapstructure:"http"`
	Log        LogConfig        `mapstructure:"log"`
	Compute    ComputeConfig    `mapstructure:"compute"`
	Workflow   WorkflowConfig   `mapstructure:"workflow"`
	Controller ControllerConfig `mapstructure:"controller"`
}

// Validate performs validation on the configuration
func (c *Config) Validate() error {
	if err := c.Database.Validate(); err != nil {
		return fmt.Errorf("database config: %w", err)
	}
	if err := c.HTTP.Validate(); err != nil {
		return fmt.Errorf("http config: %w", err)
	}
	if err := c.Log.Validate(); err != nil {
		return fmt.Errorf("log config: %w", err)
	}
	if err := c.Compute.Validate(); err != nil {
		return fmt.Errorf("compute config: %w", err)
	}
	if err := c.Workflow.Validate(); err != nil {
		return fmt.Errorf("workflow config: %w", err)
	}
	if err := c.Controller.Validate(); err != nil {
		return fmt.Errorf("controller config: %w", err)
	}
	return nil
}
