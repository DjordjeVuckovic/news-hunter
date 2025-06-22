CREATE TABLE IF NOT EXISTS articles
(
    id uuid NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    title text NOT NULL
);