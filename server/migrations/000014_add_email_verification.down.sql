-- Remove email verification columns from users table
DROP INDEX IF EXISTS idx_users_verification_token;

ALTER TABLE users
DROP COLUMN IF EXISTS email_verified,
DROP COLUMN IF EXISTS verification_token,
DROP COLUMN IF EXISTS verification_token_expires_at;
