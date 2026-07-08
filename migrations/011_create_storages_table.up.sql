
CREATE TABLE storages (
    id BIGINT PRIMARY KEY,
    code VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add storage_id column to cards table
ALTER TABLE cards ADD COLUMN storage_id INTEGER;
-- Add foreign key constraint
ALTER TABLE cards ADD CONSTRAINT fk_cards_storage_id FOREIGN KEY (storage_id) REFERENCES storages(id);
-- Add index for better performance
CREATE INDEX idx_cards_storage_id ON cards(storage_id);
