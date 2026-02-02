## 1. Configuration Files

- [ ] 1.1 Create config.example.yaml with all configuration options documented and commented
- [ ] 1.2 Verify config.example.yaml includes database section (PostgreSQL and SQLite examples)
- [ ] 1.3 Verify config.example.yaml includes HTTP section with all network options
- [ ] 1.4 Verify config.example.yaml includes logging section with level and format options
- [ ] 1.5 Verify config.example.yaml includes compute section with docker provider configuration
- [ ] 1.6 Verify config.example.yaml includes workflow section with mock, restate, and step-functions examples
- [ ] 1.7 Create test.config.yaml configured for Docker Compose with PostgreSQL, Docker compute, and Restate
- [ ] 1.8 Verify test.config.yaml is minimal and includes only necessary overrides
- [ ] 1.9 Add comments to test.config.yaml explaining Docker Compose context and service references

## 2. Docker Compose Setup

- [ ] 2.1 Create docker-compose.yml with PostgreSQL service (version 15, port 5432)
- [ ] 2.2 Configure PostgreSQL service with environment variables: POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB
- [ ] 2.3 Create PostgreSQL initialization scripts for database schema and migrations
- [ ] 2.4 Add Landlord service to docker-compose with Dockerfile or image build configuration
- [ ] 2.5 Mount test.config.yaml into Landlord container at /app/config.yaml
- [ ] 2.6 Mount /var/run/docker.sock from host into Landlord container for Docker compute provider
- [ ] 2.7 Configure Landlord service to depend on PostgreSQL with health checks
- [ ] 2.8 Add Restate service to docker-compose (appropriate image version, port 8080)
- [ ] 2.9 Configure shared Docker network so services can communicate via internal DNS
- [ ] 2.10 Add named volume for PostgreSQL persistence (landlord_postgres_data)
- [ ] 2.11 Set environment variables in docker-compose for service discovery (POSTGRES_HOST=postgres, RESTATE_HOST=restate)
- [ ] 2.12 Document all port mappings in docker-compose (8080 for Landlord, 8080 for Restate, 5432 for PostgreSQL)

## 3. Service Configuration

- [ ] 3.1 Ensure PostgreSQL listens on all interfaces (0.0.0.0) within container
- [ ] 3.2 Create database initialization script that creates landlord_db and applies migrations
- [ ] 3.3 Configure Landlord to use test.config.yaml as primary configuration
- [ ] 3.4 Verify Landlord can read Docker socket from /var/run/docker.sock path
- [ ] 3.5 Verify Landlord connects to PostgreSQL at postgres:5432 using test.config.yaml credentials
- [ ] 3.6 Verify Landlord connects to Restate at restate:8080 using test.config.yaml endpoint
- [ ] 3.7 Add health check to PostgreSQL service (psql command)
- [ ] 3.8 Add health check to Landlord service (HTTP GET /health endpoint)
- [ ] 3.9 Add health check to Restate service if applicable
- [ ] 3.10 Ensure proper startup order: PostgreSQL → Landlord, Restate independent

## 4. Documentation

- [ ] 4.1 Create LOCAL_DEVELOPMENT.md with quickstart section (clone, docker-compose up, verify)
- [ ] 4.2 Add prerequisites section (Docker Desktop/Engine version, docker-compose, system requirements)
- [ ] 4.3 Add service architecture section explaining three services, their roles, and communication
- [ ] 4.4 Document all exposed ports and their purposes (Landlord 8080, Restate 8080, PostgreSQL 5432)
- [ ] 4.5 Explain Docker socket mounting: why it's needed, security implications, how it works
- [ ] 4.6 Document PostgreSQL data persistence and how to reset database (docker-compose down -v)
- [ ] 4.7 Create troubleshooting section for port conflicts
- [ ] 4.8 Create troubleshooting section for Docker socket permission issues
- [ ] 4.9 Create troubleshooting section for service connectivity (PostgreSQL, Restate, Docker)
- [ ] 4.10 Create troubleshooting section for database initialization failures
- [ ] 4.11 Create troubleshooting section for image pulling/registry access
- [ ] 4.12 Add platform-specific guidance for macOS (Docker Desktop, socket paths)
- [ ] 4.13 Add platform-specific guidance for Linux (Docker Engine, user groups)
- [ ] 4.14 Add platform-specific guidance for Windows (WSL 2, Hyper-V, socket path)
- [ ] 4.15 Add next steps section with API testing examples (curl calls to health endpoints)
- [ ] 4.16 Add tenant provisioning example showing Docker compute provider usage
- [ ] 4.17 Add workflow testing documentation explaining how to trigger Restate workflows
- [ ] 4.18 Document log inspection using docker-compose logs with service filtering
- [ ] 4.19 Update main README.md to reference new LOCAL_DEVELOPMENT.md

## 5. Testing and Validation

- [ ] 5.1 Test docker-compose up successfully starts all three services
- [ ] 5.2 Verify PostgreSQL is accessible at postgres:5432 from Landlord container
- [ ] 5.3 Verify Landlord health check endpoint responds (GET /health returns 200)
- [ ] 5.4 Verify Landlord successfully connects to PostgreSQL database
- [ ] 5.5 Verify Landlord can read and parse test.config.yaml
- [ ] 5.6 Verify Landlord can access Docker socket and provision containers
- [ ] 5.7 Verify Restate service is running and accessible
- [ ] 5.8 Test docker-compose down properly stops and removes containers
- [ ] 5.9 Test docker-compose down -v removes named volume and cleans database
- [ ] 5.10 Test docker-compose up again creates fresh database from migrations
- [ ] 5.11 Test tenant provisioning via Docker compute provider (create container, verify running)
- [ ] 5.12 Verify all services communicate correctly on shared network
- [ ] 5.13 Test service startup order: PostgreSQL fully ready before Landlord connects
- [ ] 5.14 Test configuration inheritance: test.config.yaml overrides example defaults

## 6. Archive Preparation

- [ ] 6.1 Review all artifacts (proposal, design, specs, tasks) for completeness
- [ ] 6.2 Verify all requirements from specs are addressed in tasks
- [ ] 6.3 Ensure no gaps between design decisions and implementation tasks
- [ ] 6.4 Create .openspec.yaml metadata file with change summary
- [ ] 6.5 Move change from openspec/changes/ to openspec/changes/archive/
- [ ] 6.6 Verify archive contains: proposal.md, design.md, specs/, tasks.md, .openspec.yaml
