# Landlord CLI

This CLI interacts with the Landlord API. You can point it at a specific API URL with `--api-url` or the `LANDLORD_CLI_API_URL` environment variable.

## Create a tenant

Create a tenant with an image:

```bash
go run . create --tenant-name lbr --image nginx:latest
```

Include compute configuration (provider-specific) with `--config` as JSON. Example for Docker:

```bash
go run . create \
  --tenant-name lbr \
  --image nginx:latest \
  --config '{"env":{"FOO":"bar"},"ports":[{"container_port":8080}]}'
```

## Archive a tenant

Archive removes compute resources but keeps the tenant record:

```bash
go run . archive --tenant-name lbr
```

## Delete a tenant

Delete enqueues an archive, then the platform deletes the tenant record after archival:

```bash
go run . delete --tenant-name lbr
```

## Update a tenant (set)

Modify image or config:

```bash
go run . set --tenant-name lbr --image nginx:1.25
```

With compute config:

```bash
go run . set \
  --tenant-name lbr \
  --config '{"env":{"BAZ":"qux"}}'
```

## Discover compute config schema

Fetch the active provider schema and defaults:

```bash
go run . compute
```
