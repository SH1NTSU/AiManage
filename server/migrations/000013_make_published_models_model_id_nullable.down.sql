-- Revert model_id to NOT NULL (this may fail if there are NULL values)
-- Note: This migration will fail if there are existing NULL model_id values
-- You may need to clean up NULL values first or set them to a default model
ALTER TABLE published_models 
ALTER COLUMN model_id SET NOT NULL;

