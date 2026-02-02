## 1. Dependencies and Project Setup

- [x] 1.1 Add restate.dev Go SDK to go.mod (github.com/restatedev/sdk-go)
- [x] 1.2 Run go mod tidy to resolve dependencies
- [x] 1.3 Verify go.sum is updated with correct checksums
- [x] 1.4 Create docker-compose.yml for local Restate development environment
- [x] 1.5 Add Restate container configuration to docker-compose.yml (port 8080, volume mounts)

## 2. Configuration System Integration

- [x] 2.1 Create RestateConfig struct in internal/config/workflow.go with fields: Endpoint, ExecutionMechanism, ServiceName, AuthType, ApiKey, Timeout, RetryAttempts
- [x] 2.2 Add Restate configuration section to workflow config in internal/config/workflow.go
- [x] 2.3 Implement configuration validation function for Restate settings (endpoint URL format, mechanism values, auth type compatibility)
- [x] 2.4 Add support for environment variable overrides (LANDLORD_WORKFLOW_RESTATE_ENDPOINT, etc.)
- [x] 2.5 Add support for CLI flags for Restate configuration options
- [x] 2.6 Add default values for Restate configuration (endpoint: http://localhost:8080, execution_mechanism: local, auth_type: none)
- [x] 2.7 Validate endpoint is valid HTTP/HTTPS URL during configuration load
- [x] 2.8 Add comprehensive unit tests for Restate configuration loading and validation

## 3. Restate Provider Package Structure

- [x] 3.1 Create internal/workflow/providers/restate/ directory structure
- [x] 3.2 Create internal/workflow/providers/restate/provider.go with Provider struct and Name() method
- [x] 3.3 Create internal/workflow/providers/restate/client.go for Restate SDK client wrapper
- [x] 3.4 Create internal/workflow/providers/restate/config.go for configuration handling
- [x] 3.5 Create internal/workflow/providers/restate/errors.go for error mapping and wrapping
- [x] 3.6 Create internal/workflow/providers/restate/provider_test.go for unit tests
- [x] 3.7 Implement lazy initialization pattern for SDK client (defer creation until first use)

## 4. Core Provider Methods Implementation

- [x] 4.1 Implement Provider.Name() method (returns "restate")
- [x] 4.2 Implement Provider.Validate(ctx, spec) method with workflow spec validation
- [x] 4.3 Implement Provider.CreateWorkflow(ctx, spec) method for service registration
- [x] 4.4 Implement Provider.StartExecution(ctx, workflowID, input) method for workflow invocation
- [x] 4.5 Implement Provider.GetExecutionStatus(ctx, executionID) method for status queries
- [x] 4.6 Implement Provider.StopExecution(ctx, executionID, reason) method for execution cancellation
- [x] 4.7 Implement Provider.DeleteWorkflow(ctx, workflowID) method for service unregistration

## 5. Service Registration and Naming

- [x] 5.1 Implement service name normalization function (kebab-case to PascalCase conversion)
- [x] 5.2 Implement service name validation (alphanumeric, Restate requirements)
- [x] 5.3 Add support for optional service name override from configuration
- [x] 5.4 Ensure service registration is idempotent (duplicate registration succeeds)
- [x] 5.5 Implement service unregistration with idempotency guarantee

## 6. Error Handling and Validation

- [x] 6.1 Implement error mapping from Restate SDK errors to Provider interface errors
- [x] 6.2 Implement connection error handling with endpoint context
- [x] 6.3 Add validation for endpoint URL format at provider initialization
- [x] 6.4 Add validation for execution mechanism against supported values
- [x] 6.5 Add validation for authentication type matching execution mechanism
- [x] 6.6 Implement clear error messages for configuration mismatches
- [x] 6.7 Add error wrapping with structured logging context
- [x] 6.8 Implement graceful error handling for network failures

## 7. Authentication Support

- [x] 7.1 Implement API key authentication method (api_key auth type)
- [x] 7.2 Implement IAM role authentication method (iam auth type)
- [x] 7.3 Implement no-authentication method (none auth type for localhost)
- [x] 7.4 Add authentication validation during provider initialization
- [x] 7.5 Ensure authentication credentials are passed to Restate SDK client
- [x] 7.6 Implement secure credential storage (no logging of API keys)
- [x] 7.7 Add tests for each authentication method

## 8. Provider Registration and Initialization

- [x] 8.1 Create provider initialization function New(config RestateConfig, logger *zap.Logger)
- [x] 8.2 Add restate provider registration to application initialization code (main.go or wire setup)
- [x] 8.3 Ensure restate provider is registered before workflow manager is created
- [x] 8.4 Add logging for provider registration and initialization steps
- [x] 8.5 Test that application starts successfully with restate provider

## 9. Unit Tests - Core Functionality

- [x] 9.1 Write unit tests for Provider.Name() method
- [x] 9.2 Write unit tests for Provider.Validate() with valid and invalid specs
- [x] 9.3 Write unit tests for Provider.CreateWorkflow() including idempotency
- [x] 9.4 Write unit tests for Provider.StartExecution() with various inputs
- [x] 9.5 Write unit tests for Provider.GetExecutionStatus() with different states
- [x] 9.6 Write unit tests for Provider.StopExecution() including idempotency
- [x] 9.7 Write unit tests for Provider.DeleteWorkflow() including idempotency
- [x] 9.8 Write unit tests for RestateConfig validation
- [x] 9.9 Write tests for endpoint URL validation (valid and invalid formats)
- [x] 9.10 Write tests for authentication methods and credential handling

## 10. Backward Compatibility Verification

- [x] 10.1 Verify mock provider workflows still work with restate added
- [x] 10.2 Verify step-functions workflows still work with restate added
- [x] 10.3 Verify existing configuration without restate section still loads
- [x] 10.4 Verify default_provider defaults to mock if not configured
- [x] 10.5 Test provider selection logic when multiple providers available

## 11. Build and Final Verification

- [x] 11.1 Verify code compiles without warnings (go build)
- [x] 11.2 Run go fmt to format code
- [x] 11.3 Run go vet to check for common issues
- [x] 11.4 Verify all unit tests pass (go test)
- [x] 11.5 Verify no breaking changes to existing APIs

## 12. Documentation - Restate Provider Setup and Configuration

- [ ] 12.1 Create docs/workflows/README.md overview of all workflow providers
- [ ] 12.2 Create docs/workflows/restate/README.md overview of Restate provider
- [ ] 12.3 Create docs/workflows/restate/getting-started.md quick start guide
- [ ] 12.4 Create docs/workflows/restate/configuration.md configuration reference
- [ ] 12.5 Create docs/workflows/restate/local-development.md Docker setup guide
- [ ] 12.6 Create docs/workflows/restate/production-deployment.md deployment guide (Lambda, Fargate, Kubernetes)
- [ ] 12.7 Create docs/workflows/restate/authentication.md authentication types guide
- [ ] 12.8 Create docs/workflows/restate/troubleshooting.md common issues and solutions
- [ ] 12.9 Add example configuration files (local and production examples)
- [ ] 12.10 Update main README.md to reference workflow providers documentation
