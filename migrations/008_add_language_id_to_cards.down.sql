-- Remove foreign key constraint and index
ALTER TABLE cards DROP CONSTRAINT IF EXISTS fk_cards_language_id;
DROP INDEX IF EXISTS idx_cards_language_id;

-- Remove language_id column
ALTER TABLE cards DROP COLUMN IF EXISTS language_id;
