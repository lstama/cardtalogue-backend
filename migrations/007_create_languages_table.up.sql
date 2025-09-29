CREATE TABLE languages (
    id SERIAL PRIMARY KEY,
    code VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert some default languages
INSERT INTO languages (code, name) VALUES 
    ('JP','Japanese'),
    ('EN','English'),
    ('AE','Asian English'),
    ('ID', 'Bahasa Indonesia')
ON CONFLICT (code) DO NOTHING;
