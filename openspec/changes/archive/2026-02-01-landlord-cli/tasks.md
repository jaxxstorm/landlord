## 1. CLI Scaffolding

- [x] 1.1 Add `cmd/cli` entrypoint with Cobra root command and Fang integration
- [x] 1.2 Wire Viper configuration loading (flags, env, config file) for CLI settings
- [x] 1.3 Add API client package for Landlord tenant endpoints (create/list/delete)

## 2. Tenant Commands

- [x] 2.1 Implement `create` command and request/response handling
- [x] 2.2 Implement `list` command with formatted output
- [x] 2.3 Implement `delete` command with confirmation output

## 3. Styling and UX

- [x] 3.1 Add Lip Gloss styles for headers, success, and error output
- [x] 3.2 Add consistent error handling and exit codes across commands

## 4. Docs and Verification

- [x] 4.1 Document CLI usage and config in README or docs
- [x] 4.2 Add `go run ./cmd/cli` examples for local testing

## 5. Tests

- [x] 5.1 Add unit tests for CLI API client
- [x] 5.2 Add integration-style CLI command tests with a mock HTTP server
