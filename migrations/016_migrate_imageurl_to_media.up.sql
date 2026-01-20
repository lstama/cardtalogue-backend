-- Migrate existing card imageURLs to media entries
INSERT INTO media (url, status, provider, created_at, updated_at)
SELECT DISTINCT image_url, 'attached', 'hardlink', NOW(), NOW()
FROM cards
WHERE image_url IS NOT NULL AND image_url != '';

-- Update cards to reference the newly created media
UPDATE cards c
SET media_id = (
    SELECT id FROM media m
    WHERE m.url = c.image_url AND m.provider = 'hardlink'
    LIMIT 1
)
WHERE c.media_id IS NULL AND c.image_url IS NOT NULL AND c.image_url != '';
