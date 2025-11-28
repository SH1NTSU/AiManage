-- Create model_views table to track unique views per user
CREATE TABLE model_views (
    id SERIAL PRIMARY KEY,
    model_id INTEGER NOT NULL REFERENCES published_models(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE, -- NULL for anonymous users
    ip_address VARCHAR(45), -- For tracking anonymous users
    viewed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Prevent duplicate views from the same user
    CONSTRAINT unique_user_view UNIQUE(model_id, user_id),
    CONSTRAINT unique_anonymous_view UNIQUE(model_id, ip_address) WHERE user_id IS NULL
);

-- Create indexes for better query performance
CREATE INDEX idx_model_views_model_id ON model_views(model_id);
CREATE INDEX idx_model_views_user_id ON model_views(user_id);
CREATE INDEX idx_model_views_viewed_at ON model_views(viewed_at DESC);

-- Add comment for documentation
COMMENT ON TABLE model_views IS 'Tracks unique views per user for published models';
COMMENT ON COLUMN model_views.ip_address IS 'IP address for tracking anonymous users when user_id is NULL';
