-- Remove workflow_execution_id column from tenants table
ALTER TABLE tenants
DROP COLUMN workflow_execution_id;

-- Drop index
DROP INDEX IF EXISTS idx_tenants_workflow_execution_id;
