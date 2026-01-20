-- Revert media migration by removing newly created hardlink media entries
-- First, detach the cards from media
UPDATE cards SET media_id = NULL WHERE media_id IN (
    SELECT id FROM media WHERE provider = 'hardlink'
);

-- Then delete the hardlink media entries
DELETE FROM media WHERE provider = 'hardlink';
