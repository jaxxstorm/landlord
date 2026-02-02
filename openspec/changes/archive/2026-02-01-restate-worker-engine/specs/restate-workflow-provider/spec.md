## MODIFIED Requirements

### Requirement: Provider Configuration Structure
The system SHALL define a configuration structure for Restate-specific settings, including worker registration settings.

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
  - worker_register_on_startup (bool, optional): Whether to register worker services automatically
  - worker_admin_endpoint (string, optional): Restate admin API URL if different from endpoint
  - worker_namespace (string, optional): Restate namespace for worker registrations

#### Scenario: Configuration loading
- **WHEN** the application starts
- **THEN** the restate configuration SHALL be loaded from the workflow configuration section
- **AND** SHALL be available to the provider during initialization

#### Scenario: Configuration precedence
- **WHEN** configuration is provided via multiple sources (file, env vars, CLI flags)
- **THEN** the standard Landlord configuration precedence SHALL apply
- **AND** CLI flags SHALL override environment variables SHALL override config files
