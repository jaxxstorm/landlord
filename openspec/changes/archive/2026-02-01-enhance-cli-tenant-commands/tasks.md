## 1. Client updates

- [x] 1.1 Add client method to fetch a single tenant by ID/name
- [x] 1.2 Add client method to update tenant image/config with partial payload
- [x] 1.3 Add client tests for get/update request/response handling

## 2. CLI commands

- [x] 2.1 Add `get` command with `--tenant-id` validation and styled output
- [x] 2.2 Add `set` command with optional `--image`/`--config` flags and validation
- [x] 2.3 Add shared output helper for single-tenant rendering

## 3. Tests and docs

- [x] 3.1 Add command-level tests for get/set
- [x] 3.2 Update README CLI usage examples for get/set
- [x] 3.3 Verify `go run ./cmd/cli` flows for new commands
