-- Drop the trigger
DROP TRIGGER IF EXISTS trigger_update_article_search_vector ON articles;

-- Drop the function
DROP FUNCTION IF EXISTS update_article_search_vector();