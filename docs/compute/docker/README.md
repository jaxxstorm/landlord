# Docker Compute Provider

The Docker compute provider enables Landlord to provision and manage tenant workloads as Docker containers. Each tenant is provisioned with exactly one container, allowing for lightweight, isolated execution environments.

## Features

- **Single Container Per Tenant**: Each tenant runs exactly one Docker container with its own configuration
- **Container-Accessible API**: The Docker daemon can be accessed from within a container using Docker-in-Docker (DinD) or via a Unix socket mount
- **Configurable Docker Host**: Support for local and remote Docker daemons
- **Resource Limits**: CPU and memory constraints configurable per tenant
- **Port Mapping**: Expose container ports to the host
- **Environment Variables**: Configure container environment at provisioning time
- **Automatic Cleanup**: Containers are removed when tenants are destroyed

## Configuration

### Basic Configuration

Add Docker provider configuration to your `config.yaml`:

```yaml
compute:
  default_provider: docker
  docker:
    host: ""  # Empty defaults to Docker socket. Use "tcp://localhost:2375" for TCP connections
    network_name: bridge
    network_driver: bridge
    label_prefix: landlord
```

### Environment Variables

Docker host can be configured via the `DOCKER_HOST` environment variable:

```bash
export DOCKER_HOST=unix:///var/run/docker.sock
export DOCKER_HOST=tcp://docker-daemon:2375
```

### Configuration Options

- **host** (optional): Docker API endpoint
  - Default: Empty string (uses standard Docker socket location)
  - Examples:
    - `unix:///var/run/docker.sock` - Unix socket (local Docker)
    - `tcp://localhost:2375` - TCP connection (unsecured)
    - `tcp://localhost:2376` - TCP connection (secured with TLS)
  - Can be overridden by `DOCKER_HOST` environment variable

- **network_name** (optional): Docker network for tenant containers
  - Default: `bridge`
  - Use `overlay` for swarm deployments

- **network_driver** (optional): Network driver type
  - Default: `bridge`

- **label_prefix** (optional): Prefix for container labels
  - Default: `landlord`

## In-Container Docker Access

When running Landlord inside a Docker container, you need to enable Docker-in-Docker (DinD) or mount the host Docker socket.

### Option 1: Mount Docker Socket (Recommended)

Mount the host Docker socket into the Landlord container:

```bash
docker run -v /var/run/docker.sock:/var/run/docker.sock \
  -e DOCKER_HOST=unix:///var/run/docker.sock \
  landlord:latest
```

Or in docker-compose:

```yaml
services:
  landlord:
    image: landlord:latest
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      DOCKER_HOST: unix:///var/run/docker.sock
```

### Option 2: Docker-in-Docker

For isolated Docker environments, use Docker-in-Docker:

```bash
docker run --privileged \
  -e DOCKER_HOST=tcp://docker-dind:2375 \
  landlord:latest
```

With docker-compose:

```yaml
services:
  docker-dind:
    image: docker:dind
    privileged: true

  landlord:
    image: landlord:latest
    depends_on:
      - docker-dind
    environment:
      DOCKER_HOST: tcp://docker-dind:2375
```

## Tenant Provisioning

### Creating a Tenant with Docker Compute

```json
{
  "tenant_id": "acme-corp",
  "provider_type": "docker",
  "containers": [
    {
      "name": "app",
      "image": "myapp:1.0",
      "command": ["./start.sh"],
      "args": ["--config=/etc/app/config.yaml"],
      "env": {
        "LOG_LEVEL": "debug",
        "DATABASE_URL": "postgres://db:5432/acme"
      },
      "ports": [
        {
          "container_port": 8080,
          "host_port": 8080,
          "protocol": "tcp",
          "name": "http"
        }
      ]
    }
  ],
  "resources": {
    "cpu": 1000,      // millicores (1000 = 1 CPU)
    "memory": 512     // megabytes
  }
}
```

