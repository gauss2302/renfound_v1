-- Drop sessions table first due to foreign key constraints
DROP TABLE IF EXISTS sessions;

-- Drop users table
DROP TABLE IF EXISTS users;

-- Drop extension if needed
-- DROP EXTENSION IF EXISTS "uuid-ossp";