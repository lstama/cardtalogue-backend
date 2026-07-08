CREATE TABLE media (
    id BIGINT PRIMARY KEY,
    url VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    provider VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes for better performance
CREATE INDEX idx_media_status ON media(status);
CREATE INDEX idx_media_provider ON media(provider);
CREATE INDEX idx_media_created_at ON media(created_at);
