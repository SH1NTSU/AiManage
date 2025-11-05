-- Drop indexes
DROP INDEX IF EXISTS idx_model_purchases_buyer_model;
DROP INDEX IF EXISTS idx_model_purchases_payment_status;
DROP INDEX IF EXISTS idx_model_purchases_purchased_at;
DROP INDEX IF EXISTS idx_model_purchases_publisher_id;
DROP INDEX IF EXISTS idx_model_purchases_buyer_id;
DROP INDEX IF EXISTS idx_model_purchases_published_model_id;

-- Drop table
DROP TABLE IF EXISTS model_purchases;
