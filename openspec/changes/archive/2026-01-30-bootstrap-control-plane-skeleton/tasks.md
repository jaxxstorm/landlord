## 1. Project Setup

- [x] 1.1 Initialize Go module with appropriate module path
- [x] 1.2 Create project directory structure (cmd/, internal/, migrations/)
- [x] 1.3 Add core dependencies: chi, kong, zap, pgx, golang-migrate

## 2. Configuration Management

- [x] 2.1 Create config struct with kong tags for all configuration parameters
- [x] 2.2 Implement configuration loading from environment variables and CLI flags
- [x] 2.3 Add configuration validation logic
- [x] 2.4 Add database connection configuration (host, port, user, password, database, pool settings)
- [x] 2.5 Add HTTP server configuration (host, port, timeouts, shutdown timeout)
- [x] 2.6 Add logging configuration (level, format/mode)

## 3. Structured Logging

- [x] 3.1 Create logger initialization function with development and production modes
- [x] 3.2 Implement logger factory that returns configured zap logger
- [x] 3.3 Add helper functions for creating child loggers with contextual fields
- [x] 3.4 Create HTTP middleware for request logging with correlation IDs
- [x] 3.5 Add context-aware logging utilities for request-scoped logging

## 4. Database Persistence

- [x] 4.1 Implement database connection pooling using pgxpool
- [x] 4.2 Add connection retry logic with exponential backoff
- [x] 4.3 Create database health check function
- [x] 4.4 Set up golang-migrate integration for schema migrations
- [x] 4.5 Create initial migration file for control plane schema (can be empty placeholder)
- [x] 4.6 Implement migration application on startup with embedded files
- [x] 4.7 Add graceful database connection closure on shutdown

## 5. HTTP API Server

- [x] 5.1 Create chi router with base middleware (logging, recovery)
- [x] 5.2 Implement GET /health endpoint (liveness check)
- [x] 5.3 Implement GET /ready endpoint (readiness check with database health)
- [x] 5.4 Add HTTP server startup with configured host and port
- [x] 5.5 Implement graceful shutdown handling for SIGTERM/SIGINT
- [x] 5.6 Add shutdown timeout with force close fallback
- [x] 5.7 Add request correlation ID middleware

## 6. Main Application

- [x] 6.1 Create main.go in cmd/landlord with application entry point
- [x] 6.2 Wire up configuration loading in main
- [x] 6.3 Initialize logger in main
- [x] 6.4 Initialize database connection in main
- [x] 6.5 Apply database migrations in main
- [x] 6.6 Initialize HTTP server in main
- [x] 6.7 Set up graceful shutdown coordination (wait for HTTP and DB)
- [x] 6.8 Add error handling that bubbles errors to CLI level with exit codes

## 7. Testing

- [ ] 7.1 Add unit tests for configuration loading and validation
- [ ] 7.2 Add unit tests for logger initialization
- [ ] 7.3 Add integration test for database connection and health checks
- [ ] 7.4 Add integration test for HTTP server endpoints
- [ ] 7.5 Add test for graceful shutdown behavior

## 8. Documentation

- [ ] 8.1 Create README.md with project overview and setup instructions
- [ ] 8.2 Document environment variables and configuration options
- [ ] 8.3 Add example .env file or docker-compose for local development
- [ ] 8.4 Document database migration workflow
