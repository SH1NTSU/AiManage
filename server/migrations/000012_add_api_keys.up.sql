-- Add API key field to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS api_key VARCHAR(255);

-- Create unique index on api_key
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_api_key ON users(api_key);

-- Generate initial API keys for existing users (you can regenerate these later)
UPDATE users SET api_key = 'sk_live_' || substr(md5(random()::text || email), 1, 24) WHERE api_key IS NULL;
