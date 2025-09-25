-- Remove foreign key constraint and index
ALTER TABLE cards DROP CONSTRAINT IF EXISTS fk_cards_pack_id;
DROP INDEX IF EXISTS idx_cards_pack_id;

-- Remove language_id column
ALTER TABLE cards DROP COLUMN IF EXISTS pack_id;

-- Remove foreign key constraints and indexes
ALTER TABLE packs DROP CONSTRAINT IF EXISTS fk_packs_game_id;
DROP INDEX IF EXISTS idx_packs_game_id;
-- Remove foreign key constraints and indexes
ALTER TABLE packs DROP CONSTRAINT IF EXISTS fk_packs_language_id;
DROP INDEX IF EXISTS idx_packs_language_id;

DROP TABLE IF EXISTS packs;
