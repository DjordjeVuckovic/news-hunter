# Iter 4: Complete the Pool → Judge → Merge → Bench CLI Workflow

## Goal
Wire up the missing CLI steps so the full judgment workflow runs end-to-end with simple commands.
Judgments live in a separate file — suite YAML stays clean and readable.

## Design: `judgments_file` reference

Suite YAML references a separate judgments file:
```yaml
# fts_quality_v1.yaml — stays readable
name: "FTS Quality Benchmark v1"
judgments_file: configs/bench/trec/annotations_v1.yaml
queries:
  - id: qs-climate
    # no inline judgments — loaded from judgments_file at runtime
```

The suite loader reads `judgments_file` and maps grades to queries by `query_id`.
`--mode merge` writes/updates the judgments file only — never touches the suite YAML structure.

## Desired Command Sequence

```bash
# 1. Generate pool from all engines
./bin/bench --mode pool --spec configs/bench/spec.yaml \
  --output configs/bench/trec/pool_v1.yaml

# 2. Export annotation template (grade=-1 placeholders)
./bin/bench --mode judge \
  --pool configs/bench/trec/pool_v1.yaml \
  --output configs/bench/trec/annotations_v1.yaml

# 3. [MANUAL] Open annotations_v1.yaml, fill in grades 0-3

# 4. Run full benchmark — suite loader picks up judgments_file automatically
./bin/bench --mode bench --spec configs/bench/spec.yaml

# 5. Optionally export TREC qrels for external validation (already built)
./bin/bench --mode qrels \
  --judgments configs/bench/trec/annotations_v1.yaml \
  --output configs/bench/trec/qrels_v1.tsv
```

Note: `--mode merge` is no longer needed — the suite loader reads `judgments_file` directly at bench time.

## Changes Required

### 1. `internal/bench/suite/types.go`
- Add `JudgmentsFile string` field to `TestSuite`: `yaml:"judgments_file,omitempty"`
- Remove `Judgments []RelevanceJudgment` from `Query` (no more inline judgments)
- `JudgmentMap()` on `Query` now returns empty map (judgments injected by loader)

### 2. `internal/bench/suite/loader.go`
- After parsing suite, if `JudgmentsFile != ""`:
  - Load `judgment.JudgmentFile` from that path
  - For each `JudgmentEntry`, find matching query by ID
  - Inject `RelevanceJudgment` slice into the query
- Keep inline `judgments:` support as fallback (backwards compat)

### 3. `configs/bench/fts_quality_v1.yaml`
- Add `judgments_file: configs/bench/trec/annotations_v1.yaml` at top
- Remove all `judgments: []` lines from queries (already done)

### 4. `Makefile`
Add convenience targets:
```makefile
pool-bench: build-bench
	./bin/bench --mode pool --spec configs/bench/spec.yaml \
	  --output configs/bench/trec/pool_v1.yaml

judge-bench: build-bench
	./bin/bench --mode judge \
	  --pool configs/bench/trec/pool_v1.yaml \
	  --output configs/bench/trec/annotations_v1.yaml

qrels-bench: build-bench
	./bin/bench --mode qrels \
	  --judgments configs/bench/trec/annotations_v1.yaml \
	  --output configs/bench/trec/qrels_v1.tsv
```

## Files Touched
| File | Action |
|------|--------|
| `internal/bench/suite/types.go` | Add `JudgmentsFile`, remove inline `Judgments` from `Query` |
| `internal/bench/suite/loader.go` | Load and inject judgments from `judgments_file` |
| `configs/bench/fts_quality_v1.yaml` | Add `judgments_file` ref, remove inline `judgments: []` |
| `Makefile` | Add 3 convenience targets |

## Notes
- `--mode merge` dropped — no longer needed with this design
- Suite file is now purely query definitions, never auto-modified
- Multiple suites can reference the same judgments file
- Swap judgment sets by changing `judgments_file` path (human vs LLM annotations)
- `cmd/bench/config.go` `JudgmentsPath` flag stays (used by `--mode qrels`)