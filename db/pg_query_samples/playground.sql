-- ====================================================================
-- PostgreSQL Full-Text Search Playground
-- Testing weight label filtering and field boosting
-- ====================================================================

-- ====================================================================
-- TEST 1: Single Field Search (Weight Label Filtering)
-- ====================================================================

-- Search ONLY in title field (weight A)
SELECT
    id,
    title,
    substring(description, 1, 50) as desc_preview,
    ts_rank(
        '{0.0, 0.0, 0.0, 1.0}',  -- {D, C, B, A} - Only A weighted
        search_vector,
        plainto_tsquery('english', 'trump')
    ) as score
FROM articles
WHERE search_vector @@ (plainto_tsquery('english', 'trump')::text || ':A')::tsquery
ORDER BY score DESC
LIMIT 10;

-- Count articles with 'trump' in title only
SELECT COUNT(*) as title_count
FROM articles
WHERE search_vector @@ (plainto_tsquery('english', 'trump')::text || ':A')::tsquery;
-- Expected: 605


-- Search ONLY in description field (weight B)
SELECT COUNT(*) as description_count
FROM articles
WHERE search_vector @@ (plainto_tsquery('english', 'trump')::text || ':B')::tsquery;
-- Expected: 592


-- Search ONLY in content field (weight C)
SELECT COUNT(*) as content_count
FROM articles
WHERE search_vector @@ (plainto_tsquery('english', 'trump')::text || ':C')::tsquery;
-- Expected: 577


-- Search in ALL fields (no label filtering)
SELECT COUNT(*) as all_fields_count
FROM articles
WHERE search_vector @@ plainto_tsquery('english', 'trump');
-- Expected: 824 (union of all fields)


-- ====================================================================
-- TEST 2: Multi-Field Search with Weight Label Filtering
-- ====================================================================

-- Search in title OR description (weights A or B) - COMPACT NOTATION
SELECT
    id,
    title,
    substring(description, 1, 50) as desc_preview,
    ts_rank(
        '{0.0, 0.0, 1.5, 3.0}',  -- {D, C, B, A} - B=1.5, A=3.0
        search_vector,
        plainto_tsquery('english', 'trump')
    ) as score,
    -- Show which fields matched
    search_vector @@ (plainto_tsquery('english', 'trump')::text || ':A')::tsquery as matched_title,
    search_vector @@ (plainto_tsquery('english', 'trump')::text || ':B')::tsquery as matched_desc
FROM articles
WHERE search_vector @@ (plainto_tsquery('english', 'trump')::text || ':AB')::tsquery
ORDER BY score DESC
LIMIT 10;

-- Count articles with 'trump' in title OR description
SELECT COUNT(*) as title_or_description_count
FROM articles
WHERE search_vector @@ (plainto_tsquery('english', 'trump')::text || ':AB')::tsquery;


-- Search in title OR description OR content (weights A, B, or C)
SELECT COUNT(*) as title_desc_content_count
FROM articles
WHERE search_vector @@ (plainto_tsquery('english', 'trump')::text || ':ABC')::tsquery;


-- ====================================================================
-- TEST 3: Multi-Field with OR Operator (Alternative Syntax)
-- ====================================================================

-- Search in title OR description using || operator between tsquery
SELECT
    id,
    title,
    ts_rank(
        '{0.0, 0.0, 1.5, 3.0}',  -- {D, C, B, A}
        search_vector,
        plainto_tsquery('english', 'climate')
    ) as score
FROM articles
WHERE search_vector @@ (
    (plainto_tsquery('english', 'climate')::text || ':A')::tsquery
    ||
    (plainto_tsquery('english', 'climate')::text || ':B')::tsquery
)
ORDER BY score DESC
LIMIT 10;


-- ====================================================================
-- TEST 4: Field Boosting (ES-style title^3.0, description^1.5)
-- ====================================================================

-- Search with custom boost values
SELECT
    id,
    title,
    substring(description, 1, 100) as desc_preview,
    ts_rank(
        '{0.0, 0.0, 1.5, 3.0}',  -- {D, C, B, A} - title=3.0, description=1.5
        search_vector,
        plainto_tsquery('english', 'climate change')
    ) as score
