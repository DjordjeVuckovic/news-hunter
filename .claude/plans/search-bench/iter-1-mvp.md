# Iteration 1 — MVP: Working End-to-End Pipeline

**Goal:** `make run-benchmark` prints a comparison table to stdout for PG native vs ES.

## Scope

- `internal/benchmark/metrics/` — All 5 IR metrics (NDCG, MAP, MRR, P@K, R@K)
- `internal/benchmark/suite/` — Test suite types + YAML loader (all 5 query kinds)
- `internal/benchmark/engine/` — Thin adapter over `storage.FtsSearcher`
- `internal/benchmark/runner/` — Orchestration: queries x engines x metrics
- `internal/benchmark/report/` — Table-only report via `text/tabwriter`
- `cmd/benchmark/` — CLI entry point with flag parsing
- `configs/benchmark/fts_quality_v1.yaml` — ~10 starter queries with hardcoded judgments
- `Makefile` — `build-benchmark`, `run-benchmark` targets

## Files

| File | Purpose |
|------|---------|
| `internal/benchmark/metrics/metrics.go` | ScoreSet, ComputeAll, grade constants |
| `internal/benchmark/metrics/ndcg.go` | NDCG@K |
| `internal/benchmark/metrics/precision.go` | P@K, R@K, F1@K |
| `internal/benchmark/metrics/ranking.go` | AP, RR |
| `internal/benchmark/metrics/metrics_test.go` | Table-driven unit tests |
| `internal/benchmark/suite/types.go` | TestSuite, BenchmarkQuery, spec types |
| `internal/benchmark/suite/loader.go` | YAML loader + domain query conversion |
| `internal/benchmark/suite/loader_test.go` | Loader tests |
| `internal/benchmark/engine/engine.go` | SearchEngine type |
| `internal/benchmark/engine/adapter.go` | Execute dispatch + extractDocIDs |
| `internal/benchmark/runner/config.go` | Config with defaults |
| `internal/benchmark/runner/result.go` | Result types |
| `internal/benchmark/runner/runner.go` | Orchestration loop |
| `internal/benchmark/report/types.go` | Report types |
| `internal/benchmark/report/aggregate.go` | Mean computation |
| `internal/benchmark/report/table.go` | Text table output |
| `cmd/benchmark/main.go` | CLI entry point |
| `cmd/benchmark/config.go` | Flag parsing |
| `configs/benchmark/fts_quality_v1.yaml` | Starter suite (~10 queries) |

## Verification

1. `go test ./internal/benchmark/metrics/...` — metric unit tests pass
2. `go test ./internal/benchmark/suite/...` — loader tests pass
3. `go build ./cmd/benchmark` — compiles
4. `docker-compose up -d && make run-benchmark` — prints comparison table