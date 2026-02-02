# Proposal: Bootstrap Compute Provisioning Framework

## Problem Statement

The landlord control plane needs a pluggable framework for provisioning tenant compute resources across different cloud providers and platforms. The system should support multiple compute engines (AWS ECS, Kubernetes, Nomad) without coupling the core tenant lifecycle logic to any specific implementation.

## Proposed Solution

Create a **compute provisioning abstraction layer** that:

1. Defines a clear interface for compute providers
2. Implements a provider registry for in-tree providers
3. Provides a standardized tenant compute specification format
4. Separates provider-specific logic from core domain logic
5. Enables future providers to be added easily

## Key Design Decisions

### Pluggable Architecture
- Use Go interfaces to define the compute provider contract
- Implement a provider registry pattern for managing multiple providers
- Allow providers to be registered at startup time

### Provider Lifecycle
Each provider must implement:
- **Provision**: Create tenant compute resources
- **Update**: Modify existing tenant compute
- **Destroy**: Clean up tenant compute resources  
- **GetStatus**: Query current state of tenant compute
- **Validate**: Validate tenant compute specifications before provisioning

### Tenant Compute Specification
- Define a provider-agnostic compute spec structure
- Include provider-specific configuration as opaque JSON
- Support multi-container deployments
- Handle secrets and configuration management

### In-Tree Provider Model
- Keep all official providers in the main repository
- Organize under `internal/compute/providers/`
- Each provider in its own package
- Shared utilities in `internal/compute/common/`

## Success Criteria

- ✅ Clear interface definition for compute providers
- ✅ Provider registry implementation
- ✅ Framework for adding new providers
- ✅ No execution logic (implementation deferred to future changes)
- ✅ Documentation for adding new providers

## Out of Scope

- Actual provider implementations (ECS, Kubernetes, etc.)
- Workflow engine integration
- State persistence beyond in-memory structures
- Provider-specific error handling and retry logic
