## Context

### Current State

Landlord currently provides two workflow providers:
- **Mock Provider**: In-memory provider for testing and local development, stores workflows and executions in-memory without external dependencies
- **Step Functions Provider**: AWS Step Functions for production workflows, integrates with AWS services

Both providers implement the `workflow.Provider` interface defined in `internal/workflow/provider.go`. The provider registration pattern uses a thread-safe `Registry` component in `internal/workflow/registry.go` that maintains a map of providers and allows retrieval by name.

### Motivation

While Step Functions provides production-grade durable execution, it requires AWS credentials and infrastructure setup. The mock provider works locally but doesn't provide durable execution semantics or state persistence. This gap means developers cannot test workflows locally with production execution semantics.

Restate.dev fills this gap by providing:
- Durable execution with strong consistency guarantees (ACID properties)
- Works locally without cloud dependencies (single Docker container)
- Scales to production with multiple deployment options (Lambda, Fargate, Kubernetes, self-hosted)
- State management and recovery built-in
- Same execution semantics in local and production environments

### Constraints

- **Provider Interface**: Must implement the existing `workflow.Provider` interface without modifications
- **Configuration**: Must integrate with the existing Cobra/Viper configuration system
- **Backward Compatibility**: Addition of restate provider must not break existing mock or step-functions workflows
- **Dependencies**: Restate Go SDK must be compatible with Go 1.24.7
- **Registry Pattern**: Must register provider at application initialization through the existing Registry mechanism

### Stakeholders

- **Developers**: Local testing with production-like durable execution
- **Operations**: Flexibility to choose execution mechanism (Lambda vs Fargate vs Kubernetes vs self-hosted)
- **DevOps**: Integration with existing configuration and deployment systems

---

## Goals / Non-Goals

**Goals:**

1. Add Restate.dev as a third workflow provider option alongside mock and step-functions
2. Enable local development with production execution semantics (durable execution, state recovery)
3. Support multiple production execution mechanisms (Lambda, Fargate, Kubernetes, self-hosted)
4. Provide configurable server endpoint for local and production deployments
5. Implement proper error handling and validation for Restate-specific configuration
6. Maintain backward compatibility with existing workflows and providers
7. Document setup, configuration, and usage patterns for all deployment scenarios

**Non-Goals:**

1. Modify the existing `Provider` interface (restate provider must implement the interface as-is)
2. Add Restate-specific workflow definition language (workflows remain in existing format compatible with all providers)
3. Implement full Restate service layer abstraction (restate provider integrates with standard Restate APIs)
4. Replace mock provider for testing (mock remains lightweight for basic testing scenarios)
5. Add advanced Restate-specific features beyond the Provider interface (e.g., sagas, activities)

---

## Decisions

### 1. Package Structure for Restate Provider

**Decision**: Create restate provider at `internal/workflow/providers/restate/` following the existing mock provider pattern.

**Rationale**: Consistency with mock provider structure. Each provider is self-contained in its own package under `providers/`, making it easy to understand scope and dependencies.

**Alternatives Considered**:
- Flatten into `internal/workflow/restate/` (rejected: breaks consistency with mock structure)
- Create at `pkg/workflow/restate/` (rejected: internal workflow is already in `internal/workflow`)

**Implementation**:
```
internal/workflow/providers/restate/
├── provider.go           # Provider implementation
├── client.go             # Restate SDK client wrapper
├── config.go             # Restate-specific configuration
├── errors.go             # Restate-specific error types
└── provider_test.go      # Unit tests
```

---

### 2. Restate Client Initialization Strategy

**Decision**: Lazy-initialize the Restate SDK client on first use rather than in provider constructor.

**Rationale**: 
- Allows provider registration to succeed even if Restate server is not yet available
- Supports testing without requiring Restate server to be running during initialization
- Provides clear error messages if server becomes unavailable later
- Follows similar pattern to database clients in Landlord

**Alternatives Considered**:
- Eager initialization in constructor (rejected: fails initialization if server is down)
- No connection validation (rejected: misses connection issues until first workflow use)

