-- Create model_purchases table to track purchases and downloads
CREATE TABLE model_purchases (
    id SERIAL PRIMARY KEY,

    -- References
    published_model_id INTEGER NOT NULL REFERENCES published_models(id) ON DELETE CASCADE,
    buyer_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    publisher_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Purchase details
    price_paid INTEGER NOT NULL, -- Price paid in cents (snapshot at time of purchase)
    is_free BOOLEAN NOT NULL DEFAULT false, -- True if model was free

    -- Payment information (for future payment integration)
    payment_status VARCHAR(50) NOT NULL DEFAULT 'completed', -- 'pending', 'completed', 'failed', 'refunded'
    payment_method VARCHAR(50), -- e.g., 'stripe', 'paypal', 'free'
    transaction_id VARCHAR(255), -- External payment provider transaction ID

    -- Download tracking
    download_count INTEGER NOT NULL DEFAULT 0, -- How many times user downloaded this model
    last_downloaded_at TIMESTAMP,

    -- Timestamps
    purchased_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Constraints
    CONSTRAINT unique_user_model_purchase UNIQUE(buyer_id, published_model_id), -- User can only buy once
    CONSTRAINT price_paid_non_negative CHECK (price_paid >= 0),
    CONSTRAINT download_count_non_negative CHECK (download_count >= 0)
);

-- Create indexes for better query performance
CREATE INDEX idx_model_purchases_published_model_id ON model_purchases(published_model_id);
CREATE INDEX idx_model_purchases_buyer_id ON model_purchases(buyer_id);
CREATE INDEX idx_model_purchases_publisher_id ON model_purchases(publisher_id);
CREATE INDEX idx_model_purchases_purchased_at ON model_purchases(purchased_at DESC);
CREATE INDEX idx_model_purchases_payment_status ON model_purchases(payment_status);

-- Composite index for checking if user owns model
CREATE INDEX idx_model_purchases_buyer_model ON model_purchases(buyer_id, published_model_id);

-- Add comment for documentation
COMMENT ON TABLE model_purchases IS 'Tracks model purchases and downloads in the community marketplace';
COMMENT ON COLUMN model_purchases.price_paid IS 'Price paid in cents (snapshot at purchase time)';
COMMENT ON COLUMN model_purchases.download_count IS 'Number of times user has downloaded this model';
