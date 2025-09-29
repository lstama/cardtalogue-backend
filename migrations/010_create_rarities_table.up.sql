
CREATE TABLE rarities (
    id SERIAL PRIMARY KEY,
    code VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

INSERT INTO rarities (code, name) VALUES 
    ('N','Normal'),
    ('NP', 'Normal Parallel'),
    ('C','Common'),
    ('U','Uncommon'),
    ('UC','Uncommon'),
    ('CC','Climax Common'),
    ('R','Rare'),
    ('RR','Double Rare'),
    ('RRR','Triple Rare'),
    ('CR', 'Climax Rare'),
    ('SR','Super Rare'),
    ('SP','Special Rare'),
    ('SP','Special'),
    ('M', 'Mythic Rare'),
    ('UR','Ultra Rare'),
    ('UL','Ultimate Rare'),
    ('SCR','Secret Rare'),
    ('SE','Secret Rare'),
    ('SEC','Secret Rare'),
    ('PSCR','Prismatic Secret Rare'),
    ('PSE','Prismatic Secret Rare'),
    ('QCSR','Quarter Century Secret Rare'),
    ('QCSE','Quarter Century Secret Rare'),
    ('L','Legendary'),
    ('L', 'Leader'),
    ('TD', 'Trial Deck'),
    ('P', 'Promo'),
    ('PR', 'Promo');

-- Add rarity_id column to cards table
ALTER TABLE cards ADD COLUMN rarity_id INTEGER;
-- Set default rarity_id to 1 for existing cards
UPDATE cards SET rarity_id = 1 WHERE rarity_id IS NULL;
-- Make rarity_id NOT NULL and add foreign key constraint
ALTER TABLE cards ALTER COLUMN rarity_id SET NOT NULL;
ALTER TABLE cards ADD CONSTRAINT fk_cards_rarity_id FOREIGN KEY (rarity_id) REFERENCES rarities(id);
-- Add index for better performance
CREATE INDEX idx_cards_rarity_id ON cards(rarity_id);
