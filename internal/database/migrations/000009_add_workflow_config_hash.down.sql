-- Remove workflow_config_hash field from tenants table
DROP INDEX IF EXISTS idx_tenants_workflow_config_hash;
ALTER TABLE tenants DROP COLUMN workflow_config_hash;
