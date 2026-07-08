CREATE INDEX idx_cards_status ON cards(status);
CREATE INDEX idx_cards_name ON cards(name);
CREATE INDEX idx_cards_code ON cards(code);
CREATE INDEX idx_cards_price ON cards(price);
CREATE INDEX idx_cards_price_multiplier ON cards(price_multiplier);
CREATE INDEX idx_cards_quantity ON cards(quantity);
CREATE INDEX idx_cards_image_url ON cards(image_url);
CREATE INDEX idx_cards_for_sale ON cards(for_sale);
CREATE INDEX idx_cards_created_at ON cards(created_at);
CREATE INDEX idx_cards_deleted_at ON cards(deleted_at);

-- Add constraint for status field
ALTER TABLE cards ADD CONSTRAINT chk_cards_status CHECK (status IN ('enabled', 'disabled'));
-- Add constraint for for_sale field
ALTER TABLE cards ADD CONSTRAINT chk_cards_for_sale CHECK (for_sale IN ('yes', 'no'));
