-- Add lowercase_code column to storages table
ALTER TABLE storages ADD COLUMN lowercase_code VARCHAR(255);
-- Update existing rows to set the new column
UPDATE storages SET lowercase_code = LOWER(code);
-- Create index on the new column for faster search
CREATE INDEX idx_storages_lowercase_code ON storages(lowercase_code);

INSERT INTO storages
(code, "name", description, created_at, updated_at, lowercase_name, lowercase_code)
VALUES('DF', 'Default', 'Default', '2025-09-28 08:57:38.030', '2025-09-28 08:57:38.030', 'default', 'df');
