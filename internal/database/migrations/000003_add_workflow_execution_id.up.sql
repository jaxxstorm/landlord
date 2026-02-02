-- Add workflow_execution_id column to tenants table
ALTER TABLE tenants
ADD COLUMN workflow_execution_id VARCHAR(255);

-- Create index for efficient workflow execution lookups
CREATE INDEX idx_tenants_workflow_execution_id ON tenants(workflow_execution_id) WHERE workflow_execution_id IS NOT NULL;
