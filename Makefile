.PHONY: build swagger-docs test test-floci floci-e2e-up floci-e2e-down test-floci-e2e-stack test-floci-e2e clean help

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
	@echo "  make test-floci     - Run Docker-backed Floci integration tests"
	@echo "  make floci-e2e-up   - Start the isolated Floci EC2 E2E stack"
	@echo "  make floci-e2e-down - Stop a stack selected by LANDLORD_FLOCI_E2E_ID"
	@echo "  make test-floci-e2e-stack - Verify the stack bootstrap and readiness"
	@echo "  make test-floci-e2e - Run the Floci EC2 tenant lifecycle end-to-end test"
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

test-floci:
	@echo "Running Floci integration tests..."
	@$(GO) test -tags integration ./internal/testutil/floci ./internal/workflow/providers/stepfunctions ./internal/compute/providers/ec2runner ./internal/compute/providers/kubernetes

floci-e2e-up:
	@./test/e2e/floci/stack.sh up

floci-e2e-down:
	@./test/e2e/floci/stack.sh down

test-floci-e2e-stack:
	@./test/e2e/floci/stack.sh verify

test-floci-e2e:
	@$(GO) test -tags e2e -timeout 10m ./test/e2e/floci

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)
	@rm -f $(WORKER_BINARY_NAME)
	@rm -f coverage.out
	@$(GO) clean
