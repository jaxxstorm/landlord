# Migration Guide: Kong → Cobra/Viper

This guide helps teams migrate from the previous kong-based configuration system to the new cobra/viper system.

## Overview

**Good news**: This is a **non-breaking change**. Existing deployments continue to work without modification.

The migration from `alecthomas/kong` to `spf13/cobra` + `spf13/viper` introduces new configuration file support (YAML/JSON) while maintaining 100% backward compatibility with environment variables and CLI flags.

## What's New

### 1. Configuration Files (YAML/JSON)

You can now use configuration files instead of only environment variables and CLI flags.

**Before**:
```bash
# All configuration via environment
export DB_HOST=prod.db.example.com
export DB_PORT=5432
export DB_USER=landlord
export DB_PASSWORD=secret
export DB_DATABASE=landlord_db
export LOG_LEVEL=info
export HTTP_PORT=8080
./landlord
```

**After** (same, but could also use):
```bash
# Or with a config file
cat > config.yaml << 'EOF'
database:
  host: prod.db.example.com
  port: 5432
  user: landlord
  database: landlord_db
http:
  port: 8080
log:
  level: info
EOF

export DB_PASSWORD=secret  # Still use env for secrets
./landlord
```

### 2. Better Precedence Semantics

Configuration precedence is now explicit and documented.

**Before**:
- Kong struct tags handled precedence, but semantics were implicit
- Environment variables and CLI flags worked, but order was unclear

**After**:
- **Clear precedence order**: CLI flags > Environment variables > Config file > Defaults
- **Documented**: See [Configuration Guide](./configuration.md) for details
- **Predictable**: Always know which source wins

### 3. Standard Go Tooling

Cobra is the de facto standard CLI framework in Go.

**Benefits**:
- Used by kubectl, Docker, Hugo, and thousands of other projects
- Better ecosystem integration
- Community support and documentation
- Consistent with modern Go practices

## Breaking Changes

**ZERO breaking changes**

All environment variable names remain identical:
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_DATABASE`, etc.
- `HTTP_HOST`, `HTTP_PORT`, `HTTP_SHUTDOWN_TIMEOUT`, etc.
- `LOG_LEVEL`, `LOG_FORMAT`
- `COMPUTE_DEFAULT_PROVIDER`, `WORKFLOW_DEFAULT_PROVIDER`, etc.

All CLI flags remain identical:
- `--db-host`, `--db-port`, `--db-user`, etc.
- `--http-host`, `--http-port`, `--log-level`, etc.
- `--config` (for explicit config file path)

## Migration Timeline

### Phase 1: No Action Required (Immediate)

Your current deployment works as-is. No changes needed.

```bash
# This continues to work exactly as before
export DB_HOST=prod.db.example.com
export DB_USER=landlord
export DB_PASSWORD=secret
./landlord
```

### Phase 2: Gradual Adoption (Optional)

Adopt new features at your pace:

1. **Try config files in development**
   - Create `config.yaml` with non-sensitive configuration
   - Use `export DB_PASSWORD=...` for secrets
   - Test precedence behavior

2. **Adopt in staging**
   - Move non-secret settings to YAML
   - Verify behavior matches previous system
   - Document new workflow for team

3. **Deploy to production**
   - Create production config file
   - Keep secrets in environment variables
   - Monitor for any changes in behavior

### Phase 3: Deprecation (Future, If Any)

Kong support was removed in this version. If you're still using features that required kong, please report an issue.

## Migration Patterns

### Pattern 1: Keep Using Environment Variables (No Changes)

**Recommended for**: Docker/Kubernetes with secrets management

```bash
# Your current approach continues to work
docker run \
  -e DB_HOST=postgres.internal \
  -e DB_USER=landlord \
  -e DB_PASSWORD=$DB_PASSWORD \
  -e LOG_LEVEL=info \
  landlord:latest
```

**No action needed**. Your setup is complete.

---

### Pattern 2: Adopt Configuration Files

**Recommended for**: File-based configuration with Kubernetes ConfigMaps or local development

#### Step 1: Create Configuration File

Create `config.yaml`:

```yaml
database:
  host: prod-db.example.com
  port: 5432
  user: landlord
  database: landlord_db
  ssl_mode: require
  max_connections: 100

http:
  host: 0.0.0.0
  port: 8080

log:
  level: info
  format: production

compute:
  default_provider: kubernetes

workflow:
  default_provider: step-functions
  step_functions:
    region: us-east-1
    role_arn: arn:aws:iam::123456789012:role/landlord-executor
