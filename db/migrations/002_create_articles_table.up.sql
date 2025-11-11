DO
$$
    BEGIN
        CREATE TABLE IF NOT EXISTS articles
        (
            id            uuid        NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
            title         text        NOT NULL,
            subtitle      text,
            content       text        NOT NULL,
            author        text                             default ''::text,
            search_vector tsvector,
            url           text        NOT NULL,
            metadata      jsonb                            DEFAULT '{}'::jsonb,
            created_at    timestamptz NOT NULL             DEFAULT now(),
            language      VARCHAR(10)                      DEFAULT 'english',
            description   text                             DEFAULT ''
        );
        CREATE INDEX IF NOT EXISTS idx_articles_search_vector ON articles
            USING gin (search_vector);
    END
$$