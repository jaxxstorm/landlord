## ADDED Requirements

### Requirement: Versioned worker control plane endpoints
If worker control plane endpoints are introduced, they SHALL be served under the versioned base path (e.g., `/v1`).

#### Scenario: Worker control plane endpoint added
- **WHEN** a worker control plane endpoint is exposed
- **THEN** its path is prefixed with `/v1`

#### Scenario: Unversioned worker control plane request
- **WHEN** a client sends a request to an unversioned worker control plane path
- **THEN** the system returns HTTP 400 with error code `version_required`
