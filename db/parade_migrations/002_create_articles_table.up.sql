BEGIN;
CREATE TABLE articles
(
    id            uuid        NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    title         text        NOT NULL,
    subtitle      text,
    content       text        NOT NULL,
    author        text                             default ''::text,
    url           text        NOT NULL,
    metadata      jsonb                            DEFAULT '{}'::jsonb,
    created_at    timestamptz NOT NULL             DEFAULT now(),
    language      VARCHAR(10)                      DEFAULT 'english',
    description   text                             DEFAULT ''
);
CREATE INDEX idx_articles_search ON articles
    USING bm25 (id, title, subtitle, content, description)
    WITH (key_field='id');
COMMIT;