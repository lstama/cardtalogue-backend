-- storages
INSERT INTO storages (code, name, description)
VALUES ('BOX1', 'Main Box', 'Primary card storage box');

-- packs
INSERT INTO packs
(code, "name", image_url, description, game_id, language_id, lowercase_name, lowercase_code)
VALUES('TAMA', 'Tactical Master', 'https://static.wikia.nocookie.net/yugioh/images/0/0a/TAMA-BoosterEN.png', 'Runick, Labrynth, Vaylantz', 1, 2, 'tactical master', 'tama');
INSERT INTO packs
(code, "name", image_url, description, game_id, language_id, lowercase_name, lowercase_code)
VALUES('LEDE', 'Legacy of Destruction', 'https://static.wikia.nocookie.net/yugioh/images/8/87/LEDE-BoosterJP.png', 'Ragnaraika, Tenpai', 1, 1, 'legacy of destruction', 'lede');
