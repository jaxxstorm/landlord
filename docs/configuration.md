# Configuration Guide

Landlord supports flexible configuration through multiple sources: configuration files (YAML/JSON), environment variables, and CLI flags. This guide covers all configuration options and how to use them.

## Quick Start

### Minimal Configuration (Using Defaults)

Start landlord with defaults (all configuration from environment):

```bash
./landlord
```

### Using Environment Variables

Set configuration via environment variables:

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=landlord
export DB_PASSWORD=secret
export DB_DATABASE=landlord_db
export LOG_LEVEL=info
export HTTP_PORT=8080
./landlord
```

### Using a Configuration File

Create `config.yaml` in the current directory:

```yaml
database:
  host: localhost
  port: 5432
  user: landlord
  password: landlord_password
  database: landlord_db
http:
  port: 8080
log:
  level: info
```

Then run:

```bash
./landlord
```

### Using CLI Flags

Override specific settings with CLI flags:

```bash
./landlord --db-host prod.db.example.com --db-port 5432 --log-level debug
```

### Mixed Configuration

Combine configuration file, environment variables, and CLI flags (precedence is: CLI flags > environment variables > config file > defaults):

```bash
# config.yaml has: db_host=localhost
# Set environment variable
export DB_PORT=3306

# Run with CLI flag - takes highest priority
./landlord --log-level debug
# Result: db_host from config.yaml, db_port from env var, log_level from CLI flag
```

---

## Configuration Precedence

Configuration values are resolved in this order (highest to lowest priority):

1. **CLI flags** - Highest priority, always takes effect
2. **Environment variables** - Overrides config file values
3. **Configuration file** - YAML or JSON file values
4. **Defaults** - Built-in defaults, lowest priority

### Example Precedence Resolution

Given:
- `config.yaml`: `database.host: localhost`
- Environment: `DB_HOST=db.example.com`
- CLI flag: `--db-host cli.example.com`

**Result**: `cli.example.com` (CLI flag wins)

---

## Configuration File Format

### Supported Formats

- **YAML** (`config.yaml` or `config.yml`)
- **JSON** (`config.json`)

### File Search Locations

Landlord searches for configuration files in this order:

1. **Explicit path via `--config` flag** (if provided)
2. **`LANDLORD_CONFIG` environment variable** (if set)
3. **Current working directory**: `./config.yaml` or `./config.json`
4. **System configuration**: `/etc/landlord/config.yaml` or `/etc/landlord/config.json`
5. **User home directory**: `$XDG_CONFIG_HOME/landlord/config.yaml` or `~/.config/landlord/config.yaml`

The first file found is used; remaining locations are not checked.

### Using Explicit Configuration Path

Specify an explicit configuration file:

```bash
./landlord --config /etc/landlord/production.yaml
```

Or via environment variable:

```bash
export LANDLORD_CONFIG=/opt/landlord/custom-config.yaml
./landlord
```

---

## YAML Configuration Format

### Basic YAML Structure

```yaml
database:
  host: localhost
  port: 5432
  user: landlord
  password: landlord_password
  database: landlord_db
  ssl_mode: prefer
  max_connections: 25
  min_connections: 5
  connect_timeout: 10s
  max_conn_lifetime: 1h
  max_conn_idle_time: 30m

http:
  host: 0.0.0.0
  port: 8080
  read_timeout: 10s
  write_timeout: 10s
  idle_timeout: 120s
  shutdown_timeout: 30s

log:
  level: info
  format: development

compute:
  mock: {}

workflow:
  default_provider: mock

controller:
  enabled: true
  reconciliation_interval: 10s
  workers: 3
  workflow_trigger_timeout: 30s
  shutdown_timeout: 30s
  max_retries: 5
```

### YAML Data Types

- **String**: `host: localhost` (quoted or unquoted)
- **Integer**: `port: 8080`
- **Duration**: `timeout: 30s` (valid suffixes: s, m, h; e.g., 10s, 5m, 1h)
- **Boolean**: `debug: true` or `debug: false`
- **Nested objects**: Use indentation

### YAML Example: Development Environment

```yaml
database:
  host: localhost
  port: 5432
  user: dev_user
  password: dev_password
  database: landlord_dev

http:
  host: 127.0.0.1
  port: 8080

log:
  level: debug
  format: development

compute:
  mock: {}

workflow:
  default_provider: mock

controller:
  enabled: true
  reconciliation_interval: 5s
  workers: 2
