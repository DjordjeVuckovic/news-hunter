# PostgreSQL Full-Text Search: Design and Evaluation Through a Storage-Agnostic Interface

## Abstract
We investigate PostgreSQL's viability as a full-text search engine by:
1. Designing a storage-agnostic search interface based on fundamental operations
2. Implementing this interface for both PostgreSQL and Elasticsearch
3. Systematically comparing capabilities, performance, and limitations
4. Providing guidelines for when PostgreSQL is sufficient vs when Elasticsearch is needed

## Chapter 1: Introduction

### 1.1 Motivation
- Many applications use separate search infrastructure (Elasticsearch)
- PostgreSQL has full-text search capabilities
- Question: When is PostgreSQL "good enough"?

### 1.2 Research Questions
1. What full-text search operations can PostgreSQL support?
2. How does PostgreSQL's performance compare to Elasticsearch?
3. When should developers use PostgreSQL vs dedicated search engines?

### 1.3 Approach
- Design storage-agnostic interface (methodology)
- Implement for PostgreSQL (primary focus)
- Implement for Elasticsearch (reference/comparison)
- Benchmark and evaluate

### 1.4 Contributions
1. Systematic evaluation of PostgreSQL full-text search
2. Storage-agnostic search interface design (reusable for future work)
3. Decision framework for PostgreSQL vs Elasticsearch
4. Open-source implementation

## Chapter 2: Background

### 2.1 Full-Text Search Concepts
- Inverted indexes
- Text analysis and tokenization
- Relevance ranking
- Common operations (match, phrase, boolean)

### 2.2 PostgreSQL Full-Text Search
- ts_vector and ts_query
- Text search operators
- Ranking functions (ts_rank)
- Language support
- GIN/GiST indexes

### 2.3 Elasticsearch
- Lucene-based architecture
- Query DSL
- Capabilities overview

### 2.4 Related Work
- Comparisons of search engines
- Database-integrated search
- Query abstraction layers

## Chapter 3: Interface Design

### 3.1 Design Goals
- Storage-agnostic (not tied to one engine)
- Based on fundamental operations (not ES-specific)
- Expressive enough for common use cases
- Simple to implement

### 3.2 Core Operations
```go
- OpTextMatch: Search text in field(s) with weights
- OpPhraseMatch: Exact phrase matching
- OpAnd/Or/Not: Logical composition
```

### 3.3 Query Model
```go
type Query struct {
Operation Operation
Text      string
Fields    []WeightedField
Children  []Query // for composition
}
```

### 3.4 Capabilities Model
```go
type Capabilities struct {
SupportsTextMatch   bool
SupportsPhraseMatch bool
SupportsNesting     bool
MaxNestingDepth     int
// ...
}
```

### 3.5 Design Rationale
- Why these operations?
- Why not ES-specific terms (match_phrase, bool)?
- Trade-offs and limitations

## Chapter 4: Implementation

### 4.1 PostgreSQL Implementation
#### 4.1.1 Text Match Translation
```sql
-- Query: text_match("climate change", fields=["title"^3, "content"])
SELECT ...
WHERE setweight(to_tsvector('english', title), 'A') ||
      setweight(to_tsvector('english', content), 'C')
      @@ to_tsquery('english', 'climate & change')
```

#### 4.1.2 Phrase Match Translation
```sql
-- Query: phrase_match("climate change", fields=["title", "content"])
WHERE to_tsvector('english', title) @@ phraseto_tsquery('english', 'climate change')
   OR to_tsvector('english', content) @@ phraseto_tsquery('english', 'climate change')
```

#### 4.1.3 Logical Operations
- Simple AND/OR: Direct WHERE clause composition
- Limitation: No deep nesting (depth > 1)

#### 4.1.4 Indexing Strategy
- GIN indexes on tsvector columns
- Pre-computed vs dynamic tsvector
- Trade-offs

### 4.2 Elasticsearch Implementation
#### 4.2.1 Query Translation
```json
// OpTextMatch → multi_match query
{
  "multi_match": {
    "query": "climate change",
    "fields": ["title^3", "content"]
  }
}
```

#### 4.2.2 Boolean Composition
```json
// OpAnd → bool.must
{
  "bool": {
    "must": [...]
  }
}
```

### 4.3 Implementation Challenges
- Weight mapping (PG: A/B/C/D classes vs ES: multiplicative)
- Phrase slop (PG: limited vs ES: full support)
- Nested queries (PG: limited vs ES: unlimited)

## Chapter 5: Evaluation

