-- Create model_reviews table for ratings and reviews
CREATE TABLE model_reviews (
    id SERIAL PRIMARY KEY,

    -- References
    published_model_id INTEGER NOT NULL REFERENCES published_models(id) ON DELETE CASCADE,
    reviewer_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Review content
    rating INTEGER NOT NULL, -- 1-5 stars
    title VARCHAR(200), -- Short review title
    comment TEXT, -- Detailed review

    -- Review metadata
    is_verified_purchase BOOLEAN NOT NULL DEFAULT false, -- Did user actually buy/download?
    helpful_count INTEGER NOT NULL DEFAULT 0, -- How many found this review helpful

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Constraints
    CONSTRAINT unique_user_review UNIQUE(reviewer_id, published_model_id), -- One review per user per model
    CONSTRAINT rating_valid_range CHECK (rating >= 1 AND rating <= 5),
    CONSTRAINT helpful_count_non_negative CHECK (helpful_count >= 0)
);

-- Create indexes for better query performance
CREATE INDEX idx_model_reviews_published_model_id ON model_reviews(published_model_id);
CREATE INDEX idx_model_reviews_reviewer_id ON model_reviews(reviewer_id);
CREATE INDEX idx_model_reviews_rating ON model_reviews(rating);
CREATE INDEX idx_model_reviews_created_at ON model_reviews(created_at DESC);
CREATE INDEX idx_model_reviews_helpful_count ON model_reviews(helpful_count DESC);

-- Composite index for verified purchases with high ratings
CREATE INDEX idx_model_reviews_verified_rating ON model_reviews(is_verified_purchase, rating DESC);

-- Add updated_at trigger
CREATE TRIGGER update_model_reviews_updated_at
    BEFORE UPDATE ON model_reviews
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Function to update published_models rating when review is added/updated/deleted
CREATE OR REPLACE FUNCTION update_published_model_rating()
RETURNS TRIGGER AS $$
BEGIN
    -- Recalculate average rating for the published model
    UPDATE published_models
    SET
        rating_average = (
            SELECT COALESCE(AVG(rating)::DECIMAL(3,2), 0.00)
            FROM model_reviews
            WHERE published_model_id = COALESCE(NEW.published_model_id, OLD.published_model_id)
        ),
        rating_count = (
            SELECT COUNT(*)
            FROM model_reviews
            WHERE published_model_id = COALESCE(NEW.published_model_id, OLD.published_model_id)
        )
    WHERE id = COALESCE(NEW.published_model_id, OLD.published_model_id);

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Create triggers to auto-update ratings
CREATE TRIGGER model_review_insert_trigger
    AFTER INSERT ON model_reviews
    FOR EACH ROW
    EXECUTE FUNCTION update_published_model_rating();

CREATE TRIGGER model_review_update_trigger
    AFTER UPDATE ON model_reviews
    FOR EACH ROW
    EXECUTE FUNCTION update_published_model_rating();

CREATE TRIGGER model_review_delete_trigger
    AFTER DELETE ON model_reviews
    FOR EACH ROW
    EXECUTE FUNCTION update_published_model_rating();

-- Add comment for documentation
COMMENT ON TABLE model_reviews IS 'User reviews and ratings for published models';
COMMENT ON COLUMN model_reviews.rating IS 'Rating from 1 to 5 stars';
COMMENT ON COLUMN model_reviews.is_verified_purchase IS 'True if reviewer purchased/downloaded the model';
