# bench — IR Benchmark CLI

`bench` evaluates full-text, vector, and hybrid search queries against multiple engines (PostgreSQL variants, Elasticsearch, the news-hunter API), computes IR quality metrics and latency statistics, and writes self-attesting JSON reports.

## Track convention

Everything lives in a self-contained **track folder**:

```
tracks/<name>/
  spec.yaml          # engines, jobs, metrics config, defaults
  suite.yaml         # queries and per-engine templates
  trec/
    pool.yaml                       # candidate docs (bench pool output)
    annotations.<strategy>.yaml     # relevance grades (bench judge output)
    qrels.<strategy>.tsv            # TREC qrels export (bench qrels output)
  reports/
    <run_id>.json                   # one per bench run
    latest.json                     # pointer to most recent report
```

One track, multiple judgment strategies living side by side. Switch strategies with `--judgments <name>` on `bench run` — no YAML editing required.

## Pipeline

```
bench init <name>           1. scaffold tracks/<name>/
bench validate [<name>]     2. dry-run all queries against all engines
bench pool     [<name>]     3. gather candidate docs → trec/pool.yaml
bench judge    [<name>] --strategy <S>
                            4. grade pool → trec/annotations.<S>.yaml
bench run      [<name>]     5. execute suite + compute metrics → reports/
bench qrels    [<name>]     6. export TREC qrels (optional)
bench show     report|pool|judgments|spec [<name>]
                            inspect any artifact
```

Every command accepts a track name as a positional arg, a `--track` flag, or resolves from the current directory if you `cd tracks/<name>`.

## Strategy taxonomy

| Strategy | Class | Status | Description |
|----------|-------|--------|-------------|
| `lexical` | Heuristic | ✅ | Token-overlap baseline — fast, deterministic, no network |
| `bm25` | Heuristic | Reserved | BM25 score + threshold → grade |
| `vector` | Heuristic | Reserved | Cosine similarity + threshold → grade |
| `hybrid` | Heuristic | Reserved | Weighted combination |
| `claude-cli` | LLM | ✅ | `claude -p` subprocess per batch |
| `claude-api` | LLM | ✅ | Anthropic Messages API per batch |
| `manual` | Human | ✅ | Emits `grade: -1` placeholders for hand-grading |

File convention: `trec/annotations.<strategy>.yaml`, `trec/qrels.<strategy>.tsv`.

## Schema v1

Every produced artifact carries `schema_version: 1` and a `meta:` block. The meta block is the artifact's identity card — it records `run_id`, `tool` (with git sha), `generated_at`, and artifact-specific provenance (spec_id, strategy, judge_model, judge_prompt_version, sources).

Loading any artifact without `schema_version: 1` is a hard error — there is no silent tolerance.

## Command reference

### `bench init <name>`

Scaffolds `tracks/<name>/` with `spec.yaml`, `suite.yaml`, `trec/`, `reports/`, and `README.md`.

### `bench validate [<name>]`

Dry-runs every query through every engine using the engine's native validation endpoint (PostgreSQL `EXPLAIN`, Elasticsearch `_validate/query`). Reports per-query pass/fail — no data is stored.

### `bench pool [<name>] [--depth N]`

Runs all queries, gathers the top-N results per engine, deduplicates by doc ID, and writes `trec/pool.yaml`. Default depth is from `spec.defaults.pool_depth`.

### `bench judge [<name>] --strategy <S>`

Grades every `(query, doc)` pair in the pool using the chosen strategy. Output: `trec/annotations.<S>.yaml`.

Key flags:
- `--resume` — skip docs already graded (errors if model or prompt version changed)
- `--batch N` — override LLM batch size
- `--concurrency N` — parallel Grade calls (per-doc mode)

### `bench run [<name>] [--judgments <S|path>]`

Executes the suite against all engines, computes IR metrics and latency, writes `reports/<run_id>.json` and updates `reports/latest.json`.

Judgments resolution order:
1. `--judgments <strategy|path>` (CLI flag)
2. `spec.defaults.judgments` (per-track default)
3. None → latency-only report, warning printed

### `bench qrels [<name>] [--strategy <S>]`

Exports `trec/qrels.<S>.tsv` in standard TREC format for use with `trec_eval` or `pytrec_eval`.

### `bench show <subcommand> [<name>|path]`

Pretty-prints a one-page summary of any artifact:

| Subcommand | Reads |
|-----------|-------|
| `show spec` | `spec.yaml` |
| `show pool` | `trec/pool.yaml` |
| `show judgments [--strategy S]` | `trec/annotations.<S>.yaml` |
| `show report` | `reports/latest.json` → actual report |

## Metrics

All metrics computed per-query then averaged across judged queries:

| Metric | Description |
|--------|-------------|
| `NDCG@k` | Normalized Discounted Cumulative Gain — primary quality signal |
| `P@k` | Precision at k |
| `R@k` | Recall at k |
| `F1@k` | Harmonic mean of P@k and R@k |
| `MAP` | Mean Average Precision |
| `MRR` | Mean Reciprocal Rank |
| `Bpref` | Binary preference — robust to incomplete judgments |

Latency: per-engine p50/p75/p90/p95/p99 across all queries.

## Artifacts

All artifacts are self-attesting. A report's `provenance.sources` block records the exact paths of the spec, suite, pool, and judgments files used — you can reconstruct any run from the report alone.

For per-track documentation, see `tracks/<name>/README.md`.