## Container Naming

Containers are automatically named with the pattern:

```
{label_prefix}-tenant-{tenant_id}
```

For example, with default settings and tenant ID `acme-corp`:
```
landlord-tenant-acme-corp
```

## Status Checking

Get the current status of a tenant's container:

```bash
GET /tenants/{tenant_id}/compute/status
```

Response includes:
- Container state (running, stopped, failed)
- Health status
- Container metadata (ID, image)
- Port mappings

## Limitations

### Single Container Requirement

The Docker provider enforces exactly one container per tenant. Multi-container deployments (sidecars, etc.) are not supported. Consider using orchestration tools like Kubernetes if you need more complex deployments.

### Port Management

- Port `0` for `host_port` means no host port mapping
- Ensure no port conflicts between tenant containers
- Port mappings are on `0.0.0.0` (all interfaces)

### Resource Limits

CPU is specified in millicores (1000 = 1 CPU). Memory is in megabytes. These are hard limits enforced by Docker.

## Troubleshooting

### Docker Daemon Connection Errors

**Error**: `failed to connect to docker daemon`

Check that:
1. Docker daemon is running: `docker ps`
2. Socket permissions are correct: `ls -la /var/run/docker.sock`
3. User has Docker access: `docker info` works without sudo
4. `DOCKER_HOST` environment variable is set correctly

### In-Container Connection Issues

If running Landlord in a container:
1. Verify the socket is mounted: `ls -la /var/run/docker.sock`
2. Check Docker group membership inside container
3. Ensure firewall rules allow Docker daemon communication

### Container Startup Failures

**Error**: `failed to start container`

Check:
1. Image exists and is accessible: `docker pull {image}`
2. Sufficient resources available
3. No port conflicts with existing containers

### Image Pull Errors

**Error**: `Error response from daemon: manifest not found`

Ensure:
1. Image name and tag are correct
2. Image is public or credentials are configured
3. Image exists on the registry

## Provider-Specific Configuration

The Docker provider supports additional validation and configuration through `provider_config`:

```json
{
  "tenant_id": "example",
  "provider_type": "docker",
  "containers": [...],
  "resources": {...},
  "provider_config": {
    "restart_policy": "unless-stopped",
    "log_driver": "json-file"
  }
}
```

## Examples

### Simple Web Service

```yaml
tenant_id: web-app
provider_type: docker
containers:
  - name: web
    image: nginx:alpine
    ports:
      - container_port: 80
        host_port: 8080
        protocol: tcp
resources:
  cpu: 500
  memory: 256
```

### Application with Configuration

```yaml
tenant_id: api-service
provider_type: docker
containers:
  - name: api
    image: mycompany/api:v2.1.0
    command: ["/app/api"]
    args: ["--config", "/etc/app/config.yaml"]
    env:
      LOG_LEVEL: info
      ENVIRONMENT: production
      DATABASE_HOST: postgres.internal
      REDIS_HOST: redis.internal
    ports:
      - container_port: 3000
        host_port: 3000
        protocol: tcp
        name: http
      - container_port: 9090
        host_port: 9090
        protocol: tcp
        name: metrics
resources:
  cpu: 2000
  memory: 1024
```

## Security Considerations

1. **Socket Access**: Mounting `/var/run/docker.sock` grants full Docker access. Restrict to trusted environments.
2. **Image Trust**: Validate container images before provisioning. Consider using signed images or private registries.
3. **Network Isolation**: Use Docker networks to isolate tenant traffic.
4. **Resource Limits**: Always set CPU and memory limits to prevent resource exhaustion.
5. **Secrets Management**: Use environment variables carefully; prefer secret management systems.

## Performance Notes

- Container startup typically takes 1-5 seconds depending on image size
- Docker socket access is very fast (microseconds)
- In-container socket mounts have minimal overhead
- Network throughput is near-native (bridge mode)