```

### YAML Example: Production Environment

```yaml
database:
  host: prod-db.internal.example.com
  port: 5432
  user: landlord_prod
  password: ${DB_PASSWORD}  # Use environment variables in your deployment
  database: landlord_production
  ssl_mode: require
  max_connections: 100
  min_connections: 20
  connect_timeout: 5s
  max_conn_lifetime: 30m
  max_conn_idle_time: 10m

http:
  host: 0.0.0.0
  port: 8080
  read_timeout: 30s
  write_timeout: 30s
  shutdown_timeout: 30s

log:
  level: info
  format: production

compute:
  ecs:
    cluster_arn: arn:aws:ecs:us-east-1:123456789012:cluster/landlord
    task_definition_arn: arn:aws:ecs:us-east-1:123456789012:task-definition/tenant-app:12
    service_name_prefix: landlord-tenant-
    service_name_prefix: landlord-tenant-

workflow:
  default_provider: step-functions
  step_functions:
    region: us-east-1
    role_arn: arn:aws:iam::123456789012:role/landlord-workflow-executor

controller:
  enabled: true
  reconciliation_interval: 10s
  workers: 5
  workflow_trigger_timeout: 60s
  shutdown_timeout: 60s
  max_retries: 10
```

---

## JSON Configuration Format

### Basic JSON Structure

```json
{
  "database": {
    "host": "localhost",
    "port": 5432,
    "user": "landlord",
    "password": "landlord_password",
    "database": "landlord_db",
    "ssl_mode": "prefer",
    "max_connections": 25,
    "min_connections": 5,
    "connect_timeout": "10s",
    "max_conn_lifetime": "1h",
    "max_conn_idle_time": "30m"
  },
  "http": {
    "host": "0.0.0.0",
    "port": 8080,
    "read_timeout": "10s",
    "write_timeout": "10s",
    "idle_timeout": "120s",
    "shutdown_timeout": "30s"
  },
  "log": {
    "level": "info",
    "format": "development"
  },
  "compute": {
    "mock": {}
  },
  "workflow": {
    "default_provider": "mock",
    "step_functions": {
      "region": "us-west-2",
      "role_arn": ""
    }
  }
}
```

### JSON Data Types

- **String**: `"host": "localhost"`
- **Number**: `"port": 8080`
- **Duration string**: `"timeout": "30s"` (must be quoted)
- **Boolean**: `"debug": true` or `"debug": false`
- **Null**: `"role_arn": null`
- **Objects/Arrays**: Standard JSON nesting

### JSON Example: Minimal Configuration

```json
{
  "database": {
    "host": "localhost",
    "port": 5432
  },
  "http": {
    "port": 8080
  },
  "log": {
    "level": "info"
  }
}
```

---

## Environment Variables

All configuration can be set via environment variables. Environment variables follow the pattern: `{SECTION}_{FIELD}`.

### Database Configuration

Landlord supports multiple database providers. See [database/sqlite.md](./database/sqlite.md) for SQLite-specific configuration.

#### PostgreSQL (Default)

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `DB_PROVIDER` | string | `postgres` | Database provider: `postgres` or `sqlite` |
| `DB_HOST` | string | `localhost` | Database host (PostgreSQL only) |
| `DB_PORT` | int | `5432` | Database port (PostgreSQL only) |
| `DB_USER` | string | (required) | Database username (PostgreSQL only) |
| `DB_PASSWORD` | string | (required) | Database password (PostgreSQL only) |
| `DB_DATABASE` | string | (required) | Database name (PostgreSQL only) |
| `DB_SSLMODE` | string | `prefer` | SSL mode: disable, allow, prefer, require, verify-ca, verify-full (PostgreSQL only) |
| `DB_MAX_CONNECTIONS` | int | `25` | Maximum connection pool size |
| `DB_MIN_CONNECTIONS` | int | `5` | Minimum connection pool size (PostgreSQL only) |
| `DB_CONNECT_TIMEOUT` | duration | `10s` | Connection timeout |
| `DB_MAX_CONN_LIFETIME` | duration | `1h` | Maximum lifetime of a connection |
| `DB_MAX_CONN_IDLE_TIME` | duration | `30m` | Maximum idle time before reaping |

#### SQLite

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `SQLITE_PATH` | string | `:memory:` | Path to SQLite database file (use `:memory:` for in-memory) |
| `SQLITE_BUSY_TIMEOUT` | duration | `5s` | How long to wait when database is locked |

For more details on SQLite configuration, see [Database Provider: SQLite](./database/sqlite.md).

### HTTP Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `HTTP_HOST` | string | `0.0.0.0` | HTTP server bind address |
| `HTTP_PORT` | int | `8080` | HTTP server port |
| `HTTP_READ_TIMEOUT` | duration | `10s` | HTTP read timeout |
| `HTTP_WRITE_TIMEOUT` | duration | `10s` | HTTP write timeout |
| `HTTP_IDLE_TIMEOUT` | duration | `120s` | HTTP idle timeout |
| `HTTP_SHUTDOWN_TIMEOUT` | duration | `30s` | Graceful shutdown timeout |

### Logging Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `LOG_LEVEL` | string | `info` | Log level: debug, info, warn, error |
| `LOG_FORMAT` | string | `development` | Log format: development, production |

### Compute Configuration

Compute providers are configured in the config file via provider blocks (e.g., `compute.docker`). There is no global compute provider environment variable; use provider-specific variables like `DOCKER_HOST` as needed.

### Workflow Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `WORKFLOW_DEFAULT_PROVIDER` | string | `mock` | Default workflow provider |
| `WORKFLOW_SFN_REGION` | string | `us-west-2` | AWS Step Functions region |
| `WORKFLOW_SFN_ROLE_ARN` | string | (empty) | AWS Step Functions execution role ARN |

### Controller Configuration

The tenant reconciliation controller continuously monitors and manages tenant state transitions. These settings control how the controller operates.

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `CONTROLLER_ENABLED` | bool | `true` | Enable tenant reconciliation controller |
| `CONTROLLER_RECONCILIATION_INTERVAL` | duration | `10s` | Polling interval for discovering tenants needing reconciliation |
| `CONTROLLER_WORKERS` | int | `3` | Number of concurrent worker goroutines processing reconciliation queue |
| `CONTROLLER_WORKFLOW_TRIGGER_TIMEOUT` | duration | `30s` | Timeout for workflow trigger operations (prevents hanging on workflow provider) |
| `CONTROLLER_SHUTDOWN_TIMEOUT` | duration | `30s` | Maximum graceful shutdown duration before forcing exit |
| `CONTROLLER_MAX_RETRIES` | int | `5` | Maximum retry attempts before marking tenant as failed |

#### Detailed Configuration Explanations

**CONTROLLER_ENABLED**
- Enable or disable the automatic tenant reconciliation loop
- When disabled, tenants must be manually provisioned via API calls
- Useful for maintenance or when delegating control to external orchestrators

**CONTROLLER_RECONCILIATION_INTERVAL**
- How frequently the controller queries the database for tenants needing work
- Smaller values (e.g., `5s`) provide faster response to state changes but increase database load
- Larger values (e.g., `30s`) reduce load but increase latency in state transitions
- Should be tuned based on your typical tenant lifecycle speed and database capacity

**CONTROLLER_WORKERS**
- Number of concurrent worker goroutines that process reconciliation tasks
- Each worker can handle one tenant reconciliation simultaneously
- Higher worker counts enable faster processing of large numbers of tenants
- Set based on expected concurrency and available CPU/memory
- Rule of thumb: set to (CPU cores / 2) for balanced performance

**CONTROLLER_WORKFLOW_TRIGGER_TIMEOUT**
- Timeout for individual workflow trigger operations
- Prevents the controller from blocking indefinitely if a workflow provider becomes unresponsive
- When exceeded, the reconciliation is retried with exponential backoff
- Should be longer than typical workflow provider response time but short enough to detect failures quickly

**CONTROLLER_SHUTDOWN_TIMEOUT**
- Maximum time to wait for in-flight reconciliations to complete during shutdown
- After timeout, remaining workers are forcefully terminated
- Allows graceful draining of the queue when application needs to stop
- Longer values are safer but increase shutdown time

**CONTROLLER_MAX_RETRIES**
- Maximum number of attempts to reconcile a tenant before marking it as failed
- After exceeding this limit, the tenant transitions to `failed` status with error message
- Failed tenants can be manually retried or investigated by operators
- Prevents infinite retry loops for permanently broken tenants

#### Configuration Examples

**Development (Fast Feedback)**
```yaml
controller:
  enabled: true
  reconciliation_interval: 2s
  workers: 2
  workflow_trigger_timeout: 30s
  shutdown_timeout: 10s
  max_retries: 3
