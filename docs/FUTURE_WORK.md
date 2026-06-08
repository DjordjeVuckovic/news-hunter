# FUTURE_WORK.md

Roadmap for unifying and improving PostgreSQL vs Elasticsearch search results.
Items are tagged **DONE**, **IN PROGRESS**, or **NOT BUILT** to reflect actual state.

## Observed Differences (PG vs ES)

- **Date fields**: PG returns proper timestamps; some ES results have shown zero dates
  (`0001-01-01T00:00:00Z`). Likely a mapping/ingest issue — verify against current ES mapping.
- **Missing fields**: ES results have lacked `language` / some `sourceId` values.
- **Score scales differ**: PG `ts_rank` (0–1) vs ES BM25 (≈1–30+). No native cross-engine
  comparability — handled in evaluation by rank-based metrics, not raw scores.

PG tends to be more complete on metadata; ES tends to rank relevance better. The goal is to
keep both honest and comparable.

---

## Roadmap

### 1. Search comparison tooling — NOT BUILT
Proposed `cmd/search_compare/` CLI to query both backends with identical params and emit a
side-by-side diff (data discrepancies, missing fields, ranking drift). Does not exist yet.

> Note: cross-engine **relevance** comparison already exists via the bench CLI (see below).
> `search_compare` would be a lighter, ad-hoc field/result diff tool, not a replacement.

### 2. Data consistency — IN PROGRESS / investigate
- Verify identical source data reaches both backends; audit `configs/mappings/` for drift.
- Root-cause ES zero dates and missing `language`/`sourceId`; add ingest-time validation.
- Confirm timezone handling and analyzer/FTS-config parity.

### 3. Unified ranking / score normalization — NOT BUILT
Proposed `internal/ranking/` service: normalize engine scores to 0–1, optional composite
scoring (text relevance + recency + source authority) with configurable weights and an
explanation breakdown. Not implemented; engine scores are currently surfaced as-is and
compared via rank-based IR metrics.

### 4. IR evaluation framework — DONE (bench CLI / tracks/)
A TREC-style evaluation pipeline already exists:
- **Tooling**: `cmd/bench/` + `internal/bench/` — init → validate → pool → judge → run →
  export → show (also status, diff, clean). See `docs/bench.md`.
- **Metrics**: NDCG, MAP, MRR, Bpref, Precision/Recall/F1.
- **Tracks**: `tracks/` (`fts_quality`, `news_fuzzy`, `news_semantic`, `news_hybrid`), each
  with `spec.yaml`, `suite.yaml`, `trec/` (pool + judgments), and `reports/`.
- **Judgment strategies**: lexical, BM25, vector, hybrid, claude-cli/claude-api, manual.

Remaining ideas here: CI integration, automated ranking-drift alerts, historical tracking.

### 5. Production optimization — NOT BUILT (later)
Query routing by characteristics, per-engine index tuning, cache warming, A/B testing of
ranking changes, runtime quality monitoring.

---

## Technical Debt
- Engine-specific scoring logic embedded in storage implementations.
- Error-response shapes differ between backends.
- Limited cross-backend integration tests.
- Storage settings not centrally managed.

---

*Status: evaluation framework built; comparison tooling and ranking normalization still planned.*
