# Plan: Load Colab Parquet embeddings from S3 and ingest into PostgreSQL

## Goal
Embeddings for the canonical news corpus are generated once, offline, in Colab
(`scripts/embed_qwen3.ipynb`, Qwen3-Embedding-0.6B, last-token pool, L2-norm).
The notebook emits a single **Parquet** file with `id` + `embedding` columns and
file-level metadata. We add a Go path to pull that file from an **S3-compatible**
store, map ids to articles, and bulk-upsert into `article_embeddings`.

Online (Ollama) generation during `ds_ingest` stays optional and unchanged.

## Artifact contract (from the notebook)
Parquet `gl_news_embeddings.parquet`:
- column `id`: `string` (article UUID)
- column `embedding`: `list<float32>` (1024-dim, L2-normalised)
- file metadata: `model` (= `qwen3-embedding:0.6b`, canonical/DB name),
  `hf_model_id`, `dim`, `pooling=last_token`, `normalized=l2`, `row_count`,
  `created_at`.

`model_name` in the DB comes from the file metadata, so doc-side and query-side
(model used by the semantic searcher) always agree.

## Feature flag
`EMBEDDING_SOURCE` = `online` | `file` | `none` (default `online`).
- `online` — existing Ollama generation inside `ds_ingest` (gated by `EMBEDDING_ENABLED`).
- `file`   — load precomputed embeddings via `cmd/embed_ingest`.
- `none`   — no embeddings.

## Components
1. `internal/storage/objectstore/` — aws-sdk-go-v2 S3 client (S3-compatible via
   `BaseEndpoint` + path-style). `Download(ctx, key, dstPath)`.
2. `internal/embedding/embedfile/` — streaming Parquet reader (parquet-go).
   Exposes file `Meta` (model, dim) + batched `Read` of `{ID string, Embedding []float32}`.
3. `internal/embedding/config.go` — add `Source` + `ObjectStore` config;
   `BaseURL` required only when `Source==online`.
4. `internal/storage/pg/embedder.go` — re-runnable `SaveBulk`: COPY into a TEMP
   staging table, then `INSERT … SELECT … JOIN articles ON CONFLICT
   (article_id, model_name) DO UPDATE`. Orphan ids (no matching article) are
   skipped + counted, not fatal.
5. `cmd/embed_ingest/` — standalone command:
   download (S3) or open local Parquet → decode → validate dim==1024 → map ids →
   `EmbedIndexer.SaveBulk` in chunks. Gated by `EMBEDDING_SOURCE=file`.

## Ordering / invariants
- Articles must be ingested first (`article_embeddings.article_id` FK → `articles`).
- `embedding` width must equal `VECTOR(1024)`.
- Idempotent: safe to re-run (upsert + orphan-skip).

## Touch list
- New: `internal/storage/objectstore/`, `internal/embedding/embedfile/`,
  `cmd/embed_ingest/{main,config}.go`, `cmd/embed_ingest/.env.example`,
  `docs/embeddings.md`.
- Edit: `internal/embedding/config.go`, `internal/storage/pg/embedder.go`,
  `go.mod`, `CLAUDE.md`, notebook save cell (Parquet).
- Untouched: ds_ingest online path, Ollama `Embedder`, semantic searcher.

## Deps
- `github.com/parquet-go/parquet-go`
- `github.com/aws/aws-sdk-go-v2` (`config`, `credentials`, `service/s3`)
