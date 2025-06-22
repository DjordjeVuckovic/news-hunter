-- Create function to update search_vector based on language
CREATE OR REPLACE FUNCTION update_article_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    if NEW.search_vector IS NOT NULL AND NEW.search_vector != ''::tsvector THEN
        -- If search_vector is already set, do not update it
        RETURN NEW;
    END IF;
    NEW.search_vector := to_tsvector(
            COALESCE(NEW.language, 'english')::regconfig,
            COALESCE(NEW.title, '') || ' ' ||
            COALESCE(NEW.subtitle, '') || ' ' ||
            COALESCE(NEW.content, '') || ' ' ||
            COALESCE(NEW.author, '')
                         );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update search_vector on INSERT and UPDATE
DROP TRIGGER IF EXISTS trigger_update_article_search_vector ON articles;
CREATE TRIGGER trigger_update_article_search_vector
    BEFORE INSERT OR UPDATE ON articles
    FOR EACH ROW EXECUTE FUNCTION update_article_search_vector();

-- Update existing rows to populate search_vector with default language
UPDATE articles 
SET search_vector = to_tsvector(
    'english',
    COALESCE(title, '') || ' ' ||
    COALESCE(subtitle, '') || ' ' ||
    COALESCE(content, '') || ' ' ||
    COALESCE(author, '')
)
WHERE search_vector IS NULL OR search_vector = ''::tsvector;