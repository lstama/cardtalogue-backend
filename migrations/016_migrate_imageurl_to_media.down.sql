-- Revert media migration by removing newly created hardlink media entries
ALTER TABLE cards DROP CONSTRAINT check_image_url_consistency;

DROP INDEX IF EXISTS idx_media_url_unique_non_hardlink;

-- First, detach the cards from media
UPDATE cards SET media_id = NULL WHERE media_id IN (
    SELECT id FROM media WHERE provider = 'hardlink'
);

-- Then delete the hardlink media entries
DELETE FROM media WHERE provider = 'hardlink';
