# Step Functions Lifecycle Provider

The `step-functions` provider runs each tenant lifecycle operation in one
pre-provisioned AWS Step Functions **Standard** state machine. The Landlord
control plane starts and observes executions. The state machine invokes the
Landlord Lambda synchronously, and the Lambda uses the request payload to run
the configured compute provider without looking up tenant state.

Use this provider when Landlord, its PostgreSQL tracking database, and the
target compute provider are reachable from AWS. Use the `mock` provider for
local development without AWS infrastructure.

## Architecture

1. The reconciler creates a request containing the tenant ID, UUID, lifecycle
   operation, desired configuration, compute provider, and configuration hash.
2. The control plane starts the configured Standard state machine with a
   deterministic name. Repeated requests for the same revision return the
   existing execution instead of creating duplicate work.
3. The state machine passes the original request and its execution ARN to the
   Lambda.
4. The Lambda records the ARN in compute tracking, checks for cancellation
   before mutations, and dispatches the request to the configured compute
   provider.
5. The reconciler polls the execution ARN and transitions the tenant when the
   execution succeeds, fails, times out, or is stopped.

## Prerequisites

- An AWS account and region supporting Step Functions Standard workflows.
- A PostgreSQL database reachable by both the Landlord control plane and the
  Lambda. Apply Landlord migrations before deployment.
- At least one configured compute provider. For production AWS deployments,
  this is normally ECS.
- An artifact S3 bucket for the Lambda ZIP and ASL definition.
- Three separate IAM identities: a Landlord caller role, a state-machine role,
  and a Lambda compute role.
- AWS credentials available through the standard AWS SDK credential chain for
  the control plane and Lambda roles.

The supplied CloudFormation template currently does not attach the Lambda to a
VPC. Do not use it unchanged for a private PostgreSQL or private compute
endpoint. Add a `VpcConfig` to `LifecycleLambda` and appropriate network
egress before deploying that topology.

## Configure Landlord

Configure the control plane with the state-machine ARN produced by the
deployment. `caller_assume_role` is optional and applies only to the
control-plane process; it is not the Lambda or state-machine role.

```yaml
database:
  provider: postgres
  host: landlord-db.example.internal
  port: 5432
  user: landlord
  database: landlord
  ssl_mode: require

compute:
  ecs:
    cluster_arn: arn:aws:ecs:us-west-2:123456789012:cluster/landlord
    task_definition_arn: arn:aws:ecs:us-west-2:123456789012:task-definition/tenant:1
    service_name_prefix: landlord-tenant-

workflow:
  default_provider: step-functions
  step_functions:
    region: us-west-2
    state_machine_arn: arn:aws:states:us-west-2:123456789012:stateMachine:landlord-lifecycle
    caller_assume_role:
      role_arn: arn:aws:iam::123456789012:role/LandlordStepFunctionsCaller
      session_name: landlord-control-plane

controller:
  enabled: true
```

Keep database credentials in the deployment environment rather than this file.
The Lambda needs the same database and compute configuration. Package a
separate `config.yaml` for Lambda if its database endpoint or compute defaults
differ from the control plane.

The equivalent Step Functions environment variables are:

| Variable | Purpose |
| --- | --- |
| `WORKFLOW_SFN_REGION` | Region containing the state machine. |
| `WORKFLOW_SFN_STATE_MACHINE_ARN` | Required shared Standard state-machine ARN. |
| `WORKFLOW_SFN_CALLER_ASSUME_ROLE_ARN` | Optional control-plane caller role. |
| `WORKFLOW_SFN_CALLER_ASSUME_ROLE_EXTERNAL_ID` | Optional external ID for the caller role. |
| `WORKFLOW_SFN_CALLER_ASSUME_ROLE_SESSION_NAME` | Optional caller role session name. |
| `LANDLORD_CONFIG` | Explicit path to the Lambda or control-plane configuration file. |

See [Configuration](configuration.md) for general database, controller, and
compute-provider configuration.

## Deploy The Lambda And State Machine

The versioned deployment assets are:

- `deploy/aws/stepfunctions/v1/lifecycle.asl.json`: Standard state-machine
  definition.
- `deploy/aws/stepfunctions/v1/template.yaml`: CloudFormation reference
  template.

Build a Linux custom-runtime Lambda package. The ZIP must contain `bootstrap`
and `config.yaml` at its root.

```sh
rm -rf dist
mkdir -p dist/lambda

GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
  go build -o dist/lambda/bootstrap ./cmd/workers/stepfunctions

cp /path/to/lambda-config.yaml dist/lambda/config.yaml
(cd dist/lambda && zip -r ../landlord-step-functions.zip bootstrap config.yaml)
```

Upload both artifacts and deploy the reference stack:

```sh
aws s3 cp dist/landlord-step-functions.zip \
  s3://YOUR_ARTIFACT_BUCKET/landlord-step-functions.zip
aws s3 cp deploy/aws/stepfunctions/v1/lifecycle.asl.json \
  s3://YOUR_ARTIFACT_BUCKET/lifecycle.asl.json

aws cloudformation deploy \
  --stack-name landlord-step-functions \
  --template-file deploy/aws/stepfunctions/v1/template.yaml \
  --capabilities CAPABILITY_IAM \
  --parameter-overrides \
    LambdaCodeBucket=YOUR_ARTIFACT_BUCKET \
    LambdaCodeKey=landlord-step-functions.zip \
    StateMachineDefinitionBucket=YOUR_ARTIFACT_BUCKET \
    StateMachineDefinitionKey=lifecycle.asl.json \
    LambdaComputePolicyArn=arn:aws:iam::123456789012:policy/LandlordCompute
```

