-- Add trained_model_path column to track saved models after training
ALTER TABLE models
ADD COLUMN IF NOT EXISTS trained_model_path VARCHAR(500),
ADD COLUMN IF NOT EXISTS trained_at TIMESTAMP;

-- Create index for faster queries on trained models
CREATE INDEX IF NOT EXISTS idx_models_trained_at ON models(trained_at);
