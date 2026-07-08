CREATE TABLE cards (
    id BIGINT PRIMARY KEY,
    code VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    price INT NOT NULL DEFAULT 100,
    price_multiplier INT NOT NULL DEFAULT 120,
    quantity INT NOT NULL DEFAULT 1,
    image_url VARCHAR(255) NOT NULL DEFAULT 'https://picsum.photos/230/330.webp',
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'enabled',
    for_sale VARCHAR(20) NOT NULL DEFAULT 'yes',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
