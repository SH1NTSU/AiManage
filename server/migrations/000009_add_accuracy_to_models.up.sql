-- Add accuracy_score column to models table
ALTER TABLE models ADD COLUMN accuracy_score DECIMAL(5, 2);

-- Add index for performance
CREATE INDEX idx_models_accuracy_score ON models(accuracy_score DESC);

-- Add comment for documentation
COMMENT ON COLUMN models.accuracy_score IS 'Model accuracy score from training results (e.g., 95.50)';
