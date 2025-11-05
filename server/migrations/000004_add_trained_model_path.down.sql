-- Remove trained model tracking columns
DROP INDEX IF EXISTS idx_models_trained_at;

ALTER TABLE models
DROP COLUMN IF EXISTS trained_model_path,
DROP COLUMN IF EXISTS trained_at;
