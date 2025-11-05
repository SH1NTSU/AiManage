-- Drop indexes
DROP INDEX IF EXISTS idx_published_models_active_category;
DROP INDEX IF EXISTS idx_published_models_rating_average;
DROP INDEX IF EXISTS idx_published_models_downloads_count;
DROP INDEX IF EXISTS idx_published_models_published_at;
DROP INDEX IF EXISTS idx_published_models_is_active;
DROP INDEX IF EXISTS idx_published_models_price;
DROP INDEX IF EXISTS idx_published_models_category;
DROP INDEX IF EXISTS idx_published_models_tags;
DROP INDEX IF EXISTS idx_published_models_model_id;
DROP INDEX IF EXISTS idx_published_models_publisher_id;

-- Drop table
DROP TABLE IF EXISTS published_models;
