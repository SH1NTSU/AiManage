-- Check if api_key column exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'users'
        AND column_name = 'api_key'
    ) THEN
        -- Add api_key column if it doesn't exist
        ALTER TABLE users ADD COLUMN api_key VARCHAR(255);
        CREATE UNIQUE INDEX idx_users_api_key ON users(api_key);
        RAISE NOTICE 'api_key column added successfully';
    ELSE
        RAISE NOTICE 'api_key column already exists';
    END IF;
END $$;

-- Generate API keys for users that don't have one
UPDATE users
SET api_key = 'sk_live_' || substr(md5(random()::text || email), 1, 24)
WHERE api_key IS NULL OR api_key = '';

-- Show all users with their API keys
SELECT
    id,
    email,
    username,
    SUBSTRING(api_key, 1, 12) || '...' as api_key_preview,
    CASE
        WHEN api_key IS NULL THEN '❌ NULL'
        WHEN api_key = '' THEN '❌ EMPTY'
        ELSE '✅ SET'
    END as status
FROM users
ORDER BY id;
