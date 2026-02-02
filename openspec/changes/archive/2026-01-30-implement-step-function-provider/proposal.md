## Why

The landlord control plane requires a concrete workflow provider implementation to orchestrate tenant provisioning workflows. While the workflow provider interface and architecture are defined, we currently only have a mock provider for testing. AWS Step Functions is the initial target workflow engine and must be implemented to enable real tenant orchestration.

## What Changes

- Implement AWS Step Functions workflow provider implementing the workflow.Provider interface
- Add AWS SDK v2 client integration for Step Functions state machine management
- Implement workflow creation with ASL (Amazon States Language) validation
- Add execution tracking and state management for running workflows
- Implement execution status queries with history event translation
- Add IAM role configuration and validation
- Provide idempotent workflow creation and execution operations
- Add comprehensive tests using AWS SDK mocks
- Wire provider into main.go and registry
- Add AWS-specific configuration (region, role ARN, tags)
- Step functions uses lambda functions for its invocation, so define lambda functions and initialize them on creation of the provider, and delete them on tear down
- Store provider status in the database for persistence, including the status of the lambda functions in this case

## Capabilities

### New Capabilities
- `step-functions-provider`: AWS Step Functions workflow provider implementation with state machine management, execution tracking, ASL validation, and IAM integration

### Modified Capabilities
<!-- None - this is a new provider implementation of existing interfaces -->

## Impact

**New Code:**
- `internal/workflow/stepfunctions/` package with provider implementation
- Provider tests with AWS SDK mocks
- AWS client initialization and configuration

**Modified Code:**
- `cmd/landlord/main.go`: Register Step Functions provider in workflow registry
- `internal/config/config.go`: Add AWS workflow configuration (region, role ARN)

**Dependencies:**
- Add `github.com/aws/aws-sdk-go-v2` for Step Functions API
- Add `github.com/aws/aws-sdk-go-v2/service/sfn` for state machine operations

**No Breaking Changes** - This adds a new provider without modifying existing interfaces or behavior.
