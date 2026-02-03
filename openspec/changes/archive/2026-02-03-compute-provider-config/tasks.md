## 1. Configuration Schema & Validation

- [x] 1.1 Update config structs and mapstructure tags to remove `compute.default_provider` and `compute.defaults`
- [x] 1.2 Implement provider-keyed compute config parsing (enable provider if block present)
- [x] 1.3 Add validation errors for legacy keys and unknown provider blocks
- [x] 1.4 Add provider-specific validation on startup for enabled providers

## 2. Compute Discovery API & CLI

- [x] 2.1 Update compute discovery API to accept provider identifier and return schema/defaults for that provider
- [x] 2.2 Return clear error for unknown or disabled providers in discovery endpoint
- [x] 2.3 Update CLI `landlord-cli compute` to accept `--provider` and render new response
- [x] 2.4 Update OpenAPI/Swagger annotations and schemas for new discovery parameter

## 3. Documentation, Examples, and Tests

- [x] 3.1 Update config examples (`config.example.yaml`, `docker.config.yaml`, `ecs.config.yaml`, `test.config.yaml`) to new provider-keyed format
- [x] 3.2 Update docs to describe provider-keyed compute configuration and breaking change
- [x] 3.3 Add/adjust tests for config loading/validation with single and multiple providers
- [x] 3.4 Add/adjust tests for discovery endpoint and CLI provider parameter
