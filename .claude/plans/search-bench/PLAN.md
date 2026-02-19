# FTS Quality Benchmark — Dual-Layer Refactoring

Merges and supersedes: `search-bench/PLAN.md`, `iter-1-mvp.md`, `iter-2-reports.md`, `iter-3-judgments.md`

## What's Done (iter-1 MVP)

Already implemented and working:
- `internal/bench/metrics/` — All IR metrics (NDCG, P@K, R@K, F1@K, AP, RR). **No changes needed.**
- `internal/bench/suite/` — Types + YAML loader using `query.Kind` and domain spec types
- `internal/bench/engine/` — `SearchEngine` wrapping `storage.FtsSearcher`, `adapter.go` dispatching by `query.Kind`
- `internal/bench/runner/` — Orchestration via FtsSearcher
- `internal/bench/report/` — Aggregated + per-query text tables
- `cmd/bench/` — CLI with `--pg`, `--es-addresses`, `--k`, `--max-k`
- `internal/storage/executor.go` — `RawExecutor` interface (`Exec(ctx, query, params, opts)`)
- `internal/storage/pg/executor.go` — `pg.RawExecutor` implementation wrapping pgxpool
- `configs/benchmark/fts_quality_v1.yaml` — 10 starter queries (old format)

## What Changes

The FtsSearcher adapter layer is replaced with two execution strategies:

1. **Raw Executors** — Direct SQL/JSON queries to engines. Measures true engine capability.
2. **API Executor** — HTTP requests to the running `news_api` server. Measures application performance.

Two additional design changes:

3. **Multi-job bench specs** — A single bench run can define multiple search jobs (e.g., "PG native vs ParadeDB", "all engines vs ES"), each targeting a subset of engines. All results are gathered into one combined report.
4. **Decoupled pool and judgment** — TREC-style result pooling (retrieval of candidate docs) is separate from relevance grading (AI, manual, import). You can pool once and judge with different strategies without re-running queries.

## Architecture

```
Suite YAML (queries + judgments)     Bench Spec YAML (jobs + engines)
         │                                     │
         └──────────┐           ┌──────────────┘
                    ▼           ▼
                   Runner (orchestration)
                    │
          ┌─────────┼──────────┐
          ▼         ▼          ▼
      PgExecutor  EsExecutor  APIExecutor
          │         │          │
          └─────────┼──────────┘
                    ▼
              BenchmarkResult
                    │
              ┌─────┼──────┐
              ▼     ▼      ▼
           Report  JSON   Pool → Judge
```

**Three config files:**
- **Suite YAML** (`configs/bench/fts_quality_v1.yaml`) — queries + judgments (what to test)
- **Bench spec YAML** (`configs/bench/spec.yaml`) — jobs + engine connections (how to test)
- CLI flags for overrides and mode selection

## Implementation Plan

### Step 1: Bench Spec (`internal/bench/spec/`)

Defines search jobs and engine connections.

**CREATE** `spec/types.go`
```go
type BenchSpec struct {
    Jobs    []Job              `yaml:"jobs"`
    Engines map[string]Engine  `yaml:"engines"`
    API     *APIConfig         `yaml:"api,omitempty"`
    Metrics MetricsConfig      `yaml:"metrics"`
    Runs    RunsConfig         `yaml:"runs"`
}

type Job struct {
    Name    string   `yaml:"name"`
    Suite   string   `yaml:"suite"`   // path to suite YAML
    Engines []string `yaml:"engines"` // references keys in BenchSpec.Engines
    Layer   string   `yaml:"layer"`   // raw, api, all
}

type Engine struct {
    Type       string `yaml:"type"`       // postgres, elasticsearch
    Connection string `yaml:"connection"` // conn string or base URL
    Index      string `yaml:"index,omitempty"` // ES index name
}

type APIConfig struct {
    BaseURL string `yaml:"base_url"`
}

type MetricsConfig struct {
    KValues            []int `yaml:"k_values"`
    MaxK               int   `yaml:"max_k"`
    RelevanceThreshold int   `yaml:"relevance_threshold"`
}

type RunsConfig struct {
    Warmup     int `yaml:"warmup"`
    Iterations int `yaml:"iterations"`
}
```

**CREATE** `spec/loader.go` — `LoadFromFile(path)`, validation

