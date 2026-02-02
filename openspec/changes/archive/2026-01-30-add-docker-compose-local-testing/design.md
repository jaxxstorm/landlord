## Context

Developers currently need to manually coordinate setup of multiple services (PostgreSQL, Landlord, Restate) and manage multiple configuration files for different scenarios (local, testing, production). This creates friction during onboarding and makes it difficult to verify the complete system works end-to-end locally.

The Docker Compute Provider has been recently implemented but lacks a documented local testing setup. Restate is configured as a workflow provider. PostgreSQL is the production database. Consolidating these into a Docker Compose setup with unified configuration will accelerate local development.

**Current State:**
- Multiple config files exist: `config.yaml`, `config.json`, `config.sqlite.yaml`, `config.docker.yaml`
- No Docker Compose setup for local testing
- Documentation scattered across separate config files
- Developers must start services manually

**Stakeholders:**
- Developers (primary): Need quick local setup
- QA: Benefit from reproducible testing environment
- CI/CD: Can use docker-compose for automated integration testing
- Documentation: Single source of truth for configuration options

## Goals / Non-Goals

**Goals:**
1. Provide a `docker-compose.yml` that starts PostgreSQL, Landlord, and Restate with proper configuration
2. Create `config.example.yaml` as comprehensive reference documentation for all configuration options
3. Create `test.config.yaml` specifically tuned for local Docker Compose testing
4. Enable developers to run the complete system with a single `docker-compose up` command
5. Document how the Docker compute provider works within Docker Compose (socket mounting)
6. Consolidate configuration documentation into single reference file

**Non-Goals:**
- Production-ready Kubernetes manifests (docker-compose is for local development only)
- Database migration management (migrations run via existing setup)
- Secret management integration (local testing uses plain text credentials)
- Health check hooks (out of scope for basic setup)
- Horizontal scaling configuration

## Decisions

### Decision 1: Single docker-compose.yml for All Services

**Choice**: One `docker-compose.yml` file orchestrating PostgreSQL, Landlord, and Restate

**Rationale**: 
- Simplicity: Single file developers interact with
- Clarity: Complete system visible in one place
- Consistency: All services configured in same file

**Alternatives Considered**:
- Multiple docker-compose files (one per service): More modular but adds complexity for developers
- Separate docker-compose files for different scenarios (dev, test): More flexibility but harder to maintain

**Implementation Details**:
- PostgreSQL service: Initializes database with migrations on startup
- Landlord service: Mounts socket, volumes, and test configuration
- Restate service: Pre-configured with default ports
- Network: Shared Docker network for inter-service communication
- Volumes: Persistent PostgreSQL data, configuration file

### Decision 2: Consolidate Configurations into config.example.yaml

**Choice**: Single reference file with all options documented, commented

**Rationale**:
- Single source of truth for configuration
- All options visible and documented together
- Easier for developers to understand available options
- Cleaner repository (removes 3 other config files from root)

**Alternatives Considered**:
- Keep separate config files: More fragmented, harder to maintain
- Generate config from code: Over-engineered for current needs

**Implementation Details**:
- Comments explain each section, field, and option
- Default values clearly marked
- Examples for common scenarios (PostgreSQL, SQLite, AWS, etc.)
- Organized by component: database, http, log, compute, workflow

### Decision 3: Separate test.config.yaml for Docker Compose

**Choice**: Dedicated minimal configuration file for Docker Compose testing

**Rationale**:
- Keeps test configuration separate from example documentation
- Allows different settings for local testing vs. reference docs
- Easier to maintain test-specific values
- Can be versioned/updated independently

**Alternatives Considered**:
- Generate test config from docker-compose environment variables: Less readable, harder to maintain
- Include inline in docker-compose.yml: Mixes orchestration with configuration

**Implementation Details**:
- test.config.yaml includes only necessary overrides
- Leverages example defaults for other options
- Configured for PostgreSQL (matches docker-compose service)
- Docker compute provider enabled
- Restate workflow provider configured

### Decision 4: Docker Socket Mounting Strategy

**Choice**: Mount `/var/run/docker.sock` from host into Landlord container

**Rationale**:
- Simplest approach for local development
- No special privileges required (use docker group)
- Fast and direct access to Docker daemon
- Standard pattern in Docker development tools

**Alternatives Considered**:
- Docker-in-Docker (DinD): Heavier, more complex, unnecessary for local testing
- TCP connection to remote daemon: Adds network complexity

