// Package api provides HTTP API server and request handlers.
// @title Landlord API
// @version 1.0
// @description HTTP API for landlord application
// @basePath /v1
// @schemes http https
// @consumes application/json
// @produces application/json
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/jaxxstorm/landlord/internal/apiversion"
	"github.com/jaxxstorm/landlord/internal/compute"
	"github.com/jaxxstorm/landlord/internal/config"
	"github.com/jaxxstorm/landlord/internal/database"
	"github.com/jaxxstorm/landlord/internal/logger"
	"github.com/jaxxstorm/landlord/internal/tenant"
	"github.com/jaxxstorm/landlord/internal/workflow"
)

// Server represents the HTTP API server
type Server struct {
	router          *chi.Mux
	server          *http.Server
	provider        database.Provider
	computeProvider compute.Provider
	tenantRepo      tenant.Repository
	controller      ControllerHealthChecker
	workflowClient  WorkflowClient
	logger          *zap.Logger
}

// ControllerHealthChecker defines the interface for checking controller health
type ControllerHealthChecker interface {
	IsReady() bool
}

// WorkflowClient defines the interface for triggering workflows from API
type WorkflowClient interface {
	TriggerWorkflow(ctx context.Context, t *tenant.Tenant, action string) (string, error)
	TriggerWorkflowWithSource(ctx context.Context, t *tenant.Tenant, action, triggerSource string) (string, error)
	DetermineAction(status tenant.Status) (string, error)
	GetExecutionStatus(ctx context.Context, executionID string) (*workflow.ExecutionStatus, error)
}

// New creates a new HTTP API server
func New(cfg *config.HTTPConfig, dbProvider database.Provider, computeProvider compute.Provider, tenantRepo tenant.Repository, workflowClient WorkflowClient, log *zap.Logger) *Server {
	log = log.With(zap.String("component", "api"))

	r := chi.NewRouter()

	// Base middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(logger.HTTPMiddleware(log))
	r.Use(logger.CorrelationIDMiddleware)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	srv := &Server{
		router:          r,
		provider:        dbProvider,
		computeProvider: computeProvider,
		tenantRepo:      tenantRepo,
		controller:      nil, // Set later with SetController()
		workflowClient:  workflowClient,
		logger:          log,
		server: &http.Server{
			Addr:         cfg.Address(),
			Handler:      r,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}

	// Register routes
	srv.registerRoutes()

	return srv
}

// SetController sets the controller health checker
func (s *Server) SetController(controller ControllerHealthChecker) {
	s.controller = controller
}

// registerRoutes registers all HTTP routes
func (s *Server) registerRoutes() {
	s.router.Get("/health", s.handleHealth)
	s.router.Get("/ready", s.handleReady)

	s.router.Route("/"+apiversion.Current, func(r chi.Router) {
		r.Get("/swagger.json", s.handleSwaggerSpec)
		r.Get("/docs", s.handleDocsUI)

		// Compute config routes
		r.Get("/compute/config", s.handleComputeConfigDiscovery)

		// Tenant routes
		r.Post("/tenants", s.handleCreateTenant)
		r.Get("/tenants", s.handleListTenants)
		r.Get("/tenants/{id}", s.handleGetTenant)
		r.Put("/tenants/{id}", s.handleUpdateTenant)
		r.Post("/tenants/{id}/archive", s.handleArchiveTenant)
		r.Delete("/tenants/{id}", s.handleDeleteTenant)
	})

	s.router.Route("/api", func(r chi.Router) {
		r.Handle("/", http.HandlerFunc(s.handleVersionRequired))
		r.Handle("/*", http.HandlerFunc(s.handleVersionRequired))
	})

	s.router.Route("/v{version}", func(r chi.Router) {
		r.Handle("/", http.HandlerFunc(s.handleUnsupportedVersion))
		r.Handle("/*", http.HandlerFunc(s.handleUnsupportedVersion))
	})
}

// handleHealth is the liveness check endpoint
// @Summary Health check
// @Description Returns server health status
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string "Server health status"
// @Router /health [get]
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleReady is the readiness check endpoint
// @Summary Readiness check
// @Description Returns server readiness status and dependency health
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{} "Server is ready"
// @Failure 503 {object} map[string]interface{} "Server is unavailable"
// @Router /ready [get]
func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	checks := make(map[string]string)

	// Check database health
	if err := s.provider.Health(ctx); err != nil {
		s.logger.Warn("readiness check failed: database unhealthy", zap.Error(err))
		checks["database"] = "unhealthy"
		response := map[string]interface{}{
			"status": "unavailable",
			"checks": checks,
			"error":  err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(response)
		return
	}
	checks["database"] = "healthy"

	// Check controller health if enabled
	if s.controller != nil {
		if s.controller.IsReady() {
			checks["controller"] = "ready"
		} else {
			checks["controller"] = "not_ready"
			response := map[string]interface{}{
				"status": "unavailable",
				"checks": checks,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	response := map[string]interface{}{
		"status": "ready",
		"checks": checks,
		"time":   time.Now().UTC().Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleSwaggerSpec serves the OpenAPI specification
// @Summary OpenAPI specification
// @Description Returns the OpenAPI 3.0 specification for the API
// @Tags documentation
// @Produce json
// @Success 200 {object} map[string]interface{} "OpenAPI specification"
// @Router /v1/swagger.json [get]
func (s *Server) handleSwaggerSpec(w http.ResponseWriter, r *http.Request) {
	// Note: This handler serves the generated swagger.json file
	// The file is generated by swag init and should be served from the docs directory
	http.ServeFile(w, r, "docs/swagger.json")
}

// handleDocsUI serves the API documentation UI
// @Summary API documentation
// @Description Serves the interactive API documentation using Redoc
// @Tags documentation
// @Produce html
// @Success 200 "API documentation HTML page"
// @Router /v1/docs [get]
func (s *Server) handleDocsUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	// Language: html
	html := `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Landlord API Docs</title>
  <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
  <style>
    html, body {
      height: 100%;
      margin: 0;
      padding: 0;
      font-family: sans-serif;
    }
    #redoc-container {
      height: 100%;
    }
  </style>
</head>
<body>
  <div id="redoc-container"></div>
  <script>
    Redoc.init('/v1/swagger.json', {
      scrollYOffset: 50,
      hideLoading: false,
    }, document.getElementById('redoc-container'));
  </script>
</body>
</html>`
	w.Write([]byte(html))
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Info("starting HTTP server", zap.String("address", s.server.Addr))
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed: %w", err)
	}
	return nil
}

// Shutdown gracefully shuts down the HTTP server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down HTTP server")
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("server shutdown failed", zap.Error(err))
		return fmt.Errorf("server shutdown failed: %w", err)
	}
	s.logger.Info("HTTP server shut down successfully")
	return nil
}
