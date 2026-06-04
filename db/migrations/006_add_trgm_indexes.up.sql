BEGIN;
-- GIN trigram indexes backing the news_fuzzy pg-trgm engine.
CREATE INDEX IF NOT EXISTS idx_articles_title_trgm ON articles
    USING gin (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_articles_title_desc_trgm ON articles
    USING gin ((title || ' ' || coalesce(description,'')) gin_trgm_ops);
COMMIT;
