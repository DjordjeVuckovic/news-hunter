---
name: bench-annotate
description: Annotate / grade news articles for the IR benchmark when a human-in-the-loop pass is needed. Use this when asked to grade, annotate, label, or evaluate relevance for the bench pool — typically reviewing or overriding entries in a judgments YAML produced by `bench judge`.
---

# Bench Annotate

Grade news articles for relevance against IR benchmark queries.

## When to use this skill

You should use this skill when:
- The user asks you to **review** an `annotations_*.yaml` produced by the `bench judge` CLI and adjust grades you disagree with
- The user asks you to **fill in** a `stub` judgments file (every doc has `grade: -1`)
- The user explicitly wants a Claude pass instead of `--strategy claude-cli` / `claude-api`

For fully automated grading without human review, prefer the `bench judge` CLI — it streams one doc at a time, avoiding context-window issues.

## Grading Scale

- **3** — Highly relevant: article is centrally about the query topic
- **2** — Relevant: article clearly covers the topic
- **1** — Marginal: topic is mentioned but not the focus
- **0** — Not relevant: article is about something else

When in doubt between two grades, pick the lower one. Be consistent: the same article quality relative to the same query should always get the same grade.

## Workflow (per-query, never load the whole pool)

The pool file is large. Work one query at a time:

1. **Read the pool file header** (~30 lines) to get the suite name and confirm structure.
2. **For each query in the pool:**
   1. Use `grep -n "query_id: <id>"` on the pool YAML to find the block
   2. Extract the `doc_id`s for that query
   3. For each `doc_id`, fetch the article from Postgres if you need the content:
      ```sql
      SELECT id, title, description, content FROM articles WHERE id = '<uuid>';
      ```
      (Alternatively, ask the user to point you at an enriched per-query file.)
   4. Grade every doc using the scale above. Base your grade on title + description + content snippet together.
3. **Write the JudgmentFile** to the output path with grades 0–3 — never `-1`.

## Output schema

Match `internal/bench/judgment/types.go` exactly:

```yaml
strategy: llm        # or "manual" if you reviewed/overrode a baseline
queries:
  - query_id: qs-climate
    docs:
      - doc_id: 493384af-c5fc-4677-ad11-2081e73f0588
        grade: 2
      - doc_id: 413b1526-e74b-4ee1-95f4-4a30580fcf09
        grade: 3
```

Rules:
- Every `doc_id` from the input pool must appear in the output — no skips
- Grades are `0`, `1`, `2`, or `3` — never `-1` in the final output
- Suite YAML's `judgments_file:` should point at this file so `bench run` picks it up

## Default paths

- Pool input: `configs/bench/trec/pool_v1.yaml`
- Output: `configs/bench/trec/annotations_v1.yaml` (or `annotations_llm_v1.yaml` for a Claude pass)

Override with paths the user provides.

## After writing

Confirm: count of queries written, total docs graded, output path. Suggest running `bench run --spec configs/bench/spec.yaml` to consume the judgments.