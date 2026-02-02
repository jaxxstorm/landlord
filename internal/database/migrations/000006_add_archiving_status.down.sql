-- Remove archiving status from checks
ALTER TABLE tenants DROP CONSTRAINT IF EXISTS tenants_status_check;
ALTER TABLE tenants ADD CONSTRAINT tenants_status_check
    CHECK (status IN ('requested', 'planning', 'provisioning', 'ready', 'updating', 'deleting', 'archived', 'failed'));

ALTER TABLE tenant_state_history DROP CONSTRAINT IF EXISTS tenant_state_history_from_status_check;
ALTER TABLE tenant_state_history ADD CONSTRAINT tenant_state_history_from_status_check
    CHECK (from_status IN ('requested', 'planning', 'provisioning', 'ready', 'updating', 'deleting', 'archived', 'failed'));

ALTER TABLE tenant_state_history DROP CONSTRAINT IF EXISTS tenant_state_history_to_status_check;
ALTER TABLE tenant_state_history ADD CONSTRAINT tenant_state_history_to_status_check
    CHECK (to_status IN ('requested', 'planning', 'provisioning', 'ready', 'updating', 'deleting', 'archived', 'failed'));
