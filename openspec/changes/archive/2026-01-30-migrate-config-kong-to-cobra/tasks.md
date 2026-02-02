## 1. Dependencies and Setup

- [x] 1.1 Add `github.com/spf13/cobra` dependency to go.mod
- [x] 1.2 Add `github.com/spf13/viper` dependency to go.mod
- [x] 1.3 Run `go mod tidy` to download and hash new dependencies
- [x] 1.4 Verify cobra and viper are importable in the project

## 2. Configuration Package Refactoring

- [x] 2.1 Add mapstructure tags to Config struct fields
- [x] 2.2 Add mapstructure tags to DatabaseConfig struct fields
- [x] 2.3 Add mapstructure tags to HTTPConfig struct fields
- [x] 2.4 Add mapstructure tags to LoggingConfig struct fields
- [x] 2.5 Add mapstructure tags to WorkflowConfig struct fields
- [x] 2.6 Create `internal/config/viper.go` with viper initialization function
- [x] 2.7 Implement `newViperInstance()` to create viper with default config
- [x] 2.8 Implement `bindEnvironmentVariables()` to bind env vars with `BindEnv()`
- [x] 2.9 Implement `loadConfigFile()` to search standard locations (current dir, /etc/landlord/, XDG)
- [x] 2.10 Implement `loadFromViper()` to unmarshal viper config into Config struct
- [x] 2.11 Implement `findConfigFile()` with precedence: --config flag > LANDLORD_CONFIG env > standard locations
- [x] 2.12 Ensure existing `Config.Validate()` method is called after unmarshaling

## 3. Cobra Command Structure

- [x] 3.1 Create `cmd/landlord/root.go` with root cobra command definition
- [x] 3.2 Define cobra command with description, long help text, and usage
- [x] 3.3 Add `--config` flag with description and default value
- [x] 3.4 Add `--database-host` flag with description and binding to viper
- [x] 3.5 Add `--database-port` flag with description and binding to viper
- [x] 3.6 Add `--database-user` flag with description and binding to viper
- [x] 3.7 Add `--database-password` flag with description and binding to viper
- [x] 3.8 Add `--database-name` flag with description and binding to viper
- [x] 3.9 Add `--http-host` flag with description and binding to viper
- [x] 3.10 Add `--http-port` flag with description and binding to viper
- [x] 3.11 Add `--http-timeout` flag with description and binding to viper
- [x] 3.12 Add `--log-level` flag with description and binding to viper
- [x] 3.13 Add `--log-format` flag with description and binding to viper
- [x] 3.14 Mark required flags appropriately (if any)

## 4. Cobra PreRun Hook

- [x] 4.1 Implement `PreRun` function in root command
- [x] 4.2 In PreRun: Call viper initialization
- [x] 4.3 In PreRun: Call environment variable binding
- [x] 4.4 In PreRun: Load configuration file (if provided or found)
- [x] 4.5 In PreRun: Bind cobra flags to viper keys
- [x] 4.6 In PreRun: Unmarshal viper config to Config struct
- [x] 4.7 In PreRun: Call Config.Validate() and exit if validation fails
- [x] 4.8 Handle errors in PreRun with clear error messages to user

## 5. Main.go Refactoring

- [x] 5.1 Update `cmd/landlord/main.go` to import cobra command
- [x] 5.2 Remove kong import and usage
- [x] 5.3 Update main() function to execute root cobra command
- [x] 5.4 Wire configuration loading through cobra PreRun hook
- [x] 5.5 Remove old kong struct tag logic
- [x] 5.6 Ensure program exits with appropriate status codes on error

## 6. Environment Variable Compatibility

- [x] 6.1 Verify `DB_HOST` environment variable still works
- [x] 6.2 Verify `DB_PORT` environment variable still works
- [x] 6.3 Verify `DB_USER` environment variable still works
- [x] 6.4 Verify `DB_PASSWORD` environment variable still works
- [x] 6.5 Verify `DB_DATABASE` environment variable still works
- [x] 6.6 Verify `HTTP_HOST` environment variable still works
- [x] 6.7 Verify `HTTP_PORT` environment variable still works
- [x] 6.8 Verify `HTTP_SHUTDOWN_TIMEOUT` environment variable still works
- [x] 6.9 Verify `LOG_LEVEL` environment variable still works
- [x] 6.10 Verify `LOG_FORMAT` environment variable still works

**Status**: All 10 environment variables verified. BindEnv calls confirmed in viper.go (lines 45-103), TestBindEnvironmentVariables passing in viper_test.go, backward compatibility maintained.

## 7. Configuration File Support