```

#### Step 2: Use Secrets from Environment

```bash
# Secrets still come from environment/secrets manager
export DB_PASSWORD=$(aws secretsmanager get-secret-value --secret-id landlord/db-password --query 'SecretString' --output text)
export WORKFLOW_SFN_ROLE_ARN=$(aws secretsmanager get-secret-value --secret-id landlord/sfn-role --query 'SecretString' --output text)

./landlord
```

#### Step 3: Update CI/CD

**Before**:
```yaml
# Old: All settings as environment variables
env:
  DB_HOST: prod-db.example.com
  DB_PORT: "5432"
  DB_USER: landlord
  DB_PASSWORD: ${{ secrets.DB_PASSWORD }}
  LOG_LEVEL: info
```

**After**:
```yaml
# New: Config file with environment-specific secrets
env:
  DB_PASSWORD: ${{ secrets.DB_PASSWORD }}
  LANDLORD_CONFIG: ./config/production.yaml
```

---

### Pattern 3: Mixed Configuration

**Recommended for**: Gradual migration or advanced scenarios

Use config file for defaults, override specific values with environment variables or CLI flags.

#### Scenario: Development Environment

**config.yaml**:
```yaml
database:
  host: localhost
  port: 5432
  database: landlord_dev

http:
  port: 8080

log:
  level: info
```

**Environment Override**:
```bash
export LOG_LEVEL=debug  # Override to debug
export DB_USER=dev_user
export DB_PASSWORD=dev_password

./landlord
```

**Result**:
- DB host: localhost (from config)
- DB user: dev_user (from env, overrides config default)
- DB password: dev_password (from env)
- DB database: landlord_dev (from config)
- Log level: debug (from env, overrides config)

#### Scenario: Production Testing

**config.yaml**:
```yaml
database:
  host: prod-db.example.com
  port: 5432
  # ... rest of production config
```

**CLI Override for Testing**:
```bash
export DB_PASSWORD=...
export DB_USER=test_user  # Temporary override

# Or with CLI flag
./landlord --log-level debug

# Result: All production settings except log level is debug
```

---

## Kubernetes Migration Example

### Before (Kong)

**Deployment**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: landlord
spec:
  template:
    spec:
      containers:
      - name: landlord
        image: landlord:v0.1.0
        env:
        - name: DB_HOST
          valueFrom:
            configMapKeyRef:
              name: landlord-config
              key: db_host
        - name: DB_PORT
          valueFrom:
            configMapKeyRef:
              name: landlord-config
              key: db_port
        - name: DB_USER
          valueFrom:
            secretKeyRef:
              name: landlord-secrets
              key: db_user
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: landlord-secrets
              key: db_password
        # ... 10+ more environment variables
```

### After (Cobra/Viper)

**ConfigMap** (`config.yaml` for settings):
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
      default_provider: kubernetes
    workflow:
      default_provider: step-functions
```

**Secret** (only for sensitive values):
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: landlord-secrets
type: Opaque
stringData:
  db_password: "..."
  sfn_role_arn: "arn:aws:iam::..."
```

**Deployment**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: landlord
spec:
  template:
    spec:
      containers:
      - name: landlord
        image: landlord:v0.2.0
        env:
        # Only secrets in environment
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: landlord-secrets
              key: db_password
        - name: WORKFLOW_SFN_ROLE_ARN
          valueFrom:
            secretKeyRef:
              name: landlord-secrets
              key: sfn_role_arn
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

**Benefits**:
- Configuration file separates concerns (settings vs. secrets)
- Simpler deployment manifest (fewer env vars)
- Better version control (config file in ConfigMap)
- Easier to reason about what's where

---

## Docker Compose Migration Example

### Before (Kong)

