package config

import "fmt"

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string `mapstructure:"level" env:"LOG_LEVEL" default:"info"`
	Format string `mapstructure:"format" env:"LOG_FORMAT" default:"development"`
}

// Validate validates logging configuration
func (l *LogConfig) Validate() error {
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[l.Level] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", l.Level)
	}
	validFormats := map[string]bool{
		"development": true,
		"production":  true,
	}
	if !validFormats[l.Format] {
		return fmt.Errorf("invalid log format: %s (must be development or production)", l.Format)
	}
	return nil
}
