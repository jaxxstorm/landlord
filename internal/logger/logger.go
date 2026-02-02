package logger

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

const loggerKey contextKey = "logger"

// New creates a new logger based on the given format and level
func New(format string, level string) (*zap.Logger, error) {
	var config zap.Config
	switch format {
	case "development":
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	case "production":
		config = zap.NewProductionConfig()
	default:
		return nil, fmt.Errorf("invalid log format: %s", format)
	}

	// Parse and set log level
	zapLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}
	config.Level = zap.NewAtomicLevelAt(zapLevel)

	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return logger, nil
}

// With creates a child logger with additional fields
func With(logger *zap.Logger, fields ...zap.Field) *zap.Logger {
	return logger.With(fields...)
}

// WithComponent creates a child logger with a component name
func WithComponent(logger *zap.Logger, component string) *zap.Logger {
	return logger.With(zap.String("component", component))
}

// WithContext adds the logger to a context
func WithContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext retrieves the logger from a context
func FromContext(ctx context.Context) *zap.Logger {
	if logger, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
		return logger
	}
	// Return a no-op logger if none is found
	return zap.NewNop()
}

// WithRequestID adds a request ID field to the logger in the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	logger := FromContext(ctx)
	logger = logger.With(zap.String("request_id", requestID))
	return WithContext(ctx, logger)
}

// WithCorrelationID adds a correlation ID field to the logger in the context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	logger := FromContext(ctx)
	logger = logger.With(zap.String("correlation_id", correlationID))
	return WithContext(ctx, logger)
}
