# Docker Provider Quick Reference

## Quick Start

### 1. Enable Docker Provider in Configuration

```yaml
compute:
  docker:
    image: "nginx:latest"
    host: ""  # Uses Docker socket by default
```

### 2. Run Landlord in Docker

```bash
# Using socket mount
docker run -v /var/run/docker.sock:/var/run/docker.sock \
  -e DOCKER_HOST=unix:///var/run/docker.sock \
  landlord:latest
```

### 3. Provision a Tenant

```bash
curl -X POST http://localhost:8080/tenants \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "my-app",
    "provider_type": "docker",
    "containers": [{
      "name": "web",
      "image": "nginx:alpine",
      "ports": [{
        "container_port": 80,
        "host_port": 8080,
        "protocol": "tcp"
      }]
    }],
    "resources": {
      "cpu": 1000,
      "memory": 512
    }
  }'
```

## API Endpoints

### Provision Tenant
- **POST** `/tenants` - Create a new tenant
- Request: `TenantComputeSpec`
- Response: `ProvisionResult`

### Get Tenant Status
- **GET** `/tenants/{tenant_id}/compute/status`
- Response: `ComputeStatus`

### Update Tenant
- **PUT** `/tenants/{tenant_id}` - Update tenant configuration
- Request: `TenantComputeSpec`
- Response: `UpdateResult`

### Delete Tenant
- **DELETE** `/tenants/{tenant_id}` - Destroy tenant resources
- Response: Success/Error

## Configuration Reference

### Docker Config Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `host` | string | "" | Docker API endpoint (socket or TCP) |
| `network_name` | string | "bridge" | Docker network for containers |
| `network_driver` | string | "bridge" | Network driver type |
| `label_prefix` | string | "landlord" | Container label prefix |

### Host Examples

| Use Case | Host Value |
|----------|-----------|
| Local Docker | `""` (empty, uses socket) |
| Unix Socket | `unix:///var/run/docker.sock` |
| TCP Unsecured | `tcp://docker-host:2375` |
| TCP Secured (TLS) | `tcp://docker-host:2376` |
| Docker-in-Docker | `tcp://docker-dind:2375` |

### Environment Variables

```bash
# Override Docker host
export DOCKER_HOST=tcp://remote-docker:2375

# Landlord will automatically use this value
```

## Common Tasks

### List Provisioned Containers

```bash
docker ps -f "label=landlord.tenant=*"
```

### View Container Logs

```bash
docker logs landlord-tenant-{tenant_id}
```

### Inspect Container

```bash
docker inspect landlord-tenant-{tenant_id}
```

### Manually Stop Container

```bash
# Note: Normally Landlord manages this
docker stop landlord-tenant-{tenant_id}
```

## Troubleshooting

### Docker Daemon Not Reachable

```
Error: failed to connect to docker daemon

Solution:
1. Check Docker is running: docker ps
2. Check socket permissions: ls -la /var/run/docker.sock
3. Check user is in docker group: groups $USER
4. Verify DOCKER_HOST env var is correct
```

### Port Already in Use

```
Error: Bind address already in use

Solution:
1. Change host_port in tenant spec
2. Or stop existing container: docker stop container-name
3. Check port: lsof -i :8080
```

### Image Pull Failed

```
Error: image not found

Solution:
1. Verify image exists: docker pull {image}
2. Check registry credentials
3. Use full image reference: registry/namespace/image:tag
```

### Container Exits Immediately

```
Error: Container status is exited

Solution:
1. Check logs: docker logs container-name
2. Verify image is correct
3. Check command and arguments are valid
4. Ensure port mappings don't conflict
```

## Performance Tips

1. **Use Local Images**: Pre-pull images on the host to avoid network delays
2. **Resource Limits**: Set appropriate CPU/memory limits to prevent resource exhaustion
3. **Network**: Use host network mode if latency is critical (not recommended for isolation)
4. **Monitoring**: Track container resource usage with `docker stats`

## Security Best Practices

1. **Socket Mount**: Only mount `/var/run/docker.sock` to trusted containers
2. **Image Trust**: Use signed images and private registries
3. **Resource Limits**: Always set CPU and memory limits
4. **Secrets**: Use environment variables carefully; prefer secret management systems
5. **Network Policy**: Implement firewall rules for inter-tenant communication

## Docker Compose Integration

Run full stack with docker-compose:

```yaml
version: '3.8'
services:
  landlord:
    image: landlord:latest
    ports:
      - "8080:8080"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      DOCKER_HOST: unix:///var/run/docker.sock
      LOG_LEVEL: info
    depends_on:
      - postgres
  
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: landlord
      POSTGRES_USER: landlord
      POSTGRES_PASSWORD: secret
    volumes:
      - pgdata:/var/lib/postgresql/data

volumes:
  pgdata:
```

## Container Naming Convention

Tenants are provisioned with container names: `{label_prefix}-tenant-{tenant_id}`

Example:
- Tenant ID: `acme-corp`
- Container Name: `landlord-tenant-acme-corp`

Use this for querying or debugging specific tenant containers.
