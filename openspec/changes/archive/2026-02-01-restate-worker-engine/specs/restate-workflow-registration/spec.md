## MODIFIED Requirements

### Requirement: System registers workflows at startup
The system SHALL register all tenant lifecycle workflows and worker services with the Restate backend during the workflow provider initialization.

#### Scenario: Successful workflow and worker registration on startup
- **WHEN** the Restate workflow provider initializes at application startup
- **THEN** all tenant lifecycle workflows and worker services are registered with the Restate backend
- **AND** registration completion is logged

#### Scenario: Startup with unavailable Restate backend
- **WHEN** the Restate workflow provider initializes but the Restate backend is unavailable
- **THEN** the system logs a warning about registration failure
- **AND** the system continues startup without failing
