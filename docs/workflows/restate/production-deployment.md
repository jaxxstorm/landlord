# Restate Provider - Production Deployment

Guide for deploying Restate in production with different execution mechanisms.

## Deployment Options

Choose based on your infrastructure:

| Option | Best For | Infrastructure | Complexity |
|--------|----------|-----------------|-----------|
| **Lambda** | Serverless, pay-per-use | AWS | Medium |
| **Fargate** | Container orchestration, AWS | AWS ECS | Medium |
| **Kubernetes** | Cloud-native, multi-cloud | Kubernetes | High |
| **Self-Hosted** | Full control, on-premises | Own servers | High |

## AWS Lambda Deployment

### Prerequisites

- AWS account with Lambda permissions
- IAM role for Lambda execution
- Restate Lambda extension deployed

### Configuration

```yaml
workflow:
  default_provider: "restate"
  restate:
    endpoint: "https://restate.prod.example.com"
    execution_mechanism: "lambda"
    auth_type: "iam"  # Uses Lambda execution role
    timeout: 1h
    retry_attempts: 3
```

### Environment Variables

```bash
LANDLORD_WORKFLOW_RESTATE_ENDPOINT=https://restate.prod.example.com
LANDLORD_WORKFLOW_RESTATE_EXECUTION_MECHANISM=lambda
LANDLORD_WORKFLOW_RESTATE_AUTH_TYPE=iam
```

### Setup Steps

