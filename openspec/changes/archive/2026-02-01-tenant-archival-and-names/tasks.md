## 1. Persistence and domain model

- [x] 1.1 Update tenant domain model for archived status and hard delete semantics
- [x] 1.2 Add unique tenant name column/constraint and ensure UUID generation on create
- [x] 1.3 Update repository methods to lookup by UUID or name and to hard delete records
- [x] 1.4 Update migrations and seed/test data for new schema

## 2. API handlers and validation

- [x] 2.1 Change create request to accept tenant-name and return UUID + name
- [x] 2.2 Update get/update/delete handlers to accept UUID or name identifiers
- [x] 2.3 Update request validation for name uniqueness and identifier parsing
- [x] 2.4 Adjust delete flow to archive after compute teardown and hard delete on final delete

## 3. Workflow and reconciliation

- [x] 3.1 Update lifecycle status transitions to include archived state
- [x] 3.2 Update delete workflow to transition to archived and treat archived as terminal for compute
- [x] 3.3 Update reconciler and workflow status checks for archived vs deleted

## 4. CLI and client

- [x] 4.1 Update CLI create to accept tenant-name and display generated UUID
- [x] 4.2 Update CLI get/set/delete to accept name or UUID identifiers
- [x] 4.3 Update CLI client resolution logic for name vs UUID lookups

## 5. Tests and docs

- [x] 5.1 Update API tests for new create/get/update/delete behaviors
- [x] 5.2 Add persistence tests for archived vs deleted behavior and name uniqueness
- [x] 5.3 Update CLI tests for name-based workflows
- [x] 5.4 Update README/API docs for new naming and lifecycle semantics
