# FTS Quality Benchmark Framework

## Overview

Build an extensible bench framework to evaluate full-text search quality across PostgreSQL native, ParadeDB, and Elasticsearch using industry-standard IR metrics. The framework uses TREC-style pooling with LLM-assisted relevance judgments, ~30-50 curated queries across all 5 search paradigms, and produces comparison reports. Latency is captured as metadata; dedicated load testing deferred to k6 later on.

## Package Structure

```
internal/bench/
    metrics/        -- Pure IR metric functions (no project deps)
    suite/          -- Test suite types and YAML loader
    engine/         -- Thin adapter over FtsSearcher
    runner/         -- Orchestration: queries x engines x metrics
    report/         -- JSON + text table report generation
    judgment/       -- TREC-style pooling + LLM judgment helper
    types/          -- Shared types (e.g., ScoreSet)

cmd/bench/      -- CLI entry point
configs/bench/  -- Test suite YAML files
```

## Implementation Steps

### Step 1: Metrics Package (`internal/bench/metrics/`)

Pure functions with zero project dependencies. All take `[]uuid.UUID` (ranked doc IDs) and `map[uuid.UUID]int` (relevance judgments) as input.

**Files:**

- `metrics.go` - Grade constants (0-3), `ScoreSet` struct, `ComputeAll()` convenience function
- `ndcg.go` - `NDCGAtK(rankedDocs, judgments, k)` using `DCG = sum((2^rel - 1) / log2(i+2))`
- `precision.go` - `PrecisionAtK()`, `RecallAtK()`, `F1AtK()` with configurable relevance threshold
- `ranking.go` - `AveragePrecision()` (single-query MAP component), `ReciprocalRank()` (MRR component)
- `metrics_test.go` - Table-driven tests: perfect ranking, inverse ranking, empty results, graded relevance, edge cases

**Key type:**
```go
package bench

type ScoreSet struct {
    NDCG      map[int]float64  // K -> NDCG@K
    Precision map[int]float64  // K -> P@K
    Recall    map[int]float64  // K -> R@K
    F1        map[int]float64  // K -> F1@K
    AP        float64          // Average Precision
    RR        float64          // Reciprocal Rank
}
```

### Step 2: Suite Package (`internal/bench/suite/`)

Defines test queries and loads them from YAML. Uses separate spec types (like existing DTO pattern) to keep domain types clean.

**Files:**

- `types.go` - `TestSuite`, `BenchmarkQuery`, `RelevanceJudgment`, query spec types (`QueryStringSpec`, `MatchSpec`, `MultiMatchSpec`, `PhraseSpec`, `BooleanSpec`)
- `loader.go` - `LoadFromFile(path)`, `Parse(data)`, `ToDomainQuery(bq)` converting specs to `query.*` domain types
- `loader_test.go` - Parsing and validation tests

**Key types:**
```go
type BenchmarkQuery struct {
    ID          string              `yaml:"id"`
    Description string              `yaml:"description"`
    Kind        query.Kind          `yaml:"kind"`
    QueryString *QueryStringSpec    `yaml:"query_string,omitempty"`
    Match       *MatchSpec          `yaml:"match,omitempty"`
    // ... one spec per query kind
    Judgments   []RelevanceJudgment `yaml:"judgments"`
}

type RelevanceJudgment struct {
    DocID uuid.UUID `yaml:"doc_id"`
    Grade int       `yaml:"grade"`  // 0=not relevant, 1=marginal, 2=relevant, 3=highly relevant
}
```

### Step 3: Engine Adapter (`internal/bench/engine/`)

Thin wrapper that bridges bench queries to the existing `FtsSearcher` interface and extracts ranked doc IDs from `SearchResult.Hits`.

**Files:**

- `engine.go` - `SearchEngine` struct (Name + `storage.FtsSearcher`)
- `adapter.go` - `Execute(ctx, engine, query, maxK) QueryExecution` dispatching by `query.Kind`, plus `extractDocIDs()` pulling `Article.ID` from `dto.ArticleSearchResult`
- `factory.go` - `CreateEngines(ctx, []EngineSpec)` reusing `factory.NewSearcher()`

**Key type:**
```go
type QueryExecution struct {
    RankedDocIDs []uuid.UUID
    TotalMatches int64
    Latency      time.Duration
    Error        error
}
```

### Step 4: Runner (`internal/bench/runner/`)

Orchestrates: for each query, for each engine, execute query, compute all metrics.

**Files:**

- `config.go` - `Config` struct: KValues (`[]int{5,10,20}`), MaxK (100), RelevanceThreshold (1), WarmupRuns (1), Runs (3)
- `result.go` - `QueryResult` (per query-engine pair), `BenchmarkResult` (map[queryID][engineName]QueryResult)
- `runner.go` - `Runner.Run(ctx)` iterating queries x engines, calling `engine.Execute()` then `metrics.ComputeAll()`

### Step 5: Report (`internal/bench/report/`)

Converts `BenchmarkResult` into human-readable and machine-readable output.

**Files:**

- `types.go` - `Report`, `Entry` (per query-engine metrics), `AggregatedEntry` (mean across queries per engine)
- `aggregate.go` - `Generate(result, suite)` computing means: Mean NDCG@K, MAP (mean of AP), MRR (mean of RR), etc.
- `json.go` - `WriteJSON(report, path)` with `json.MarshalIndent`
- `table.go` - `WriteTable(report, writer)` using `text/tabwriter` for aligned output:

