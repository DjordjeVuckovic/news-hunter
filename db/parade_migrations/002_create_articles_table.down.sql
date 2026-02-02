DO
$$
    BEGIN
        DROP TABLE IF EXISTS articles;
        DROP INDEX IF EXISTS idx_articles_search;
    END
$$;