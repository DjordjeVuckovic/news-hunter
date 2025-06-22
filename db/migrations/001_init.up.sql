CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
-- fuzzy search and similarity functions
CREATE EXTENSION IF NOT EXISTS pg_trgm;
-- remove accents from text, improves search quality. Removes diacritics from text, improving search for international content.
CREATE EXTENSION IF NOT EXISTS unaccent;