# Quickstart (Local)

This quickstart uses local defaults with SQLite and mock providers.

## Prerequisites

- Go 1.24+
- Docker (optional, only if you want the Docker compute provider)

## 1. Create a local config

Copy the example config and update it for local use:

```bash
cp config.example.yaml config.yaml
```

Edit `config.yaml` and set:

```yaml
database:
  provider: sqlite
  sqlite:
    path: landlord.db

compute:
  mock: {}

workflow:
  default_provider: mock
```

## 2. Build and run

```bash
make build
./landlord
```

The API will be available at `http://localhost:8080`.

## 3. Create a tenant

```bash
go run ./cmd/cli create --tenant-name demo-tenant \
  --config '{"image":"nginx:alpine"}'
```

## 4. Inspect tenant state

```bash
go run ./cmd/cli list
go run ./cmd/cli get --tenant-id <tenant-id>
```

## Next steps

- Review the architecture overview in `README.md`.
- Explore providers in `compute-providers.md` and `workflow-providers.md`.
- Browse the API in `api.md`.
