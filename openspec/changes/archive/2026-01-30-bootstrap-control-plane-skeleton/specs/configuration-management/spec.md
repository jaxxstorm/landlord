## ADDED Requirements

### Requirement: Configuration is loaded from environment variables

The system SHALL load configuration from environment variables using kong tags.

#### Scenario: Required configuration is present
- **WHEN** all required environment variables are set
- **THEN** the configuration struct is populated with values
- **AND** the application starts successfully

#### Scenario: Required configuration is missing
- **WHEN** a required environment variable is not set
- **THEN** the application fails to start with a validation error
- **AND** the error message indicates which variable is missing

### Requirement: Configuration supports CLI flag overrides

The system SHALL allow environment variables to be overridden by CLI flags.

#### Scenario: CLI flag overrides environment variable
- **WHEN** both an environment variable and CLI flag are provided
- **THEN** the CLI flag value takes precedence
- **AND** the configuration reflects the CLI flag value

### Requirement: Configuration includes database connection parameters

The system SHALL provide configuration for database connectivity.

#### Scenario: Database configuration parameters
- **WHEN** configuration is loaded
- **THEN** database host, port, username, password, and database name are available
- **AND** optional parameters like SSL mode and connection pool settings are supported

### Requirement: Configuration includes HTTP server parameters

The system SHALL provide configuration for HTTP server settings.

#### Scenario: HTTP server configuration
- **WHEN** configuration is loaded
- **THEN** HTTP host, port, and timeout settings are available
- **AND** graceful shutdown timeout is configurable

### Requirement: Configuration includes logging parameters

The system SHALL provide configuration for logging behavior.

#### Scenario: Logging configuration
- **WHEN** configuration is loaded
- **THEN** log level and output format are configurable
- **AND** development vs production mode can be specified

### Requirement: Configuration validation occurs at startup

The system SHALL validate all configuration values before starting services.

#### Scenario: Valid configuration passes validation
- **WHEN** configuration is loaded and all values are valid
- **THEN** validation succeeds
- **AND** the application proceeds to start services

#### Scenario: Invalid configuration fails validation
- **WHEN** configuration contains invalid values
- **THEN** validation fails with specific error messages
- **AND** the application exits before starting services

### Requirement: Configuration help text is auto-generated

The system SHALL generate help text from configuration struct tags.

#### Scenario: Help flag displays configuration options
- **WHEN** the application is run with --help flag
- **THEN** all configuration options are displayed with descriptions
- **AND** environment variable names and default values are shown
