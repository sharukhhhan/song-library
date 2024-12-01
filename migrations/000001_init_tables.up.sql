CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS groups (
                        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                        name VARCHAR(255) NOT NULL UNIQUE,
                        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS songs (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       title TEXT NOT NULL,
                       group_id UUID REFERENCES groups(id) ON DELETE CASCADE,
                       release_date DATE,
                       link TEXT,
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                       updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                        UNIQUE(title, group_id)
);

CREATE TABLE IF NOT EXISTS lyrics_verses (
                        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                        song_id UUID REFERENCES songs(id) ON DELETE CASCADE,
                        verse TEXT NOT NULL,
                        verse_number INTEGER NOT NULL,
                        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
