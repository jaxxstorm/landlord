# Restate Provider - Troubleshooting Guide

Common issues and solutions when using Restate provider with Landlord.

## Connection Issues

### Restate Server Not Reachable

**Error**:
```
failed to connect to restate server: Get "http://localhost:8080/health": 
dial tcp 127.0.0.1:8080: connection refused
```

**Causes**:
- Restate not running
- Wrong endpoint URL
- Port mismatch
- Firewall blocking connection

**Solutions**:

1. **Check Restate is running**
   ```bash
   # Local development
   docker-compose ps
   
   # Should show:
   # NAME    STATUS
   # restate Up 2 minutes (healthy)
   ```

2. **Verify endpoint configuration**
   ```bash
   # Check your config.yaml
   grep -A5 "restate:" config.yaml
   ```

3. **Test connection manually**
   ```bash
   curl -v http://localhost:8080/health
   # Should return 200 OK
   ```

4. **Check firewall/network**
   ```bash
   # From Landlord machine
   telnet localhost 8080
   # Should connect (Ctrl+] to exit)
   ```

---

### Timeout Connecting to Restate

**Error**:
```
context deadline exceeded when connecting to restate
```

**Causes**:
- Network latency too high
- Restate overloaded
- Firewall dropping packets

**Solutions**:

1. **Increase timeout in configuration**
   ```yaml
   restate:
     timeout: 60s  # Increase from 30s
   ```

2. **Check network latency**
   ```bash
   ping restate.example.com
   # If latency > 1000ms, investigate network
   ```

3. **Check Restate health**
   ```bash
   # Monitor Restate resources
   docker-compose stats restate
   ```

---

## Authentication Issues

### Connection Refused with No Auth Required

**Error**:
```
failed to connect: connection refused
even with auth_type: "none"
```

**Causes**:
- Restate expects authentication but not configured
- Network issue (same as above)

**Solution**: Verify Restate is running and accessible.

---

### 401 Unauthorized with API Key

**Error**:
```
401 Unauthorized: authentication failed
```

**Causes**:
- API key invalid or expired
- API key not set
- Wrong API key format

**Solutions**:

1. **Verify API key is set**
   ```bash
   echo $LANDLORD_WORKFLOW_RESTATE_API_KEY
   # Should print your API key
   ```

2. **Check format**
   ```bash
   # Key should start with rk_
   $LANDLORD_WORKFLOW_RESTATE_API_KEY | grep "^rk_"
   ```

3. **Regenerate if needed**
   - Contact Restate admin to regenerate key
   - Update environment variable
   - Restart Landlord

4. **Verify in config vs env var**
   ```yaml
   # Avoid putting key directly in YAML
   # Instead use env var reference
   api_key: "${RESTATE_API_KEY}"
   ```

---

### IAM Permission Denied

**Error**:
```
AccessDenied: User is not authorized to perform: lambda:InvokeFunction
```

**Causes**:
- IAM role missing required permissions
- Execution role not attached
- Policy too restrictive

**Solutions**:

1. **Verify role is attached**
   ```bash
   # For Lambda
   aws lambda get-function --function-name landlord \
     --query 'Configuration.Role'
   
   # For Fargate
   aws ecs describe-task-definition \
     --task-definition landlord \
     --query 'taskDefinition.taskRoleArn'
   ```

2. **Check IAM policy**
   ```bash
   aws iam get-role-policy --role-name landlord-role \
     --policy-name restate-policy
   ```

3. **Verify policy allows Restate operations**
   ```json
   {
     "Effect": "Allow",
     "Action": [
       "lambda:InvokeFunction"
     ],
     "Resource": "arn:aws:lambda:*:*:function:restate-*"
   }
   ```

4. **Update policy if needed**
   ```bash
   aws iam put-role-policy --role-name landlord-role \
     --policy-name restate-policy \
     --policy-document file://restate-policy.json
   ```

---

## Configuration Issues

### Invalid Configuration at Startup

**Error**:
```
invalid endpoint URL: endpoint must be http or https
```

**Causes**:
- Malformed endpoint URL
- Missing protocol (http/https)
- Typo in configuration

**Solution**:
```yaml
# Wrong
restate:
  endpoint: "localhost:8080"

# Correct
restate:
  endpoint: "http://localhost:8080"
```

---

### Invalid Execution Mechanism

**Error**:
```
invalid execution_mechanism: "foo" 
(must be local, lambda, fargate, kubernetes, or self-hosted)
```

**Causes**:
- Typo in mechanism name
- Case sensitivity issue

