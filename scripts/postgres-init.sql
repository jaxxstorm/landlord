-- PostgreSQL initialization script for Landlord
-- This script is automatically executed when the PostgreSQL container starts

-- Create extensions required by Landlord
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Set up basic permissions
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO landlord;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO landlord;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT EXECUTE ON FUNCTIONS TO landlord;

-- Log initialization
SELECT 'PostgreSQL initialization complete for Landlord' AS status;
