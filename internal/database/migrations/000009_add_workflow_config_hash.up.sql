-- Add workflow_config_hash field to tenants table for config change detection
ALTER TABLE tenants
ADD COLUMN workflow_config_hash VARCHAR(64);

-- Index for efficient config change queries
CREATE INDEX idx_tenants_workflow_config_hash ON tenants(workflow_config_hash) WHERE workflow_config_hash IS NOT NULL;