Example `configs/bench/spec.yaml`:
```yaml
engines:
  pg-native:
    type: postgres
    connection: "postgresql://news_user:news_password@localhost:54320/news_db"
  paradedb:
    type: postgres
    connection: "postgresql://news_user:news_password@localhost:54321/news_db"
  elasticsearch:
    type: elasticsearch
    connection: "http://localhost:9200"
    index: news

api:
  base_url: "http://localhost:8080"

metrics:
  k_values: [3, 5, 10]
  max_k: 100
  relevance_threshold: 1

runs:
  warmup: 1
  iterations: 3

jobs:
  - name: "pg-native-vs-paradedb"
    suite: configs/bench/fts_quality_v1.yaml
    engines: [pg-native, paradedb]
    layer: raw

  - name: "all-engines-raw"
    suite: configs/bench/fts_quality_v1.yaml
    engines: [pg-native, paradedb, elasticsearch]
    layer: raw

  - name: "api-layer"
    suite: configs/bench/fts_quality_v1.yaml
    engines: [pg-native, elasticsearch]
    layer: api
```

### Step 2: Engine Package (`internal/bench/engine/`)

**DELETE** `engine/engine.go`, `engine/adapter.go`

**CREATE** `engine/executor.go`
```go
type Executor interface {
    Execute(ctx context.Context, rawQuery string) (*Execution, error)
    Name() string
    Close() error
}

type Execution struct {
    RankedDocIDs []uuid.UUID
    TotalMatches int64
    Latency      time.Duration
}
```

**CREATE** `engine/pg_executor.go`
- Wraps `storage.RawExecutor` (from `internal/storage/pg/executor.go`)
- `Execute()` calls `rawExecutor.Exec(ctx, sql, nil, opts)`, extracts `"id"` from rows
- Handles pgx UUID types (`[16]byte` -> `uuid.UUID`)
- Constructor: `NewPgExecutor(name string, pool *pg.ConnectionPool) *PgExecutor`

**CREATE** `engine/es_executor.go`
- `net/http` directly: `POST /{index}/_search` with raw JSON body
- Parses `hits.hits[]._source.id` (per `es/document.go:13`)
- Constructor: `NewEsExecutor(name string, baseURL string, index string) *EsExecutor`

**CREATE** `engine/api_executor.go`
- `ExecuteAPI(ctx, query *suite.APIQuery) (*Execution, error)`
- Parses `dto.SearchResponse` JSON, extracts `hits[].article.id`
- Constructor: `NewAPIExecutor(baseURL string) *APIExecutor`

**CREATE** `engine/factory.go`
- `CreateFromSpec(engines map[string]spec.Engine) (map[string]Executor, func(), error)`
- Builds executors based on engine type, manages connection pool lifecycle

**CREATE** `engine/pg_executor_test.go`

### Step 3: Suite Types (`internal/bench/suite/`)

**REWRITE** `suite/types.go`
- Remove: `Kind`, `QueryStringSpec`, `MatchSpec`, `MultiMatchSpec`, `PhraseSpec`, `BooleanSpec`, `query` import
- Keep: `RelevanceJudgment`, `JudgmentMap()`

```go
type TestSuite struct {
    Name        string     `yaml:"name"`
    Description string     `yaml:"description"`
    Version     string     `yaml:"version"`
    RawQueries  []RawQuery `yaml:"raw_queries"`
    APIQueries  []APIQuery `yaml:"api_queries"`
}

type RawQuery struct {
    ID          string              `yaml:"id"`
    Description string              `yaml:"description"`
    Engines     map[string]string   `yaml:"engines"`
    Judgments   []RelevanceJudgment `yaml:"judgments"`
}

type APIQuery struct {
    ID          string              `yaml:"id"`
    Description string              `yaml:"description"`
    Method      string              `yaml:"method"`
    Path        string              `yaml:"path"`
    Body        string              `yaml:"body,omitempty"`
    Params      map[string]string   `yaml:"params,omitempty"`
    Headers     map[string]string   `yaml:"headers,omitempty"`
    Backends    []string            `yaml:"backends"`
    Judgments   []RelevanceJudgment `yaml:"judgments"`
}
```

**REWRITE** `suite/loader.go` — remove `ToDomainQuery()`, update validation
**REWRITE** `suite/loader_test.go` — tests for new format

### Step 4: Pool Package (`internal/bench/pool/`)

TREC-style result pooling — retrieval of candidate docs only.

