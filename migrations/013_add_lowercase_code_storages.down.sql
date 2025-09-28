-- Drop index and column from storages table
DROP INDEX IF EXISTS idx_storages_lowercase_code;
ALTER TABLE storages DROP COLUMN IF EXISTS lowercase_code;
