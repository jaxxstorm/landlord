# Restate Provider - Authentication Guide

Guide for configuring authentication with Restate provider in different environments.

## Authentication Types

Restate provider supports three authentication methods:

| Type | Use Case | Mechanism |
|------|----------|-----------|
| `none` | Local development | No authentication |
| `api_key` | Self-hosted, direct HTTP | API key in headers |
| `iam` | AWS (Lambda, Fargate) | AWS IAM roles |

## No Authentication (Development)

### Configuration

```yaml
workflow:
  restate:
    auth_type: "none"
    endpoint: "http://localhost:8080"
```

### When to Use

- **Local development** with Docker Compose
- **Testing** environments
- **Development/staging** on private networks

### Security Note

Only use `none` on:
- Localhost (`127.0.0.1`, `localhost`)
- Private networks (VPN, corporate network)
- Development machines

Never expose Restate without authentication to the internet.

---

## API Key Authentication

### Configuration

```yaml
workflow:
  restate:
    auth_type: "api_key"
    api_key: "rk_abcd1234..."
    endpoint: "https://restate.example.com"
```

### Environment Variable

```bash
LANDLORD_WORKFLOW_RESTATE_API_KEY=rk_abcd1234...
```

### Obtaining an API Key

#### Self-Hosted Restate

1. Start Restate admin console (or API)
2. Generate API key:
   ```bash
   curl -X POST https://restate.example.com/api/keys \
     -H "Content-Type: application/json" \
     -d '{"name": "landlord-key"}'
   ```
3. Copy the returned key

#### Restate Cloud (if available)

Access through console dashboard and generate key.

### Key Format

API keys typically look like:
```
rk_prod_1234567890abcdef
```

### When to Use

- **Self-hosted** Restate on public internet
- **Kubernetes** deployments (via secrets)
- **Custom infrastructure** with direct HTTP access

### Best Practices

1. **Store securely** - Use environment variables or secrets management
2. **Rotate regularly** - Plan for key rotation every 90 days
3. **Use scoped keys** - If Restate supports it, limit permissions
4. **Monitor usage** - Track which applications use which keys
5. **Alert on misuse** - Monitor for unusual API patterns

### Examples

#### Kubernetes Secret

```bash
# Create secret
kubectl create secret generic restate-api-key \
  --from-literal=api-key=rk_prod_1234567890abcdef

# Reference in deployment
env:
  - name: LANDLORD_WORKFLOW_RESTATE_API_KEY
    valueFrom:
      secretKeyRef:
        name: restate-api-key
        key: api-key
```

#### GitHub Actions Secret

```yaml
env:
  LANDLORD_WORKFLOW_RESTATE_API_KEY: ${{ secrets.RESTATE_API_KEY }}
```

#### Docker Compose

```yaml
services:
  landlord:
    environment:
      LANDLORD_WORKFLOW_RESTATE_API_KEY: ${RESTATE_API_KEY}
```

Then run:
```bash
RESTATE_API_KEY=rk_prod_... docker-compose up
```

---

## AWS IAM Authentication

### Configuration

```yaml
workflow:
  restate:
    auth_type: "iam"
    endpoint: "https://restate.prod.internal"
    execution_mechanism: "lambda"  # or "fargate"
```

### How It Works

When `auth_type: "iam"`:
1. Provider uses AWS SDK for authentication
2. SDK automatically fetches temporary credentials from:
   - Lambda execution role (for Lambda)
   - ECS task role (for Fargate)
   - EC2 instance role (for EC2)
   - Environment variables (for local testing)
4. Credentials are automatically rotated

### No Configuration Needed

Unlike API key, you don't configure credentials explicitly. The provider uses:

1. **Lambda**: Execution role attached to function
2. **Fargate**: Task role attached to task definition
3. **EC2**: Instance role attached to instance
4. **Local**: `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` env vars

### IAM Policy for Lambda

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "lambda:InvokeFunction",
        "lambda:InvokeAsync"
      ],
      "Resource": "arn:aws:lambda:*:*:function:restate-*"
    }
  ]
}
```

### IAM Policy for Fargate

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecs:DescribeTaskDefinition",
        "ecs:DescribeServices",
        "ecs:UpdateService"
      ],
      "Resource": "*"
    }
  ]
}
```

### When to Use

- **AWS Lambda** deployment
- **AWS ECS Fargate** deployment
- **EC2 instances** with instance roles
- **Strong security** requirement (temporary credentials)

### Benefits

- No API keys to manage
- Automatic credential rotation
- Fine-grained IAM policies
- Audit trail in CloudTrail
- Works seamlessly in AWS services

### Local Testing with IAM

For local development with IAM:

```bash
export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
export AWS_DEFAULT_REGION=us-west-2

landlord server \
  --workflow.restate.auth_type=iam \
  --workflow.restate.endpoint=http://localhost:8080 \
  --workflow.restate.execution_mechanism=lambda
```

---

## Authentication Troubleshooting

### Connection Refused

**Problem**: `Connection refused` error when connecting to Restate

**Causes**:
- Restate server not running
- Endpoint URL incorrect
- Network connectivity issue
- Firewall blocking connection

**Solution**:
1. Verify Restate is running: `curl -v https://restate.example.com/health`
2. Check endpoint configuration
3. Verify network connectivity: `ping restate.example.com`
4. Check firewall rules

### Authentication Failed

**Problem**: `401 Unauthorized` error

**Causes**:
- API key incorrect or expired
- IAM role doesn't have permissions
- Authentication type mismatch

**Solution for API Key**:
1. Verify API key: `echo $LANDLORD_WORKFLOW_RESTATE_API_KEY`
2. Check key format (should start with `rk_`)
3. Verify key is still valid in Restate
4. Regenerate if necessary

**Solution for IAM**:
1. Verify IAM role attached: `aws sts get-caller-identity`
2. Check IAM policy permissions
3. Verify Restate can read AWS credentials
4. Check CloudTrail for API calls

### Mixed Authentication Error

**Problem**: `auth_type: "api_key"` but no `api_key` configured

**Causes**:
- Missing API key environment variable
- Typo in variable name
- Configuration not reloaded

**Solution**:
1. Verify environment variable is set: `echo $LANDLORD_WORKFLOW_RESTATE_API_KEY`
2. Check YAML configuration: `auth_type` and `api_key` fields
3. Restart Landlord after configuration change

---

## Security Best Practices

### 1. Store Credentials Securely

**Never**:
- Commit API keys to git
- Log credentials
- Put keys in error messages

**Instead**:
- Use environment variables
- Use secrets management (Kubernetes, AWS Secrets Manager)
- Use IAM roles when possible

### 2. Rotate Credentials Regularly

- API keys: Every 90 days
- IAM credentials: Automatic (managed by AWS)
- Passwords: Never used directly, use IAM

### 3. Principle of Least Privilege

IAM policies should grant minimum permissions:
- Only allow Restate operations needed
- Limit resources to specific services
- Avoid wildcard permissions

### 4. Monitor and Alert

- Monitor API usage patterns
- Alert on authentication failures
- Review access logs regularly
- Setup rate limiting if available

### 5. Revoke Compromised Credentials

If API key or credentials are compromised:
1. Immediately revoke in Restate
2. Generate new key
3. Rotate in all environments
4. Review access logs for misuse

---

## Related Documentation

- [Configuration Reference](restate-configuration.md)
- [Getting Started](restate-getting-started.md)
- [Production Deployment](restate-production-deployment.md)
- [Troubleshooting](restate-troubleshooting.md)
