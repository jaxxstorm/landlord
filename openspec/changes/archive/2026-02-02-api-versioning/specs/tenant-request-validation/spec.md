## ADDED Requirements

### Requirement: Validate API version in request path
The system SHALL validate that tenant API requests include a supported version prefix in the URL path.

#### Scenario: Supported version prefix
- **WHEN** a request targets `/v1/...`
- **THEN** the system accepts the version and proceeds with validation

#### Scenario: Unsupported version prefix
- **WHEN** a request targets `/v2/...` and v2 is not supported
- **THEN** the system returns HTTP 400 with error code `unsupported_version`

#### Scenario: Missing version prefix
- **WHEN** a request targets an unversioned path (e.g., `/api/tenants`)
- **THEN** the system returns HTTP 400 with error code `version_required`
