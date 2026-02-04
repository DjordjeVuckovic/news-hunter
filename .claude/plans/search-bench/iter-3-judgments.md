# Iteration 3 — Judgment Workflow

**Goal:** Systematic relevance annotation using TREC-style pooling + LLM assist.

## Scope

- `internal/bench/judgment/pooler.go` — TREC-style result pooling (merge top-K from all engines)
- `internal/bench/judgment/exporter.go` — Export pool for annotation (TSV/CSV)
- `internal/bench/judgment/importer.go` — Import completed annotations back into suite YAML
- Documented workflow for manual + LLM-assisted annotation

## Files

| File | Purpose |
|------|---------|
| `internal/bench/judgment/pooler.go` | TREC-style result pooling |
| `internal/bench/judgment/exporter.go` | Export pool for annotation |
| `internal/bench/judgment/importer.go` | Import completed annotations |