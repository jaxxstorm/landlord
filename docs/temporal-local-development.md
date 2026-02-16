# Temporal Local Development

This guide runs Landlord locally with Temporal as the workflow provider using Docker Compose.

## Prerequisites

- Docker and Docker Compose
- Access to `/var/run/docker.sock` (for Docker compute provider)

## Start the Temporal stack

```bash
docker compose -f docker-compose.temporal.yml up --build -d
```

## Verify service readiness

```bash
docker compose -f docker-compose.temporal.yml ps
curl -sS http://localhost:8081/health
```

Temporal endpoints:

- gRPC: `localhost:7233`
- Temporal UI: `http://localhost:8233`

Landlord API endpoint:

- `http://localhost:8081`

## Run a quick tenant lifecycle smoke test

```bash
curl -sS -X POST http://localhost:8081/v1/tenants \
  -H 'Content-Type: application/json' \
  -d '{"name":"acme","email":"admin@acme.test"}'

curl -sS http://localhost:8081/v1/tenants
```

## Follow logs

```bash
docker compose -f docker-compose.temporal.yml logs -f landlord worker-temporal temporal
```

## Stop and clean up

```bash
docker compose -f docker-compose.temporal.yml down -v
```
