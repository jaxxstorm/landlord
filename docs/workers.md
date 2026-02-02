# Worker Types

Workers execute compute actions for workflow steps. The worker runtime is tied to the selected workflow provider.

## Supported worker types

| Worker type | Related workflow provider | Notes |
| --- | --- | --- |
| restate | restate | Runs as a Restate service that performs compute actions |

## Restate worker

The Restate worker runs as a separate process and connects to the Restate admin endpoint to register services.

Example environment configuration:

```bash
export WORKFLOW_DEFAULT_PROVIDER=restate
export WORKFLOW_RESTATE_ENDPOINT=http://localhost:8080
export WORKFLOW_RESTATE_ADMIN_ENDPOINT=http://localhost:9070
export WORKFLOW_RESTATE_WORKER_REGISTER_ON_STARTUP=true
export WORKFLOW_RESTATE_WORKER_ADMIN_ENDPOINT=http://localhost:9070
export WORKFLOW_RESTATE_WORKER_COMPUTE_PROVIDER=mock
export WORKFLOW_RESTATE_WORKER_LANDLORD_API_URL=http://localhost:8080
```

Run the worker:

```bash
go run ./cmd/workers/restate
```

Public Restate documentation:
- https://docs.restate.dev/