```
=== FTS Quality Benchmark ===
Suite: fts_quality_v1 | Queries: 35

                  | PG Native | ParadeDB  | Elasticsearch
------------------+-----------+-----------+--------------
NDCG@10           |     0.720 |     0.785 |         0.812
MAP               |     0.650 |     0.710 |         0.745
MRR               |     0.890 |     0.920 |         0.940
Precision@10      |     0.550 |     0.620 |         0.650
Recall@10         |     0.450 |     0.500 |         0.540
```

### Step 6: Judgment Helper (`internal/bench/judgment/`)

Supports the TREC-style pooling + LLM-assist workflow.

**Files:**

- `pooler.go` - `PoolResults(results map[engineName]QueryExecution, depth int) []uuid.UUID` merges top-K docs from all engines into deduplicated pool
- `exporter.go` - `ExportForAnnotation(pool, articles, outputPath)` writes a YAML/JSON file with document ID, title, snippet for annotation
- `importer.go` - `ImportJudgments(path) []RelevanceJudgment` reads completed annotations back

**Workflow:**
1. Run bench with empty judgments -> collect `QueryExecution` results
2. `PoolResults()` merges top-20 from each engine per query
3. `ExportForAnnotation()` creates annotation file with doc metadata
4. Annotate: LLM labels + manual review/correction
5. `ImportJudgments()` merges annotations back into suite YAML
6. Re-run bench with populated judgments -> compute metrics

### Step 7: CLI (`cmd/bench/`)

**Files:**

- `main.go` - Wires everything: load suite, create engines, run bench, generate reports
- `config.go` - CLI flags: `--suite`, `--output`, `--k`, `--max-k`, `--pg-native`, `--pg-parade`, `--es-addresses`, `--es-index`

```bash
# Run against all engines
./bin/bench \
  --pg-native "postgresql://news_user:news_password@localhost:54320/news_db" \
  --pg-parade "postgresql://news_user:news_password@localhost:54321/news_db" \
  --es-addresses "http://localhost:9200" \
  --suite configs/bench/fts_quality_v1.yaml \
  --output bench_results
```

### Step 8: Test Suite YAML (`configs/bench/fts_quality_v1.yaml`)

~35 queries distributed across paradigms:

| Paradigm     | Count | Examples                                           |
|-------------|-------|----------------------------------------------------|
| query_string | 8     | "climate change", "election results 2024"          |
| match        | 7     | title:"renewable energy", content:"trade deficit"  |
| multi_match  | 7     | "inflation rate" (title^3, description^2, content) |
| phrase       | 7     | "World Cup" (slop=0), "artificial intelligence" (slop=2) |
| boolean      | 6     | "(climate OR weather) AND NOT politics"             |

### Step 9: Makefile + Infra

- Add `build-bench`, `run-bench`, `run-bench-all` targets
- Add `cmd/bench/.env` template
- Add `bench_results/` to `.gitignore`

## Files Modified (existing)

| File | Change |
|------|--------|
| `Makefile` | Add bench build/run targets |
| `.gitignore` | Add `bench_results/` |

## Files Created (new)

| File | Purpose |
|------|---------|
| `internal/bench/metrics/metrics.go` | ScoreSet type, ComputeAll, grade constants |
| `internal/bench/metrics/ndcg.go` | NDCG@K computation |
| `internal/bench/metrics/precision.go` | P@K, R@K, F1@K |
| `internal/bench/metrics/ranking.go` | AP, RR |
| `internal/bench/metrics/metrics_test.go` | Comprehensive unit tests |
| `internal/bench/suite/types.go` | TestSuite, BenchmarkQuery, spec types |
| `internal/bench/suite/loader.go` | YAML loading and domain conversion |
| `internal/bench/suite/loader_test.go` | Loader unit tests |
| `internal/bench/engine/engine.go` | SearchEngine type |
| `internal/bench/engine/adapter.go` | Execute + extractDocIDs |
| `internal/bench/engine/factory.go` | CreateEngines from specs |
| `internal/bench/runner/config.go` | Runner config with defaults |
| `internal/bench/runner/result.go` | QueryResult, BenchmarkResult |
| `internal/bench/runner/runner.go` | Orchestration loop |
| `internal/bench/report/types.go` | Report, Entry, AggregatedEntry |
| `internal/bench/report/aggregate.go` | Mean computation |
| `internal/bench/report/json.go` | JSON output |
| `internal/bench/report/table.go` | Text table output |
| `internal/bench/judgment/pooler.go` | TREC-style result pooling |
| `internal/bench/judgment/exporter.go` | Export pool for annotation |
| `internal/bench/judgment/importer.go` | Import completed annotations |
| `cmd/bench/main.go` | CLI entry point |
| `cmd/bench/config.go` | Flag parsing, engine spec creation |
| `configs/bench/fts_quality_v1.yaml` | Initial test suite (~35 queries) |

## Verification

1. `go test ./internal/bench/metrics/...` - All metric unit tests pass
2. `go test ./internal/bench/suite/...` - Suite loader tests pass with sample YAML
3. `go build ./cmd/bench` - CLI compiles
4. `docker-compose up -d` then `make run-bench-all` - Runs against live PG + ES, produces `bench_results/report.json` and table output
5. Spot-check: for a query with known relevant docs, verify NDCG@10 = 1.0 when engine returns them in ideal order
