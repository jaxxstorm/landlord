## ADDED Requirements

### Requirement: Return HTTP 400 for unsupported API version
The system SHALL return HTTP 400 Bad Request when a request targets an unsupported API version.

#### Scenario: Unsupported version requested
- **WHEN** a client sends a request to `/v2/...` and v2 is not supported
- **THEN** the system returns HTTP 400 with error code `unsupported_version` and a list of supported versions

#### Scenario: Missing version prefix
- **WHEN** a client sends a request to an unversioned path (e.g., `/api/tenants`)
- **THEN** the system returns HTTP 400 with error code `version_required` and a list of supported versions
