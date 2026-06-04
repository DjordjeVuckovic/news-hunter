# Plan: Serbian full-text search support (deferred — post-v1)

## Status
Deferred. We WANT Serbian support but it is not the v1 focus. Captured here so
it is not lost.

## The problem (verified)
- `SupportedLanguages` accepts `"serbian"` (`internal/types/query/language.go`).
- The ingest trigger (`db/migrations/003_add_search_vector_trigger.up.sql`) and the
  native searcher (`internal/storage/pg/native/searcher.go:477,489,504,703`,
  `fts_helpers.go:113,118`) build `to_tsvector('serbian'::regconfig, …)` /
  `plainto_tsquery('serbian'::regconfig, …)` / `websearch_to_tsquery('serbian'…)`.
- **No `serbian` text-search configuration is created in any migration.** PostgreSQL
  ships no built-in `serbian` config, so any `language='serbian'` row (ingest) or
  query throws `text search configuration "serbian" does not exist`.
- Also cosmetic: the constant is misnamed `LanguageSpanish Language = "serbian"`.

Net: the "multi-language (English, Serbian)" claim in CLAUDE.md / docs is currently
false for PostgreSQL.

## What v1 does instead
- Rename `LanguageSpanish` → `LanguageSerbian` (value stays `"serbian"`).
- For v1, treat the corpus as English (the Kaggle GlobalNews dataset is English).
  Do NOT ingest `serbian` rows until the config below exists.

## Implementation when picked up
1. Migration `00X_create_serbian_tsconfig`:
   - Create a `serbian` text-search configuration. Options, simplest → richest:
     - **simple + unaccent**: copy `simple`, prepend `unaccent` dictionary →
       diacritic-insensitive, no stemming. Cheapest; good baseline for Serbian
       Latin/Cyrillic.
     - **snowball**: PG has no Serbian snowball stemmer built in; would need a
       custom dictionary (e.g. Hunspell `sr` affix/dic files) loaded via
       `CREATE TEXT SEARCH DICTIONARY … (TEMPLATE = ispell, …)`. Best quality.
   - Wire `unaccent` (already installed, currently unused) into the mapping so
     accents/Cyrillic-Latin variants normalise.
   - Re-populate `search_vector` for existing serbian rows.
2. Add an integration test: ingest a serbian doc, FTS roundtrip via the native
   searcher, assert hit + stemming/diacritic behaviour.
3. ES side: add a `serbian` analyzer to the index mapping for parity (ICU /
   custom analyzer); benchmark PG-serbian vs ES-serbian.
4. Flip docs/CLAUDE.md to truthfully claim Serbian once green.

## Caveats
- Cyrillic vs Latin Serbian: decide whether to transliterate at ingest or index
  both. `unaccent` does not transliterate scripts — a transliteration step
  (e.g. ICU) may be needed for cross-script recall.
