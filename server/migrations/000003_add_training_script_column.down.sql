-- Remove training_script column from models table
ALTER TABLE models
DROP COLUMN IF EXISTS training_script;