```

**Production (Optimized Throughput)**
```yaml
controller:
  enabled: true
  reconciliation_interval: 10s
  workers: 8
  workflow_trigger_timeout: 60s
  shutdown_timeout: 30s
  max_retries: 5
```

**High-Availability (Conservative)**
```yaml
controller:
  enabled: true
  reconciliation_interval: 15s
  workers: 4
  workflow_trigger_timeout: 120s
  shutdown_timeout: 60s
  max_retries: 10
```

### Example: Setting Database Configuration via Environment

```bash
export DB_HOST=prod-db.example.com
export DB_PORT=5432
export DB_USER=landlord
export DB_PASSWORD=secure_password_here
export DB_DATABASE=landlord_prod
export DB_SSLMODE=require
export LOG_LEVEL=info
export HTTP_PORT=8080

./landlord
```

---

## CLI Flags

All configuration can be overridden via command-line flags. Flags take precedence over environment variables and config files.

### Database Flags

```bash
./landlord \
  --db-host localhost \
  --db-port 5432 \
  --db-user landlord \
  --db-password secret \
  --db-database landlord_db
```

### HTTP Flags

```bash
./landlord \
  --http-host 0.0.0.0 \
  --http-port 8080 \
  --http-read-timeout 10s \
  --http-write-timeout 10s \
  --http-idle-timeout 120s \
  --http-shutdown-timeout 30s
