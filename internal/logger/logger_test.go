package logger

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		level   string
		wantErr bool
	}{
		{
			name:    "development mode with info level",
			format:  "development",
			level:   "info",
			wantErr: false,
		},
		{
			name:    "production mode with warn level",
			format:  "production",
			level:   "warn",
			wantErr: false,
		},
		{
			name:    "invalid format",
			format:  "invalid",
			level:   "info",
			wantErr: true,
		},
		{
			name:    "invalid level",
			format:  "development",
			level:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.format, tt.level)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if logger == nil {
				t.Error("expected logger but got nil")
			}
		})
	}
}

func TestWithComponent(t *testing.T) {
	logger, err := New("development", "info")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	componentLogger := WithComponent(logger, "test-component")
	if componentLogger == nil {
		t.Error("expected logger with component but got nil")
	}

	// Verify the logger has the component field
	// This is a basic check - in production you'd want more detailed validation
	if componentLogger == logger {
		t.Error("expected new logger instance but got same instance")
	}
}

func TestWith(t *testing.T) {
	logger, err := New("development", "info")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	field := zap.String("key", "value")
	childLogger := With(logger, field)

	if childLogger == nil {
		t.Error("expected child logger but got nil")
	}
	if childLogger == logger {
		t.Error("expected new logger instance but got same instance")
	}
}

func TestLoggerLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error"}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			logger, err := New("production", level)
			if err != nil {
				t.Errorf("failed to create logger with level %s: %v", level, err)
			}

			if logger == nil {
				t.Errorf("expected logger for level %s but got nil", level)
			}

			// Verify the log level is set correctly
			expectedLevel, _ := zapcore.ParseLevel(level)
			if !logger.Core().Enabled(expectedLevel) {
				t.Errorf("logger with level %s should be enabled for %s level", level, level)
			}
		})
	}
}
