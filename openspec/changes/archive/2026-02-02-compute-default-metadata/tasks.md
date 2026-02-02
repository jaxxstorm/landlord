## 1. Define default metadata conventions

- [x] 1.1 Decide canonical metadata keys (owner, tenant ID, optional tenant name) and namespace prefix
- [x] 1.2 Add helper for default metadata map in compute domain layer

## 2. Apply metadata in Docker provider

- [x] 2.1 Attach default labels to Docker containers during provision
- [x] 2.2 Ensure any other Docker-managed resources include the default labels

## 3. Tests and verification

- [x] 3.1 Add or update Docker provider tests to assert default labels
- [x] 3.2 Verify metadata is applied for tenant provisioning flows
