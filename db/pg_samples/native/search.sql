-- count
SELECT COUNT(*) AS total
FROM articles;
-- basic search(plainto_tsquery)
SELECT id, title, ts_rank(search_vector, query) as rank
FROM articles, plainto_tsquery('english', 'trump') query
WHERE search_vector @@ query
ORDER BY rank DESC
LIMIT 100;

-- advanced search with OR operator(plainto_tsquery)
SELECT id, title, ts_rank(search_vector, query) as rank
FROM articles, plainto_tsquery('english', 'trump | obama') query
WHERE search_vector @@ query
ORDER BY rank DESC
LIMIT 100;

-- phrase search(phraseto_tsquery)
SELECT id, title, ts_rank(search_vector, query) as rank
FROM articles, phraseto_tsquery('english', 'donald trump') query
WHERE search_vector @@ query
ORDER BY rank DESC
LIMIT 100;

-- bool advanced search with AND, OR, NOT operators(to_tsquery)
SELECT id, title, content,  ts_rank(search_vector, query) as rank
FROM articles, to_tsquery('english', '(trump | biden) & !obama') query
WHERE search_vector @@ query
ORDER BY rank DESC
LIMIT 100;

-- weighted search
-- 1) match search(only search title)
SELECT id, title, description, ts_rank(search_vector, query) as rank
FROM articles, to_tsquery('english', 'trump' || ':A') query
WHERE search_vector @@ query
ORDER BY rank DESC
LIMIT 100;
-- 2) multi match search(search title and description)
SELECT id, title, description, ts_rank(search_vector, query) as rank
FROM articles, to_tsquery('english', 'trump' || ':AB') query
WHERE search_vector @@ query
ORDER BY rank DESC
LIMIT 100;
-- 3) custom weights
SELECT
    id,
    title,
    ts_rank('{0.0, 0.0, 0.6, 1.0}', search_vector, plainto_tsquery('english', 'trump')) as score_custom
FROM articles
WHERE search_vector @@ (plainto_tsquery('english', 'trump')::text || ':AB')::tsquery
ORDER BY score_custom DESC
LIMIT 100;

-- proximity search with <N> operator(to_tsquery)
SELECT id, title, ts_rank(search_vector, query) as rank
FROM articles, to_tsquery('english', 'trump <5> election') query
WHERE search_vector @@ query
ORDER BY rank DESC
LIMIT 100;

-- highlight search results
SELECT
    ts_headline(
            'english',
            content,
            plainto_tsquery('english', 'trump'),
            'StartSel=<b>, StopSel=</b>, MaxWords=35, MinWords=15, MaxFragments=3'
    ) as highlighted_content
FROM articles
WHERE search_vector @@ plainto_tsquery('english', 'trump')
LIMIT 50;