**Implementation Details**:
- Mount: `- /var/run/docker.sock:/var/run/docker.sock`
- Ownership: Container process inherits host docker permissions
- Network: "bridge" mode for tenant containers

### Decision 5: File Organization in Root Directory

**Choice**: Create new files in root; keep all configs at root level

**Rationale**:
- docker-compose.yml: Standard location for Docker development
- config files at root: Easy discoverability for developers
- Consistency with existing file layout

**Alternatives Considered**:
- Create configs/ subdirectory: Adds nesting complexity
- Keep configs distributed: Current problem we're solving

## Risks / Trade-offs

**Risk 1: Docker Socket Permissions**
- **Issue**: Host docker.sock may require root access or specific group membership
- **Mitigation**: Document group membership requirement. Provide troubleshooting guide for permission issues.

**Risk 2: Port Conflicts on Host**
- **Issue**: Docker Compose services (Landlord 8080, Restate 8080) may conflict with running services
- **Mitigation**: Document port mappings. Include override instructions for custom ports. Provide `docker-compose down` cleanup steps.

**Risk 3: Database Initialization Timing**
- **Issue**: Landlord service may start before PostgreSQL is ready
- **Mitigation**: Use `depends_on` with health checks. Add retry logic in Landlord startup.

**Risk 4: Configuration File Evolution**
- **Issue**: config.example.yaml may become stale as options change
- **Mitigation**: Link config.example.yaml evolution to code documentation. Add validation to catch stale examples.

**Risk 5: Platform-Specific Issues**
- **Issue**: Docker socket path and docker group availability differs between macOS (Docker Desktop) and Linux
- **Mitigation**: Document platform-specific setup. Provide macOS-specific instructions.

**Trade-off 1: Consolidation vs. Granularity**
- **Decision**: Single config.example.yaml
- **Benefit**: Single source of truth, easier maintenance
- **Cost**: Slightly larger file, may be overwhelming for simple use cases
- **Mitigation**: Good commenting and organization minimize this cost

**Trade-off 2: Minimal vs. Full Configuration**
- **Decision**: test.config.yaml is minimal; inherits example defaults
- **Benefit**: Easier to maintain, clear what's test-specific
- **Cost**: Less explicit (requires understanding defaults)
- **Mitigation**: Comments in test.config.yaml clarify inherited values

## Migration Plan

**Phase 1: Addition (No Breaking Changes)**
1. Create `docker-compose.yml` (new file)
2. Create `config.example.yaml` (new file)
3. Create `test.config.yaml` (new file)
4. Update documentation to reference new files

**Phase 2: Deprecation (Optional, Future)**
- Existing `config.yaml`, `config.json`, etc. remain for backward compatibility
- Documentation points to `config.example.yaml` as primary reference
- Eventually deprecate old files in future release

**Rollout Strategy**:
1. Review and test docker-compose locally
2. Test with CI/CD pipeline
3. Document in contributing guide
4. Update onboarding documentation
5. Add to README

**Rollback Strategy**:
- If docker-compose issues arise, can remove new files
- Existing configurations continue to work
- No breaking changes to Landlord itself

## Open Questions

1. **Restate Configuration**: Should Restate configuration options be exposed in test.config.yaml or are defaults sufficient?
   - **Proposal**: Use Restate defaults for now; expose only if developers need customization
   
2. **PostgreSQL Version**: Which PostgreSQL version should docker-compose use? 12, 13, 14, 15?
   - **Proposal**: Use PostgreSQL 15 (recent, well-supported)
   
3. **Data Persistence**: Should PostgreSQL data persist between docker-compose down/up cycles?
   - **Proposal**: Yes, use named volume for development convenience
   
4. **Credentials**: Should test.config.yaml use hardcoded credentials or environment variables?
   - **Proposal**: Hardcoded for simplicity; document security implications for non-test use
   
5. **Build Strategy**: Should docker-compose build Landlord image or use pre-built?
   - **Proposal**: Build from current codebase (docker-compose build flag), ensures latest code

## Success Criteria

1. ✅ `docker-compose up` successfully starts all three services
2. ✅ Landlord successfully connects to PostgreSQL database
3. ✅ Landlord successfully connects to Restate service
4. ✅ Docker socket mounting works (Landlord can provision containers)
5. ✅ Health checks confirm all services are running
6. ✅ config.example.yaml documents all configuration options
7. ✅ test.config.yaml is minimal and focused
8. ✅ Documentation includes setup and troubleshooting guides
