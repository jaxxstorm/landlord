## MODIFIED Requirements

### Requirement: API handlers stage tenants for reconciliation
The API handlers SHALL persist tenant changes and set lifecycle status for reconciliation. Workflow execution is triggered by the reconciler, not the API.

#### Scenario: POST /tenants stages provisioning
- **WHEN** a client creates a new tenant via POST /api/tenants
- **THEN** the API handler SHALL persist the tenant with StatusRequested
- **AND** the tenant SHALL be picked up by the reconciliation loop to trigger provisioning

#### Scenario: PUT /tenants stages update
- **WHEN** a client updates an existing tenant via PUT /api/tenants/{id}
- **THEN** the API handler SHALL set StatusUpdating when appropriate
- **AND** the reconciler SHALL trigger the update workflow for the tenant

#### Scenario: DELETE /tenants stages deletion
- **WHEN** a client deletes a tenant via DELETE /api/tenants/{id}
- **THEN** the API handler SHALL set StatusDeleting
- **AND** the reconciler SHALL trigger the deletion workflow for the tenant
