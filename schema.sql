CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    country TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS flags (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    rule TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_users_country ON users(country);
