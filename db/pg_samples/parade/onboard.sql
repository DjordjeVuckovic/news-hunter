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

-- 2)

