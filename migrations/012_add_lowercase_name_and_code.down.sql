-- Drop indexes and columns from cards table
DROP INDEX IF EXISTS idx_cards_lowercase_name;
DROP INDEX IF EXISTS idx_cards_lowercase_code;
ALTER TABLE cards DROP COLUMN IF EXISTS lowercase_name;
ALTER TABLE cards DROP COLUMN IF EXISTS lowercase_code;

-- Drop index and column from games table
DROP INDEX IF EXISTS idx_games_lowercase_name;
ALTER TABLE games DROP COLUMN IF EXISTS lowercase_name;

-- Drop index and column from storages table
DROP INDEX IF EXISTS idx_storages_lowercase_name;
ALTER TABLE storages DROP COLUMN IF EXISTS lowercase_name;

-- Drop indexes and columns from languages table
DROP INDEX IF EXISTS idx_languages_lowercase_name;
DROP INDEX IF EXISTS idx_languages_lowercase_code;
ALTER TABLE languages DROP COLUMN IF EXISTS lowercase_name;
ALTER TABLE languages DROP COLUMN IF EXISTS lowercase_code;

-- Drop indexes and columns from packs table
DROP INDEX IF EXISTS idx_packs_lowercase_name;
DROP INDEX IF EXISTS idx_packs_lowercase_code;
ALTER TABLE packs DROP COLUMN IF EXISTS lowercase_name;
ALTER TABLE packs DROP COLUMN IF EXISTS lowercase_code;

-- Drop indexes and columns from rarities table
DROP INDEX IF EXISTS idx_rarities_lowercase_name;
DROP INDEX IF EXISTS idx_rarities_lowercase_code;
ALTER TABLE rarities DROP COLUMN IF EXISTS lowercase_name;
ALTER TABLE rarities DROP COLUMN IF EXISTS lowercase_code;
