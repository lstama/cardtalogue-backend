-- Add game_id column to cards table
ALTER TABLE cards ADD COLUMN game_id INTEGER;

-- Set default game_id to 1 for existing cards
UPDATE cards SET game_id = 1 WHERE game_id IS NULL;

-- Make game_id NOT NULL and add foreign key constraint
ALTER TABLE cards ALTER COLUMN game_id SET NOT NULL;
ALTER TABLE cards ADD CONSTRAINT fk_cards_game_id FOREIGN KEY (game_id) REFERENCES games(id);

-- Add index for better performance
CREATE INDEX idx_cards_game_id ON cards(game_id);