1. **Deploy Restate to AWS Lambda**
   - Follow [Restate AWS Lambda Guide](https://restate.dev/deploy/lambda/)
   - Deploy Restate Lambda function
   - Configure API Gateway or ALB for access

2. **Configure IAM Role**
   ```json
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Action": [
           "lambda:InvokeFunction"
         ],
         "Resource": "arn:aws:lambda:region:account:function:restate-*"
       }
     ]
   }
   ```

3. **Deploy Landlord**
   - Set Restate endpoint in configuration
   - Assign IAM role to Landlord Lambda function (or EC2/ECS task)
   - Deploy

### Benefits

- Fully managed service
- Auto-scaling
- Pay-per-use pricing
- Integrates with AWS ecosystem

### Considerations

- Cold start latency
- Lambda timeout limits (15 minutes max)
- Additional Lambda invocation costs

---

## AWS ECS Fargate Deployment

### Prerequisites

- AWS account with ECS permissions
- VPC and security groups configured
- ECR repository for container images

### Configuration

```yaml
workflow:
  default_provider: "restate"
  restate:
    endpoint: "http://restate-nlb.prod.internal:8080"
    execution_mechanism: "fargate"
    auth_type: "iam"  # Uses Fargate task role
    timeout: 1h
    retry_attempts: 3
```

### Setup Steps

1. **Deploy Restate to Fargate**
   - Create ECS task definition for Restate
   - Configure networking (security groups, load balancer)
   - Deploy service with auto-scaling

2. **Configure Task Role**
   ```json
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Action": "ecs:DescribeTaskDefinition",
         "Resource": "*"
       }
     ]
   }
   ```

3. **Deploy Landlord on Fargate**
   - Create Landlord ECS task
   - Assign task role with Restate permissions
   - Configure service discovery to reach Restate

### Benefits

- Container-based (easier to manage)
- Better cost than Lambda for long-running processes
- Full control over resources
- Integrates with AWS monitoring (CloudWatch)

### Considerations

- Requires VPC and networking setup
- More resources to manage than Lambda
- Need to configure auto-scaling

---

## Kubernetes Deployment

### Prerequisites

- Kubernetes cluster (1.24+)
- kubectl configured
- Helm (optional but recommended)

### Configuration

```yaml
workflow:
  default_provider: "restate"
  restate:
    endpoint: "http://restate-headless.default.svc.cluster.local:8080"
    execution_mechanism: "kubernetes"
    auth_type: "api_key"
    api_key: "${RESTATE_API_KEY}"  # From secret
    timeout: 1h
    retry_attempts: 3
```

### Setup Steps

1. **Deploy Restate using Helm**

   ```bash
   helm repo add restate https://restate-helm.dev
   helm install restate restate/restate \
     --namespace restate --create-namespace \
     --values restate-values.yaml
   ```

   Example `restate-values.yaml`:
   ```yaml
   replication:
     enabled: true
     replicas: 3
   
   persistence:
     enabled: true
     size: 50Gi
   
   service:
     type: ClusterIP
     port: 8080
   ```

2. **Create API Key Secret**

   ```bash
   kubectl create secret generic restate-api-key \
     --from-literal=api-key=$(openssl rand -hex 32) \
     -n restate
   ```

3. **Deploy Landlord**

   ```bash
   kubectl create configmap landlord-config \
     --from-file=config.yaml \
     -n default
   
   kubectl apply -f landlord-deployment.yaml
   ```

   Example `landlord-deployment.yaml`:
   ```yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: landlord
   spec:
     replicas: 2
     template:
       spec:
         containers:
         - name: landlord
           image: landlord:latest
           env:
           - name: LANDLORD_WORKFLOW_RESTATE_ENDPOINT
             value: "http://restate-headless.default.svc.cluster.local:8080"
           - name: LANDLORD_WORKFLOW_RESTATE_EXECUTION_MECHANISM
             value: "kubernetes"
           - name: LANDLORD_WORKFLOW_RESTATE_AUTH_TYPE
             value: "api_key"
           - name: LANDLORD_WORKFLOW_RESTATE_API_KEY
             valueFrom:
               secretKeyRef:
                 name: restate-api-key
                 key: api-key
   ```

### Benefits

- Cloud-native, multi-cloud support
- Excellent observability and monitoring
- Service mesh integration (Istio)
- GitOps-friendly (declarative config)
- Multi-region and multi-az deployment

### Considerations

- Steeper learning curve
- Requires Kubernetes expertise
- More operational overhead

---

## Self-Hosted Deployment

### Prerequisites

- Dedicated servers or VMs
- Network connectivity between servers
- HTTP/HTTPS endpoint access

### Configuration

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

### Setup Steps

1. **Download Restate Binary**
   ```bash
   wget https://releases.restate.dev/restate-latest-linux-x86_64
   chmod +x restate
   ```

2. **Configure Restate**
   ```bash
   ./restate config init
   # Edit restate.toml with your settings
   ```

3. **Start Restate**
   ```bash
   ./restate server --config restate.toml
   ```

4. **Setup Reverse Proxy (HTTPS)**

   Example with Nginx:
   ```nginx
   server {
     listen 443 ssl;
     server_name restate.mycompany.com;
     
     ssl_certificate /etc/ssl/certs/restate.crt;
     ssl_certificate_key /etc/ssl/private/restate.key;
     
     location / {
       proxy_pass http://localhost:8080;
       proxy_set_header Host $host;
       proxy_set_header X-Forwarded-For $remote_addr;
     }
   }
   ```

5. **Deploy Landlord**
   ```bash
   ./landlord server \
     --workflow.restate.endpoint=https://restate.mycompany.com \
     --workflow.restate.execution_mechanism=self-hosted \
     --workflow.restate.auth_type=api_key
   ```

### Benefits

- Complete control over infrastructure
- No vendor lock-in
- Can run anywhere (on-premises, private cloud)
- Custom optimization options

### Considerations

- Requires operational expertise
- Must manage updates, security patches
- Must setup monitoring and backups
- Scaling requires manual configuration

---

## Production Checklist

Before deploying to production:

- [ ] Configure HTTPS with valid certificates
- [ ] Setup authentication (API keys or IAM)
- [ ] Configure logging and monitoring
- [ ] Setup backup and recovery procedures
- [ ] Configure health checks and auto-recovery
- [ ] Setup load balancing (if multiple Restate instances)
- [ ] Configure resource limits (memory, CPU)
- [ ] Test failover and recovery scenarios
- [ ] Document procedures and runbooks
- [ ] Setup alerting for failures
- [ ] Perform load testing
- [ ] Plan for scaling strategy

---

## Monitoring and Observability

### Key Metrics to Monitor

- Workflow execution success rate
- Execution latency (p50, p95, p99)
- Error rates by type
- Restate server health
- Disk usage (for state storage)
- Network latency to Restate

### Logging

Configure structured logging to capture:
- Workflow creation/execution events
- Error details with context
- Performance metrics
- Connection issues

### Health Checks

Restate provides health endpoint:
```bash
curl https://restate.mycompany.com/health
```

Include in load balancer health checks and monitoring systems.

---

## Disaster Recovery

### Backup Strategy

1. **Regular snapshots** of Restate state storage
2. **Offsite backups** (S3, separate datacenter)
3. **Backup frequency** based on RPO requirements
4. **Restore testing** in staging environment

### Recovery Procedures

1. Document step-by-step recovery procedures
2. Test recovery regularly (at least quarterly)
3. Maintain runbooks for different failure scenarios
4. Train team on recovery procedures

---

## Performance Tuning

### Scaling Restate

- **Vertical**: Increase CPU/memory per instance
- **Horizontal**: Add more Restate instances with load balancing

### Scaling Landlord

- Deploy multiple Landlord instances behind load balancer
- Use connection pooling to Restate
- Monitor and adjust based on metrics

---

## Related Documentation

- [Getting Started](restate-getting-started.md)
- [Configuration Reference](restate-configuration.md)
- [Authentication](restate-authentication.md)
- [Troubleshooting](restate-troubleshooting.md)
