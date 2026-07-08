CREATE TABLE languages (
    id BIGINT PRIMARY KEY,
    code VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert some default languages
INSERT INTO languages (code, name) VALUES 
    ('JP','Japanese'),
    ('EN','English'),
    ('AE','Asian English'),
    ('ID', 'Bahasa Indonesia');
