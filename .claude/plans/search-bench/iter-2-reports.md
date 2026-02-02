# Iteration 2 — Full Suite + JSON Reports

**Goal:** Production-quality test suite and machine-readable output for thesis data analysis.

## Scope

- Expand `configs/benchmark/fts_quality_v1.yaml` to ~35 queries across all paradigms
- `internal/benchmark/report/json.go` — JSON report output
- `internal/benchmark/report/table.go` — Per-query detail table mode
- `internal/benchmark/engine/factory.go` — `CreateEngines()` helper for cleaner multi-engine setup

## Files

| File | Purpose |
|------|---------|
| `configs/benchmark/fts_quality_v1.yaml` | Expanded to ~35 queries |
| `internal/benchmark/report/json.go` | JSON report output |
| `internal/benchmark/engine/factory.go` | Engine factory helper |
| `internal/benchmark/report/table.go` | (modified) per-query detail mode |