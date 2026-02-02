## 1. Compute config discovery API

- [x] 1.1 Define compute config discovery response model (provider, schema, defaults) in API models
- [x] 1.2 Add compute config discovery endpoint and handler in API server
- [x] 1.3 Wire endpoint into router and update Swagger annotations
- [x] 1.4 Add unit tests for compute config discovery handler

## 2. Compute provider schema exposure

- [x] 2.1 Extend compute provider interface to expose config JSON schema and optional defaults
- [x] 2.2 Implement Docker provider schema (and defaults if applicable)
- [x] 2.3 Add provider registry method to fetch active provider schema
- [x] 2.4 Add validation helper to validate compute_config against active provider schema

## 3. Request validation for compute_config

- [x] 3.1 Update tenant create request validation to parse compute_config as JSON object
- [x] 3.2 Validate compute_config against active provider schema at API ingress
- [x] 3.3 Update tenant update request validation to accept compute_config and validate schema
- [x] 3.4 Add error mapping for schema validation failures (clear 400 responses)

## 4. CLI compute config support

- [x] 4.1 Add `compute` CLI command to call compute config discovery endpoint
- [x] 4.2 Update create command to accept `--config` JSON and send compute_config
- [x] 4.3 Update set command to accept `--config` JSON and send compute_config
- [x] 4.4 Update CLI output formatting for compute config schema and defaults
- [x] 4.5 Add CLI tests for compute command and config payloads

## 5. Documentation and integration

- [x] 5.1 Update API docs to include compute config discovery endpoint schema
- [x] 5.2 Update CLI docs/examples for compute command and config payloads
- [x] 5.3 Add end-to-end test for create/set with compute_config