FROM articles
WHERE search_vector @@ (plainto_tsquery('english', 'climate change')::text || ':AB')::tsquery
ORDER BY score DESC
LIMIT 10;


-- ====================================================================
-- TEST 5: Zero Boost (Matches but contributes 0 to score)
-- ====================================================================

-- Search in title AND description, but only title contributes to score
SELECT
    id,
    title,
    description,
    ts_rank(
        '{0.0, 0.0, 0.0, 3.0}',  -- Only A (title) weighted
        search_vector,
        plainto_tsquery('english', 'trump')
    ) as score,
    -- Verify which fields matched
    search_vector @@ (plainto_tsquery('english', 'trump')::text || ':A')::tsquery as in_title,
    search_vector @@ (plainto_tsquery('english', 'trump')::text || ':B')::tsquery as in_desc
FROM articles
WHERE search_vector @@ (plainto_tsquery('english', 'trump')::text || ':AB')::tsquery
ORDER BY score DESC
LIMIT 10;
-- Note: Documents with 'trump' in description (but not title) will have score=0


-- ====================================================================
-- TEST 6: OR Operator with websearch_to_tsquery
-- ====================================================================

-- Search for "climate OR change" in all fields
SELECT
    id,
    title,
    ts_rank(
        search_vector,
        websearch_to_tsquery('english', 'climate OR change')
    ) as score
FROM articles
WHERE search_vector @@ websearch_to_tsquery('english', 'climate OR change')
ORDER BY score DESC
LIMIT 10;

-- Count
SELECT COUNT(*) as or_query_count
FROM articles
WHERE search_vector @@ websearch_to_tsquery('english', 'climate OR change');


-- Search for "climate OR change" in title and description only
SELECT
    id,
    title,
    ts_rank(
        '{0.0, 0.0, 1.0, 2.0}',  -- Title boosted 2x
        search_vector,
        websearch_to_tsquery('english', 'climate OR change')
    ) as score
FROM articles
WHERE search_vector @@ (
    websearch_to_tsquery('english', 'climate OR change')::text || ':AB'
)::tsquery
ORDER BY score DESC
LIMIT 10;


-- ====================================================================
-- TEST 7: Phrase Search
-- ====================================================================

-- Search for exact phrase "climate change" using websearch_to_tsquery
SELECT
    id,
    title,
    description,
    ts_rank(
        search_vector,
        websearch_to_tsquery('english', '"climate change"')
    ) as score
FROM articles
WHERE search_vector @@ websearch_to_tsquery('english', '"climate change"')
ORDER BY score DESC
LIMIT 10;


-- ====================================================================
-- TEST 8: Complex Query (OR + NOT + Phrase)
-- ====================================================================

-- Search for ("climate change" OR "global warming") but NOT "hoax"
SELECT
    id,
    title,
    ts_rank(
        search_vector,
        websearch_to_tsquery('english', '"climate change" OR "global warming" -hoax')
    ) as score
FROM articles
WHERE search_vector @@ websearch_to_tsquery('english', '"climate change" OR "global warming" -hoax')
ORDER BY score DESC
LIMIT 10;


-- ====================================================================
-- TEST 9: Comparison - Dynamic vs Pre-computed tsvector
-- ====================================================================

-- Count using pre-computed search_vector (FAST with GIN index)
SELECT COUNT(*) FROM articles
WHERE search_vector @@ plainto_tsquery('english', 'trump');

-- Count using dynamic to_tsvector on title (SLOW - no index)
SELECT COUNT(*) FROM articles
WHERE to_tsvector('english', title) @@ plainto_tsquery('english', 'trump');


-- Check that search_vector contains proper weight labels
SELECT
    id,
    title,
    search_vector,
    -- Extract weighted terms for 'trump'
    ts_filter(search_vector, '{a}') as title_vector,      -- A-weighted terms
    ts_filter(search_vector, '{b}') as desc_vector,       -- B-weighted terms
    ts_filter(search_vector, '{c}') as content_vector,    -- C-weighted terms
    ts_filter(search_vector, '{d}') as subtitle_author_vector  -- D-weighted terms
