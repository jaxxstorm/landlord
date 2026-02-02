## ADDED Requirements

### Requirement: Comprehensive Configuration Example

The system SHALL provide a `config.example.yaml` file that documents every configuration option available in Landlord with clear explanations, defaults, and usage examples for different scenarios (PostgreSQL, SQLite, AWS Step Functions, Restate workflows).

#### Scenario: Example file contains all top-level sections
- **WHEN** config.example.yaml is reviewed
- **THEN** it includes sections for: database, http, log, compute, and workflow providers

#### Scenario: Database options are fully documented
- **WHEN** reviewing database section of config.example.yaml
- **THEN** it shows examples for both PostgreSQL (production) and SQLite (development) with comments explaining when to use each

#### Scenario: Compute provider options are documented
- **WHEN** reviewing compute section
- **THEN** it includes examples for mock, docker providers with configuration options and comments

#### Scenario: Workflow provider options are documented
- **WHEN** reviewing workflow section
- **THEN** it includes examples for mock, restate, and step-functions providers with their respective configuration options

#### Scenario: Every field has explanatory comments
- **WHEN** reading any configuration section
- **THEN** each field includes a comment explaining its purpose, accepted values, and default if applicable

#### Scenario: Configuration is organized logically
- **WHEN** viewing config.example.yaml structure
- **THEN** sections are ordered: database → http → log → compute → workflow, with subsections grouped by provider type

### Requirement: Configuration File Consolidation

The system SHALL replace multiple scattered configuration examples (`config.yaml`, `config.json`, `config.sqlite.yaml`, `config.docker.yaml`) with a single `config.example.yaml` file, reducing duplication and confusion.

#### Scenario: All historical config options are represented
- **WHEN** comparing new config.example.yaml with old config files
- **THEN** every option from the old files appears in the new consolidated example

#### Scenario: YAML format is used for readability
- **WHEN** reviewing config.example.yaml
- **THEN** format is YAML (not JSON) for better human readability and comments

#### Scenario: Old config files are documented for reference
- **WHEN** reading documentation
- **THEN** it explains that `config.example.yaml` is the authoritative reference and old files are maintained for backwards compatibility only

### Requirement: Scenario-Based Examples

The system SHALL provide commented example configurations for common scenarios within the single config.example.yaml file, showing how to configure the system for different use cases.

#### Scenario: Development with Docker and Restate is documented
- **WHEN** looking for local development configuration example
- **THEN** config.example.yaml includes a complete example with Docker compute provider and Restate workflows

#### Scenario: Production with PostgreSQL and Step Functions is documented
- **WHEN** looking for production configuration example
- **THEN** config.example.yaml includes a complete example with PostgreSQL database and AWS Step Functions workflows

#### Scenario: Testing with SQLite is documented
- **WHEN** looking for test configuration example
- **THEN** config.example.yaml includes a simple example using SQLite and mock providers

### Requirement: Default Values Clarity

The system SHALL clearly indicate which values are defaults, which are required, and which are optional throughout the configuration example file.

#### Scenario: Required fields are marked
- **WHEN** reading config.example.yaml
- **THEN** required fields are marked with comments like "# REQUIRED" or "# must be set"

#### Scenario: Optional fields are marked
- **WHEN** reading config.example.yaml
- **THEN** optional fields are marked with comments like "# OPTIONAL: defaults to X if not set"

#### Scenario: Default values are shown
- **WHEN** seeing a field with a default value
- **THEN** the comment explains what the default is and when it's used
