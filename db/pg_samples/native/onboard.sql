-- Full Text Search Playground
select * from articles limit 10;

-- tsvector: a document representation for full-text search. Stemmed and normalized.
SELECT to_tsvector('english', 'Sikkim warning: Hydroelectricity push must be accompanied by safety measures');
-- multiple occurrences
SELECT to_tsvector('english', 'Livin la vida loca: Living the life in the fast lane. Living it up in the city.');

-- tsvector with multiple columns
SELECT to_tsvector('english', title || ' ' || content) AS combined_tsvector
FROM (
         SELECT 'News Search Guide' as title,
                'Learn how to search in News effectively' as content
     ) t;
-- tsvector with multiple columns and weights
SELECT setweight(to_tsvector('english', title), 'A') ||
       setweight(to_tsvector('english', content), 'B') AS weighted_combined_tsvector
FROM (
         SELECT 'News Search Guide' as title,
                'Learn how to search in News effectively' as content
        ) t;

-- tsquery: a query representation for full-text search. It can be used to match against tsvector.

SELECT to_tsquery('english', 'postgresql & database');  -- Match both words
SELECT to_tsquery('english', 'postgresql | mysql');     -- Match either word
SELECT to_tsquery('english', 'database & !mysql');      -- Match database but not mysql
SELECT to_tsquery('english', 'postgresql <-> database'); -- Match words in sequence

WITH sample_queries AS (
    SELECT
        to_tsquery('english', 'postgresql & database') as q1,
        to_tsquery('english', 'postgresql <-> database') as q2,
        to_tsquery('english', 'fast & !slow & database') as q3
)
SELECT
    'PostgreSQL is a database' as text,
    to_tsvector('english', 'PostgreSQL is a database') @@ q1 as matches_and,
    to_tsvector('english', 'PostgreSQL is a database') @@ q2 as matches_phrase,
    to_tsvector('english', 'PostgreSQL is a fast database') @@ q3 as matches_not
FROM sample_queries;

-- lexemes: the basic units of meaning in full-text search. They are the normalized forms of words. Represent the core meaning of words.
-- Pg text search normalizes words by:
-- 1. Converting to lowercase
-- 2. Removing stop words
-- 3. Applying stemming (reducing words to their root form)
-- 4. Handling special characters

WITH lexemes_examples AS (
    SELECT unnest(ARRAY[
        'running',
        'Runs',
        'ran',
        'PostgreSQL''s',
        'databases',
        'faster!'
        ]) AS word
)
SELECT
    word as original_word,
    to_tsvector('english', word) as lexeme
FROM lexemes_examples;