# viper-configuration-files Specification

## Purpose
Enable landlord to load configuration from YAML and JSON configuration files with proper precedence handling (files < environment variables < CLI flags) and seamless struct binding.

## ADDED Requirements

### Requirement: Configuration files are loaded from standard locations
The system SHALL search for configuration files in standard locations and load the first file found.

#### Scenario: Config file in current directory is loaded
- **WHEN** a `config.yaml` or `config.json` file exists in the current working directory
- **THEN** the system loads and parses the configuration from that file
- **AND** the parsed values are made available to the application

#### Scenario: Config file in /etc/landlord/ is loaded
- **WHEN** no config file is found in the current directory but `config.yaml` exists in `/etc/landlord/`
- **THEN** the system loads configuration from `/etc/landlord/config.yaml`
- **AND** the file is properly parsed and applied

#### Scenario: Config file via XDG_CONFIG_HOME is loaded
- **WHEN** no config file is found in current or system directories but `$XDG_CONFIG_HOME/landlord/config.yaml` exists
- **THEN** the system loads configuration from the XDG location
- **AND** the file is properly parsed and applied

#### Scenario: No config file is optional
- **WHEN** no configuration file is found in any standard location
- **THEN** the system continues with default values and environment variable/CLI flag configuration
- **AND** the application starts normally without error

### Requirement: Explicit config file path can be specified
The system SHALL allow the user to specify an explicit configuration file path via the `--config` flag or `LANDLORD_CONFIG` environment variable.

#### Scenario: Config file specified via CLI flag
- **WHEN** the user runs `landlord --config /path/to/custom-config.yaml`
- **THEN** the system loads configuration exclusively from the specified file path
- **AND** standard location search is bypassed

#### Scenario: Config file specified via environment variable
- **WHEN** the `LANDLORD_CONFIG` environment variable is set to a file path
- **THEN** the system loads configuration from that file path
- **AND** CLI flag takes precedence if both are provided

#### Scenario: Invalid config file path produces error
- **WHEN** the specified config file does not exist or is not readable
- **THEN** the system reports a clear error indicating the file cannot be loaded
- **AND** the application exits without starting services

### Requirement: YAML configuration files are parsed and bound to struct
The system SHALL parse YAML configuration files and bind the content to the Config struct using mapstructure tags.

#### Scenario: Nested YAML structure is bound correctly
- **WHEN** a `config.yaml` file contains nested configuration (e.g., database and http sections)
- **THEN** the system parses the YAML structure and populates the corresponding struct fields
- **AND** all nested values are accessible via the configuration struct

#### Scenario: YAML with environment variable references is handled
- **WHEN** a YAML file contains configuration values
- **THEN** the values are read as-is (no variable interpolation at file parse time)
- **AND** environment variable overrides still apply per precedence rules

### Requirement: JSON configuration files are parsed and bound to struct
The system SHALL parse JSON configuration files and bind the content to the Config struct using mapstructure tags.

#### Scenario: Nested JSON structure is bound correctly
- **WHEN** a `config.json` file contains nested configuration (e.g., database and http sections)
- **THEN** the system parses the JSON structure and populates the corresponding struct fields
- **AND** all nested values are accessible via the configuration struct

#### Scenario: JSON parsing errors are reported clearly
- **WHEN** a `config.json` file contains invalid JSON syntax
- **THEN** the system reports a JSON parsing error with line and column information
- **AND** the application exits without starting services

### Requirement: Configuration precedence is enforced correctly
The system SHALL apply configuration precedence in the correct order: CLI flags > environment variables > config file > defaults.

#### Scenario: CLI flag overrides all other sources
- **WHEN** the same configuration value is specified in CLI flag, environment variable, and config file
- **THEN** the CLI flag value takes precedence
- **AND** the application uses the CLI flag value

#### Scenario: Environment variable overrides config file
- **WHEN** the same configuration value is specified in environment variable and config file, but not CLI flag
- **THEN** the environment variable value takes precedence
- **AND** the application uses the environment variable value

#### Scenario: Config file provides values when higher sources not set
- **WHEN** a configuration value is only present in the config file (not in CLI flags or environment variables)
- **THEN** the value from the config file is used
- **AND** the application uses the file value

#### Scenario: Defaults apply when no other source provides value
- **WHEN** a configuration value is not present in CLI flags, environment variables, or config file
- **THEN** the default value (if defined) is used
- **AND** the application continues normally

### Requirement: Configuration errors are reported clearly
The system SHALL validate configuration and report errors with helpful messages indicating the source of the problem.

#### Scenario: Missing required field is reported
- **WHEN** a required configuration field is not provided via any source (CLI, env, file, or defaults)
- **THEN** the system reports which field is missing
- **AND** the application exits with a clear validation error

#### Scenario: Invalid value type is reported
- **WHEN** a configuration value has an incorrect type (e.g., string instead of integer)
- **THEN** the system reports the field name and expected type
- **AND** the application does not start

#### Scenario: Configuration source is logged for debugging
- **WHEN** the application starts successfully
- **THEN** the effective configuration (with sources) can be determined from logs or debug output
- **AND** sensitive values (passwords, tokens) are redacted in output
