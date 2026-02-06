BEGIN;

-- Create function to update search_vector with proper field weights
CREATE OR REPLACE FUNCTION update_article_search_vector()
    RETURNS TRIGGER AS $$
BEGIN
    -- Only compute if not already set by application
    IF NEW.search_vector IS NULL OR NEW.search_vector = ''::tsvector THEN
        NEW.search_vector :=
            setweight(to_tsvector(COALESCE(NEW.language, 'english')::regconfig,
                                  COALESCE(NEW.title, '')), 'A') ||

            setweight(to_tsvector(COALESCE(NEW.language, 'english')::regconfig,
                                  COALESCE(NEW.description, '')), 'B') ||

            setweight(to_tsvector(COALESCE(NEW.language, 'english')::regconfig,
                                  COALESCE(NEW.content, '')), 'C') ||

            setweight(to_tsvector(COALESCE(NEW.language, 'english')::regconfig,
                                  COALESCE(NEW.subtitle, '') || ' ' || COALESCE(NEW.author, '')), 'D');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


-- Create trigger to automatically update search_vector on INSERT and UPDATE
DROP TRIGGER IF EXISTS trigger_update_article_search_vector ON articles;
CREATE TRIGGER trigger_update_article_search_vector
    BEFORE INSERT OR UPDATE ON articles
    FOR EACH ROW EXECUTE FUNCTION update_article_search_vector();


-- Update existing rows to populate search_vector with proper weights
UPDATE articles
SET search_vector =
        setweight(to_tsvector(COALESCE(language, 'english')::regconfig,
                              COALESCE(title, '')), 'A') ||
        setweight(to_tsvector(COALESCE(language, 'english')::regconfig,
                              COALESCE(description, '')), 'B') ||
        setweight(to_tsvector(COALESCE(language, 'english')::regconfig,
                              COALESCE(content, '')), 'C') ||
        setweight(to_tsvector(COALESCE(language, 'english')::regconfig,
                              COALESCE(subtitle, '') || ' ' || COALESCE(author, '')), 'D')
WHERE search_vector IS NULL OR search_vector = ''::tsvector;

COMMIT;