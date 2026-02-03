select id, title, content
from articles
limit 10;

-- Tokenizers
-- 1) word boundary: splits on unicode word boundaries. Respects language-specific rules.
SELECT 'South Africa beat France 29-28 to reach Rugby World Cup semifinals and didn''t beat Serbia!'::pdb.unicode_words::text[];
-- 2) literal/literal normalized: indexes entire input as a single token. Normalized version lowercases and removes diacritics.
-- useful for term search like IDs, codes, etc. Default for doc IDs.
SELECT 'South Africa beat France 29-28 to reach Rugby World Cup semifinals and didn''t beat Serbia!'::pdb.literal::text[];
SELECT 'South Africa beat France 29-28 to reach Rugby World Cup semifinals and didn''t beat Serbia!'::pdb.literal_normalized::text[];
-- 3) n-gram: breaks input into overlapping substrings of length n. Useful for partial matching and autocomplete.
-- Here we use n-grams of length 2 to 3.
SELECT 'South Africa beat France'::pdb.ngram(2, 3)::text[];


-- Stemmer
-- 1) English stemmer
SELECT 'South Africa beat France 29-28 to reach Rugby World Cup semifinals and didn''t beat Serbia!'::pdb.simple('stemmer=english')::text[];

-- Search

-- 1) basic search
-- Find articles with 'trump OR biden' in the title
SELECT *, pdb.score(id) as score
FROM articles
WHERE title @@@ 'trump biden'
ORDER BY score DESC
LIMIT 10;

--- 2) match search
-- Match disjunction(|||) in the content(trump OR obama) - Find any article which title contains one or more of the terms tokenized from search query
SELECT *, pdb.score(id) as score
FROM articles
WHERE title ||| 'trump obama'
ORDER BY score DESC
LIMIT 10;
-- Match conjunction(&&&) in the content(trump AND election) - Find any article which title contains all the terms tokenized from search query
SELECT *, pdb.score(id) as score
FROM articles
WHERE title &&& 'trump obama'
ORDER BY score DESC
LIMIT 10;

---3) phrase search
-- Find articles with the exact phrase 'donald trump' in the title - Same as conjunction search but requires exact order of terms
SELECT *, pdb.score(id) as score
FROM articles
WHERE title ### 'donald trump'
ORDER BY score DESC
LIMIT 10;
-- with slop - Find articles with the phrase 'donald trump' allowing up to 3 intervening words
SELECT *, pdb.score(id) as score
FROM articles
WHERE title ### 'donald trump'::pdb.slop(3)
ORDER BY score DESC
LIMIT 10;

---4) term search
-- Find articles with the exact term 'trump' in the title - exact string match, but at the token level
SELECT *, pdb.score(id) as score
FROM articles
WHERE title === 'trump'
ORDER BY score DESC
LIMIT 10;
-- Term set - find articles with any of the exact terms 'trump', 'biden', or 'obama' in the title
SELECT *, pdb.score(id) as score
FROM articles
WHERE title === ARRAY['trump', 'biden', 'obama']
ORDER BY score DESC
LIMIT 10;

-- 5) Fuzzy search
-- Find articles with terms similar to 'trum' in the title (edit distance of 2). 2 is max edit distance allowed (performance reasons from docs).
-- Fuzziness is supported for match and term queries. Should be used with n-gram tokenizer for better results.
SELECT *, pdb.score(id) as score
FROM articles
WHERE title ||| 'trum'::pdb.fuzzy(2)
ORDER BY score DESC
LIMIT 10;

--- 6) Highlight search results
-- Highlight matches for 'trump' in the title
SELECT id, pdb.snippet(title)
from articles
WHERE title @@@ 'trump'
LIMIT 10;

-- 7) Proximity search
-- Find articles with 'trump' within 5 words of 'election' in the title
-- Order of terms does not matter: just distance between them
SELECT *, pdb.score(id) as score
FROM articles
WHERE title @@@ ('trump' ## 5 ## 'election')
ORDER BY score DESC
LIMIT 10;
-- Ordered proximity search - Find articles with 'trump' within 5 words before 'election' in the title
SELECT *, pdb.score(id) as score
FROM articles
WHERE title @@@ ('trump' ##> 5 ##> 'election')
ORDER BY score DESC
LIMIT 10;

-- 8) Boosted search
-- Boost matches for 'trump' by a factor of 2 in the title while searching for 'biden' in the content
SELECT *, pdb.score(id) as score
FROM articles
WHERE title ||| 'trump'::pdb.boost(2) OR content ||| 'biden'
ORDER BY score DESC
LIMIT 10;
-- Works with multiple casts too: boost and fuzzy
SELECT *, pdb.score(id) as score
FROM articles
WHERE title ||| 'trump'::pdb.fuzzy(2)::pdb.boost(2) OR content ||| 'biden'
ORDER BY score DESC
LIMIT 10;
-- Constant scoring assigns a fixed score to all matching documents
SELECT *, pdb.score(id) as score
FROM articles
WHERE title ||| 'trump'::pdb.const(1)
ORDER BY score DESC
LIMIT 10;

-- Top N search