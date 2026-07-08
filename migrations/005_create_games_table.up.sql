CREATE TABLE games (
    id BIGINT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert some default games
INSERT INTO games (name) VALUES 
    ('Yu-Gi-Oh!'),
    ('Magic: The Gathering'),
    ('Pokémon'),
    ('Digimon'),
    ('Cardfight!! Vanguard'),
    ('Weiß Schwarz');