- [x] 7.1 Create example `config.yaml` file in project root with sample configuration
- [x] 7.2 Create example `config.json` file in project root with sample configuration
- [x] 7.3 Verify YAML file is parsed correctly with nested structure
- [x] 7.4 Verify JSON file is parsed correctly with nested structure
- [x] 7.5 Test config file loading from current directory
- [ ] 7.6 Test config file loading from /etc/landlord/ (if applicable for testing)
- [ ] 7.7 Test config file loading from XDG_CONFIG_HOME location
- [x] 7.8 Test --config flag with explicit file path
- [ ] 7.9 Test LANDLORD_CONFIG environment variable override
- [x] 7.10 Verify error handling for missing config files (should be optional)
- [x] 7.11 Verify error handling for invalid YAML syntax
- [x] 7.12 Verify error handling for invalid JSON syntax

## 8. Precedence Testing

- [x] 8.1 Test CLI flag precedence over environment variable
- [x] 8.2 Test CLI flag precedence over config file value
- [x] 8.3 Test environment variable precedence over config file value
- [x] 8.4 Test config file value when no CLI flag or env var present
- [x] 8.5 Test default value when no source provides value
- [x] 8.6 Test partial config (some values from file, some from env, some from CLI)

## 9. Testing and Validation

- [x] 9.1 Create unit tests for viper initialization function
- [x] 9.2 Create unit tests for environment variable binding
- [x] 9.3 Create unit tests for config file discovery and loading
- [x] 9.4 Create unit tests for unmarshal and struct binding
- [x] 9.5 Create integration tests for full config loading pipeline
- [x] 9.6 Create tests for all precedence scenarios
- [x] 9.7 Create tests for error cases (invalid files, missing fields, wrong types)
- [x] 9.8 Run `go test ./cmd/... ./internal/config/... -v` and verify all tests pass
- [x] 9.9 Run `go test -race ./cmd/... ./internal/config/...` to check for race conditions
- [ ] 9.10 Verify test coverage is 80%+ for config-related code

## 10. Documentation

- [x] 10.1 Update README.md with configuration section
- [x] 10.2 Document YAML configuration format with examples
- [x] 10.3 Document JSON configuration format with examples
- [x] 10.4 Document environment variable names and usage
- [x] 10.5 Document CLI flags with examples for each flag
- [x] 10.6 Document configuration precedence order (CLI > env > file > defaults)
- [x] 10.7 Document configuration file search locations
- [x] 10.8 Document how to use --config flag for explicit file path
- [x] 10.9 Document LANDLORD_CONFIG environment variable usage
- [x] 10.10 Add examples of common configuration scenarios
- [x] 10.11 Document migration path for users upgrading from kong

**Status**: All 11 documentation tasks completed. Configuration guide created in `docs/configuration.md` (18KB) with comprehensive YAML/JSON examples, environment variable reference, CLI flag documentation, precedence rules, common scenarios, and troubleshooting. Migration guide created in `docs/migration-kong-to-cobra.md` (14KB) with detailed migration patterns, Kubernetes examples, Docker Compose examples, GitHub Actions examples, and rollback plan.

## 11. Build and Integration Testing

- [x] 11.1 Run `go build ./cmd/landlord` and verify no compile errors
- [x] 11.2 Run built binary with --help and verify help text is displayed
- [x] 11.3 Run built binary with --version and verify version is displayed
- [x] 11.4 Run built binary with invalid flag and verify error message
- [x] 11.5 Run built binary with environment variables only
- [x] 11.6 Run built binary with CLI flags only
- [x] 11.7 Run built binary with config file only
- [x] 11.8 Run built binary with mix of config file, env vars, and CLI flags
- [x] 11.9 Verify application starts successfully with valid configuration
- [x] 11.10 Verify application fails cleanly with invalid configuration

**Status**: All 10 integration tests verified through manual testing

## 12. Cleanup and Finalization

- [x] 12.1 Remove kong struct tags from Config and related structs
- [x] 12.2 Remove any remaining kong imports from code
- [x] 12.3 Update go.mod to remove kong dependency (if no other code uses it)
- [x] 12.4 Run `go mod tidy` to clean up unused dependencies
- [x] 12.5 Verify no deprecated code or warnings in build output
- [ ] 12.6 Run linter (`golangci-lint` if available) and fix any issues
- [x] 12.7 Final `go build ./cmd/landlord` to ensure clean build
- [x] 12.8 Create short migration guide for team members

**Status**: Kong dependency successfully removed from go.mod and go.sum. Clean build verified. Migration guide created in docs/migration-kong-to-cobra.md.
