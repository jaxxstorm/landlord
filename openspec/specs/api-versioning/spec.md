# Specification: API Versioning

## Purpose

Define the API versioning scheme and supported versions. (TBD)

## ADDED Requirements

### Requirement: Versioned base path required
The system SHALL serve all HTTP APIs under a versioned base path of the form `/v{major}` (e.g., `/v1`).

#### Scenario: Request targets versioned base path
- **WHEN** a client sends a request to `/v1/...`
- **THEN** the system routes the request to the v1 handlers

### Requirement: Supported API versions are enforced
The system SHALL define an explicit set of supported API versions, with v1 as the initial stable version.

#### Scenario: Supported version request
- **WHEN** a client sends a request to a supported version (e.g., `/v1/...`)
- **THEN** the system processes the request normally

#### Scenario: Unsupported version request
- **WHEN** a client sends a request to `/v2/...` and v2 is not supported
- **THEN** the system returns HTTP 400 with error code `unsupported_version` and a list of supported versions

### Requirement: Unversioned requests are rejected
The system SHALL reject requests that do not include a version prefix in the path.

#### Scenario: Missing version prefix
- **WHEN** a client sends a request to an unversioned path (e.g., `/api/tenants`)
- **THEN** the system returns HTTP 400 with error code `version_required` and a list of supported versions