```

### Logging Flags

```bash
./landlord \
  --log-level info \
  --log-format development
```

### Configuration File Flag

```bash
./landlord --config /etc/landlord/production.yaml
```

### Help and Version

```bash
# Display help text
./landlord --help
./landlord -h

# Display version
./landlord --version
```

---

## Common Configuration Scenarios

### Scenario 1: Local Development with Defaults

Use all default values with just database credentials from environment:

```bash
export DB_USER=dev_user
export DB_PASSWORD=dev_password
export DB_DATABASE=landlord_dev

./landlord
```

**Configuration resolved**:
- Database: localhost:5432, user: dev_user, password: dev_password, db: landlord_dev
- HTTP: 0.0.0.0:8080
- Logging: info level, development format
- Compute: mock provider
- Workflow: mock provider

---

### Scenario 2: Docker Container with Environment Variables

Dockerfile:

```dockerfile
FROM golang:1.24-alpine
WORKDIR /app
COPY . .
RUN go build ./cmd/landlord
ENTRYPOINT ["./landlord"]
```

Run with environment variables:

```bash
docker run -e DB_HOST=postgres.internal \
  -e DB_USER=landlord \
  -e DB_PASSWORD=secret \
  -e DB_DATABASE=landlord \
  -e LOG_LEVEL=info \
  -p 8080:8080 \
  landlord:latest
```

---

### Scenario 3: Kubernetes with ConfigMap

Create a ConfigMap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: landlord-config
data:
  config.yaml: |
    database:
      host: postgres.default.svc.cluster.local
      port: 5432
      ssl_mode: require
    http:
      port: 8080
    log:
      level: info
    compute:
      ecs:
        cluster_arn: arn:aws:ecs:us-east-1:123456789012:cluster/landlord
        task_definition_arn: arn:aws:ecs:us-east-1:123456789012:task-definition/tenant-app:12
        service_name_prefix: landlord-tenant-
    workflow:
      default_provider: step-functions
```

Pod manifest:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: landlord
spec:
  containers:
  - name: landlord
    image: landlord:latest
    ports:
    - containerPort: 8080
    env:
    - name: DB_USER
      valueFrom:
        secretKeyRef:
          name: landlord-secrets
          key: db-user
    - name: DB_PASSWORD
      valueFrom:
        secretKeyRef:
          name: landlord-secrets
          key: db-password
    - name: LANDLORD_CONFIG
      value: /etc/landlord/config.yaml
    volumeMounts:
    - name: config
      mountPath: /etc/landlord
  volumes:
  - name: config
    configMap:
      name: landlord-config
```

---

### Scenario 4: Production with Explicit Config File

Create `/etc/landlord/production.yaml`:

```yaml
database:
  host: prod-db.internal.example.com
  port: 5432
  user: landlord_prod
  database: landlord_production
  ssl_mode: require
  max_connections: 100

http:
  host: 0.0.0.0
  port: 8080
  shutdown_timeout: 30s

log:
  level: warn
  format: production

compute:
  ecs:
    cluster_arn: arn:aws:ecs:us-east-1:123456789012:cluster/landlord
    task_definition_arn: arn:aws:ecs:us-east-1:123456789012:task-definition/tenant-app:12

workflow:
  default_provider: step-functions
  step_functions:
    region: us-east-1
    role_arn: arn:aws:iam::123456789012:role/landlord-executor
```

Set database password via environment (never commit to config):

```bash
export DB_PASSWORD=$(aws secretsmanager get-secret-value --secret-id landlord/db-password --query 'SecretString' --output text)
./landlord --config /etc/landlord/production.yaml
```

---

### Scenario 5: Override Single Value for Testing

Use config file for most settings, override one value with CLI flag:

```bash
# Using config.yaml for database, HTTP, logging settings
# Override just the log level for debugging
./landlord --log-level debug
```

---

### Scenario 6: Mixed Sources (Config File + Env + CLI)

**config.yaml**:
```yaml
database:
  host: localhost
  database: landlord_dev

