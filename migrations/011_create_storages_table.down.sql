-- Remove foreign key constraint and index
ALTER TABLE cards DROP CONSTRAINT IF EXISTS fk_cards_storage_id;
DROP INDEX IF EXISTS idx_cards_storage_id;

-- Remove pack_id column
ALTER TABLE cards DROP COLUMN IF EXISTS storage_id;

DROP TABLE IF EXISTS storages;
