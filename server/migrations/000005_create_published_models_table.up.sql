-- Create published_models table for community marketplace
CREATE TABLE published_models (
    id SERIAL PRIMARY KEY,

    -- Reference to original model
    model_id INTEGER NOT NULL REFERENCES models(id) ON DELETE CASCADE,
    publisher_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Model information (copied from models table)
    name VARCHAR(255) NOT NULL,
    picture TEXT,
    trained_model_path VARCHAR(500) NOT NULL, -- Path to the trained model file
    training_script VARCHAR(255),

    -- Marketplace-specific fields
    description TEXT NOT NULL, -- Detailed description of the model
    short_description VARCHAR(500), -- Short tagline/summary
    price INTEGER NOT NULL DEFAULT 0, -- Price in cents (0 = free)
    category VARCHAR(100), -- e.g., "Image Classification", "NLP", "Computer Vision"
    tags TEXT[], -- Array of tags for filtering

    -- Model metadata
    model_type VARCHAR(100), -- e.g., "CNN", "Transformer", "ResNet"
    framework VARCHAR(50), -- e.g., "pytorch", "tensorflow", "keras"
    file_size BIGINT, -- Size in bytes
    accuracy_score DECIMAL(5, 2), -- Model accuracy if available (e.g., 95.50)

    -- Licensing
    license_type VARCHAR(100) DEFAULT 'personal_use', -- e.g., "personal_use", "commercial", "mit", "apache"

    -- Statistics
    downloads_count INTEGER NOT NULL DEFAULT 0,
    views_count INTEGER NOT NULL DEFAULT 0,
    rating_average DECIMAL(3, 2) DEFAULT 0.00, -- Average rating (0.00 - 5.00)
    rating_count INTEGER NOT NULL DEFAULT 0,

    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true, -- Can be unpublished
    is_featured BOOLEAN NOT NULL DEFAULT false, -- Featured on homepage

    -- Timestamps
    published_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Constraints
    CONSTRAINT price_non_negative CHECK (price >= 0),
    CONSTRAINT rating_valid_range CHECK (rating_average >= 0 AND rating_average <= 5),
    CONSTRAINT downloads_non_negative CHECK (downloads_count >= 0),
    CONSTRAINT views_non_negative CHECK (views_count >= 0)
);

-- Create indexes for better query performance
CREATE INDEX idx_published_models_publisher_id ON published_models(publisher_id);
CREATE INDEX idx_published_models_model_id ON published_models(model_id);
CREATE INDEX idx_published_models_category ON published_models(category);
CREATE INDEX idx_published_models_price ON published_models(price);
CREATE INDEX idx_published_models_is_active ON published_models(is_active);
CREATE INDEX idx_published_models_published_at ON published_models(published_at DESC);
CREATE INDEX idx_published_models_downloads_count ON published_models(downloads_count DESC);
CREATE INDEX idx_published_models_rating_average ON published_models(rating_average DESC);

-- GIN index for tags array (for fast tag searches)
CREATE INDEX idx_published_models_tags ON published_models USING GIN(tags);

-- Composite index for common queries
CREATE INDEX idx_published_models_active_category ON published_models(is_active, category);

-- Add updated_at trigger
CREATE TRIGGER update_published_models_updated_at
    BEFORE UPDATE ON published_models
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comment for documentation
COMMENT ON TABLE published_models IS 'Community marketplace for published AI models';
COMMENT ON COLUMN published_models.price IS 'Price in cents (0 means free)';
COMMENT ON COLUMN published_models.rating_average IS 'Average rating from 0.00 to 5.00';
