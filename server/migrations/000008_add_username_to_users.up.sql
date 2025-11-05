-- Add username column to users table
ALTER TABLE users ADD COLUMN username VARCHAR(50) UNIQUE;

-- Add index on username for fast lookups
CREATE INDEX idx_users_username ON users(username);

-- Add comment for documentation
COMMENT ON COLUMN users.username IS 'Unique username for the user, displayed as creator name';
