# Follow-up plan: Parquet as the canonical dataset format

## Motivation
The canonical (post-preprocessor) dataset is currently JSONL/CSV, read row-wise
into `map[string]string` and mapped via `ArticleDirectMapper`. For the thesis'
scaling analysis (multiple corpus sizes) Parquet is a strict upgrade:
- typed schema → decode straight into `document.Article`, no stringly-typed map,
  no CSV quoting/escaping bugs;
- columnar + zstd → smaller files, faster ingest;
- column pruning → read only needed fields;
- one toolchain (`parquet-go`) shared with the embeddings loader, and both keyed
  on the same `id` column so corpus + vectors line up.

Scope this as a **separate PR** from the embeddings loader to keep reviews tight.
`parquet-go` will already be a dependency after the embeddings work.

## Boundary
- Raw heterogeneous Kaggle sources → keep CSV/JSONL + YAML field-mapping
  (`ArticleMapper`). Parquet is for the **canonical interchange format only**:
  `raw CSV → mapper → canonical Parquet → ds_ingest`.

## Changes
1. `internal/ingest/reader/parquet_reader.go` — typed reader producing
   `document.Article` directly (implements the reader contract used by the
   collector; bypasses `map[string]string`). Streaming via parquet-go
   `GenericReader[articleRow]`.
2. `cmd/ds_ingest/main.go` — add `.parquet` case to the extension switch.
   When the dataset is canonical Parquet, skip the YAML mapper entirely.
3. `cmd/preprocessor` — emit `.parquet` as canonical output (alongside or
   instead of JSONL).
4. Schema: define the canonical Parquet schema mirroring `document.Article`
   (id, title, subtitle, content, url, source, language, createdAt, …).
   Document it next to the data-mapping schema.

## Open questions
- Keep JSONL canonical output too (for human inspection / diffing) or fully
  switch to Parquet?
- Reuse the `reflect`-based mapper for Parquet rows, or hand-write the struct
  mapping (cleaner, type-safe)? Recommend hand-written struct + parquet tags.

## Tests
- Round-trip: write canonical Parquet fixture → read → assert `document.Article`
  fields; include a NULL/optional-field case.
