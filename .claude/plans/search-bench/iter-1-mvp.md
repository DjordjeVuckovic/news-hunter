# Iteration 1 — MVP: Working End-to-End Pipeline

**Goal:** `make run-bench` prints a comparison table to stdout for PG native vs ES.

## Scope

- `internal/bench/metrics/` — All 5 IR metrics (NDCG, MAP, MRR, P@K, R@K)
- `internal/bench/suite/` — Test suite types + YAML loader (all 5 query kinds)
- `internal/bench/engine/` — Thin adapter over `storage.FtsSearcher`
- `internal/bench/runner/` — Orchestration: queries x engines x metrics
- `internal/bench/report/` — Table-only report via `text/tabwriter`
- `cmd/bench/` — CLI entry point with flag parsing
- `configs/bench/fts_quality_v1.yaml` — ~10 starter queries with hardcoded judgments
- `Makefile` — `build-bench`, `run-bench` targets

## Files

| File                                     | Purpose |
|------------------------------------------|---------|
| `internal/bench/metrics/metrics.go`      | ScoreSet, ComputeAll, grade constants |
| `internal/bench/metrics/ndcg.go`         | NDCG@K |
| `internal/bench/metrics/precision.go`    | P@K, R@K, F1@K |
| `internal/bench/metrics/ranking.go`      | AP, RR |
| `internal/bench/metrics/metrics_test.go` | Table-driven unit tests |
| `internal/bench/suite/types.go`          | TestSuite, BenchmarkQuery, spec types |
| `internal/bench/suite/loader.go`         | YAML loader + domain query conversion |
| `internal/bench/suite/loader_test.go`    | Loader tests |
| `internal/bench/engine/engine.go`        | SearchEngine type |
| `internal/bench/engine/adapter.go`       | Execute dispatch + extractDocIDs |
| `internal/bench/runner/config.go`        | Config with defaults |
| `internal/bench/runner/result.go`        | Result types |
| `internal/bench/runner/runner.go`        | Orchestration loop |
| `internal/bench/report/types.go`         | Report types |
| `internal/bench/report/aggregate.go`     | Mean computation |
| `internal/bench/report/table.go`         | Text table output |
| `cmd/bench/main.go`                      | CLI entry point |
| `cmd/bench/config.go`                    | Flag parsing |
| `configs/bench/fts_quality_v1.yaml`      | Starter suite (~10 queries) |

## Verification

1. `go test ./internal/bench/metrics/...` — metric unit tests pass
2. `go test ./internal/bench/suite/...` — loader tests pass
3. `go build ./cmd/bench` — compiles
4. `docker-compose up -d && make run-bench` — prints comparison table