**Implementation**:
- `New()` stores configuration but does not connect
- `CreateWorkflow()` triggers lazy initialization if not already done
- Connection errors are wrapped with context about the endpoint

---

### 3. Endpoint Configuration for Local vs Production

**Decision**: Single `workflow.restate.endpoint` configuration supports both local and production deployments with automatic detection.

**Rationale**:
- Simple configuration model: single field controls environment
- Local development: `http://localhost:8080` (default for developers)
- Production: `https://restate.example.com` (operator-configured)
- Matches existing Landlord configuration patterns (one configuration, environment-specific values)

**Alternatives Considered**:
- Separate local/production configuration sections (rejected: adds complexity, inconsistent with existing patterns)
- Environment variable detection (rejected: mixing config and environment is confusing)

**Implementation**:
- Configuration validation confirms endpoint is valid HTTP/HTTPS URL
- SDK client connects to endpoint as-is (no transformation)
- Local development uses Docker Compose with Restate container on localhost:8080

---

### 4. Execution Mechanism Configuration

**Decision**: Configuration field `workflow.restate.execution_mechanism` specifies deployment target; provider validates compatibility and restate server handles deployment details.

**Rationale**:
- Execution mechanism is deployment/infrastructure concern, not core provider concern
- Restate server itself manages the deployment based on configuration
- Provider focuses on workflow interface, leaves deployment to infrastructure layer
- Matches how mock provider works (simple in-memory execution)

**Supported Mechanisms**:
- `local`: Use Restate's local compute options (default, best for development)
- `lambda`: AWS Lambda invocation
- `fargate`: AWS ECS Fargate
- `kubernetes`: Kubernetes service discovery and invocation
- `self-hosted`: Direct HTTP to Restate instances

**Implementation**:
- Configuration field stored but primarily for documentation/deployment validation
- Provider passes mechanism to Restate SDK during service registration
- Restate server determines how to invoke services based on mechanism

---

### 5. Service Registration and Naming

**Decision**: Service name is derived from workflow ID with optional override via `workflow.restate.service_name`.

**Rationale**:
- Workflow ID uniquely identifies workflow in Landlord
- Provides default service name automatically
- Allows override for complex scenarios (e.g., versioned services)
- Maintains consistency: one workflow → one service

**Naming Convention**:
- Default: Convert workflow ID `tenant-provisioning` → service name `TenantProvisioning` (PascalCase, required by Restate)
- Override: Use configured `service_name` if provided

**Implementation**:
- `normalizeServiceName()` function converts kebab-case workflow ID to PascalCase
- Configuration override in `workflow.restate.service_name`
- Validation ensures name is alphanumeric (Restate requirement)

---

### 6. Authentication Approach

**Decision**: Support multiple authentication methods per execution mechanism; provider validates auth type matches mechanism.

**Rationale**:
- Different mechanisms require different auth (Lambda uses IAM, self-hosted uses API keys)
- Authentication is deployment-concern, not core workflow concern
- Validation catches misconfigurations early

**Supported Auth Types**:
- `api_key`: API key for direct HTTP calls (self-hosted)
- `iam`: AWS IAM roles (Lambda, Fargate)
- `none`: No authentication (local development)

**Validation Rules**:
- Local endpoint → `none` auth (no validation required)
- Self-hosted mechanism → `api_key` required
- Lambda/Fargate → IAM role configured via SDK
- Provider validates at initialization, returns clear error if misconfigured

**Implementation**:
- Auth configuration stored separately from endpoint
- Provider passes auth to Restate SDK client
- Connection test at initialization verifies auth works

---

### 7. Configuration Integration with Viper

**Decision**: Restate configuration lives in `workflow.restate` section, integrated with existing Cobra/Viper system.

**Rationale**: Consistent with existing workflow provider configuration patterns; supports environment variable overrides and CLI flags through Landlord's standard precedence rules.

**Configuration Structure**:
```yaml
workflow:
  default_provider: "restate"  # or "mock", "step-functions"
  restate:
    endpoint: "http://localhost:8080"
    execution_mechanism: "local"
    service_name: ""  # optional override
    auth_type: "none"
    timeout: 30m
    retry_attempts: 3
```

