# Restate Provider - Local Development Setup

Complete guide for setting up Restate locally for development using Docker Compose.

## Prerequisites

- Docker and Docker Compose installed
- 4GB RAM available (minimum for Restate container)
- Landlord source code

## Docker Compose Setup

### 1. Docker Compose Configuration

A `docker-compose.yml` file is included with Landlord for easy local setup:

```yaml
version: '3.8'

services:
  restate:
    image: restatedev/restate:latest
    ports:
      - "8080:8080"
    environment:
      - LOG_LEVEL=info
    volumes:
      - restate-data:/tmp/restate
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 5s
      timeout: 3s
      retries: 5
      start_period: 10s

volumes:
  restate-data:
```

### 2. Start Restate

From the Landlord project root:

```bash
docker-compose up restate
```

Wait for the health check to pass:

```
restate_1  | 2026-01-30 14:28:22 INFO restate server started on 0.0.0.0:8080
```

You can verify Restate is running by visiting:
```
http://localhost:8080/health
```

Should return:
```json
{"status":"ok"}
```

### 3. Configuration for Local Development

Create or update `config.yaml` in your Landlord directory:

```yaml
workflow:
  default_provider: "restate"
  restate:
    endpoint: "http://localhost:8080"
    execution_mechanism: "local"
    auth_type: "none"
    timeout: 30m
    retry_attempts: 3
```

Or use environment variables:

```bash
export LANDLORD_WORKFLOW_DEFAULT_PROVIDER=restate
export LANDLORD_WORKFLOW_RESTATE_ENDPOINT=http://localhost:8080
export LANDLORD_WORKFLOW_RESTATE_EXECUTION_MECHANISM=local
export LANDLORD_WORKFLOW_RESTATE_AUTH_TYPE=none
```

### 4. Start Landlord

In another terminal:

```bash
cd /path/to/landlord
go run ./cmd/landlord server
```

Landlord should connect to Restate:

```
2026-01-30 14:28:25 INFO  registering Restate workflow provider endpoint=http://localhost:8080 execution_mechanism=local
2026-01-30 14:28:25 INFO  registered workflow providers providers=[mock restate step-functions]
```

### 5. Create and Execute a Workflow

```bash
# Create a simple workflow
cat > workflow.yaml << 'EOF'
workflows:
  - id: hello-world
    definition: echo "Hello, Restate!"
EOF

# Register the workflow
landlord workflow create hello-world --definition workflow.yaml

# Start execution
landlord workflow execute hello-world

# Check status
landlord workflow status hello-world
```

## Common Commands

### Stop Restate

```bash
docker-compose stop restate
```

### Restart Restate (preserves data)

```bash
docker-compose restart restate
```

### Remove Restate and Data

```bash
docker-compose down
docker-compose down -v  # Also remove volumes
```

### View Restate Logs

```bash
docker-compose logs -f restate
```

### Access Restate Console

Open your browser to:
```
http://localhost:8080
```

The console shows:
- Running workflows and executions
- Service registry
- Execution history
- State inspection

## Troubleshooting

### Container Won't Start

Check Docker is running:
```bash
docker ps
```

Check logs for errors:
```bash
docker-compose logs restate
```

### Port 8080 Already in Use

Either stop the conflicting service or map to a different port:

```yaml
ports:
  - "9080:8080"  # Use port 9080 instead
```

Then update Landlord config:
```yaml
restate:
  endpoint: "http://localhost:9080"
```

### Connection Refused

Ensure Restate container is running:
```bash
docker-compose ps
```

Should show:
```
NAME      STATUS
restate   Up 2 minutes (healthy)
```

Wait for health check to pass (initially takes ~10 seconds).

### Workflow Execution Fails

Check Landlord logs for error messages and Restate logs:

```bash
docker-compose logs restate | tail -20
```

### Data Persistence Issues

Restate data is stored in a Docker volume. To preserve data between restarts:

```bash
# Already configured in docker-compose.yml
volumes:
  - restate-data:/tmp/restate
```

To start fresh (clear all state):
```bash
docker-compose down -v  # Remove volumes
docker-compose up restate  # Start with clean state
```

## Performance Tuning

### Increase Memory Allocation

Edit `docker-compose.yml`:

```yaml
services:
  restate:
    deploy:
      resources:
        limits:
          memory: 8G  # Increase from 4G
```

### Enable Debug Logging

```yaml
services:
  restate:
    environment:
      - LOG_LEVEL=debug  # More verbose logs
```

## Integration with Tests

### Run Tests Against Local Restate

Unit tests can run against the local Restate container:

```bash
# Start Restate
docker-compose up -d restate

# Run tests
go test ./... -v

# Stop Restate
docker-compose down
```

### CI/CD Integration

For GitHub Actions or other CI systems:

```yaml
services:
  restate:
    image: restatedev/restate:latest
    ports:
      - "8080:8080"
    options: --health-cmd="curl -f http://localhost:8080/health"
```

## Next Steps

- [Configuration Reference](restate-configuration.md) - Customize configuration
- [Production Deployment](restate-production-deployment.md) - Deploy to production
- [Getting Started](restate-getting-started.md) - Basic usage guide
