BEGIN;
-- GIN trigram indexes backing the news_fuzzy pg-trgm engine (similarity() queries).
-- idx_articles_title_trgm backs the pg_trgm_title template.
-- idx_articles_title_desc_trgm backs the pg_trgm_multi template (title + description).
CREATE INDEX IF NOT EXISTS idx_articles_title_trgm ON articles
    USING gin (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_articles_title_desc_trgm ON articles
    USING gin ((title || ' ' || coalesce(description,'')) gin_trgm_ops);
COMMIT;
