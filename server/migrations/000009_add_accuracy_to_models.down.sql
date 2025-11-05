-- Drop index
DROP INDEX IF EXISTS idx_models_accuracy_score;

-- Drop accuracy_score column
ALTER TABLE models DROP COLUMN IF EXISTS accuracy_score;
