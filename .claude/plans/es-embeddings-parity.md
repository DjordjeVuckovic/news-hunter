# Plan: Elasticsearch embeddings parity

Bring ES to parity with Postgres for the embedding/vector stack. Today PG
implements all three concerns and ES is stubbed everywhere:

| Interface | Role | PG | ES (now) |
|---|---|---|---|
| `storage.EmbedIndexer` | write/ingest embeddings (`embed_ingest`, online `ds_ingest`) | `pg.Embedder` | factory error "not yet implemented" |
| `storage.VectorStore` (`QueryVector`, `DocVectors`) | read side for bench / semantic | `pg.VectorStore` | `es.VectorStore` stub |
| `storage.SemanticSearcher` (`SearchSemantic`) | query-time semantic search API | `pg.SemanticSearcher` | factory error "not yet implemented" |

## Core design decision: vector lives ON the article document
ES is not relational. Instead of a parallel `article_embeddings` index keyed by
`article_id` (the PG table shape), store a `dense_vector` field **on the existing
article document**. The article `_id` is already the article UUID
(`es/indexer.go:47` → `.Id(doc.ID)`), so `embedding.Vec.ID.String()` maps
directly to `_id`. This is what makes ES kNN + hybrid (BM25 + vector) work and
keeps one document per article.

Consequence: "loading embeddings" on ES = **bulk partial-update of existing
article docs**, not inserting into a side index.

## Pieces (in build order)

### 1. Mapping: add `dense_vector` (`internal/storage/es/document.go`)
In `IndexBuilder.buildMapping`, add:
```go
"embedding": types.DenseVectorProperty{
    Dims:       some.Int(1024),
    Index:      some.Bool(true),
    Similarity: &densevectorsimilarity.Cosine, // L2-normalised Qwen vectors
    // optional: IndexOptions{ Type: hnsw, M, EfConstruction }
},
```
- New indices get it via `EnsureIndex`.
- **Existing indices**: `EnsureIndex` early-returns if the index exists, so the
  field won't appear. Add a step that issues `PUT <index>/_mapping` to add the
  `embedding` field when missing (ES allows adding new fields without reindex).
  The article docs still need the value written (step 2).

### 2. Write side: `es.Embedder` implements `storage.EmbedIndexer`
New `internal/storage/es/embedder.go`. Use `esutil.BulkIndexer` (mirror
`es/indexer.go:66`) with `Action: "update"`:
```
DocumentID: vec.ID.String()
Body:       {"doc": {"embedding": [...]}}
```
- **Idempotent for free** — partial update overwrites; safe to re-run (ES's
  native upsert-by-`_id`, the PG-parity property).
- **Orphan tolerance** — updating a missing `_id` returns a per-item 404;
  `BulkIndexer` continues. Collect failures in `OnFailure`, count + log skipped
  (mirror the PG orphan-skip semantics).
- Do NOT use `doc_as_upsert` — we must not create article-less docs.
- Wire `factory.NewEmbedderIndexer` ES branch → `es.NewEmbedder(cfg.Es)`.
- `embed_ingest` itself needs **no change**: `STORAGE_TYPE=es EMBEDDING_SOURCE=file`.

### 3. Read side: `es.VectorStore` (replace stub)
`internal/storage/es/vector_store.go`:
- `DocVectors(ids)` — `_mget` (or `_source` filtered to `embedding`) for the ids;
  return `map[uuid]→[]float32`, ids without a vector simply absent (per the
  interface contract). Mirror `pg.VectorStore.DocVectors`.
- `QueryVector(text)` — embed query text via `embedding.Client` exactly like
  `pg.VectorStore` (needs an embedder injected). The query model must match the
  stored model.
- Wire `factory/vector.go`: the ES branch currently returns `es.NewVectorStore()`
  with no args — extend it to pass the ES client + an embedder (same
  `EmbeddingClient` + `Model` plumbing the PG branch already has).

### 4. Query side: `es.SemanticSearcher` implements `SearchSemantic`
New `internal/storage/es/semantic_searcher.go`:
- Embed `query.Semantic.Query` → run a **kNN search** over the `embedding` field
  (`knn: { field: "embedding", query_vector, k, num_candidates }`), optionally
  with a `similarity`/score threshold analogous to PG's `query.Threshold`.
- Map hits → `storage.VectorSearchResult` (reuse the `dto.ArticleSearchResult`
  shape the FTS searcher already builds in `es/searcher.go`).
- Wire `factory.NewSemanticSearcher` ES branch.

## model_name handling (the one real asymmetry)
PG keys on `(article_id, model_name)` → multiple models per article. An ES doc
has a single `embedding` field. For v1:
- one model → one `embedding` field;
- record the model in the index `_meta` (and/or a `model_name` keyword field) for
  provenance;
- multiple models later → distinct fields (`embedding_qwen`, …) or a nested type.
`es.Embedder` should accept the model name and reject/loudly-warn if a second,
different model is loaded into the same field.

## Config
- ES already has `ClientConfig{Addresses, IndexName, …}` (`es/client.go`).
- Query-time (`VectorStore.QueryVector`, `SemanticSearcher`) needs an
  `embedding.Client` (Ollama) just like PG — same wiring already present in
  `factory/vector.go` and `factory.NewSemanticSearcher`.
- No new env for `embed_ingest` write path beyond existing `ES_*` + `EMBEDDING_*`.

## Tests (testcontainers ES already exists: `pkgtesting.NewESContainer`)
- `es.Embedder`: bulk update sets `embedding` on existing docs; missing `_id`
  is skipped + counted (orphan); re-run overwrites (idempotent).
- `es.VectorStore.DocVectors`: round-trips vectors; absent ids omitted.
- `es.SemanticSearcher`: kNN returns nearest doc for a known query vector.
- Reuse the `embedfile` parquet reader as-is (engine-agnostic).

## Ops / migration
- Fresh index: mapping change is enough.
- Existing index: `PUT _mapping` to add `embedding`, then run `embed_ingest`
  (`STORAGE_TYPE=es`) to populate via bulk update.
- Bulk-load cost: ES `update` = get-merge-reindex of the whole doc + dense_vector
  HNSW segment merges — heavier than fresh inserts. If reloading the whole
  corpus, prefer **reindex articles+vectors together** (single index op per doc)
  over update. Document both.

## Caveats
- ES indexed `dense_vector` dims limit is 4096 → 1024 is fine.
- Hybrid (BM25 + kNN) search is a natural follow-up once the field exists, but is
  out of scope for parity.
- Keep PG as the precedence engine (`factory/vector.go` already prefers PG);
  parity makes ES selectable, not default.

## Sequencing (suggested PRs)
1. Mapping + `es.Embedder` + factory wiring + tests → `embed_ingest` works on ES.
2. `es.VectorStore` (fill stub) + factory/vector.go wiring + tests → bench works.
3. `es.SemanticSearcher` + factory wiring + tests → API semantic search on ES.
