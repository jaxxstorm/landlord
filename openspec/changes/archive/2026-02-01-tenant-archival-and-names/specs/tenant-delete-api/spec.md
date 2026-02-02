## MODIFIED Requirements

### Requirement: Delete tenant by ID
The system SHALL provide a DELETE endpoint at `/api/tenants/{id}` that archives a tenant by UUID or name and removes it from the database when deletion is finalized.

#### Scenario: Valid delete request
- **WHEN** a client sends a DELETE request to `/api/tenants/{id}` with a valid UUID or name
- **THEN** the system attempts to delete that tenant

#### Scenario: Tenant exists
- **WHEN** the delete request specifies an existing tenant identifier
- **THEN** the system archives the tenant and triggers compute deletion

### Requirement: Confirm successful deletion
The system SHALL return appropriate response upon successful tenant deletion.

#### Scenario: Successful delete request
- **WHEN** a tenant delete request is accepted
- **THEN** the system returns HTTP 200 with the archived tenant or HTTP 202 Accepted

### Requirement: Handle non-existent tenant
The system SHALL return appropriate error when attempting to delete non-existent tenant.

#### Scenario: Delete non-existent tenant
- **WHEN** a delete request specifies a tenant identifier that does not exist
- **THEN** the system returns HTTP 404 Not Found
