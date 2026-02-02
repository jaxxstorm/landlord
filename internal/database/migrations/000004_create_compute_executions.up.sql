-- Create compute_executions table to track compute provisioning operations
CREATE TABLE compute_executions (
  id SERIAL PRIMARY KEY,
  execution_id VARCHAR(255) UNIQUE NOT NULL,
  tenant_id UUID NOT NULL,
  workflow_execution_id VARCHAR(255) NOT NULL,
  operation_type VARCHAR(50) NOT NULL,
  status VARCHAR(50) NOT NULL,
  resource_ids JSONB,
  error_code VARCHAR(100),
  error_message TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Create indexes for efficient querying
CREATE INDEX idx_compute_executions_execution_id ON compute_executions(execution_id);
CREATE INDEX idx_compute_executions_tenant_id ON compute_executions(tenant_id);
CREATE INDEX idx_compute_executions_status ON compute_executions(status);
CREATE INDEX idx_compute_executions_workflow_execution_id ON compute_executions(workflow_execution_id);
CREATE INDEX idx_compute_executions_tenant_status ON compute_executions(tenant_id, status);