**CREATE** `pool/pooler.go`
```go
// PoolResults merges top-K docs from all engine executions into a
// deduplicated candidate set for relevance judgment.
func PoolResults(results map[string]*engine.Execution, depth int) []uuid.UUID

// PoolFile is the serializable output of a pooling run.
type PoolFile struct {
    SuiteName string       `yaml:"suite_name"`
    Queries   []PoolEntry  `yaml:"queries"`
}

type PoolEntry struct {
    QueryID   string      `yaml:"query_id"`
    QueryDesc string      `yaml:"query_desc"`
    Docs      []PooledDoc `yaml:"docs"`
}

type PooledDoc struct {
    DocID   uuid.UUID `yaml:"doc_id"`
    Sources []string  `yaml:"sources"` // which engines returned this doc
}
```

**CREATE** `pool/writer.go` — `WritePoolFile(poolFile, path)` writes YAML
**CREATE** `pool/reader.go` — `ReadPoolFile(path)` reads YAML
**CREATE** `pool/pooler_test.go`

### Step 5: Judgment Package (`internal/bench/judgment/`)

Separate from pooling. Takes a pool file and produces graded judgments.

**CREATE** `judgment/types.go`
```go
// Judge produces relevance grades for pooled documents.
type Judge interface {
    Grade(ctx context.Context, entry pool.PoolEntry) ([]GradedDoc, error)
}

type GradedDoc struct {
    DocID uuid.UUID `yaml:"doc_id"`
    Grade int       `yaml:"grade"` // 0-3
}

// JudgmentFile is the output of a judgment pass.
type JudgmentFile struct {
    Strategy string            `yaml:"strategy"` // "manual", "ai", "import"
    Queries  []JudgmentEntry   `yaml:"queries"`
}

type JudgmentEntry struct {
    QueryID string      `yaml:"query_id"`
    Docs    []GradedDoc `yaml:"docs"`
}
```

**CREATE** `judgment/manual.go`
- `ExportForAnnotation(poolFile, outputPath)` — writes YAML template with `grade: -1` placeholders
- `ImportAnnotations(path) (*JudgmentFile, error)` — reads completed manual annotations

**CREATE** `judgment/merger.go`
- `MergeIntoSuite(judgmentFile, suite) *suite.TestSuite` — populates suite queries' `Judgments` from a judgment file

### Step 6: Runner (`internal/bench/runner/`)

**MODIFY** `runner/config.go`
- Add `Layer`, `WarmupRuns`, `Runs` fields
- Add `Layer` type with constants (`raw`, `api`, `all`)

**MODIFY** `runner/result.go`
- Remove `QueryKind query.Kind`, add `Layer string`, add `JobName string`
- `BenchmarkResult.Results` keyed by `[jobName][queryID][engineName]`

**REWRITE** `runner/runner.go`
```go
type Runner struct {
    config Config
}

// RunJob executes a single search job against its engines.
func (r *Runner) RunJob(ctx context.Context, job spec.Job, s *suite.TestSuite,
    executors map[string]engine.Executor, apiExec *engine.APIExecutor) (*JobResult, error)

// RunAll executes all jobs from a bench spec.
func (r *Runner) RunAll(ctx context.Context, bs *spec.BenchSpec,
    executors map[string]engine.Executor, apiExec *engine.APIExecutor) (*BenchmarkResult, error)
```

`RunJob()` logic:
- Filter executors to only those listed in `job.Engines`
- Raw layer: iterate `suite.RawQueries`, warmup + runs, median latency, compute metrics
- API layer: iterate `suite.APIQueries`, warmup + runs, compute metrics
- Returns `*JobResult` with per-query results for this job

`RunAll()`: iterates jobs, loads each suite, calls `RunJob()`, merges into `BenchmarkResult`

### Step 7: Report (`internal/bench/report/`)

**MODIFY** `report/types.go` — add `Layer`, `JobName` to `Entry`/`AggregatedEntry`
**MODIFY** `report/aggregate.go` — group by (job, engine, layer) before aggregating
**MODIFY** `report/table.go` — section per job, layer subsections when both present
**CREATE** `report/json.go` — `WriteJSON(report, path)`

### Step 8: CLI (`cmd/bench/`)

