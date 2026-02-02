## Why

Unversioned APIs lock us into existing behavior and make it risky to evolve the surface area. Introducing explicit versions now lets us ship future changes (v2+) without breaking existing clients, while keeping documentation and CLI usage consistent.

## What Changes

- **BREAKING**: All HTTP APIs move to a versioned base path (e.g., `/v1/...`), with a defined policy for any legacy unversioned paths.
- Establish a single, documented versioning scheme (path prefix, supported versions, and compatibility expectations).
- CLI commands default to the current stable API version and include clear versioned endpoint references.
- Docs, examples, and references updated to reflect versioned URLs and version selection guidance.
- API error responses cover unsupported/unknown versions consistently.

## Capabilities

### New Capabilities
- `api-versioning`: Define the versioning scheme, supported versions (v1/v2), and compatibility/deprecation expectations for all HTTP APIs.

### Modified Capabilities
- `http-api-server`: Routing must enforce and handle versioned API prefixes.
- `tenant-create-api`: Endpoint paths and behaviors must be versioned.
- `tenant-get-api`: Endpoint paths and behaviors must be versioned.
- `tenant-list-api`: Endpoint paths and behaviors must be versioned.
- `tenant-update-api`: Endpoint paths and behaviors must be versioned.
- `tenant-delete-api`: Endpoint paths and behaviors must be versioned.
- `worker-control-plane-api`: Endpoint paths and behaviors must be versioned.
- `tenant-api-errors`: Error responses must cover unsupported/unknown API versions.
- `tenant-request-validation`: Validation must consider versioned paths and version rules.
- `landlord-cli`: CLI references and defaults must align with versioned APIs.
- `public-docs`: Public documentation must reflect the versioned API surface and guidance.

## Impact

- API routing, handlers, and tests across the HTTP server.
- CLI commands and any API client codepaths.
- Documentation, examples, and configuration references.
- Backward compatibility expectations and upgrade guidance for users.
