## REMOVED Requirements

### Requirement: Worker control plane API (deferred)
Worker registration, heartbeat, polling, and status update endpoints are out of scope for the reconciliation-based architecture in this change.

#### Scenario: Deferred worker control plane
- **WHEN** workers are started
- **THEN** they SHALL register directly with the workflow backend (e.g., Restate)
- **AND** Landlord SHALL rely on reconciliation and provider status polling instead of worker-control APIs
