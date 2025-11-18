-- Make model_id nullable to support imported models from HuggingFace
-- Imported models don't have a local model entry, so model_id should be NULL
ALTER TABLE published_models 
ALTER COLUMN model_id DROP NOT NULL;

-- Update foreign key constraint to handle NULL values
-- The existing foreign key constraint already allows NULL, so no change needed
-- But we'll add a comment for clarity
COMMENT ON COLUMN published_models.model_id IS 'Reference to local model (NULL for imported models from HuggingFace)';

