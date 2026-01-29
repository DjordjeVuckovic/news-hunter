-- count
SELECT COUNT(*) AS total
FROM articles;
-- basic search
SELECT id, title, ts_rank(search_vector, query) as rank
FROM articles, to_tsquery('english', 'trump') query
WHERE search_vector @@ query
ORDER BY rank DESC
LIMIT 100;


SELECT id, title, ts_rank(search_vector, query) as rank
FROM articles, to_tsquery('english', 'clim & change') query
WHERE search_vector @@ query
ORDER BY rank DESC
LIMIT 100;