**Environment Variable Support**:
- `LANDLORD_WORKFLOW_RESTATE_ENDPOINT=http://localhost:8080`
- `LANDLORD_WORKFLOW_RESTATE_EXECUTION_MECHANISM=lambda`
- Follows Landlord's standard env var naming convention

**CLI Flag Support**:
- `--workflow.restate.endpoint=http://localhost:8080`
- Highest precedence, overrides all other sources

---

### 8. Error Handling Strategy

**Decision**: Wrap Restate SDK errors with context and map to common Provider interface errors.

**Rationale**:
- Provider interface defines standard errors (ErrProviderNotFound, ErrInvalidSpec, etc.)
- Restate SDK uses its own error types
- Wrapping provides consistency with other providers and clear error messages
- Includes Restate-specific context (endpoint, service name) in error messages

**Error Mapping**:
- Restate connection error → Provider context error + endpoint info
- Service not found → ErrWorkflowNotFound
- Invalid workflow definition → ErrInvalidSpec
- Authentication failure → Descriptive error with auth type

**Implementation**:
- `wrapRestateError()` function handles mapping
- Errors include structured logging context
- Error messages suitable for end users

---

### 9. Validation and Idempotency

**Decision**: Provider explicitly validates all inputs and ensures all operations are idempotent as required by Provider interface.

**Rationale**:
- Provider interface requires idempotent operations
- Validation catches configuration/input errors early with clear messages
- Idempotency allows safe retries and deployment updates

**Validation Points**:
- Configuration validation at initialization
- Workflow spec validation before CreateWorkflow
- Service name normalization validation
- Endpoint URL format validation

**Idempotency Implementation**:
- `CreateWorkflow()`: Check if service already registered, return success if so
- `DeleteWorkflow()`: Return success if service doesn't exist
- `StopExecution()`: Return success if execution already stopped
- Uses Restate SDK's built-in idempotency where available

---

### 10. State Management and Recovery

**Decision**: Leverage Restate's built-in state management; provider uses SDK APIs for state operations.

**Rationale**:
- Restate handles state persistence and recovery automatically
- Provider focuses on workflow interface, delegates state to Restate
- Provides strong consistency guarantees through Restate
- Automatic recovery from failures without provider intervention

**State Operations**:
- `GetExecutionStatus()` queries Restate for current execution state
- State includes workflow input, output, and execution history
- Execution history provides visibility into execution progress
- Restate manages durability; provider doesn't need to implement state storage

---

### 11. Testing Strategy

**Decision**: Local development uses Docker Compose with Restate container; tests use mock provider or local Restate instance.

**Rationale**:
- Docker Compose provides simple, reproducible local environment
- Real Restate instance for integration testing (better than mocking Restate itself)
- Unit tests can use existing mock provider for Provider interface contract validation
- Matches patterns used in Landlord for other services (database, etc.)

**Test Scenarios**:
- Unit tests: Restate provider implements Provider interface correctly
- Integration tests: Local Restate container via Docker Compose
- Configuration tests: Validation of all configuration options
- Error case tests: Network failures, invalid endpoints, auth failures

**Implementation**:
- `docker-compose.yml` includes Restate container for local development
- Integration tests spawn container if not already running
- Unit tests mock Restate SDK client for testing provider logic

---

### 12. Documentation and Setup Guides

**Decision**: Provide setup guides for each deployment scenario (local, Lambda, Fargate, Kubernetes).

**Rationale**:
- Each scenario has different configuration and prerequisites
- Clear documentation reduces setup time and errors
- Covers both developer (local) and operator (production) use cases

**Documentation Sections**:
1. Local Development (Docker Compose)
2. AWS Lambda Deployment
3. AWS ECS Fargate Deployment
4. Kubernetes Deployment
5. Self-Hosted Deployment
6. Configuration Reference
7. Migration Guide (from mock or step-functions)

---

## Risks / Trade-offs

