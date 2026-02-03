# Mock Compute Provider

The mock compute provider is used for tests and demos. It does not provision real resources.

## Tenant compute_config reference

The mock provider does not define any provider-specific fields. Any keys provided are accepted and ignored.

### Full JSON example

```json
{}
```

### Full YAML example

```yaml
{}
```

### Using file:// with the CLI

```bash
go run ./cmd/cli create --tenant-name demo \
  --config file:///path/to/mock-compute-config.yaml
```
