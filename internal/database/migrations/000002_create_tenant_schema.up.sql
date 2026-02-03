-- Create tenants table for current state tracking
CREATE TABLE tenants (
    -- Identity
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE, -- Human-friendly identifier
    
    -- Status
    status VARCHAR(20) NOT NULL,
    status_message TEXT,
    
    -- Desired state
    desired_config JSON NOT NULL DEFAULT '{}',
    
    -- Observed state
    observed_config JSON NOT NULL DEFAULT '{}',
    observed_resource_ids JSON NOT NULL DEFAULT '{}',
    
    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- Optimistic locking
    version INTEGER NOT NULL DEFAULT 1,
    
    -- Organization
    labels JSON NOT NULL DEFAULT '{}',
    annotations JSON NOT NULL DEFAULT '{}',
    
    -- Constraints
    CHECK (version >= 1),
    CHECK (length(name) >= 1),
    CHECK (status IN ('requested', 'planning', 'provisioning', 'ready', 'updating', 'deleting', 'archived', 'failed'))
);

-- Indexes for common queries
CREATE INDEX idx_tenants_status ON tenants(status);
CREATE INDEX idx_tenants_created_at ON tenants(created_at DESC);

-- Create tenant state history table for audit trail
CREATE TABLE tenant_state_history (
    -- Identity
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    
    -- Transition details
    from_status VARCHAR(20) NOT NULL,
    to_status VARCHAR(20) NOT NULL,
    
    -- Context
    reason TEXT,
    triggered_by VARCHAR(255),
    
    -- State snapshots
    desired_state_snapshot JSON,
    observed_state_snapshot JSON,
    
    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CHECK (from_status IN ('requested', 'planning', 'provisioning', 'ready', 'updating', 'deleting', 'archived', 'failed')),
    CHECK (to_status IN ('requested', 'planning', 'provisioning', 'ready', 'updating', 'deleting', 'archived', 'failed'))
);

-- Indexes for history queries
CREATE INDEX idx_tenant_state_history_tenant_id ON tenant_state_history(tenant_id);
CREATE INDEX idx_tenant_state_history_created_at ON tenant_state_history(created_at DESC);
CREATE INDEX idx_tenant_state_history_to_status ON tenant_state_history(to_status);
