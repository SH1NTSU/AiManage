-- Create model_likes table for simple like/unlike functionality
CREATE TABLE model_likes (
    id SERIAL PRIMARY KEY,

    -- References
    published_model_id INTEGER NOT NULL REFERENCES published_models(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Constraints - one like per user per model
    CONSTRAINT unique_user_like UNIQUE(user_id, published_model_id)
);

-- Create indexes
CREATE INDEX idx_model_likes_published_model_id ON model_likes(published_model_id);
CREATE INDEX idx_model_likes_user_id ON model_likes(user_id);
CREATE INDEX idx_model_likes_created_at ON model_likes(created_at DESC);

-- Create model_comments table for simple comments (separate from reviews)
CREATE TABLE model_comments (
    id SERIAL PRIMARY KEY,

    -- References
    published_model_id INTEGER NOT NULL REFERENCES published_models(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_comment_id INTEGER REFERENCES model_comments(id) ON DELETE CASCADE, -- For nested replies

    -- Content
    comment_text TEXT NOT NULL,

    -- Metadata
    edited BOOLEAN NOT NULL DEFAULT false,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Constraints
    CONSTRAINT comment_text_not_empty CHECK (LENGTH(TRIM(comment_text)) > 0),
    CONSTRAINT comment_text_max_length CHECK (LENGTH(comment_text) <= 2000)
);

-- Create indexes
CREATE INDEX idx_model_comments_published_model_id ON model_comments(published_model_id);
CREATE INDEX idx_model_comments_user_id ON model_comments(user_id);
CREATE INDEX idx_model_comments_parent_id ON model_comments(parent_comment_id);
CREATE INDEX idx_model_comments_created_at ON model_comments(created_at DESC);

-- Add updated_at trigger for comments
CREATE TRIGGER update_model_comments_updated_at
    BEFORE UPDATE ON model_comments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE model_likes IS 'User likes for published models (simple like/unlike)';
COMMENT ON TABLE model_comments IS 'User comments on published models (separate from reviews)';
COMMENT ON COLUMN model_comments.parent_comment_id IS 'NULL for top-level comments, or ID of parent comment for replies';
