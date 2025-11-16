# Implementation Goal: Multi-Paradigm Search Architecture

> **Purpose**: Reference architecture for implementing multiple search paradigms (lexical, boolean, fuzzy, vector, hybrid) across different storage backends (PostgreSQL, Elasticsearch, Solr, etc.)

## Table of Contents
1. [Architecture Overview](#architecture-overview)
2. [Design Principles](#design-principles)
3. [Query Type Definitions](#query-type-definitions)
4. [Interface Segregation](#interface-segregation)
5. [Implementation Examples](#implementation-examples)
6. [API Design](#api-design)
7. [Capability Discovery](#capability-discovery)
8. [Testing Strategy](#testing-strategy)

---

## Architecture Overview

This architecture enables the News Hunter project to:
- Support **multiple search paradigms** as defined in `SEARCH_TERMINOLOGY.md`
- Allow **different storage backends** to implement only what they support
- Maintain **type safety** while providing runtime flexibility
- Enable **clear benchmarking** between PostgreSQL and Elasticsearch across comparable search paradigms

### Core Principle: Interface Segregation

Instead of forcing all storage backends to implement all search types, we use **interface segregation**:
- **Base interface** (`Reader`): Minimal required capability (lexical search)
- **Optional interfaces**: Each search paradigm has its own interface
- **Runtime capability detection**: Type assertions to check if backend supports specific features

---

## Design Principles

### 1. Interface Segregation Principle (ISP)
**Problem**: A single `Search(query SearchQuery)` method forces all backends to handle all query types.

**Solution**: Separate interface for each search paradigm.

```go
// ❌ BAD: Monolithic interface
type Reader interface {
    Search(ctx context.Context, query *SearchQuery, ...) (*SearchResult, error)
}
// Solr must implement vector search even if it doesn't support it

// ✅ GOOD: Segregated interfaces
type Reader interface {
    SearchFullText(...) (*SearchResult, error)  // Required
}

type VectorSearcher interface {
    SearchVector(...) (*SearchResult, error)  // Optional
}
// Solr only implements Reader, not VectorSearcher
```

### 2. Explicit is Better Than Implicit
Each search paradigm has:
- Dedicated query type struct
- Dedicated interface method
- Clear capability declaration

### 3. Compile-Time Safety, Runtime Flexibility
- Compile-time: All backends must implement `Reader` (lexical search)
- Runtime: Optional capabilities checked via type assertions

### 4. Academic Rigor
Clear separation enables:
- Benchmarking specific paradigms
- Comparing PG vs ES on same paradigm
- Documenting capability matrices

---

## Query Type Definitions

Located in `internal/types/query/query.go`:

```go
type QueryType string

const (
    QueryTypeFullText   QueryType = "full_text"
    QueryTypeBoolean     QueryType = "boolean"
    QueryTypeFieldLevel QueryType = "field_level"
    QueryTypeFuzzy      QueryType = "fuzzy"
    QueryTypeVector     QueryType = "vector"
    QueryTypeHybrid     QueryType = "hybrid"
)

// SearchQuery is the top-level query container
// Only one field should be non-nil based on Type
type SearchQuery struct {
    Type     QueryType
    FullText *FullTextQuery
    Boolean  *BooleanQuery
    Field    *FieldLevelQuery
    Fuzzy    *FuzzyQuery
    Vector   *VectorQuery
    Hybrid   *HybridQuery
}

// FullTextQuery: Token-based full-text search with relevance ranking
// PostgreSQL: plainto_tsquery + ts_rank
// Elasticsearch: multi_match with BM25
type FullTextQuery struct {
    Text string  // Plain text query: "climate change"
}

// BooleanQuery: Structured queries with logical operators
// PostgreSQL: to_tsquery with &, |, ! operators
// Elasticsearch: bool query with must, should, must_not
type BooleanQuery struct {
    Expression string  // "climate AND (change OR warming) AND NOT politics"
}

// FieldLevelQuery: Targeted search on specific fields
// PostgreSQL: Multiple tsvector columns or direct field matching
// Elasticsearch: field-specific queries in bool context
type FieldLevelQuery struct {
    Fields map[string]string  // {"title": "Trump", "description": "election"}
}

// FuzzyQuery: Typo-tolerant approximate matching
// PostgreSQL: pg_trgm similarity, Levenshtein distance
// Elasticsearch: fuzzy query with edit distance
type FuzzyQuery struct {
    Text      string
    Fuzziness int  // Edit distance threshold (0-2 typically)
    // Future: MinSimilarity, PrefixLength, etc.
}

// VectorQuery: Semantic search using embeddings
// PostgreSQL: pgvector with <=> (cosine), <-> (L2), <#> (dot product)
// Elasticsearch: kNN query with dense_vector
type VectorQuery struct {
    Embedding []float64  // Dense vector (384, 768, or 1536 dimensions typically)
    K         int        // Top-K nearest neighbors
    // Future: MinScore, IndexType (IVFFlat, HNSW), etc.
}

// HybridQuery: Combines multiple search paradigms
// Typically full-text + vector with score fusion (RRF or weighted)
type HybridQuery struct {
    FullText *FullTextQuery
    Vector   *VectorQuery
    Weights  HybridWeights
}

// HybridWeights: Score fusion configuration
type HybridWeights struct {
    Method        string   // "rrf" (Reciprocal Rank Fusion) or "weighted"
    FullTextWeight float64 // Weight for full-text scores (used in weighted fusion)
    VectorWeight  float64  // Weight for vector scores (used in weighted fusion)
    RRFConstant   int      // k parameter for RRF (typically 60)
}
```

### Query Type Usage Matrix

| Query Type   | PostgreSQL | Elasticsearch | Solr (hypothetical) |
|--------------|------------|---------------|---------------------|
| FullText     | ✅ tsvector | ✅ multi_match | ✅ standard         |
| Boolean      | ✅ tsquery  | ✅ bool query  | ✅ query parser     |
| Field-Level  | ✅ columns  | ✅ native      | ✅ field queries    |
| Fuzzy        | ✅ pg_trgm  | ✅ fuzzy       | ✅ fuzzy            |
| Vector       | ✅ pgvector | ✅ kNN         | ⚠️ limited          |
| Hybrid       | ✅ RRF SQL  | ✅ hybrid API  | ❌ not supported    |

---

## Interface Segregation

Located in `internal/storage/reader.go`:

```go
// Reader: Base interface - ALL storage backends must implement
type Reader interface {
    SearchFullText(ctx context.Context, query *query.FullTextQuery, cursor *query.Cursor, size int) (*SearchResult, error)
}

// Optional capability interfaces

type BooleanSearcher interface {
    SearchBoolean(ctx context.Context, query *query.BooleanQuery, cursor *query.Cursor, size int) (*SearchResult, error)
}

type FieldLevelSearcher interface {
    SearchFieldLevel(ctx context.Context, query *query.FieldLevelQuery, cursor *query.Cursor, size int) (*SearchResult, error)
}

type FuzzySearcher interface {
    SearchFuzzy(ctx context.Context, query *query.FuzzyQuery, cursor *query.Cursor, size int) (*SearchResult, error)
}

type VectorSearcher interface {
    SearchVector(ctx context.Context, query *query.VectorQuery, cursor *query.Cursor, size int) (*SearchResult, error)
}

type HybridSearcher interface {
    SearchHybrid(ctx context.Context, query *query.HybridQuery, cursor *query.Cursor, size int) (*SearchResult, error)
}
```

### Compile-Time Assertions

```go
// PostgreSQL - implements ALL interfaces
var _ Reader = (*pg.Reader)(nil)
var _ BooleanSearcher = (*pg.Reader)(nil)
var _ FieldLevelSearcher = (*pg.Reader)(nil)
var _ FuzzySearcher = (*pg.Reader)(nil)
var _ VectorSearcher = (*pg.Reader)(nil)
var _ HybridSearcher = (*pg.Reader)(nil)

// Elasticsearch - implements ALL interfaces
var _ Reader = (*es.Reader)(nil)
var _ BooleanSearcher = (*es.Reader)(nil)
var _ VectorSearcher = (*es.Reader)(nil)
// ... etc

// Hypothetical Solr - implements SUBSET
var _ Reader = (*solr.Reader)(nil)
var _ BooleanSearcher = (*solr.Reader)(nil)
var _ FuzzySearcher = (*solr.Reader)(nil)
// Does NOT implement VectorSearcher or HybridSearcher
```

---

## Implementation Examples

### PostgreSQL Reader

```go
// internal/storage/pg/pg_reader.go

type Reader struct {
    db *pgxpool.Pool
}

// Required: Base lexical search
func (r *Reader) SearchFullText(
    ctx context.Context,
    query *query.FullTextQuery,
    cursor *query.Cursor,
    size int,
) (*storage.SearchResult, error) {
    // Use plainto_tsquery + ts_rank
    // Current implementation (renamed from SearchFullText)
}

// Optional: Boolean search
func (r *Reader) SearchBoolean(
    ctx context.Context,
    query *query.BooleanQuery,
    cursor *query.Cursor,
    size int,
) (*storage.SearchResult, error) {
    // Parse boolean expression to PostgreSQL tsquery syntax
    // "climate AND change" → to_tsquery('english', 'climate & change')
    // Use ts_rank for scoring
}

// Optional: Field-level search
func (r *Reader) SearchFieldLevel(
    ctx context.Context,
    query *query.FieldLevelQuery,
    cursor *query.Cursor,
    size int,
) (*storage.SearchResult, error) {
    // Option 1: Use separate tsvector columns
    // WHERE title_vector @@ plainto_tsquery('Trump')
    //   AND description_vector @@ plainto_tsquery('election')

    // Option 2: Direct field matching
    // WHERE title ILIKE '%Trump%' AND description ILIKE '%election%'
}

// Optional: Fuzzy search
func (r *Reader) SearchFuzzy(
    ctx context.Context,
    query *query.FuzzyQuery,
    cursor *query.Cursor,
    size int,
) (*storage.SearchResult, error) {
    // Use pg_trgm extension
    // WHERE title % $1  -- similarity operator
    // ORDER BY similarity(title, $1) DESC
}

// Optional: Vector search
func (r *Reader) SearchVector(
    ctx context.Context,
    query *query.VectorQuery,
    cursor *query.Cursor,
    size int,
) (*storage.SearchResult, error) {
    // Use pgvector extension
    // ORDER BY embedding <=> $1::vector  -- cosine distance
    // LIMIT K
}

// Optional: Hybrid search
func (r *Reader) SearchHybrid(
    ctx context.Context,
    query *query.HybridQuery,
    cursor *query.Cursor,
    size int,
) (*storage.SearchResult, error) {
    // Implement RRF (Reciprocal Rank Fusion) in SQL
    // WITH lexical AS (...), semantic AS (...)
    // SELECT ..., (1/(60+lex_rank) + 1/(60+sem_rank)) as rrf_score
}
```

### Elasticsearch Reader

```go
// internal/storage/es/es_reader.go

type Reader struct {
    client    *elasticsearch.TypedClient
    indexName string
}

func (r *Reader) SearchFullText(...) (*storage.SearchResult, error) {
    // Use MultiMatchQuery with BM25
    // Current implementation (renamed from SearchFullText)
}

func (r *Reader) SearchBoolean(...) (*storage.SearchResult, error) {
    // Use BoolQuery with must, should, must_not clauses
    // Parse expression to ES bool query structure
}

func (r *Reader) SearchVector(...) (*storage.SearchResult, error) {
    // Use kNN query with dense_vector field
}

func (r *Reader) SearchHybrid(...) (*storage.SearchResult, error) {
    // Use ES native hybrid query with RRF
}
```

---

## API Design

### Endpoint: Single Search Endpoint with Query Type

**POST /v1/articles/search**

The API uses a single endpoint with a discriminated union based on `query.type`:

#### Full-Text Search Request
```json
{
  "type": "lexical",
  "query": {
    "text": "climate change renewable energy"
  },
  "cursor": "eyJzY29yZSI6MC45LCJpZCI6IjEyMyJ9",
  "size": 20
}
```

#### Boolean Search Request
```json
{
  "type": "boolean",
  "query": {
    "expression": "climate AND (change OR warming) AND NOT politics"
  },
  "size": 20
}
```

#### Field-Level Search Request
```json
{
  "type": "field_level",
  "query": {
    "fields": {
      "title": "Trump election",
      "description": "fraud investigation"
    }
  },
  "size": 20
}
```

#### Fuzzy Search Request
```json
{
  "type": "fuzzy",
  "query": {
    "text": "Obema",
    "fuzziness": 2
  },
  "size": 20
}
```

#### Vector Search Request
```json
{
  "type": "vector",
  "query": {
    "embedding": [0.123, 0.456, 0.789, ...],  // 768-dimensional vector
    "k": 10
  },
  "size": 20
}
```

#### Hybrid Search Request
```json
{
  "type": "hybrid",
  "query": {
    "lexical": {
      "text": "renewable energy"
    },
    "vector": {
      "embedding": [0.123, 0.456, ...],
      "k": 100
    },
    "weights": {
      "method": "rrf",
      "rrf_constant": 60
    }
  },
  "size": 20
}
```

### Response Format (Same for All Query Types)

```json
{
  "hits": [
    {
      "article": {
        "id": "uuid-here",
        "title": "Climate Change Impact on Renewable Energy",
        "description": "...",
        "content": "...",
        "author": "John Doe",
        "url": "https://...",
        "language": "en",
        "created_at": "2024-01-15T10:30:00Z",
        "metadata": {
          "source_id": "cnn",
          "source_name": "CNN",
          "published_at": "2024-01-15T08:00:00Z",
          "category": "Environment",
          "imported_at": "2024-01-15T10:00:00Z"
        }
      },
      "score": 0.95,
      "score_normalized": 1.0
    }
  ],
  "next_cursor": "eyJzY29yZSI6MC44NSwiaWQiOiJ1dWlkLTIifQ==",
  "has_more": true,
  "max_score": 0.95,
  "page_max_score": 0.95,
  "total_matches": 1234,
  "query_type": "lexical"  // Echoed back for client confirmation
}
```

### Error Responses

#### Unsupported Query Type
```json
HTTP 400 Bad Request
{
  "error": "boolean search not supported by current storage backend",
  "query_type": "boolean",
  "backend": "in_memory"
}
```

#### Invalid Query Structure
```json
HTTP 400 Bad Request
{
  "error": "invalid query: lexical query requires 'text' field",
  "query_type": "lexical"
}
```

---

## Capability Discovery

### Endpoint: GET /v1/capabilities

Returns which query types are supported by the current storage backend.

```json
{
  "supported_query_types": [
    "lexical",
    "boolean",
    "field_level",
    "fuzzy",
    "vector",
    "hybrid"
  ],
  "storage_backend": "postgresql",
  "version": "1.0.0",
  "features": {
    "cursor_pagination": true,
    "max_page_size": 10000,
    "vector_dimensions": [384, 768, 1536],
    "supported_languages": ["en", "es", "fr", "de"]
  }
}
```

### Router Implementation

```go
// internal/router/search.go

func (r *SearchRouter) searchHandler(c echo.Context) error {
    var req struct {
        Type   query.QueryType    `json:"type"`
        Query  json.RawMessage     `json:"query"`
        Cursor *string             `json:"cursor,omitempty"`
        Size   int                 `json:"size,omitempty"`
    }

    if err := c.Bind(&req); err != nil {
        return c.JSON(400, map[string]string{"error": "invalid request"})
    }

    // Parse cursor
    var cursor *query.Cursor
    if req.Cursor != nil {
        cursor, err = query.DecodeCursor(*req.Cursor)
        if err != nil {
            return c.JSON(400, map[string]string{"error": "invalid cursor"})
        }
    }

    size := req.Size
    if size <= 0 {
        size = 100
    }

    // Dispatch based on query type
    var result *storage.SearchResult
    var err error

    switch req.Type {
    case query.QueryTypeFullText:
        var fullTextQuery query.FullTextQuery
        if err := json.Unmarshal(req.Query, &fullTextQuery); err != nil {
            return c.JSON(400, map[string]string{"error": "invalid full-text query"})
        }
        result, err = r.storage.SearchFullText(ctx, &fullTextQuery, cursor, size)

    case query.QueryTypeBoolean:
        bs, ok := r.storage.(storage.BooleanSearcher)
        if !ok {
            return c.JSON(400, map[string]string{
                "error": "boolean search not supported",
                "backend": getBackendType(r.storage),
            })
        }
        var boolQuery query.BooleanQuery
        if err := json.Unmarshal(req.Query, &boolQuery); err != nil {
            return c.JSON(400, map[string]string{"error": "invalid boolean query"})
        }
        result, err = bs.SearchBoolean(ctx, &boolQuery, cursor, size)

    case query.QueryTypeVector:
        vs, ok := r.storage.(storage.VectorSearcher)
        if !ok {
            return c.JSON(400, map[string]string{
                "error": "vector search not supported",
                "backend": getBackendType(r.storage),
            })
        }
        var vecQuery query.VectorQuery
        if err := json.Unmarshal(req.Query, &vecQuery); err != nil {
            return c.JSON(400, map[string]string{"error": "invalid vector query"})
        }
        result, err = vs.SearchVector(ctx, &vecQuery, cursor, size)

    // ... other query types ...

    default:
        return c.JSON(400, map[string]string{
            "error": "unsupported query type",
            "type": string(req.Type),
        })
    }

    if err != nil {
        slog.Error("Search failed", "error", err, "type", req.Type)
        return c.JSON(500, map[string]string{"error": "search failed"})
    }

    return c.JSON(200, buildResponse(result, req.Type))
}

func (r *SearchRouter) capabilitiesHandler(c echo.Context) error {
    caps := []string{string(query.QueryTypeFullText)} // always supported

    if _, ok := r.storage.(storage.BooleanSearcher); ok {
        caps = append(caps, string(query.QueryTypeBoolean))
    }
    if _, ok := r.storage.(storage.VectorSearcher); ok {
        caps = append(caps, string(query.QueryTypeVector))
    }
    if _, ok := r.storage.(storage.FuzzySearcher); ok {
        caps = append(caps, string(query.QueryTypeFuzzy))
    }
    if _, ok := r.storage.(storage.FieldLevelSearcher); ok {
        caps = append(caps, string(query.QueryTypeFieldLevel))
    }
    if _, ok := r.storage.(storage.HybridSearcher); ok {
        caps = append(caps, string(query.QueryTypeHybrid))
    }

    return c.JSON(200, map[string]interface{}{
        "supported_query_types": caps,
        "storage_backend": getBackendType(r.storage),
    })
}
```

---

## Testing Strategy

### Unit Tests per Query Type

```go
// internal/storage/pg/pg_reader_test.go

func TestPGReader_SearchFullText(t *testing.T) {
    // Test plainto_tsquery + ts_rank
}

func TestPGReader_SearchBoolean(t *testing.T) {
    // Test to_tsquery with boolean operators
}

func TestPGReader_SearchVector(t *testing.T) {
    // Test pgvector operations
}
```

### Interface Compliance Tests

```go
func TestPGReader_ImplementsAllInterfaces(t *testing.T) {
    var reader storage.Reader = &pg.Reader{}

    _, ok := reader.(storage.BooleanSearcher)
    assert.True(t, ok, "PG should support boolean search")

    _, ok = reader.(storage.VectorSearcher)
    assert.True(t, ok, "PG should support vector search")
}
```

### Benchmark Tests

```go
func BenchmarkSearchFullText_PG(b *testing.B) {
    // Benchmark PostgreSQL lexical search
}

func BenchmarkSearchFullText_ES(b *testing.B) {
    // Benchmark Elasticsearch lexical search
}

func BenchmarkSearchBoolean_PG(b *testing.B) {
    // Benchmark PostgreSQL boolean search
}

func BenchmarkSearchBoolean_ES(b *testing.B) {
    // Benchmark Elasticsearch boolean search
}
```

### Integration Tests

```go
func TestE2E_FullTextSearch(t *testing.T) {
    // POST /v1/articles/search with lexical query
    // Verify response structure
}

func TestE2E_BooleanSearch(t *testing.T) {
    // POST /v1/articles/search with boolean query
}

func TestE2E_UnsupportedQuery(t *testing.T) {
    // Test graceful degradation when backend doesn't support query type
}

func TestE2E_Capabilities(t *testing.T) {
    // GET /v1/capabilities
    // Verify correct capabilities reported
}
```

---

## Migration Path

### Phase 1: Refactor Existing Code ✅
- [x] Create `internal/types/query/query.go` with `FullTextQuery` and `BooleanQuery`
- [x] Rename `SearchLexical` → `SearchFullText` in PG and ES readers
- [x] Update router to handle `QueryTypeFullText`

### Phase 2: Add Boolean Search 🚧
- [ ] Implement `SearchBoolean` in PG reader (use `to_tsquery`)
- [ ] Implement `SearchBoolean` in ES reader (use `BoolQuery`)
- [ ] Add boolean query parser
- [ ] Update router to handle `QueryTypeBoolean`
- [ ] Add tests and benchmarks

### Phase 3: Add Field-Level Search
- [ ] Define `FieldLevelQuery` and `FieldLevelSearcher`
- [ ] Implement in PG (multiple tsvector columns)
- [ ] Implement in ES (field-specific queries)
- [ ] Update router

### Phase 4: Add Fuzzy Search
- [ ] Define `FuzzyQuery` and `FuzzySearcher`
- [ ] Implement in PG (pg_trgm)
- [ ] Implement in ES (fuzzy query)
- [ ] Update router

### Phase 5: Add Vector Search
- [ ] Define `VectorQuery` and `VectorSearcher`
- [ ] Implement in PG (pgvector)
- [ ] Implement in ES (kNN)
- [ ] Add embedding generation pipeline
- [ ] Update router

### Phase 6: Add Hybrid Search
- [ ] Define `HybridQuery` and `HybridSearcher`
- [ ] Implement RRF fusion in PG (SQL)
- [ ] Implement RRF fusion in ES (native API)
- [ ] Add fusion configuration options
- [ ] Update router

---

## Benefits Summary

### For Development
- ✅ **Type safety**: Compile-time interface checks
- ✅ **Flexibility**: Runtime capability detection
- ✅ **Maintainability**: Clear separation of concerns
- ✅ **Extensibility**: Easy to add new backends or query types

### For Research
- ✅ **Clear taxonomy**: Aligns with SEARCH_TERMINOLOGY.md
- ✅ **Benchmarking**: Compare same paradigm across engines
- ✅ **Documentation**: Explicit capability matrix
- ✅ **Reproducibility**: Well-defined query types

### For API Consumers
- ✅ **Capability discovery**: Know what's supported
- ✅ **Consistent responses**: Same format across query types
- ✅ **Clear errors**: Explicit unsupported feature messages
- ✅ **Type-driven**: Query type determines validation

---

*Last updated: 2025-11-08*
*Master Thesis: "PostgreSQL as a Search Engine"*