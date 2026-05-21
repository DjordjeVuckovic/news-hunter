# fts_quality

A self-contained search benchmark workspace.

## Layout

```
fts_quality/
  spec.yaml          # engines + jobs
  suite.yaml         # query suite (templates + queries)
  trec/
    pool.yaml        # candidate pool (generated)
    annotations.yaml # relevance grades (generated/edited)
    qrels.tsv        # TREC qrels (optional, exported)
```

## Workflow

```bash
# 1. Dry-run every query through each engine — catch broken templates early.
bench validate --spec fts_quality/spec.yaml

# 2. Generate the candidate pool from all engines.
bench pool --spec fts_quality/spec.yaml --output fts_quality/trec/pool.yaml

# 3. Grade the pool. Pick one:
#    keyword     — fast deterministic baseline
#    claude-api  — Anthropic Messages API in batches (set ANTHROPIC_API_KEY)
#    claude-cli  — `claude -p` per batch
#    stub        — write -1 placeholders for manual grading
bench judge \
  --pool fts_quality/trec/pool.yaml \
  --strategy claude-api \
  --pg "$PG_CONNECTION_STRING" \
  --output fts_quality/trec/annotations.yaml

# 4. Run the full benchmark — judgments auto-loaded via suite.judgments_file.
bench run --spec fts_quality/spec.yaml

# 5. (Optional) Export TREC qrels for trec_eval and friends.
bench qrels \
  --judgments fts_quality/trec/annotations.yaml \
  --output fts_quality/trec/qrels.tsv

# Inspect intermediates at any time:
bench show pool       fts_quality/trec/pool.yaml
bench show judgments  fts_quality/trec/annotations.yaml
bench show spec       fts_quality/spec.yaml
```

## Resume

`bench judge` is resumable: re-run with `--resume` and the same `--output` to
skip docs already graded.

## Editing judgments by hand

Open `trec/annotations.yaml` and adjust `grade:` values (0–3). `bench run`
will pick up your edits on the next invocation.
