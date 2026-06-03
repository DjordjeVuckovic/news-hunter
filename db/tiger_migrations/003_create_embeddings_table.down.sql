BEGIN;
DROP INDEX IF EXISTS idx_article_embedding;
DROP TABLE IF EXISTS article_embeddings;
DROP EXTENSION IF EXISTS vector;
COMMIT;
