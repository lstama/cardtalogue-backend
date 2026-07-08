
CREATE TABLE packs (
    id BIGINT PRIMARY KEY,
    code VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL UNIQUE,
    image_url VARCHAR(255) NOT NULL DEFAULT 'https://picsum.photos/230/330.webp',
    description TEXT,
    game_id INTEGER NOT NULL,
    language_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uniquePacks UNIQUE (code, game_id, language_id)
);

-- Add foreign key constraints
ALTER TABLE packs ADD CONSTRAINT fk_packs_game_id FOREIGN KEY (game_id) REFERENCES games(id);
ALTER TABLE packs ADD CONSTRAINT fk_packs_language_id FOREIGN KEY (language_id) REFERENCES languages(id);

-- Add index for better performance
CREATE INDEX idx_packs_game_id ON packs(game_id);
CREATE INDEX idx_packs_language_id ON packs(language_id);

-- Add pack_id column to cards table
ALTER TABLE cards ADD COLUMN pack_id INTEGER;
-- Set default pack_id to NULL for existing cards
UPDATE cards SET pack_id = NULL WHERE pack_id IS NULL;
-- Make pack_id nullable and add foreign key constraint
ALTER TABLE cards ADD CONSTRAINT fk_cards_pack_id FOREIGN KEY (pack_id) REFERENCES packs(id);
-- Add index for better performance
CREATE INDEX idx_cards_pack_id ON cards(pack_id);
