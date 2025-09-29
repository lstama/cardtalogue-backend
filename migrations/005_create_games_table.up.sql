CREATE TABLE games (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert some default games
INSERT INTO games (name) VALUES 
    ('Yu-Gi-Oh!'),
    ('Magic: The Gathering'),
    ('Pokémon'),
    ('Digimon'),
    ('Cardfight!! Vanguard'),
    ('Weiß Schwarz')
ON CONFLICT (name) DO NOTHING;