Retrieve the ARN and configure it in the control plane:

```sh
aws cloudformation describe-stacks \
  --stack-name landlord-step-functions \
  --query 'Stacks[0].Outputs[?OutputKey==`StateMachineArn`].OutputValue' \
  --output text

go run ./cmd/landlord --config /path/to/control-plane.yaml
```

The template creates CloudWatch log groups, a Lambda role, a state-machine
role, the Lambda function, and the Standard state machine. It expects a
customer-managed compute policy because the minimum permissions vary by compute
provider and resource naming policy.

## IAM Roles

Use distinct roles and scope resources wherever AWS supports it.

| Role | Minimum permissions |
| --- | --- |
| Landlord caller | `states:StartExecution` and `states:ListExecutions` on the lifecycle state machine; `states:DescribeExecution`, `states:GetExecutionHistory`, and `states:StopExecution` on its execution ARNs. |
| State machine | `lambda:InvokeFunction` on the lifecycle Lambda and Step Functions CloudWatch Logs delivery permissions. |
| Lambda compute | CloudWatch Logs writes, `states:DescribeExecution` for cooperative cancellation, PostgreSQL network access, and the least-privilege mutations for its configured compute provider. |

Do not grant compute-mutation permissions to the Landlord caller role. For ECS,
the Lambda compute policy should be restricted to the configured cluster, task
definition, and Landlord-managed service resources where possible.

## Smoke Test And Operations

After deployment, verify the state machine without mutating compute resources
by starting a `plan` execution:

```sh
cat > /tmp/landlord-plan.json <<'EOF'
{
  "tenant_id": "smoke-test",
  "tenant_uuid": "smoke-test",
  "operation": "plan",
  "compute_provider": "ecs",
  "desired_config": {},
  "metadata": {"config_hash": "smoke-test"}
}
EOF

aws stepfunctions start-execution \
  --state-machine-arn "$STATE_MACHINE_ARN" \
  --name "landlord-smoke-$(date +%s)" \
  --input file:///tmp/landlord-plan.json
```

Use the returned execution ARN to inspect or stop executions:

```sh
aws stepfunctions describe-execution --execution-arn "$EXECUTION_ARN"
aws stepfunctions get-execution-history --execution-arn "$EXECUTION_ARN"
aws stepfunctions stop-execution --execution-arn "$EXECUTION_ARN" --error OperatorStop
```

The Step Functions log group records state transitions and Lambda invocation
errors. Lambda logs include `execution_arn`, and PostgreSQL compute execution
records store the same value as `workflow_execution_id`; use this identifier to
correlate a controller request, state-machine execution, and compute mutation.

Stopping an execution is cooperative. The Lambda checks the execution status
before each mutable operation and avoids subsequent mutations after a stop. It
cannot safely undo a cloud-provider request that already started.

To roll back, choose a previously deployed provider such as `temporal`,
`restate`, or `mock` in `workflow.default_provider`. Existing Step Functions
executions remain observable and can be stopped with AWS tooling. Do not delete
the state machine or logs until active executions are terminal.

## Troubleshooting

| Symptom | Check |
| --- | --- |
| Startup rejects the provider | Set both `workflow.step_functions.region` and `state_machine_arn` when `step-functions` is the default. |
| Start fails with access denied | Verify the caller role has `states:StartExecution` on the configured ARN and any configured assume-role trust policy permits the control plane. |
| Lambda fails during initialization | Confirm the ZIP contains root-level `bootstrap` and `config.yaml`, and that the Lambda configuration validates with PostgreSQL and a compute provider enabled. |
| Lambda cannot reach PostgreSQL | Verify DNS, security groups, routes, and VPC attachment. The supplied template requires extension for private networking. |
| Execution stops but cloud resources remain | This is expected for a stop racing an in-flight mutation; inspect the execution ARN and compute tracking record before remediating. |
| Duplicate execution name | Landlord uses deterministic names per desired-state revision. A repeated request returns the existing execution; change the desired configuration to create a new revision. |

## Testing

Run the unit, controller, and asset tests locally:

```sh
go test ./internal/workflow/providers/stepfunctions \
  ./internal/workflow/lifecycle \
  ./internal/controller \
  ./deploy/aws/stepfunctions/v1
```

The opt-in integration test runs against AWS or an AWS-compatible endpoint. It
requires a deployed state machine, credentials with start/describe/list/stop
permissions, and optionally `AWS_ENDPOINT_URL` for LocalStack:

```sh
LANDLORD_SFN_INTEGRATION_REGION=us-west-2 \
LANDLORD_SFN_INTEGRATION_STATE_MACHINE_ARN="$STATE_MACHINE_ARN" \
AWS_ENDPOINT_URL=http://localhost:4566 \
go test -tags=integration ./internal/workflow/providers/stepfunctions
```
