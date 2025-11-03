-- Create a function that sends a notification when models table changes
CREATE OR REPLACE FUNCTION notify_models_change()
RETURNS TRIGGER AS $$
BEGIN
    -- Send notification on the 'models_changes' channel
    -- Include the operation type and row data
    PERFORM pg_notify(
        'models_changes',
        json_build_object(
            'operation', TG_OP,
            'table', TG_TABLE_NAME,
            'data', CASE
                WHEN TG_OP = 'DELETE' THEN row_to_json(OLD)
                ELSE row_to_json(NEW)
            END
        )::text
    );

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Create triggers for INSERT, UPDATE, and DELETE on models table
CREATE TRIGGER models_insert_trigger
    AFTER INSERT ON models
    FOR EACH ROW
    EXECUTE FUNCTION notify_models_change();

CREATE TRIGGER models_update_trigger
    AFTER UPDATE ON models
    FOR EACH ROW
    EXECUTE FUNCTION notify_models_change();

CREATE TRIGGER models_delete_trigger
    AFTER DELETE ON models
    FOR EACH ROW
    EXECUTE FUNCTION notify_models_change();

-- Add comment for documentation
COMMENT ON FUNCTION notify_models_change() IS 'Sends PostgreSQL notification when models table is modified';
