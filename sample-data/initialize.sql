-- packs
INSERT INTO packs
(code, "name", image_url, description, game_id, language_id, created_at, updated_at, lowercase_name, lowercase_code)
VALUES('TAMA', 'Tactical Master', 'https://static.wikia.nocookie.net/yugioh/images/0/0a/TAMA-BoosterEN.png', 'Runick, Labrynth, Vaylantz', 1, 2, '2025-09-28 07:05:44.834', '2025-09-28 07:05:44.834', 'tactical master', 'tama');
INSERT INTO packs
(code, "name", image_url, description, game_id, language_id, created_at, updated_at, lowercase_name, lowercase_code)
VALUES('LEDE', 'Legacy of Destruction', 'https://static.wikia.nocookie.net/yugioh/images/8/87/LEDE-BoosterJP.png', 'Ragnaraika, Tenpai', 1, 1, '2025-09-28 07:05:44.834', '2025-09-28 07:05:44.834', 'legacy of destruction', 'lede');


-- cards
INSERT INTO cards
(code, "name", price, price_multiplier, quantity, image_url, description, status, for_sale, created_at, updated_at, deleted_at, game_id, language_id, pack_id, rarity_id, storage_id, lowercase_name, lowercase_code)
VALUES('TAMA-EN014', 'Lovely Labrynth of the Silver Castle', 100, 120, 1, 'https://static.wikia.nocookie.net/yugioh/images/2/2a/LovelyLabrynthoftheSilverCastle-TAMA-EN-1E-OP.png', 'Dark/Fiend/Effect/8*/2900/1900', 'enabled', 'no', '2025-09-28 07:10:22.362', '2025-09-28 07:10:22.362', NULL, 1, 2, 1, 4, 1, 'lovelylabrynthofthesilvercastle', 'tama-en014');
INSERT INTO cards
(code, "name", price, price_multiplier, quantity, image_url, description, status, for_sale, created_at, updated_at, deleted_at, game_id, language_id, pack_id, rarity_id, storage_id, lowercase_name, lowercase_code)
VALUES('LEDE-JP022', 'Selettrice Vaalmonica', 100, 120, 1, 'https://static.wikia.nocookie.net/yugioh/images/e/e8/SelettriceVaalmonica-LEDE-JP-OP.png', 'Water/Spellcaster/Effect/4*/1200/1200', 'enabled', 'no', '2025-09-28 07:10:22.362', '2025-09-28 07:10:22.362', NULL, 1, 1, 2, 2, 1, 'selettricevaalmonica', 'lede-jp022');
INSERT INTO cards
(code, "name", price, price_multiplier, quantity, image_url, description, status, for_sale, created_at, updated_at, deleted_at, game_id, language_id, pack_id, rarity_id, storage_id, lowercase_name, lowercase_code)
VALUES('TAMA-EN037', 'Hugin the Runick Wings', 100, 120, 2, 'https://static.wikia.nocookie.net/yugioh/images/9/9f/HugintheRunickWings-TAMA-EN-UR-1E.png', 'Light/Fairy/Fusion/Effect/2*/0/0', 'enabled', 'yes', '2025-09-28 07:10:22.362', '2025-09-28 07:10:22.362', NULL, 1, 2, 1, 4, 1, 'hugintherunickwings', 'tama-en037');
INSERT INTO cards
(code, "name", price, price_multiplier, quantity, image_url, description, status, for_sale, created_at, updated_at, deleted_at, game_id, language_id, pack_id, rarity_id, storage_id, lowercase_name, lowercase_code)
VALUES('LEDE-JP016', 'Tenpai Dragon Paidra', 100, 120, 3, 'https://static.wikia.nocookie.net/yugioh/images/2/2c/TenpaiDragonPaidra-LEDE-EN-SR-1E.png', 'Fire/Dragon/Effect/3*/1700/1000', 'enabled', 'yes', '2025-09-28 07:10:22.362', '2025-09-28 07:10:22.362', NULL, 1, 1, 2, 2, 1, 'tenpaidragonpaidra', 'lede-jp016');
