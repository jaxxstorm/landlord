package logger

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// HTTPMiddleware creates middleware that logs HTTP requests
func HTTPMiddleware(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Get or generate request ID
			requestID := middleware.GetReqID(r.Context())
			if requestID == "" {
				requestID = fmt.Sprintf("%d", middleware.NextRequestID())
			}

			// Get correlation ID from header if present
			correlationID := r.Header.Get("X-Correlation-ID")
			if correlationID == "" {
				correlationID = requestID
			}

			// Create request-scoped logger
			reqLogger := logger.With(
				zap.String("request_id", requestID),
				zap.String("correlation_id", correlationID),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
			)

			// Add logger to context
			ctx := WithContext(r.Context(), reqLogger)
			r = r.WithContext(ctx)

			// Wrap response writer to capture status code
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Process request
			next.ServeHTTP(ww, r)

			// Log request completion
			duration := time.Since(start)
			reqLogger.Info("http request",
				zap.Int("status", ww.Status()),
				zap.Int("bytes", ww.BytesWritten()),
				zap.Duration("duration", duration),
				zap.String("duration_ms", fmt.Sprintf("%.2f", float64(duration.Milliseconds()))),
			)
		})
	}
}

// CorrelationIDMiddleware adds correlation ID to response headers
func CorrelationIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get("X-Correlation-ID")
		if correlationID == "" {
			correlationID = middleware.GetReqID(r.Context())
		}
		w.Header().Set("X-Correlation-ID", correlationID)
		next.ServeHTTP(w, r)
	})
}
