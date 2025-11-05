-- Drop index
DROP INDEX IF EXISTS idx_users_username;

-- Drop username column
ALTER TABLE users DROP COLUMN IF EXISTS username;
