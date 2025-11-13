-- Rollback subscription fields
DROP INDEX IF EXISTS idx_users_subscription;
DROP INDEX IF EXISTS idx_users_stripe_customer;

ALTER TABLE users DROP COLUMN IF EXISTS subscription_tier;
ALTER TABLE users DROP COLUMN IF EXISTS subscription_status;
ALTER TABLE users DROP COLUMN IF EXISTS stripe_customer_id;
ALTER TABLE users DROP COLUMN IF EXISTS stripe_subscription_id;
ALTER TABLE users DROP COLUMN IF EXISTS subscription_start_date;
ALTER TABLE users DROP COLUMN IF EXISTS subscription_end_date;
ALTER TABLE users DROP COLUMN IF EXISTS training_credits;
