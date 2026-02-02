# Specification: Workflow Provisioning (Delta)

## Overview

This delta specification documents changes to the workflow provisioning capability to add Restate as a third workflow provider option alongside step-functions and mock.

## MODIFIED Requirements

### Requirement: Provider Interface Examples

The Provider interface documentation SHALL include restate as a valid provider name example.

**Modified in**: `openspec/specs/workflow-provisioning/01-provider-interface.md`

The Name() method comment SHALL be updated from:
```go
// Examples: "mock", "step-functions", "temporal"
```

To:
```go
// Examples: "mock", "step-functions", "restate"
```

#### Scenario: Provider name examples
- **WHEN** developers reference the Provider interface documentation
- **THEN** they SHALL see "restate" listed as a valid provider name alongside "mock" and "step-functions"

### Requirement: Registry Example Provider List

The Registry.List() documentation SHALL include restate in example output.

**Modified in**: `openspec/specs/workflow-provisioning/03-registry-and-manager.md`

The List() method example SHALL be updated from:
```go
providers := registry.List()
// Returns: ["mock", "step-functions", "temporal"]
```

To:
```go
providers := registry.List()
// Returns: ["mock", "restate", "step-functions"]
```

#### Scenario: Registry list output
- **WHEN** List() is called with mock, restate, and step-functions registered
- **THEN** the output SHALL include all three providers in sorted order

### Requirement: Default Provider Configuration

The workflow configuration SHALL support "restate" as a valid default_provider value.

**Modified in**: Configuration documentation (to be created/updated)

The workflow.default_provider configuration SHALL accept:
- "mock" - for testing
- "step-functions" - for AWS Step Functions
- "restate" - for Restate.dev

#### Scenario: Restate as default provider
- **WHEN** workflow.default_provider is set to "restate"
- **THEN** the system SHALL use the restate provider for all workflow operations
- **AND** SHALL load restate-specific configuration from workflow.restate section

#### Scenario: Provider selection validation
- **WHEN** a WorkflowSpec specifies provider_type as "restate"
- **THEN** the Manager SHALL route the request to the restate provider
- **AND** SHALL validate the provider is registered

### Requirement: Provider Registration

The system SHALL register the restate provider alongside existing providers during initialization.

**Modified in**: Application initialization code

#### Scenario: Restate provider registration
- **WHEN** the application initializes the workflow registry
- **THEN** it SHALL call registry.Register() for the restate provider
- **AND** the registration SHALL occur before the workflow manager is created
- **AND** the restate provider SHALL be available for subsequent workflow operations

#### Scenario: Multiple provider registration
- **WHEN** all providers are registered
- **THEN** registry.List() SHALL return a list containing "mock", "restate", and "step-functions"
- **AND** each provider SHALL be independently accessible via registry.Get()

### Requirement: Configuration Schema

The workflow configuration schema SHALL include a restate section for restate-specific settings.

**Modified in**: Configuration types and validation

#### Scenario: Restate configuration section
- **WHEN** workflow configuration is loaded
- **THEN** it SHALL include a workflow.restate section with fields:
  - endpoint (string): Restate server URL
  - execution_mechanism (string): Deployment target
  - service_name (string): Service identifier
  - auth_type (string): Authentication method
  - api_key (string): API key for authentication
  - timeout (duration): Default workflow timeout
  - retry_attempts (int): Retry attempts for operations

#### Scenario: Configuration validation
- **WHEN** restate is specified as the provider
- **THEN** the system SHALL validate workflow.restate configuration is present
- **AND** SHALL validate required fields (endpoint) are configured
- **AND** SHALL validate optional fields have sensible defaults

### Requirement: Provider Documentation

The workflow provisioning documentation SHALL include restate provider setup and usage examples.

**Modified in**: Documentation files (to be created/updated)

#### Scenario: Provider options documentation
- **WHEN** developers consult workflow provider documentation
- **THEN** they SHALL see restate listed alongside mock and step-functions
- **AND** SHALL see configuration examples for all three providers

#### Scenario: Restate-specific documentation
- **WHEN** developers want to use the restate provider
- **THEN** documentation SHALL provide:
  - Local development setup (Docker Compose example)
  - Configuration reference for all restate settings
  - Production deployment patterns for each execution mechanism
  - Migration guide from mock or step-functions to restate

## ADDED Requirements

### Requirement: Provider Count

The system SHALL support exactly three workflow providers: mock, step-functions, and restate.

#### Scenario: Provider count validation
- **WHEN** the registry is initialized with all providers
- **THEN** registry.List() SHALL return exactly 3 providers
- **AND** SHALL include "mock", "restate", and "step-functions"

#### Scenario: Provider independence
- **WHEN** workflows are created using different providers
- **THEN** each provider SHALL operate independently
- **AND** workflows on one provider SHALL not affect workflows on other providers

### Requirement: Restate Configuration Loading

The system SHALL load restate configuration from the workflow configuration section.

#### Scenario: Configuration structure
- **WHEN** configuration is loaded from YAML/JSON
- **THEN** restate settings SHALL be nested under workflow.restate
- **AND** SHALL follow the same structure as other provider configurations

#### Scenario: Environment variable override
- **WHEN** LANDLORD_WORKFLOW_RESTATE_ENDPOINT is set as an environment variable
- **THEN** it SHALL override the endpoint value from config files
- **AND** SHALL follow standard Landlord configuration precedence rules

#### Scenario: CLI flag override
- **WHEN** --workflow.restate.endpoint is provided as a CLI flag
- **THEN** it SHALL override both environment variables and config files
- **AND** SHALL be the highest precedence configuration source

### Requirement: Provider Selection Logic

The system SHALL route workflow operations to the correct provider based on ProviderType or default_provider configuration.

#### Scenario: Explicit provider type
- **WHEN** a WorkflowSpec specifies provider_type as "restate"
- **THEN** the Manager SHALL use registry.Get("restate") to retrieve the provider
- **AND** SHALL execute the operation using the restate provider

#### Scenario: Default provider fallback
- **WHEN** a WorkflowSpec does not specify provider_type
- **THEN** the Manager SHALL use the configured workflow.default_provider
- **AND** SHALL use "mock" if default_provider is not configured

#### Scenario: Provider not found
- **WHEN** a WorkflowSpec specifies a provider_type that is not registered
- **THEN** the Manager SHALL return ErrProviderNotFound
- **AND** SHALL include the requested provider name in the error message

### Requirement: Backward Compatibility

The addition of the restate provider SHALL maintain backward compatibility with existing mock and step-functions configurations.

#### Scenario: Existing configurations unchanged
- **WHEN** existing configurations for mock or step-functions are loaded
- **THEN** they SHALL continue to work without modification
- **AND** SHALL not require restate configuration to be present

#### Scenario: Default provider unchanged
- **WHEN** workflow.default_provider is not configured
- **THEN** the system SHALL continue to default to "mock"
- **AND** SHALL not change existing behavior

#### Scenario: Migration path
- **WHEN** teams want to migrate from mock or step-functions to restate
- **THEN** they SHALL be able to change workflow.default_provider to "restate"
- **AND** SHALL configure workflow.restate settings
- **AND** SHALL not need to modify existing workflow definitions
