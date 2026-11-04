-- storages
INSERT INTO storages (code, name, description)
VALUES ('BOX1', 'Main Box', 'Primary card storage box');

-- packs
INSERT INTO packs
(code, "name", image_url, description, game_id, language_id, lowercase_name, lowercase_code)
VALUES('TAMA', 'Tactical Master', 'https://example.com/packs/tama-booster-en.png', 'Runick, Labrynth, Vaylantz', 1, 2, 'tactical master', 'tama');
INSERT INTO packs
(code, "name", image_url, description, game_id, language_id, lowercase_name, lowercase_code)
VALUES('LEDE', 'Legacy of Destruction', 'https://example.com/packs/lede-booster-jp.png', 'Ragnaraika, Tenpai', 1, 1, 'legacy of destruction', 'lede');
