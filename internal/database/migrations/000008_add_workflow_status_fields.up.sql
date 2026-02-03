-- Add workflow status fields to tenants table
ALTER TABLE tenants
ADD COLUMN workflow_sub_state VARCHAR(50),
ADD COLUMN workflow_retry_count INTEGER,
ADD COLUMN workflow_error_message TEXT;
