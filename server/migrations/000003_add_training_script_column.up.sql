-- Add training_script column to models table
ALTER TABLE models
ADD COLUMN IF NOT EXISTS training_script VARCHAR(255) DEFAULT 'train.py';

-- Update existing records to have a default value
UPDATE models
SET training_script = 'train.py'
WHERE training_script IS NULL OR training_script = '';
