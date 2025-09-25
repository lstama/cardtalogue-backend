-- Add language_id column to cards table
ALTER TABLE cards ADD COLUMN language_id INTEGER;

-- Set default language_id to 1 for existing cards
UPDATE cards SET language_id = 1 WHERE language_id IS NULL;

-- Make language_id NOT NULL and add foreign key constraint
ALTER TABLE cards ALTER COLUMN language_id SET NOT NULL;
ALTER TABLE cards ADD CONSTRAINT fk_cards_language_id FOREIGN KEY (language_id) REFERENCES languages(id);

-- Add index for better performance
CREATE INDEX idx_cards_language_id ON cards(language_id);