**REWRITE** `cmd/bench/config.go`
- Primary flag: `--spec` (path to bench spec YAML)
- Override flags: `--suite`, `--pg`, `--es-addresses`, `--es-index`, `--k`, `--max-k` (for quick single-job runs without a spec file)
- New: `--layer`, `--api-url`, `--warmup`, `--runs`, `--output`
- Mode: `--mode bench` (default), `--mode pool`, `--mode judge`

**REWRITE** `cmd/bench/main.go`
- If `--spec` provided: load spec, create executors from spec engines, run all jobs
- If no `--spec`: build a single-job spec from CLI flags (backward-compatible quick mode)
- Mode `pool`: run queries, export pool file
- Mode `judge`: export annotation template from pool file (or later: AI grading)

```bash
# Multi-job via spec
./bin/bench --spec configs/bench/spec.yaml --output results/

# Quick single-job (no spec file needed)
./bin/bench --pg "postgresql://..." --suite configs/bench/fts_quality_v1.yaml

# Pool then judge workflow
./bin/bench --mode pool --spec configs/bench/spec.yaml --output pool.yaml
./bin/bench --mode judge --pool pool.yaml --output judgments.yaml
# Edit judgments.yaml manually or via AI...
# Then re-run with populated judgments
./bin/bench --spec configs/bench/spec.yaml --output results/
```

### Step 9: Suite YAML (`configs/bench/fts_quality_v1.yaml`)

Rewrite existing `configs/benchmark/fts_quality_v1.yaml` in new format. ~10 starter queries with `raw_queries` + `api_queries`, empty judgments.

## Files Summary

| Action | File |
|--------|------|
| Keep | `internal/bench/metrics/*` |
| Create | `internal/bench/spec/types.go` |
| Create | `internal/bench/spec/loader.go` |
| Delete | `internal/bench/engine/engine.go` |
| Delete | `internal/bench/engine/adapter.go` |
| Create | `internal/bench/engine/executor.go` |
| Create | `internal/bench/engine/pg_executor.go` |
| Create | `internal/bench/engine/es_executor.go` |
| Create | `internal/bench/engine/api_executor.go` |
| Create | `internal/bench/engine/factory.go` |
| Create | `internal/bench/engine/pg_executor_test.go` |
| Rewrite | `internal/bench/suite/types.go` |
| Rewrite | `internal/bench/suite/loader.go` |
| Rewrite | `internal/bench/suite/loader_test.go` |
| Create | `internal/bench/pool/pooler.go` |
| Create | `internal/bench/pool/writer.go` |
| Create | `internal/bench/pool/reader.go` |
| Create | `internal/bench/pool/pooler_test.go` |
| Create | `internal/bench/judgment/types.go` |
| Create | `internal/bench/judgment/manual.go` |
| Create | `internal/bench/judgment/merger.go` |
| Modify | `internal/bench/runner/config.go` |
| Modify | `internal/bench/runner/result.go` |
| Rewrite | `internal/bench/runner/runner.go` |
| Modify | `internal/bench/report/types.go` |
| Modify | `internal/bench/report/aggregate.go` |
| Modify | `internal/bench/report/table.go` |
| Create | `internal/bench/report/json.go` |
| Rewrite | `cmd/bench/config.go` |
| Rewrite | `cmd/bench/main.go` |
| Create | `configs/bench/spec.yaml` |
| Create | `configs/bench/fts_quality_v1.yaml` |

## Existing Code Referenced

- `internal/storage/executor.go` — `RawExecutor` interface (reused by PgExecutor)
- `internal/storage/pg/executor.go` — `pg.RawExecutor` impl (wraps pgxpool)
- `internal/storage/pg/pool.go` — `ConnectionPool`
- `internal/storage/es/document.go:13` — ES stores `id` as string in `_source`
- `internal/api/dto/query.go` — `SearchResponse` for API executor parsing
- `configs/benchmark/fts_quality_v1.yaml` — existing queries to migrate

## Verification

1. `go test ./internal/bench/...` — all tests pass
2. `go build ./cmd/bench` — compiles
3. `go vet ./...` — no issues
4. Single-job quick mode: `./bin/bench --pg "postgresql://..." --suite configs/bench/fts_quality_v1.yaml`
5. Multi-job spec: `./bin/bench --spec configs/bench/spec.yaml --output results/`
6. Pool workflow: `./bin/bench --mode pool --spec configs/bench/spec.yaml --output pool.yaml`
7. Spot-check: query with known relevant docs at top -> NDCG@K ~ 1.0
