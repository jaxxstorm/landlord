# Landlord

Landlord is a tenant provisioning control plane that manages long-lived tenants through declarative desired state, workflow orchestration, and pluggable compute.

> Disclosure: This project was built with substantial assistance from OpenSpec and AI tools. All outputs are reviewed and curated by the maintainers.

## What it does

Landlord accepts desired tenant definitions, reconciles them to actual infrastructure, and tracks lifecycle state in a durable database. It is designed to be pluggable so different workflow, compute, and database backends can be used without changing the API.

## How it works (quick example)

1. A tenant definition is submitted to the API.
2. The controller reconciles desired vs observed state.
3. A workflow provider executes the provisioning plan.
4. A worker performs compute actions and reports status.
5. The database persists state, history, and transitions.

Example CLI flow:

```bash
go run ./cmd/cli create --tenant-name demo-tenant \
  --config '{"image":"nginx:alpine"}'
```

Multi-line JSON example:

```bash
go run ./cmd/cli create --tenant-name demo-tenant \
  --config '{
    "image": "nginx:alpine",
    "env": {
      "FOO": "bar"
    },
    "ports": [
      {
        "container_port": 8080
      }
    ]
  }'
```

YAML example:

```bash
go run ./cmd/cli create --tenant-name demo-tenant \
  --config 'image: "nginx:alpine"\nenv:\n  FOO: "bar"\nports:\n  - container_port: 8080'
```

## API versioning

All HTTP APIs are versioned with a path prefix. The current stable version is `v1`, so endpoints look like `/v1/tenants` and `/v1/compute/config`.

## Documentation

Public docs live in `docs/` and are viewable with Docsify. Start with:

- `docs/README.md` for the architecture overview and component matrix
- `docs/quickstart.md` to run locally
- `docs/api.md` to browse the OpenAPI spec in Swagger UI

## Components

- **API**: Validates requests and exposes tenant lifecycle and status APIs.
- **Controller**: Reconciles desired state and drives workflow execution.
- **Workflow provider**: Executes provisioning plans (pluggable).
- **Worker**: Performs compute actions for workflows (pluggable).
- **Compute provider**: Provisions runtime infrastructure (pluggable).
- **Database**: Persists tenants, transitions, and audit history (pluggable).

## Local quickstart (short)

```bash
make build
./landlord
```

See `docs/quickstart.md` for a complete local setup.
