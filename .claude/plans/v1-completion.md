# Plan: Complete v1 of News Hunter ("PostgreSQL as a Search Engine")

Scope decision: **Phases 0–2** = the v1 cut line. Phase 3 (scaling chapter) and
Phase 4 (hygiene) are tracked but out of v1 scope. Serbian FTS is deferred — see
[serbian-fts.md](serbian-fts.md).

All findings below were verified against code/schema (no speculation).

## Current state (verified)
- Build/vet/unit-tests green. CI = lint + build + vet + `go test -race -short`.
- Paradigms wired: FTS (PG+ES, API + bench) is the ONLY fully-executed track
  (`fts_quality` has pool/qrels/reports). Fuzzy/Semantic/Hybrid tracks are
  scaffolded but `trec/` is empty (never pooled/judged/run).
- Engines provisioned in docker-compose: native pgvector `:54320`, ParadeDB
  `:54321`, TigerData `pg_textsearch` `:54322`, ES `:9200`. Bench tracks wire
  pg-seq/pg-gin/paradedb/elasticsearch — **TigerData is never benched.**
- Semantic API: PG only. ES `SemanticSearcher` is a factory stub
  (`internal/storage/factory/factory.go:131`); `es.Embedder` + `es.VectorStore`
  already exist. Hybrid: no interface/endpoint (bench-only via raw RRF templates).

---

## Phase 0 — Correctness fixes (unblocks the paradigm tracks)
- [ ] **BUG-1** GIN trigram indexes. New migration `006_add_trgm_indexes.up.sql`:
      `CREATE INDEX … ON articles USING gin (title gin_trgm_ops);` plus an
      expression index matching `pg_trgm_multi` (`title || ' ' || coalesce(description,'')`).
      Without it `news_fuzzy` similarity queries seq-scan → invalid latency numbers.
      `pg_trgm` ext already exists (`001_init`). Mirror into tiger/parade sets if
      those engines run the fuzzy track.
- [ ] **BUG-2** Fix `tracks/news_semantic/suite.yaml` PG templates: they query
      `articles.embedding` which does not exist. Rewrite to
      `SELECT article_id AS id FROM article_embeddings ORDER BY embedding <=> '{{embedding}}'::vector LIMIT {{limit}}`
      (copy the correct pattern from `tracks/news_hybrid/suite.yaml`). Fix header
      comments: real model is Qwen3-0.6B / **1024-dim**, not text-embedding-3-small/1536.
- [ ] **BUG-3** Serbian: rename constant `LanguageSpanish` → `LanguageSerbian`
      (`internal/types/query/language.go`). Restrict v1 to English (don't ingest
      serbian rows). Full config deferred → serbian-fts.md.
- [ ] **BUG-5** `internal/bench/judgment/hybrid.go:83` — guard `len(res)==0`
      before `res[0]` (panics on embedding-less doc via the single-Grade path).
- [ ] Validate: `bench validate fts_quality news_fuzzy news_semantic news_hybrid`
      (PG path runs `EXPLAIN`, so all templates must pass). This is the gate.

## Phase 1 — Run all four paradigm benchmarks end-to-end
- [ ] Wire **TigerData** (`:54322`, pg_textsearch) as a bench engine in
      `fts_quality` (and others where pg_textsearch applies). Engine name e.g.
      `pg-tiger`; add to relevant jobs.
- [ ] For each track run the pipeline: `bench pool → judge → run → export`.
      Judgment strategies per spec: fuzzy/FTS lexical+bm25; semantic/hybrid claude-cli.
      Commit `trec/` (pool, annotations, qrels) + `reports/`.
- [ ] Deliverable: PG-variants-vs-ES across all 5 paradigms (NDCG/MAP/MRR/Bpref + latency).

## Phase 2 — API parity
- [ ] **ES semantic searcher** (kNN over `dense_vector`): implement
      `internal/storage/es/semantic_searcher.go`, wire `factory.go:131`. Finishes
      step 4 of `.claude/plans/es-embeddings-parity.md` (steps 1–3 already done:
      embedder, vector_store, mapping).
- [ ] **Hybrid search**: add `storage.HybridSearcher` interface + PG RRF SQL
      helper + ES native RRF; expose as a new query type on `POST /v1/articles/_search`.
- [ ] **`/v1/capabilities`** endpoint (type `internal/types/query/capabilities.go`
      exists; route not bound).

---

## Out of v1 (tracked)
### Phase 3 — Thesis-grade rigor (new features)
- Multiple corpus sizes (10k/100k/1M) scaling sweep (impr.txt: a thesis chapter).
- Index footprint + build-time metrics per engine (GIN vs BM25 vs HNSW).
- Extend statistical-significance reporting (Wilcoxon via `bench diff`) to all tracks.
- `search_compare` CLI (FUTURE_WORK.md) for qualitative PG/ES diffing.

### Phase 4 — Hygiene
- Reconcile `docs/IMPLEMENTATION_GOAL.md` / `FUTURE_WORK.md` with real interfaces
  (`FtsSearcher`/`SemanticSearcher`/`VectorStore` — the documented `Reader.SearchFullText`/
  `VectorSearcher`/`HybridSearcher` shapes don't exist).
- CI: `pg/embedder_test.go` + `es/embedder_test.go` + `pkg/testing` container
  helpers lack a `testing.Short()` guard, so `-short` does NOT skip them
  (ci.yml:42 comment is wrong). Add guards or a dedicated integration job.
- `unaccent` installed (`001:5`) but unused in the tsvector pipeline.

## v1 cut line
Phases 0–2 complete = defensible thesis v1: all five paradigms benchmarked across
PG variants + ES, with API parity for semantic and hybrid.
