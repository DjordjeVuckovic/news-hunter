# Search Terminology

## Current Implementation

### Full-Text Search (`SearchFullText`)

**What it does:**
- Analyzes and tokenizes text into searchable terms
- Performs relevance ranking using scoring algorithms
- Searches across multiple fields (title, description, content)
- Handles stemming, stop words, and text normalization

**Implementation:**
- **PostgreSQL**: Uses `tsvector` with `plainto_tsquery` and `ts_rank` for scoring
- **Elasticsearch**: Uses `MultiMatchQuery` with field boosting (title^3, description^2, content)

**Example queries:**
```
"climate change"        → Matches: "Climate", "changing", "climatic"
"Trump fraud trial"     → Matches across title, description, content with relevance ranking
"renewable energy"      → Tokenizes and searches for both terms
```

**API Interface:**
```go
package storage
SearchFullText(ctx context.Context, query string, page int, size int) (*SearchResult, error)
```

---

## Future Search Types

### Boolean Search (Planned)
- Uses logical operators: AND, OR, NOT
- Example: `"climate AND change OR warming NOT denial"`

### Advanced/Structured Search (Planned)
- Field-specific queries with structured syntax
- Example: `"title:Trump AND content:fraud AND published_at:[2024-01-01 TO 2024-12-31]"`

### Keyword Filtering (Planned)
- Exact match filtering on specific fields
- Example: `category="politics" AND source="CNN"`
- Uses `keyword` field type in Elasticsearch

---

## Key Concepts

### Full-Text Search vs Keyword Search

| Aspect | Full-Text Search | Keyword Search |
|--------|------------------|----------------|
| **Matching** | Analyzed, tokenized, fuzzy | Exact, case-sensitive |
| **Use Case** | Content search, articles | Filters, IDs, tags |
| **Field Type** | `text` (ES), `tsvector` (PG) | `keyword` (ES) |
| **Example** | "Climate Change" matches "climatic" | author.keyword = "John Smith" |

### Relevance Ranking

Both implementations calculate relevance scores:
- **PostgreSQL**: `ts_rank()` returns 0-1 normalized scores
- **Elasticsearch**: BM25 algorithm returns scores typically 1-30+

**Note:** Scores are not directly comparable between backends. See `docs/FUTURE_WORK.md` for ranking unification roadmap.

### Field Boosting (Elasticsearch)

Different fields have different importance weights:
```
title^3        → Title matches are 3x more important
description^2  → Description matches are 2x more important
content        → Content matches have base weight
```

---

## Search Quality

Current full-text search provides:
- ✅ Relevance ranking based on term frequency and field importance
- ✅ Multi-field matching with configurable weights
- ✅ Language-specific text analysis
- ✅ Pagination with accurate total counts
- ⚠️ Different ranking scales between PG and ES (see FUTURE_WORK.md)

---

*Last updated: 2025-11-02*
