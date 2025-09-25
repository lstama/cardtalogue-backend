
CREATE TABLE rarities (
    id SERIAL PRIMARY KEY,
    code VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

INSERT INTO rarities (code, name) VALUES 
    ('N','Normal'),
    ('SR','Super Rare'),
    ('R','Rare'),
    ('UR','Ultra Rare'),
    ('UL','Ultimate Rare'),
    ('C','Common'),
    ('U','Uncommon'),
    ('SCR','Secret Rare'),
    ('L','Legendary')
ON CONFLICT (code) DO NOTHING;

-- Add rarity_id column to cards table
ALTER TABLE cards ADD COLUMN rarity_id INTEGER;
-- Set default rarity_id to 1 for existing cards
UPDATE cards SET rarity_id = 1 WHERE rarity_id IS NULL;
-- Make rarity_id NOT NULL and add foreign key constraint
ALTER TABLE cards ALTER COLUMN rarity_id SET NOT NULL;
ALTER TABLE cards ADD CONSTRAINT fk_cards_rarity_id FOREIGN KEY (rarity_id) REFERENCES rarities(id);
-- Add index for better performance
CREATE INDEX idx_cards_rarity_id ON cards(rarity_id);
