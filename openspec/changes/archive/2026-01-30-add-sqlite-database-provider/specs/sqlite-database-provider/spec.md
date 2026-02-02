# sqlite-database-provider Specification

## Purpose
Defines the SQLite implementation of the database provider interface, enabling local development and lightweight deployments without external database dependencies while maintaining full compatibility with the database-persistence specification.

## ADDED Requirements

### Requirement: SQLite provider supports same interface as PostgreSQL provider

The system SHALL implement the database provider interface with SQLite as the backing store, maintaining identical behavior guarantees.

#### Scenario: Initialize SQLite provider
- **WHEN** the application starts with SQLite provider configuration
- **THEN** a SQLite database file is created or opened
- **AND** the connection is configured with appropriate pragmas
- **AND** all repository operations function identically to PostgreSQL

#### Scenario: Repository operations produce identical results
- **WHEN** any repository operation is executed against SQLite
- **THEN** the operation produces the same result as PostgreSQL
- **AND** the same domain errors are returned for failures
- **AND** tests written for PostgreSQL pass against SQLite

### Requirement: SQLite provider enables WAL mode for concurrency

The system SHALL configure SQLite to use Write-Ahead Logging (WAL) mode for improved concurrent read performance.

#### Scenario: WAL mode enabled on startup
- **WHEN** SQLite provider is initialized
- **THEN** WAL mode is enabled with PRAGMA journal_mode=WAL
- **AND** multiple readers can access the database concurrently
- **AND** writers do not block readers

#### Scenario: WAL checkpoint behavior is configured
- **WHEN** WAL mode is active
- **THEN** automatic checkpointing occurs at configured intervals
- **AND** checkpoint operations do not block normal operations
- **AND** database file remains consistent

### Requirement: SQLite provider handles busy timeouts gracefully

The system SHALL configure appropriate busy timeout to handle concurrent write contention.

#### Scenario: Busy timeout prevents immediate failures
- **WHEN** a write operation encounters a locked database
- **THEN** SQLite retries for the configured busy timeout duration
- **AND** the operation succeeds if lock is released within timeout
- **AND** database locked error is returned only after timeout expires

#### Scenario: Busy timeout is configurable
- **WHEN** SQLite provider is configured with busy timeout value
- **THEN** PRAGMA busy_timeout is set to the configured milliseconds
- **AND** the default timeout is 5000ms if not specified

### Requirement: SQLite provider supports in-memory mode for testing

The system SHALL support in-memory SQLite databases for fast test execution.

#### Scenario: Initialize in-memory database
- **WHEN** SQLite provider is configured with ":memory:" path
- **THEN** database exists only in memory
- **AND** no disk I/O occurs
- **AND** database is destroyed when connection closes

#### Scenario: Shared in-memory database for testing
- **WHEN** SQLite provider uses "file::memory:?cache=shared" URI
- **THEN** multiple connections share the same in-memory database
- **AND** test setup can create schema once
- **AND** test cases can run with isolated transactions

### Requirement: SQLite migrations are compatible with PostgreSQL migrations

The system SHALL execute the same migration files as PostgreSQL, with SQLite-compatible syntax.

#### Scenario: Apply migrations to SQLite
- **WHEN** SQLite provider starts and migrations are pending
- **THEN** migration files are applied in order
- **AND** schema version tracking works identically to PostgreSQL
- **AND** migration syntax is compatible with both databases

#### Scenario: Handle PostgreSQL-specific syntax
- **WHEN** migrations contain PostgreSQL-specific features
- **THEN** SQLite-compatible equivalents are used automatically
- **AND** UUID type is mapped to TEXT
- **AND** JSONB is mapped to JSON
- **AND** NOW() is mapped to CURRENT_TIMESTAMP

### Requirement: SQLite provider enforces foreign key constraints

The system SHALL enable SQLite foreign key constraint enforcement.

#### Scenario: Foreign key constraints are enforced
- **WHEN** SQLite provider is initialized
- **THEN** PRAGMA foreign_keys=ON is set
- **AND** foreign key violations prevent operations
- **AND** CASCADE deletes function correctly

#### Scenario: Foreign key constraint violations return errors
- **WHEN** an operation violates a foreign key constraint
- **THEN** the operation fails with constraint violation error
- **AND** the error is mapped to domain error appropriately

### Requirement: SQLite provider uses connection pooling compatible with single-writer model

The system SHALL implement connection pooling that respects SQLite's single-writer constraint.

#### Scenario: Connection pool configured for SQLite
- **WHEN** SQLite provider creates connection pool
- **THEN** max open connections is set to appropriate value for SQLite
- **AND** connection pool prevents write contention
- **AND** read operations can use multiple connections

#### Scenario: Write serialization prevents database locked errors
- **WHEN** multiple goroutines attempt concurrent writes
- **THEN** connection pool serializes write operations appropriately
- **AND** busy timeout handles remaining contention
- **AND** operations succeed without manual retry logic

### Requirement: SQLite provider supports file-based and URI-based paths

The system SHALL accept both simple file paths and SQLite URI strings.

#### Scenario: File path configuration
- **WHEN** SQLite provider is configured with file path "/data/landlord.db"
- **THEN** database is created at the specified path
- **AND** parent directories are created if they don't exist
- **AND** file permissions are set appropriately

#### Scenario: URI configuration with query parameters
- **WHEN** SQLite provider is configured with URI "file:/data/landlord.db?mode=rwc&cache=shared"
- **THEN** URI parameters control connection behavior
- **AND** mode parameter determines read/write permissions
- **AND** cache parameter controls connection sharing

### Requirement: SQLite provider maintains connection health checks

The system SHALL provide health check capabilities for SQLite connections.

#### Scenario: Health check on active connection
- **WHEN** health check is performed on SQLite provider
- **THEN** a simple SELECT query verifies database accessibility
- **AND** health check returns success if query succeeds
- **AND** health check times out after configured duration

#### Scenario: Health check with inaccessible database file
- **WHEN** health check is performed and database file is inaccessible
- **THEN** health check returns failure
- **AND** failure includes descriptive error message
- **AND** connection pool attempts recovery

### Requirement: SQLite provider uses pragmas for performance tuning

The system SHALL apply performance-oriented pragmas on connection initialization.

#### Scenario: Performance pragmas are applied
- **WHEN** SQLite connection is established
- **THEN** PRAGMA synchronous is set to NORMAL for durability balance
- **AND** PRAGMA temp_store is set to MEMORY for performance
- **AND** PRAGMA mmap_size is configured for memory-mapped I/O
- **AND** pragmas are logged at startup for observability

#### Scenario: Pragmas are configurable via provider settings
- **WHEN** SQLite provider configuration includes pragma overrides
- **THEN** custom pragma values are applied instead of defaults
- **AND** invalid pragma values cause startup failure with clear error
- **AND** pragma configuration is validated before application

### Requirement: SQLite provider closes connections cleanly on shutdown

The system SHALL close all SQLite connections during graceful shutdown.

#### Scenario: Graceful shutdown with active connections
- **WHEN** application shutdown is initiated
- **THEN** all active SQLite connections complete current operations
- **AND** connection pool is drained cleanly
- **AND** WAL checkpoint finalizes on close
- **AND** database file is left in consistent state

#### Scenario: Shutdown timeout handling
- **WHEN** shutdown occurs with long-running operations
- **THEN** operations are given configured timeout to complete
- **AND** connections are forcibly closed if timeout expires
- **AND** warning is logged for incomplete operations
