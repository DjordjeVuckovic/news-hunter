# Iteration 2 — Full Suite + JSON Reports

**Goal:** Production-quality test suite and machine-readable output for thesis data analysis.

## Scope

- Expand `configs/bench/fts_quality_v1.yaml` to ~35 queries across all paradigms
- `internal/bench/report/json.go` — JSON report output
- `internal/bench/report/table.go` — Per-query detail table mode
- `internal/bench/engine/factory.go` — `CreateEngines()` helper for cleaner multi-engine setup

## Files

| File | Purpose |
|------|---------|
| `configs/bench/fts_quality_v1.yaml` | Expanded to ~35 queries |
| `internal/bench/report/json.go` | JSON report output |
| `internal/bench/engine/factory.go` | Engine factory helper |
| `internal/bench/report/table.go` | (modified) per-query detail mode |