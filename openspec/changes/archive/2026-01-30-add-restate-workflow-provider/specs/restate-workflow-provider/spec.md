# Specification: Restate Workflow Provider

## Overview

This specification defines the Restate workflow provider implementation for Landlord. Restate.dev provides durable execution with strong consistency guarantees and supports both local development (without cloud dependencies) and production deployment across multiple execution mechanisms (Lambda, ECS Fargate, Kubernetes, self-hosted).

The provider integrates with the Restate.dev Go SDK to implement the Provider interface, enabling workflows to run on Restate's durable execution platform.

## ADDED Requirements

### Requirement: Provider Implementation

The system SHALL implement a Restate provider that satisfies the workflow.Provider interface.

#### Scenario: Provider registration
- **WHEN** the application initializes the workflow registry
- **THEN** the restate provider SHALL be registered with name "restate"

#### Scenario: Provider identification
- **WHEN** Name() is called on the restate provider
- **THEN** it SHALL return "restate"

#### Scenario: Workflow creation
- **WHEN** CreateWorkflow() is called with a valid WorkflowSpec
- **THEN** the provider SHALL register the workflow with the Restate server
- **AND** SHALL return a CreateWorkflowResult with the workflow registration details

#### Scenario: Execution start
- **WHEN** StartExecution() is called with a workflow ID and execution input
- **THEN** the provider SHALL invoke the workflow on the Restate server
- **AND** SHALL return an ExecutionResult with execution ID and initial state

#### Scenario: Execution status query
- **WHEN** GetExecutionStatus() is called with an execution ID
- **THEN** the provider SHALL query the Restate server for execution state
- **AND** SHALL return ExecutionStatus with current state, input, output, and history

#### Scenario: Execution stop
- **WHEN** StopExecution() is called with an execution ID and reason
- **THEN** the provider SHALL send a cancellation request to the Restate server
- **AND** the execution SHALL transition to stopped state

#### Scenario: Workflow deletion
- **WHEN** DeleteWorkflow() is called with a workflow ID
- **THEN** the provider SHALL unregister the workflow from the Restate server

#### Scenario: Workflow validation
- **WHEN** Validate() is called with a WorkflowSpec
- **THEN** the provider SHALL validate the workflow definition is compatible with Restate
- **AND** SHALL validate any restate-specific configuration in ProviderConfig

### Requirement: Restate SDK Integration

The system SHALL integrate with the Restate.dev Go SDK for all Restate API interactions.

#### Scenario: SDK initialization
- **WHEN** the restate provider is created
- **THEN** it SHALL initialize the Restate Go SDK client with the configured endpoint

#### Scenario: Service registration
- **WHEN** a workflow is created
- **THEN** the provider SHALL use the SDK to register the workflow as a Restate service

#### Scenario: Workflow invocation
- **WHEN** a workflow execution is started
- **THEN** the provider SHALL use the SDK to invoke the registered Restate service

#### Scenario: State management
- **WHEN** workflow state is queried or updated
- **THEN** the provider SHALL use the SDK's state management APIs

### Requirement: Endpoint Configuration

The system SHALL support configurable Restate server endpoints for local development and production deployments.

#### Scenario: Local development endpoint
- **WHEN** workflow.restate.endpoint is configured to "http://localhost:8080"
- **THEN** the provider SHALL connect to the local Restate server
- **AND** SHALL enable developers to test workflows without cloud dependencies

#### Scenario: Production endpoint
- **WHEN** workflow.restate.endpoint is configured to a production URL
- **THEN** the provider SHALL connect to the production Restate server

#### Scenario: Endpoint validation
- **WHEN** the provider is initialized with an endpoint configuration
- **THEN** it SHALL validate the endpoint is a valid HTTP/HTTPS URL
- **AND** SHALL return an error if the endpoint is malformed

#### Scenario: Endpoint connection test
- **WHEN** the provider is initialized
- **THEN** it SHOULD attempt to connect to the configured endpoint
- **AND** SHOULD log a warning if the endpoint is unreachable

### Requirement: Execution Mechanism Support

The system SHALL support multiple execution mechanisms for deploying Restate workflows.

