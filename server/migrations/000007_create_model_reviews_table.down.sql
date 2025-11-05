-- Drop triggers
DROP TRIGGER IF EXISTS model_review_delete_trigger ON model_reviews;
DROP TRIGGER IF EXISTS model_review_update_trigger ON model_reviews;
DROP TRIGGER IF EXISTS model_review_insert_trigger ON model_reviews;

-- Drop function
DROP FUNCTION IF EXISTS update_published_model_rating();

-- Drop indexes
DROP INDEX IF EXISTS idx_model_reviews_verified_rating;
DROP INDEX IF EXISTS idx_model_reviews_helpful_count;
DROP INDEX IF EXISTS idx_model_reviews_created_at;
DROP INDEX IF EXISTS idx_model_reviews_rating;
DROP INDEX IF EXISTS idx_model_reviews_reviewer_id;
DROP INDEX IF EXISTS idx_model_reviews_published_model_id;

-- Drop table
DROP TABLE IF EXISTS model_reviews;
