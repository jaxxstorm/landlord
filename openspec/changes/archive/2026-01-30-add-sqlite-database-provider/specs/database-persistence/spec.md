# database-persistence Specification Delta

## Purpose
Extends database-persistence to support provider abstraction, allowing PostgreSQL or SQLite backends with identical interface contract.

## MODIFIED Requirements

### Requirement: Application connects to configured database provider

The system SHALL establish a connection to the configured database provider (PostgreSQL or SQLite) using provider-specific configuration parameters.

#### Scenario: Successful database connection with PostgreSQL
- **WHEN** the application starts with provider type "postgres" and valid database configuration
- **THEN** a connection pool is established to PostgreSQL
- **AND** the pool is ready to execute queries

#### Scenario: Successful database connection with SQLite
- **WHEN** the application starts with provider type "sqlite" and valid file path
- **THEN** a SQLite database is opened or created
- **AND** the connection is ready to execute queries

#### Scenario: Connection fails due to invalid credentials (PostgreSQL)
- **WHEN** the application starts with PostgreSQL provider and invalid credentials
- **THEN** the application fails to start with a clear authentication error
- **AND** the error message indicates credential issues

#### Scenario: Connection fails due to invalid file path (SQLite)
- **WHEN** the application starts with SQLite provider and inaccessible file path
- **THEN** the application fails to start with a clear file access error
- **AND** the error message indicates path permission or existence issues

#### Scenario: Connection fails due to unreachable host (PostgreSQL)
- **WHEN** the PostgreSQL database host is unreachable
- **THEN** the application retries connection with exponential backoff
- **AND** fails after maximum retry attempts with a clear error

### Requirement: Database provider type is configurable

The system SHALL allow configuration of the database provider type via configuration.

#### Scenario: Configure PostgreSQL provider
- **WHEN** configuration specifies provider type "postgres"
- **THEN** PostgreSQL driver is initialized
- **AND** PostgreSQL-specific configuration options are used
- **AND** PostgreSQL connection string is constructed from config

#### Scenario: Configure SQLite provider
- **WHEN** configuration specifies provider type "sqlite"
- **THEN** SQLite driver is initialized
- **AND** SQLite-specific configuration options are used
- **AND** SQLite file path or URI is constructed from config

#### Scenario: Invalid provider type is rejected
- **WHEN** configuration specifies an unknown provider type
- **THEN** application fails to start with validation error
- **AND** error message lists supported provider types

#### Scenario: Provider defaults to PostgreSQL for backward compatibility
- **WHEN** configuration does not specify provider type
- **THEN** PostgreSQL provider is used by default
- **AND** startup logs indicate PostgreSQL provider selection

### Requirement: Database migrations adapt to provider capabilities

The system SHALL apply migrations with provider-appropriate syntax and features.

#### Scenario: Migrations with provider-specific types (PostgreSQL)
- **WHEN** migrations are applied to PostgreSQL
- **THEN** PostgreSQL-native types are used (UUID, JSONB, TIMESTAMPTZ)
- **AND** PostgreSQL-specific functions are available (gen_random_uuid, NOW)
- **AND** GIN indexes are created for JSONB columns

#### Scenario: Migrations with provider-compatible types (SQLite)
- **WHEN** migrations are applied to SQLite
- **THEN** SQLite-compatible types are used (TEXT for UUID, JSON, DATETIME)
- **AND** type mappings maintain semantic equivalence
- **AND** indexes are created with SQLite-compatible syntax

#### Scenario: Migration compatibility validation
- **WHEN** new migrations are added
- **THEN** migrations must be tested against all supported providers
- **AND** provider-specific syntax is avoided in shared migrations
- **AND** provider-specific migrations use separate files if needed

### Requirement: Repository interface is provider-agnostic

The system SHALL provide a repository interface that works identically across all database providers.

#### Scenario: Repository operations with PostgreSQL
- **WHEN** repository operations are executed against PostgreSQL provider
- **THEN** all operations function correctly with PostgreSQL-specific optimizations
- **AND** JSONB operators can be used for advanced queries
- **AND** connection pooling uses PostgreSQL best practices

#### Scenario: Repository operations with SQLite
- **WHEN** repository operations are executed against SQLite provider
- **THEN** all operations function correctly with SQLite-specific optimizations
- **AND** JSON functions provide equivalent capabilities
- **AND** connection management respects SQLite single-writer model

#### Scenario: Test suite passes against all providers
- **WHEN** repository test suite is executed
- **THEN** tests pass against PostgreSQL provider
- **AND** tests pass against SQLite provider
- **AND** test assertions are provider-agnostic

## ADDED Requirements

### Requirement: Database provider factory creates appropriate provider implementation

The system SHALL use a factory pattern to instantiate the correct database provider based on configuration.

#### Scenario: Factory creates PostgreSQL provider
- **WHEN** NewProvider is called with type "postgres"
- **THEN** PostgreSQL provider implementation is returned
- **AND** provider implements the Database interface
- **AND** PostgreSQL-specific configuration is validated

#### Scenario: Factory creates SQLite provider
- **WHEN** NewProvider is called with type "sqlite"
- **THEN** SQLite provider implementation is returned
- **AND** provider implements the Database interface
- **AND** SQLite-specific configuration is validated

#### Scenario: Factory rejects unknown provider types
- **WHEN** NewProvider is called with unsupported type
- **THEN** factory returns error
- **AND** error message indicates supported provider types

### Requirement: Provider interface defines common database operations

The system SHALL define a Database interface that abstracts provider-specific implementations.

#### Scenario: Database interface methods are provider-agnostic
- **WHEN** Database interface methods are called
- **THEN** methods work identically regardless of provider
- **AND** return types are consistent across providers
- **AND** error semantics are consistent across providers

#### Scenario: Provider exposes underlying connection for advanced usage
- **WHEN** advanced database operations are needed
- **THEN** provider exposes method to access underlying connection
- **AND** type assertion allows provider-specific features
- **AND** standard operations remain provider-agnostic

### Requirement: Configuration includes provider-specific sections

The system SHALL support provider-specific configuration while maintaining common configuration structure.

#### Scenario: PostgreSQL-specific configuration
- **WHEN** PostgreSQL provider is configured
- **THEN** configuration includes host, port, username, password, database
- **AND** configuration includes SSL settings
- **AND** configuration includes connection pool settings

#### Scenario: SQLite-specific configuration
- **WHEN** SQLite provider is configured
- **THEN** configuration includes file path or URI
- **AND** configuration includes pragmas for tuning
- **AND** configuration includes busy timeout setting

#### Scenario: Common configuration applies to all providers
- **WHEN** any provider is configured
- **THEN** configuration includes connect timeout
- **AND** configuration includes migration source path
- **AND** configuration includes logging preferences
