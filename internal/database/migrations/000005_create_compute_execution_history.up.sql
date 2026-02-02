-- Create compute_execution_history table for audit trail
CREATE TABLE compute_execution_history (
  id SERIAL PRIMARY KEY,
  compute_execution_id VARCHAR(255) NOT NULL,
  status VARCHAR(50) NOT NULL,
  details JSONB,
  timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (compute_execution_id) REFERENCES compute_executions(execution_id) ON DELETE CASCADE
);

-- Create indexes for efficient querying
CREATE INDEX idx_compute_execution_history_execution_id ON compute_execution_history(compute_execution_id);
CREATE INDEX idx_compute_execution_history_timestamp ON compute_execution_history(timestamp);
