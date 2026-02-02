# Landlord Documentation

Landlord is an open-source tenant provisioning control plane. It provides a declarative API for creating, managing, and reconciling long-lived tenants backed by workflow orchestration, compute provisioning, and persistent state.

## How Landlord works

Landlord follows a reconciliation model:

- The API accepts a desired tenant definition.
- The controller compares desired and observed state and creates a plan.
- A workflow provider executes the plan and tracks execution status.
- A worker performs compute actions for workflow steps.
- A database persists state, transitions, and audit history.

## API versioning

Landlord uses path-based API versioning. The current stable version is `v1`, and all HTTP endpoints are served under `/v1`.

## Components

- **Compute**: Provisions runtime resources for tenants.
- **API**: Validates and exposes tenant lifecycle operations.
- **Worker**: Executes workflow tasks and talks to compute providers.
- **Database**: Stores tenants, transitions, and history.
- **Workflow**: Orchestrates provisioning and lifecycle transitions.

## Pluggable design

Landlord uses registries for compute, workflow, and database providers. Each component can be swapped by configuration without changing the API surface or core controller logic.

## Supported plugins

| Component | Supported plugins | Notes |
| --- | --- | --- |
| Compute | docker, mock | Docker for local/dev, mock for tests |
| Workflow | restate, step-functions, mock | Choose based on infrastructure and durability needs |
| Database | postgres, sqlite | Postgres for production, sqlite for local/dev |
| Worker | restate | Worker runtime for Restate workflows |

## Where to go next

- Quickstart: `quickstart.md`
- Compute providers: `compute-providers.md`
- Workflow providers: `workflow-providers.md`
- Database types: `database.md`
- Worker types: `workers.md`
- API browser: `api.md`
