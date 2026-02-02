.PHONY: build swagger-docs test clean help

# Build variables
BINARY_NAME=landlord
WORKER_BINARY_NAME=landlord-worker
GO=go

help:
	@echo "Available targets:"
	@echo "  make build          - Build the application"
	@echo "  make build-worker   - Build the workflow worker"
	@echo "  make swagger-docs   - Generate Swagger/OpenAPI documentation"
	@echo "  make test           - Run tests"
	@echo "  make clean          - Clean build artifacts"

# Generate Swagger documentation
swagger-docs:
	@echo "Generating Swagger documentation..."
	@~/go/bin/swag init -g internal/api/server.go
	@echo "Swagger documentation generated in docs/"

# Build the application with swagger docs
build: swagger-docs
	@echo "Building $(BINARY_NAME)..."
	@$(GO) build -o $(BINARY_NAME) ./cmd/landlord

# Build the worker
build-worker:
	@echo "Building $(WORKER_BINARY_NAME)..."
	@$(GO) build -o $(WORKER_BINARY_NAME) ./cmd/workers/restate

# Run tests
test:
	@echo "Running tests..."
	@$(GO) test -v -race -coverprofile=coverage.out ./...

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)
	@rm -f $(WORKER_BINARY_NAME)
	@rm -f coverage.out
	@$(GO) clean