#### Scenario: Lambda execution
- **WHEN** workflow.restate.execution_mechanism is "lambda"
- **THEN** the provider SHALL configure workflows to run on AWS Lambda
- **AND** SHALL use Lambda-specific invocation methods

#### Scenario: Fargate execution
- **WHEN** workflow.restate.execution_mechanism is "fargate"
- **THEN** the provider SHALL configure workflows to run on AWS ECS Fargate
- **AND** SHALL use Fargate-specific configuration

#### Scenario: Kubernetes execution
- **WHEN** workflow.restate.execution_mechanism is "kubernetes"
- **THEN** the provider SHALL configure workflows to run on Kubernetes
- **AND** SHALL use Kubernetes-specific service discovery

#### Scenario: Self-hosted execution
- **WHEN** workflow.restate.execution_mechanism is "self-hosted"
- **THEN** the provider SHALL configure workflows to run on self-managed infrastructure
- **AND** SHALL use direct HTTP connections to the Restate server

#### Scenario: Local execution
- **WHEN** workflow.restate.execution_mechanism is "local" or unset during development
- **THEN** the provider SHALL configure workflows to run using Restate's local compute options
- **AND** SHALL enable testing without external infrastructure

### Requirement: Service Registration

The system SHALL register workflows as Restate services with proper configuration.

#### Scenario: Service name configuration
- **WHEN** workflow.restate.service_name is configured
- **THEN** the provider SHALL use this name when registering services with Restate

#### Scenario: Default service name
- **WHEN** workflow.restate.service_name is not configured
- **THEN** the provider SHALL generate a service name from the workflow ID
- **AND** SHALL ensure the generated name is valid for Restate

#### Scenario: Service registration success
- **WHEN** a workflow is created successfully
- **THEN** the service SHALL be registered with the Restate server
- **AND** the registration SHALL be idempotent (repeated registration succeeds)

#### Scenario: Service registration failure
- **WHEN** service registration fails due to Restate server error
- **THEN** CreateWorkflow() SHALL return an error with details
- **AND** the error SHALL include the Restate server response

### Requirement: Authentication Configuration

The system SHALL support authentication methods appropriate for each execution mechanism.

#### Scenario: API key authentication
- **WHEN** workflow.restate.auth_type is "api_key"
- **THEN** the provider SHALL use the configured API key for Restate server authentication
- **AND** SHALL include the API key in all Restate API requests

#### Scenario: IAM role authentication for Lambda
- **WHEN** execution_mechanism is "lambda" and auth_type is "iam"
- **THEN** the provider SHALL use AWS IAM roles for authentication
- **AND** SHALL configure Lambda functions with appropriate IAM policies

#### Scenario: No authentication for local development
- **WHEN** endpoint is localhost and auth_type is not configured
- **THEN** the provider SHALL connect without authentication
- **AND** SHALL enable frictionless local development

#### Scenario: Authentication validation
- **WHEN** authentication configuration is provided
- **THEN** the provider SHALL validate the configuration is appropriate for the execution mechanism
- **AND** SHALL return an error if authentication is misconfigured

### Requirement: State Management

The system SHALL leverage Restate's durable state management for workflow data.

#### Scenario: Workflow state persistence
- **WHEN** a workflow execution is running
- **THEN** the provider SHALL use Restate's state management to persist workflow state
- **AND** SHALL ensure state is durable across restarts

#### Scenario: State recovery
- **WHEN** a workflow execution is interrupted
- **THEN** Restate SHALL automatically recover the workflow from persisted state
- **AND** the workflow SHALL continue from the last consistent checkpoint

#### Scenario: State query
- **WHEN** GetExecutionStatus() is called
- **THEN** the provider SHALL query Restate for current execution state
- **AND** SHALL return state including workflow input, output, and execution history

### Requirement: Error Handling

The system SHALL handle Restate-specific errors appropriately.

#### Scenario: Connection failure
- **WHEN** the provider cannot connect to the Restate server
- **THEN** it SHALL return an error describing the connection issue
- **AND** SHALL include the endpoint in the error message

#### Scenario: Workflow definition validation error
- **WHEN** a workflow definition is incompatible with Restate
- **THEN** Validate() SHALL return ErrInvalidSpec with details
- **AND** SHALL describe what aspect of the definition is invalid

