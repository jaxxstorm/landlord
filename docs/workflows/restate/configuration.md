# Restate Provider - Configuration Reference

Complete reference for all Restate provider configuration options.

## Configuration Structure

```yaml
workflow:
  default_provider: "restate"        # Set restate as the default provider
  restate:
    endpoint: "http://localhost:8080"  # Restate server endpoint
    execution_mechanism: "local"       # Where to execute workflows
    service_name: ""                   # (optional) Override service name
    auth_type: "none"                  # Authentication method
    api_key: ""                        # API key for api_key auth
    timeout: 30m                       # Workflow execution timeout
    retry_attempts: 3                  # Number of retries on failure
```

## Configuration Options

### `endpoint` (required)

The URL where Restate server is accessible.

**Format**: `http://host:port` or `https://host:port`

**Examples**:
- Local development: `http://localhost:8080`
- Production (AWS): `https://restate.example.com`
- Production (Kubernetes): `http://restate-cluster.default.svc.cluster.local:8080`

**Default**: `http://localhost:8080`

---

### `execution_mechanism` (required)

Where and how Restate should execute your services.

**Allowed Values**:
- `local` - Use Restate's local compute (development)
- `lambda` - Invoke AWS Lambda functions
- `fargate` - Invoke AWS ECS Fargate tasks
- `kubernetes` - Invoke Kubernetes services
- `self-hosted` - Direct HTTP to Restate instances

**Default**: `local`

**Guidance**:
- Development: Use `local`
- AWS Lambda: Use `lambda` with `auth_type: "iam"`
- AWS Fargate: Use `fargate` with `auth_type: "iam"`
- Kubernetes: Use `kubernetes` with appropriate service discovery
- Self-managed infrastructure: Use `self-hosted` with `auth_type: "api_key"`

---

### `service_name` (optional)

Override the service name used for service registration.

**Default**: Derived from workflow ID (e.g., `my-workflow` → `MyWorkflow`)

**Use Cases**:
- Versioned services: `MyWorkflowV2`
- Custom naming convention: `service_namespace_my_workflow`
- Multi-tenant scenarios: `tenant_1_workflow`

**Example**:
```yaml
restate:
  service_name: "CustomServiceName"
```

---

### `auth_type` (optional)

Authentication method for Restate server connection.

**Allowed Values**:
- `none` - No authentication (local development)
- `api_key` - API key authentication (self-hosted, production)
- `iam` - AWS IAM role authentication (Lambda, Fargate)

**Default**: `none`

**Authentication Rules**:
| Execution Mechanism | Recommended Auth | Reason |
|---|---|---|
| local | none | Local development, no auth needed |
| self-hosted | api_key | API key secures HTTP endpoint |
| lambda | iam | Lambda execution role provides credentials |
| fargate | iam | Fargate task role provides credentials |
| kubernetes | api_key | API key through secret injection |

---

### `api_key` (required if `auth_type: "api_key"`)

API key for authentication with Restate server.

**Format**: String

**How to Obtain**:
- Self-hosted: Generate in Restate admin console
- Production: Set via environment variable `LANDLORD_WORKFLOW_RESTATE_API_KEY`

**Example**:
```yaml
restate:
  auth_type: "api_key"
  api_key: "${RESTATE_API_KEY}"  # Reference env var
```

**Security**: Never commit API keys to version control. Use environment variables or secrets management.

---

### `timeout` (optional)

Maximum duration a workflow execution can run.

**Format**: Go duration string (e.g., `30m`, `1h`, `5m30s`)

**Default**: `30m` (30 minutes)

**Range**: Must be positive (e.g., `1s` to `24h`)

**Example**:
```yaml
restate:
  timeout: 1h  # 1 hour timeout
```

---

### `retry_attempts` (optional)

Number of times to retry a failed workflow execution.

**Format**: Non-negative integer

**Default**: `3`

**Example**:
```yaml
restate:
  retry_attempts: 5  # Retry up to 5 times
```

---

## Configuration via Environment Variables

All configuration options can be set via environment variables with the `LANDLORD_` prefix:

```bash
# Endpoint
LANDLORD_WORKFLOW_RESTATE_ENDPOINT=http://localhost:8080

# Execution mechanism
LANDLORD_WORKFLOW_RESTATE_EXECUTION_MECHANISM=lambda

# Authentication
LANDLORD_WORKFLOW_RESTATE_AUTH_TYPE=iam
LANDLORD_WORKFLOW_RESTATE_API_KEY=your-api-key

# Service name
LANDLORD_WORKFLOW_RESTATE_SERVICE_NAME=CustomService

# Timeout
LANDLORD_WORKFLOW_RESTATE_TIMEOUT=1h

# Retry attempts
LANDLORD_WORKFLOW_RESTATE_RETRY_ATTEMPTS=5
```

---

## Configuration via CLI Flags

Override configuration with command-line flags:

```bash
landlord server \
  --workflow.default_provider=restate \
  --workflow.restate.endpoint=http://localhost:8080 \
  --workflow.restate.execution_mechanism=local \
  --workflow.restate.auth_type=none
```

---

## Configuration Precedence

Configuration is loaded in this order (highest to lowest precedence):

1. **CLI Flags** (highest priority)
2. **Environment Variables**
3. **Configuration File** (config.yaml)
4. **Default Values** (lowest priority)

Example: If `config.yaml` sets `endpoint: "http://localhost:8080"` but environment variable `LANDLORD_WORKFLOW_RESTATE_ENDPOINT=https://prod.example.com` is set, the environment variable takes precedence.

---

## Configuration Examples

### Local Development (Recommended for Development)

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

### AWS Lambda Production

```yaml
workflow:
  default_provider: "restate"
  restate:
    endpoint: "https://restate.prod.internal"
    execution_mechanism: "lambda"
    auth_type: "iam"  # Uses AWS Lambda execution role
    timeout: 1h
    retry_attempts: 3
```

### AWS Fargate Production

```yaml
workflow:
  default_provider: "restate"
  restate:
    endpoint: "https://restate.prod.internal"
    execution_mechanism: "fargate"
    auth_type: "iam"  # Uses Fargate task role
    timeout: 1h
    retry_attempts: 5
```

### Kubernetes Production

```yaml
workflow:
  default_provider: "restate"
  restate:
    endpoint: "http://restate-cluster.default.svc.cluster.local:8080"
    execution_mechanism: "kubernetes"
    auth_type: "api_key"
    api_key: "${RESTATE_API_KEY}"  # From Kubernetes secret
    timeout: 1h
    retry_attempts: 3
```

### Self-Hosted Production

```yaml
workflow:
  default_provider: "restate"
  restate:
    endpoint: "https://restate.mycompany.com"
    execution_mechanism: "self-hosted"
    auth_type: "api_key"
    api_key: "${RESTATE_API_KEY}"
    timeout: 30m
    retry_attempts: 3
```

---

## Configuration Validation

Landlord validates Restate configuration at startup. If any configuration is invalid, the application will fail with a descriptive error message:

```
Error: restate config: invalid auth_type: "invalid" (must be none, api_key, or iam)
```

This helps catch configuration issues early before runtime.

---

## Related Documentation

- [Getting Started](restate-getting-started.md)
- [Local Development](restate-local-development.md)
- [Production Deployment](restate-production-deployment.md)
- [Authentication](restate-authentication.md)
