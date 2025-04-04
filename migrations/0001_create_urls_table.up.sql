CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS urls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    short_url TEXT NOT NULL UNIQUE,
    original_url TEXT NOT NULL UNIQUE
);