http:
  port: 8080

log:
  level: info
```

**Environment**:
```bash
export DB_USER=dev_user
export DB_PASSWORD=dev_password
export LOG_LEVEL=debug  # This will be overridden
```

**Command**:
```bash
./landlord --log-level warn
```

**Resolved configuration**:
- DB host: localhost (from config file)
- DB user: dev_user (from environment variable)
- DB password: dev_password (from environment variable)
- DB database: landlord_dev (from config file)
- HTTP port: 8080 (from config file)
- Log level: warn (from CLI flag) ← Takes precedence

---

## Migration Path from Kong

### What Changed

The configuration system has migrated from `alecthomas/kong` to `spf13/cobra` + `spf13/viper`. This provides:

- **YAML and JSON file support** - Previously only env vars and CLI flags
- **Better precedence semantics** - Clear CLI > env > file > defaults
- **Standard Go tooling** - Cobra is used by kubectl, Docker, Hugo
- **Backward compatible** - All environment variable names remain unchanged

### What Stays the Same

- **All environment variable names** - `DB_HOST`, `HTTP_PORT`, `LOG_LEVEL`, etc. all still work
- **All CLI flag names** - `--db-host`, `--http-port`, `--log-level`, etc. unchanged
- **Configuration parameters** - All available options remain the same
- **Default values** - Default behavior is identical

### Migration Steps for Users

**No action required for most users!** Your existing configurations continue to work:

```bash
# This continues to work exactly as before
export DB_HOST=prod.db.example.com
export DB_USER=landlord
export DB_PASSWORD=secret
./landlord
```

**Optional: Take advantage of new features**

If you want to use the new YAML/JSON configuration files:

1. Create a `config.yaml` file (see examples above)
2. Store sensitive values in environment variables
3. Reference the new precedence rules in your documentation

### Example Migration for Teams

**Before (Kong)**:
- All configuration via environment variables in CI/CD
- Limited validation and documentation
- Error messages tied to struct tags

**After (Cobra/Viper)**:
- Use YAML config files for most settings
- Environment variables for secrets (DB_PASSWORD, API keys)
- CLI flags for override values in specific environments
- Clear error messages with configuration source information

**Gradual adoption**:
1. Create a `config.yaml` in your repository with non-secret settings
2. Continue using environment variables for secrets
3. Remove hardcoded defaults, rely on config file
4. Use CLI flags for environment-specific overrides

---

## Troubleshooting

### Configuration Not Taking Effect

**Problem**: Settings not applying as expected

**Solution**: Check precedence order. CLI flags override everything, so verify:

```bash
# Check if you have a conflicting CLI flag
./landlord --help | grep database

# Check environment variables
env | grep DB_

# Check config file path
./landlord --config /path/to/config.yaml  # Explicitly specify
```

### Config File Not Found

**Problem**: "Config file not found" error

**Solution**: File is searched in order:

1. `--config` flag path (if provided)
2. `LANDLORD_CONFIG` env var path (if set)
3. `./config.yaml` or `./config.json` (current directory)
4. `/etc/landlord/config.yaml` or `/etc/landlord/config.json`
5. `$XDG_CONFIG_HOME/landlord/config.yaml` (if XDG_CONFIG_HOME is set)

Try explicit path:

```bash
./landlord --config ./config.yaml
```

### Invalid Configuration

**Problem**: Validation error at startup

**Solution**: Check the error message for which field is invalid:

```bash
# Example: Invalid port
# Error: http config: invalid port: 99999 (must be 1-65535)

# Fix in config.yaml:
http:
  port: 8080  # Must be between 1 and 65535
```

### Database Connection Fails

**Problem**: Can't connect to database

**Solution**: Verify database configuration:

```bash
# Check resolved configuration by using --help
./landlord --help

# Test with explicit values
./landlord \
  --db-host localhost \
  --db-port 5432 \
  --db-user test \
  --db-password test \
  --db-database test

# Verify connection string format
# Format: host=X port=Y user=Z password=W dbname=Q sslmode=R
```

### Log Level Not Changing

**Problem**: `--log-level debug` not showing debug logs

**Solution**: Verify precedence:

```bash
# Check all sources for log level
./landlord \
  --log-level debug \
  -h  # Help to see resolved value
```

Remember: CLI flag should override env var and config file.

---

## Reference

For more information:
- See `config.yaml` and `config.json` in project root for full examples
- Run `./landlord --help` for current available flags
- Check `internal/config/` source code for validation rules