**Solution**:
```yaml
# Valid values (case-sensitive)
execution_mechanism: "local"        # valid
execution_mechanism: "lambda"       # valid
execution_mechanism: "fargate"      # valid
execution_mechanism: "kubernetes"   # valid
execution_mechanism: "self-hosted"  # valid

# Invalid
execution_mechanism: "Local"        # invalid case
execution_mechanism: "aws-lambda"   # invalid name
```

---

### Invalid Auth Type

**Error**:
```
invalid auth_type: "bearer"
(must be none, api_key, or iam)
```

**Causes**:
- Typo in auth type
- Unsupported authentication method

**Solution**:
```yaml
# Valid values
auth_type: "none"       # valid
auth_type: "api_key"    # valid
auth_type: "iam"        # valid

# Invalid
auth_type: "bearer"     # invalid
auth_type: "oauth"      # invalid
```

---

## Workflow Execution Issues

### Service Registration Fails

**Error**:
```
failed to register service: service already exists
```

**Causes**:
- Service already registered with same name
- Network issue during first attempt

**Solutions**:

1. **Check existing services**
   ```bash
   curl http://localhost:8080/services
   ```

2. **Delete and recreate**
   ```bash
   landlord workflow delete my-workflow
   landlord workflow create my-workflow
   ```

3. **Or override service name**
   ```yaml
   restate:
     service_name: "MyWorkflowV2"  # Use different name
   ```

---

### Workflow Execution Never Completes

**Error**:
```
execution stuck in RUNNING state
```

**Causes**:
- Workflow timeout too short
- Restate server overloaded
- Workflow has infinite loop

**Solutions**:

1. **Increase timeout**
   ```yaml
   restate:
     timeout: 60m  # Increase if workflow takes time
   ```

2. **Check Restate resources**
   ```bash
   docker-compose stats restate
   # If CPU/Memory maxed, scale up
   ```

3. **Review workflow definition**
   ```bash
   landlord workflow show my-workflow
   # Look for infinite loops or blocking operations
   ```

---

### Execution Status Always Returns Empty

**Error**:
```
execution found but no status information
```

**Causes**:
- Execution ID incorrect
- Execution already completed
- Restate data corruption

**Solutions**:

1. **Verify execution ID**
   ```bash
   landlord workflow executions my-workflow
   # Get correct execution ID
   ```

2. **Check if execution completed**
   ```bash
   landlord workflow status my-workflow <execution-id>
   # May show completed instead of running
   ```

---

## Performance Issues

### Slow Workflow Registration

**Problem**: `landlord workflow create` takes > 5 seconds

**Causes**:
- Network latency
- Restate overloaded
- Large workflow definition

**Solutions**:

1. **Check network latency**
   ```bash
   time curl http://localhost:8080/health
   # Should be < 100ms
   ```

2. **Monitor Restate**
   ```bash
   docker-compose logs -f restate | grep -i slow
   ```

3. **Reduce workflow size** if very large

---

### High Latency Starting Executions

**Problem**: Executions take > 2 seconds to start

**Causes**:
- Restate latency
- Network issue
- Landlord overloaded

**Solutions**:

1. **Profile Restate**
   ```bash
   curl -w "Time taken: %{time_total}s\n" \
     http://localhost:8080/health
   ```

2. **Check Landlord logs** for where time is spent

3. **Verify database connection** (Landlord database, not Restate)

---

## Log Analysis

### Enable Debug Logging

For more detailed logs:

```bash
export RUST_LOG=debug  # For Restate
export LOG_LEVEL=debug  # For Landlord

landlord server
```

### Common Log Patterns

**Normal startup**:
```
INFO registering Restate workflow provider endpoint=http://localhost:8080
INFO registered workflow providers providers=[mock restate step-functions]
```

**Connection issues**:
```
WARN restate server unreachable at initialization
      (will retry on first use)
```

**Authentication issues**:
```
ERROR authentication failed
      error=401 Unauthorized
```

**Successful operation**:
```
INFO workflow created workflow_id=my-workflow service_name=MyWorkflow
INFO execution started execution_id=123 workflow_id=my-workflow
```

---

## Getting Help

If you encounter an issue not listed here:

1. **Check recent logs**
   ```bash
   docker-compose logs restate | tail -50
   ```

2. **Enable debug mode** and capture logs

3. **Test manually** with curl:
   ```bash
   curl -v http://localhost:8080/health
   curl -v -X POST http://localhost:8080/services \
     -H "Content-Type: application/json" \
     -d '{"name": "TestService"}'
   ```

4. **Check Restate status page**:
   ```bash
   http://localhost:8080  # Web UI
   ```

5. **Consult Restate documentation**: https://restate.dev/

---

## Related Documentation

- [Getting Started](restate-getting-started.md)
- [Configuration Reference](restate-configuration.md)
- [Authentication Guide](restate-authentication.md)
- [Local Development](restate-local-development.md)
