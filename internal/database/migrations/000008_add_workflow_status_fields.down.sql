-- Remove workflow status fields from tenants table
ALTER TABLE tenants
DROP COLUMN workflow_sub_state,
DROP COLUMN workflow_retry_count,
DROP COLUMN workflow_error_message;