FROM articles
WHERE search_vector @@ plainto_tsquery('english', 'trump')
LIMIT 5;

-- Get total count and page of results
WITH ranked AS (
    SELECT
        id,
        title,
        description,
        ts_rank(
            '{0.0, 0.0, 1.5, 3.0}',
            search_vector,
            plainto_tsquery('english', 'trump')
        ) as score
    FROM articles
    WHERE search_vector @@ (plainto_tsquery('english', 'trump')::text || ':AB')::tsquery
)
SELECT
    (SELECT COUNT(*) FROM ranked) as total_count,
    (SELECT COALESCE(MAX(score), 0.0) FROM ranked) as max_score,
    *
FROM ranked
ORDER BY score DESC
LIMIT 10 OFFSET 0;

-- Test 1: Without custom boosts (default 1.0 for both fields)
SELECT COUNT(*) as count_no_custom_boost
FROM articles
WHERE search_vector @@ (plainto_tsquery('english', 'trump')::text || ':AB')::tsquery;

-- Test 2: With custom boosts (title=3.0, description=1.5)
SELECT COUNT(*) as count_with_custom_boost
FROM articles
WHERE search_vector @@ (plainto_tsquery('english', 'trump')::text || ':AB')::tsquery;
-- Should return SAME count as Test 1!

-- Test 3: Check if ts_rank affects counts (it shouldn't!)
SELECT
    COUNT(*) as total,
    COUNT(*) FILTER (WHERE ts_rank('{0.0, 0.0, 1.0, 1.0}', search_vector, plainto_tsquery('english', 'trump')) > 0) as with_default_weights,
    COUNT(*) FILTER (WHERE ts_rank('{0.0, 0.0, 1.5, 3.0}', search_vector, plainto_tsquery('english', 'trump')) > 0) as with_custom_weights
FROM articles
WHERE search_vector @@ (plainto_tsquery('english', 'trump')::text || ':AB')::tsquery;

-- Test 4: Compare scores with different weights (same docs, different scores)
SELECT
    id,
    title,
    ts_rank('{0.0, 0.0, 1.0, 1.0}', search_vector, plainto_tsquery('english', 'trump')) as score_default,
    ts_rank('{0.0, 0.0, 1.5, 3.0}', search_vector, plainto_tsquery('english', 'trump')) as score_custom
FROM articles
WHERE search_vector @@ (plainto_tsquery('english', 'trump')::text || ':AB')::tsquery
ORDER BY score_custom DESC
LIMIT 10;


-- validate field weighting by counting matches per field
SELECT
    'Title Only' as field,
    (SELECT COUNT(*) FROM articles
     WHERE search_vector @@ (plainto_tsquery('english', 'trump')::text || ':A')::tsquery) as count
UNION ALL
SELECT
    'Description Only' as field,
    (SELECT COUNT(*) FROM articles
     WHERE search_vector @@ (plainto_tsquery('english', 'trump')::text || ':B')::tsquery) as count
UNION ALL
SELECT
    'Content Only' as field,
    (SELECT COUNT(*) FROM articles
     WHERE search_vector @@ (plainto_tsquery('english', 'trump')::text || ':C')::tsquery) as count
UNION ALL
SELECT
    'Subtitle/Author Only' as field,
    (SELECT COUNT(*) FROM articles
     WHERE search_vector @@ (plainto_tsquery('english', 'trump')::text || ':D')::tsquery) as count
UNION ALL
SELECT
    'Title OR Description' as field,
    (SELECT COUNT(*) FROM articles
     WHERE search_vector @@ (plainto_tsquery('english', 'trump')::text || ':AB')::tsquery) as count
UNION ALL
SELECT
    'All Fields' as field,
    (SELECT COUNT(*) FROM articles
     WHERE search_vector @@ plainto_tsquery('english', 'trump')) as count
ORDER BY count DESC;

-- highlighting example
SELECT
    ts_headline(
            'english',
            content,
            plainto_tsquery('english', 'trump'),
            'StartSel=<b>, StopSel=</b>, MaxWords=35, MinWords=15, MaxFragments=3'
    ) as highlighted_content
FROM articles
WHERE search_vector @@ plainto_tsquery('english', 'trump')