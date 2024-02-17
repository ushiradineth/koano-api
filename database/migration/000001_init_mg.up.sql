CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS "user"
(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    name TEXT,
    email VARCHAR(255),
    password VARCHAR(255),

    UNIQUE (email)
);

CREATE TABLE IF NOT EXISTS event (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES "user"(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    title TEXT,
    start TIMESTAMP,
    "end" TIMESTAMP,
    tz TEXT,
    repeated TEXT
);