```yaml
version: '3.8'
services:
  landlord:
    image: landlord:v0.1.0
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: landlord
      DB_PASSWORD: secret_password
      DB_DATABASE: landlord_db
      DB_SSLMODE: disable
      DB_MAX_CONNECTIONS: 25
      DB_MIN_CONNECTIONS: 5
      DB_CONNECT_TIMEOUT: 10s
      HTTP_HOST: 0.0.0.0
      HTTP_PORT: 8080
      HTTP_READ_TIMEOUT: 10s
      HTTP_WRITE_TIMEOUT: 10s
      HTTP_IDLE_TIMEOUT: 120s
      HTTP_SHUTDOWN_TIMEOUT: 30s
      LOG_LEVEL: info
      LOG_FORMAT: development
    ports:
      - "8080:8080"
    depends_on:
      - postgres

  postgres:
    image: postgres:15
    environment:
      POSTGRES_USER: landlord
      POSTGRES_PASSWORD: secret_password
      POSTGRES_DB: landlord_db
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

### After (Cobra/Viper)

```yaml
version: '3.8'
services:
  landlord:
    image: landlord:v0.2.0
    environment:
      DB_PASSWORD: secret_password
      LANDLORD_CONFIG: /etc/landlord/config.yaml
    volumes:
      - ./config.yaml:/etc/landlord/config.yaml:ro
    ports:
      - "8080:8080"
    depends_on:
      - postgres

  postgres:
    image: postgres:15
    environment:
      POSTGRES_USER: landlord
      POSTGRES_PASSWORD: secret_password
      POSTGRES_DB: landlord_db
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

**config.yaml**:
```yaml
database:
  host: postgres
  port: 5432
  user: landlord
  database: landlord_db
  ssl_mode: disable
  max_connections: 25
  min_connections: 5
  connect_timeout: 10s

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
```

**Benefits**:
- Configuration in version-controlled file
- Simpler docker-compose.yaml
- Easier to maintain multiple environments

---

## GitHub Actions Migration Example

### Before (Kong)

```yaml
name: Deploy Landlord
on: [push]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Deploy
        run: |
          docker build -t landlord:${{ github.sha }} .
          docker run \
            -e DB_HOST=${{ secrets.PROD_DB_HOST }} \
            -e DB_PORT=${{ secrets.PROD_DB_PORT }} \
            -e DB_USER=${{ secrets.PROD_DB_USER }} \
            -e DB_PASSWORD=${{ secrets.PROD_DB_PASSWORD }} \
            -e DB_DATABASE=${{ secrets.PROD_DB_NAME }} \
            -e LOG_LEVEL=info \
            -e HTTP_PORT=8080 \
            landlord:${{ github.sha }}
```

### After (Cobra/Viper)

```yaml
name: Deploy Landlord
on: [push]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Create config file
        run: |
          cat > config.yaml << 'EOF'
          database:
            host: ${{ secrets.PROD_DB_HOST }}
            port: ${{ secrets.PROD_DB_PORT }}
            user: ${{ secrets.PROD_DB_USER }}
            database: ${{ secrets.PROD_DB_NAME }}
            ssl_mode: require
          http:
            port: 8080
          log:
            level: info
          EOF
      
      - name: Deploy
        env:
          DB_PASSWORD: ${{ secrets.PROD_DB_PASSWORD }}
          LANDLORD_CONFIG: config.yaml
        run: |
          docker build -t landlord:${{ github.sha }} .
          docker run \
            -e DB_PASSWORD \
            -e LANDLORD_CONFIG \
            -v $(pwd)/config.yaml:/etc/landlord/config.yaml:ro \
            landlord:${{ github.sha }}
```

**Benefits**:
- Secrets only passed for sensitive values
- Configuration versioned in repository
- Easier to review and audit

---

## Rollback Plan (If Needed)

If you encounter issues, rollback is simple since environment variables still work:

```bash
# Simply revert to environment-variable only approach
export DB_HOST=...
export DB_USER=...
export DB_PASSWORD=...
export LOG_LEVEL=...

# Run without --config flag
./landlord

# Previous version (with kong) can still be deployed if needed
```

All changes are backward compatible, so you can always revert.

---

## Support and Questions

- **Configuration Guide**: See [configuration.md](./configuration.md) for detailed reference
- **Example Configs**: Check `config.yaml` and `config.json` in project root
- **CLI Help**: Run `./landlord --help` for available flags
- **Report Issues**: Open an issue on the project repository

---

## Summary

| Aspect | Before (Kong) | After (Cobra/Viper) |
|--------|---------------|-------------------|
| **Config Files** | Not supported | YAML/JSON files |
| **Environment Variables** | Yes | Yes (unchanged) |
| **CLI Flags** | Yes | Yes (unchanged) |
| **Precedence** | Implicit | Documented, explicit |
| **Ecosystem** | Kong-specific | Standard Go (kubectl-like) |
| **Secrets Management** | Environment only | Env + file support |
| **Backward Compatible** | N/A | 100% compatible |
| **Breaking Changes** | N/A | Zero |

**Action Required**: **NONE** - Your deployment works as-is. Adopt new features at your pace.
