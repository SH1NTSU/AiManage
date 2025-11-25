-- Add email verification columns to users table
ALTER TABLE users
ADD COLUMN email_verified BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN verification_token VARCHAR(255),
ADD COLUMN verification_token_expires_at TIMESTAMP;

-- Create index for verification token lookup
CREATE INDEX idx_users_verification_token ON users(verification_token);