#### Scenario: Execution timeout
- **WHEN** a workflow execution exceeds its configured timeout
- **THEN** Restate SHALL stop the execution
- **AND** GetExecutionStatus() SHALL return state "timeout"

#### Scenario: Service not found
- **WHEN** StartExecution() is called for a workflow that isn't registered
- **THEN** the provider SHALL return ErrWorkflowNotFound
- **AND** SHALL include the workflow ID in the error

### Requirement: Workflow Definition Format

The system SHALL support Restate-specific workflow definitions.

#### Scenario: Go handler workflow definition
- **WHEN** a workflow definition specifies a Go handler function
- **THEN** the provider SHALL register the handler with Restate
- **AND** SHALL ensure the handler is invoked for workflow executions

#### Scenario: Workflow definition validation
- **WHEN** Validate() is called with a WorkflowSpec
- **THEN** the provider SHALL verify the definition contains required Restate workflow fields
- **AND** SHALL verify the definition format matches Restate's expectations

#### Scenario: Definition storage
- **WHEN** CreateWorkflow() is called
- **THEN** the provider SHALL store the workflow definition in Restate
- **AND** SHALL make the definition available for execution

### Requirement: Idempotency

The system SHALL ensure all provider operations are idempotent as required by the Provider interface.

#### Scenario: Duplicate workflow creation
- **WHEN** CreateWorkflow() is called twice with the same WorkflowID
- **THEN** the second call SHALL succeed without error
- **AND** SHALL not modify the existing workflow registration

#### Scenario: Duplicate workflow deletion
- **WHEN** DeleteWorkflow() is called on a non-existent workflow
- **THEN** the call SHALL succeed without error

#### Scenario: Stop already-stopped execution
- **WHEN** StopExecution() is called on an execution that is already stopped
- **THEN** the call SHALL succeed without error
- **AND** SHALL not modify the execution state

### Requirement: Configuration Validation

The system SHALL validate all restate-specific configuration at provider initialization.

#### Scenario: Missing endpoint configuration
- **WHEN** the provider is initialized without workflow.restate.endpoint configured
- **THEN** it SHALL return an error indicating endpoint is required

#### Scenario: Invalid execution mechanism
- **WHEN** workflow.restate.execution_mechanism is set to an unsupported value
- **THEN** the provider SHALL return an error listing valid execution mechanisms

#### Scenario: Complete configuration validation
- **WHEN** the provider is initialized with a complete configuration
- **THEN** it SHALL validate all fields are properly formatted
- **AND** SHALL validate the configuration is internally consistent (e.g., auth_type matches execution_mechanism)

### Requirement: Logging and Observability

The system SHALL provide comprehensive logging for Restate operations.

#### Scenario: Operation logging
- **WHEN** any provider operation is called
- **THEN** the provider SHALL log the operation with structured fields (workflow ID, execution ID, etc.)

#### Scenario: Error logging
- **WHEN** a Restate operation fails
- **THEN** the provider SHALL log the error with full context
- **AND** SHALL include the Restate server response if available

#### Scenario: State transition logging
- **WHEN** a workflow execution changes state
- **THEN** the provider SHALL log the state transition with timestamps

### Requirement: Provider Configuration Structure

The system SHALL define a configuration structure for Restate-specific settings.

#### Scenario: Configuration fields
- **WHEN** workflow.restate configuration is loaded
- **THEN** it SHALL include fields for:
  - endpoint (string, required): Restate server URL
  - execution_mechanism (string, optional): Deployment target (lambda/fargate/kubernetes/self-hosted/local)
  - service_name (string, optional): Service identifier for registration
  - auth_type (string, optional): Authentication method (api_key/iam/none)
  - api_key (string, optional): API key if auth_type is api_key
  - timeout (duration, optional): Default workflow timeout
  - retry_attempts (int, optional): Number of retry attempts for operations

#### Scenario: Configuration loading
- **WHEN** the application starts
- **THEN** the restate configuration SHALL be loaded from the workflow configuration section
- **AND** SHALL be available to the provider during initialization

#### Scenario: Configuration precedence
- **WHEN** configuration is provided via multiple sources (file, env vars, CLI flags)
- **THEN** the standard Landlord configuration precedence SHALL apply
- **AND** CLI flags SHALL override environment variables SHALL override config files
