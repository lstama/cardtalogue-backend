-- Remove foreign key constraint and index
ALTER TABLE cards DROP CONSTRAINT IF EXISTS fk_cards_game_id;
DROP INDEX IF EXISTS idx_cards_game_id;

-- Remove game_id column
ALTER TABLE cards DROP COLUMN IF EXISTS game_id;
