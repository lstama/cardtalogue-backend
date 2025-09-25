-- Remove foreign key constraint and index
ALTER TABLE cards DROP CONSTRAINT IF EXISTS fk_cards_rarity_id;
DROP INDEX IF EXISTS idx_cards_rarity_id;

-- Remove pack_id column
ALTER TABLE cards DROP COLUMN IF EXISTS rarity_id;

DROP TABLE IF EXISTS rarities;
