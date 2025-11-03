-- Drop triggers
DROP TRIGGER IF EXISTS models_insert_trigger ON models;
DROP TRIGGER IF EXISTS models_update_trigger ON models;
DROP TRIGGER IF EXISTS models_delete_trigger ON models;

-- Drop function
DROP FUNCTION IF EXISTS notify_models_change();
