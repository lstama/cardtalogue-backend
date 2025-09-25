-- Add lowercase_name and lowercase_code columns to cards table
ALTER TABLE cards ADD COLUMN lowercase_name VARCHAR(255);
ALTER TABLE cards ADD COLUMN lowercase_code VARCHAR(255);
-- Update existing rows to set the new columns
UPDATE cards SET lowercase_name = LOWER(name), lowercase_code = LOWER(code);
-- Create indexes on the new columns for faster search
CREATE INDEX idx_cards_lowercase_name ON cards(lowercase_name);
CREATE INDEX idx_cards_lowercase_code ON cards(lowercase_code);

-- Add lowercase_name column to games table
ALTER TABLE games ADD COLUMN lowercase_name VARCHAR(255);
-- Update existing rows to set the new column
UPDATE games SET lowercase_name = LOWER(name);
-- Create index on the new column for faster search
CREATE INDEX idx_games_lowercase_name ON games(lowercase_name);

-- Add lowercase_name column to storages table
ALTER TABLE storages ADD COLUMN lowercase_name VARCHAR(255);
-- Update existing rows to set the new column
UPDATE storages SET lowercase_name = LOWER(name);
-- Create index on the new column for faster search
CREATE INDEX idx_storages_lowercase_name ON storages(lowercase_name);

-- Add lowercase_name column to languages table
ALTER TABLE languages ADD COLUMN lowercase_name VARCHAR(255);
ALTER TABLE languages ADD COLUMN lowercase_code VARCHAR(255);
-- Update existing rows to set the new column
UPDATE languages SET lowercase_name = LOWER(name);
UPDATE languages SET lowercase_code = LOWER(code);
-- Create index on the new column for faster search
CREATE INDEX idx_languages_lowercase_name ON languages(lowercase_name);
CREATE INDEX idx_languages_lowercase_code ON languages(lowercase_code);

-- Add lowercase_name column to packs table
ALTER TABLE packs ADD COLUMN lowercase_name VARCHAR(255);
ALTER TABLE packs ADD COLUMN lowercase_code VARCHAR(255);
-- Update existing rows to set the new column
UPDATE packs SET lowercase_name = LOWER(name);
UPDATE packs SET lowercase_code = LOWER(code);
-- Create index on the new column for faster search
CREATE INDEX idx_packs_lowercase_name ON packs(lowercase_name);
CREATE INDEX idx_packs_lowercase_code ON packs(lowercase_code);

-- Add lowercase_name column to rarities table
ALTER TABLE rarities ADD COLUMN lowercase_name VARCHAR(255);
ALTER TABLE rarities ADD COLUMN lowercase_code VARCHAR(255);
-- Update existing rows to set the new column
UPDATE rarities SET lowercase_name = LOWER(name);
UPDATE rarities SET lowercase_code = LOWER(code);
-- Create index on the new column for faster search
CREATE INDEX idx_rarities_lowercase_name ON rarities(lowercase_name);
CREATE INDEX idx_rarities_lowercase_code ON rarities(lowercase_code);