### Risk 1: Restate Server Availability
**Impact**: If Restate server is down, workflows fail.

**Mitigation**:
- Documentation recommends high-availability Restate deployment (auto-recovery)
- Connection validation at provider initialization detects issues early
- Clear error messages guide troubleshooting
- Option to stick with mock provider for testing if Restate unavailable

---

### Risk 2: Execution Mechanism Complexity
**Impact**: Operators must understand deployment mechanics for each mechanism (Lambda, Fargate, Kubernetes).

**Mitigation**:
- Comprehensive documentation with examples for each mechanism
- Configuration templates for common scenarios
- Setup guides provide step-by-step instructions
- Start with local development, progress to production mechanisms

---

### Risk 3: New Dependency (Restate SDK)
**Impact**: Adds dependency on external Restate SDK; SDK updates could break compatibility.

**Mitigation**:
- Pin Restate SDK version in go.mod for stability
- Monitor Restate SDK releases for breaking changes
- Isolation in separate package makes updates easier
- Fallback to mock/step-functions if needed

---

### Risk 4: Authentication Configuration Complexity
**Impact**: Misconfigured authentication prevents workflow execution.

**Mitigation**:
- Validation at provider initialization catches misconfigurations
- Error messages indicate what auth is required for the mechanism
- Examples in documentation show correct patterns
- Simpler defaults (no auth for localhost)

---

### Trade-off 1: Lazy Client Initialization
**Trade-off**: Provider initialization succeeds even if Restate is unavailable, errors happen later at first workflow use.

**Justification**: Allows provider registration to succeed during app startup; better to fail explicitly when trying to use provider. Mirrors database client pattern.

---

### Trade-off 2: Service Name Normalization
**Trade-off**: Automatically converts workflow IDs to service names; override available but not automatic.

**Justification**: Most workflows don't need override; simplifies configuration. Override available for complex scenarios.

---

### Trade-off 3: Single Endpoint Configuration
**Trade-off**: One endpoint supports both local and production; environment-specific values in deployment configuration.

**Justification**: Simpler configuration model; aligns with Landlord's pattern. Environment differences handled through deployment config (Docker Compose, Helm, etc.).

---

## Migration Plan

### Phase 1: Local Development
1. Developers add `docker-compose.yml` to project
2. Start Restate container: `docker-compose up restate`
3. Configure `workflow.default_provider = "restate"`
4. Set `workflow.restate.endpoint = "http://localhost:8080"`
5. Existing workflows continue working with Restate semantics

### Phase 2: Testing Infrastructure
1. Update CI/CD to include Restate container
2. Update integration tests to use Restate provider
3. Run full workflow test suite against Restate
4. Maintain mock provider for unit tests

### Phase 3: Staging Deployment
1. Deploy Restate to staging environment (choose mechanism)
2. Configure endpoint for staging Restate
3. Run production-like workflows on staging
4. Validate error handling and recovery

### Phase 4: Production Deployment
1. Deploy Restate infrastructure (Lambda/Fargate/Kubernetes/self-hosted)
2. Configure endpoint and execution mechanism
3. Set up monitoring and alerting for Restate
4. Deploy Landlord with restate provider enabled
5. Monitor workflow execution and state recovery

### Rollback Strategy
- Keep step-functions provider available as fallback
- If issues with Restate, switch `workflow.default_provider` back to step-functions
- No code changes needed; only configuration change required
- Workflows created on Restate won't transfer to step-functions (infrastructure-specific state)

---

## Open Questions

1. **Service Discovery**: For Kubernetes deployment, how should Landlord discover Restate services? DNS? Service mesh?
2. **Monitoring Integration**: Should provider expose Restate metrics to Landlord's monitoring system?
3. **State Export**: Should provider provide capability to export Restate state for backup/restore?
4. **Advanced Features**: Should future iterations add support for Restate-specific features (sagas, activities)?
5. **Cost Considerations**: Should documentation include cost comparison across execution mechanisms?
6. **Multi-Region**: Does operator need multi-region Restate setup? How is replication handled?