### 5.1 Experimental Setup
- Dataset: [e.g., Wikipedia articles, news corpus]
- Size: X documents, Y GB
- Hardware: ...
- Versions: PostgreSQL 16, Elasticsearch 8.x

### 5.2 Capability Comparison

| Operation | PostgreSQL | Elasticsearch |
|-----------|-----------|---------------|
| Text match (single field) | ✅ Full | ✅ Full |
| Text match (multi-field) | ✅ Full | ✅ Full |
| Field weighting | ⚠️ 4 classes | ✅ Continuous |
| Phrase match | ✅ Basic | ✅ Full (slop) |
| Boolean AND/OR | ✅ Simple | ✅ Full |
| Nested boolean | ❌ Depth 1 | ✅ Unlimited |
| Fuzzy matching | ❌ No | ✅ Yes |

### 5.3 Performance Benchmarks

#### 5.3.1 Query Latency
- Simple text match
- Multi-field weighted
- Phrase queries
- Boolean composition

#### 5.3.2 Throughput
- Queries per second
- Concurrent query handling

#### 5.3.3 Indexing Performance
- Insert/update speed
- Index size
- Memory usage

#### 5.3.4 Scalability
- Dataset size: 100K, 1M, 10M, 100M documents
- How does performance degrade?

### 5.4 Results
[Your actual benchmark results]

### 5.5 Analysis
- Where PG performs well
- Where ES has advantages
- Surprising findings

## Chapter 6: Discussion

### 6.1 When PostgreSQL is Sufficient
✅ Use PostgreSQL when:
- Dataset < 10M documents
- Simple to moderate queries (text match, basic boolean)
- Cost sensitivity (no separate infrastructure)
- Already using PostgreSQL
- Don't need advanced features (fuzzy, aggregations)

### 6.2 When Elasticsearch is Better
✅ Use Elasticsearch when:
- Dataset > 50M documents
- Complex nested queries required
- Need fuzzy matching, suggestions
- Need aggregations/faceting
- Need very low latency (<10ms)
- Dedicated search features (highlighting, more-like-this)

### 6.3 The Gray Area (10M-50M docs)
- Depends on query complexity
- Performance requirements
- Team expertise
- Infrastructure costs

### 6.4 Storage-Agnostic Interface Evaluation
- Did abstraction work?
- What was lost in translation?
- Overhead of abstraction
- Would this work for other engines? (speculation)

### 6.5 Limitations of This Study
- Only two engines evaluated
- Specific dataset characteristics
- Single-node deployments
- Didn't evaluate: Solr, Meilisearch, Typesense, etc.

## Chapter 7: Related Work & Future Work

### 7.1 Related Work
- Other PG vs ES comparisons
- Search abstraction layers
- Database-integrated search

### 7.2 Future Work
- Extend interface to other engines (Solr, Meilisearch)
- Distributed/sharded PostgreSQL search
- Hybrid approaches (PG + ES)
- Advanced features (suggestions, typo tolerance)

## Chapter 8: Conclusion

### 8.1 Summary
- PostgreSQL CAN be a viable search engine for many use cases
- Storage-agnostic interface enabled systematic evaluation
- Clear guidelines for when to use each engine

### 8.2 Contributions
1. **Empirical evaluation** of PostgreSQL full-text search
2. **Storage-agnostic interface** design (reusable methodology)
3. **Decision framework** for practitioners
4. **Open-source implementation** for community

### 8.3 Closing Remarks
PostgreSQL is "good enough" for X% of use cases, but Elasticsearch
still has advantages for Y% of workloads. The storage-agnostic
interface provides a foundation for future comparisons with other
search technologies.

### Key Framing Points 🎯
```text
In introduction:

"While a truly universal search abstraction would require evaluating many search engines (Solr, Meilisearch, Typesense, etc.), this thesis focuses on PostgreSQL as the primary subject and uses Elasticsearch as a reference point. The storage-agnostic interface serves as both a methodology for systematic evaluation and a potential foundation for future work extending to other search technologies."

In Your Limitations Section:

"This study evaluates only PostgreSQL and Elasticsearch. While the interface design is intended to be storage-agnostic, validating its applicability to other search engines (Solr, Meilisearch, Typesense, OpenSearch) is left for future work. Our contribution is the methodology and PostgreSQL evaluation, not a universal search abstraction."

In Your Future Work:

"The storage-agnostic interface designed in this thesis could be extended to evaluate other search engines:

Apache Solr: Another Lucene-based engine
Meilisearch: Lightweight alternative
Typesense: Modern typo-tolerant search
OpenSearch: Elasticsearch fork

Such extensions would validate the generality of our interface design and provide broader comparison data."
```
