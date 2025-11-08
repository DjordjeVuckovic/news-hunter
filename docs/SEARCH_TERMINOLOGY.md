# Search Terminology

> **Academic Context**: This document defines search terminology for the master thesis "PostgreSQL as a Search Engine" - comparing PostgreSQL's search capabilities with Elasticsearch across multiple search paradigms.

## Table of Contents
1. [Search Paradigms](#search-paradigms)
2. [Lexical Search (Full-Text)](#lexical-search-full-text)
3. [Boolean Search](#boolean-search)
4. [Field-Level Search](#field-level-search)
5. [Fuzzy and Approximate Search](#fuzzy-and-approximate-search)
6. [Vector and Semantic Search](#vector-and-semantic-search)
7. [Hybrid Search](#hybrid-search)
8. [Core Search Concepts](#core-search-concepts)
9. [PostgreSQL-Specific Terms](#postgresql-specific-terms)
10. [Elasticsearch-Specific Terms](#elasticsearch-specific-terms)
11. [Performance and Benchmarking](#performance-and-benchmarking)

---

## Search Paradigms

### 1. Lexical Search
Token-based search that matches documents based on exact or stemmed word matches. Relies on term frequency, inverse document frequency, and field importance.

### 2. Boolean Search
Structured queries using logical operators (AND, OR, NOT) to combine multiple search terms with explicit logic.

### 3. Field-Level Search
Targeted search within specific document fields using field:value syntax (e.g., `title:"Trump"`).

### 4. Fuzzy/Approximate Search
Character-level matching that tolerates typos, spelling variations, and approximate matches using edit distance or n-gram similarity.

### 5. Semantic Search (Vector Search)
Embedding-based search that finds semantically similar documents regardless of exact keyword matches, using dense vector representations.

### 6. Hybrid Search
Combination of multiple search paradigms (typically lexical + semantic) with score fusion techniques like Reciprocal Rank Fusion (RRF).

---

## Lexical Search (Full-Text)

### Full-Text Search (`SearchFullText`) - ✓ IMPLEMENTED

**What it does:**
- Analyzes and tokenizes text into searchable terms
- Performs relevance ranking using scoring algorithms
- Searches across multiple fields (title, description, content)
- Handles stemming, stop words, and text normalization

**Implementation:**
- **PostgreSQL**: Uses `tsvector` with `plainto_tsquery` and `ts_rank` for scoring
- **Elasticsearch**: Uses `MultiMatchQuery` with field boosting (title, description, content)

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

## Boolean Search

**Status**: Planned

**Definition**: Search queries that use logical operators to combine terms and create complex search expressions.

**Operators**:
- `AND` (∧): All terms must be present
- `OR` (∨): At least one term must be present
- `NOT` (¬): Term must not be present
- `()`: Grouping for precedence

**Example Queries**:
```
climate AND change
(renewable OR sustainable) AND energy
Trump AND NOT biden
(climate OR weather) AND (change OR warming)
```

**Implementation Approaches**:

### PostgreSQL
- **tsquery operators**: `&` (AND), `|` (OR), `!` (NOT)
- **to_tsquery()**: Converts boolean expressions to tsquery
- **Example**: `to_tsquery('english', 'climate & change | warming')`

### Elasticsearch
- **Bool Query**: Combines multiple query clauses
  - `must`: Must match (AND)
  - `should`: Should match (OR)
  - `must_not`: Must not match (NOT)
- **Query String Query**: Parses boolean syntax directly

**Use Cases**:
- Legal document search
- Research literature queries
- Complex news filtering

---

## Field-Level Search

**Status**: Nice to have

**Definition**: Targeted search that restricts queries to specific document fields, providing more precise control over search scope.

**Syntax**: `field:"search terms"` or `field:value`

**Example Queries**:
```
title:"Trump"
title:"climate change" AND description:"renewable"
author:"John Smith" OR author:"Jane Doe"
title:"election" AND NOT content:"fraud"
category:politics AND language:en
```

**Implementation Approaches**:

### PostgreSQL
Need to create separate tsvector columns or use field-specific queries:
```
-- Option 1: Separate tsvector columns
WHERE title_vector @@ plainto_tsquery('english', 'Trump')
  AND description_vector @@ plainto_tsquery('english', 'Melania')

-- Option 2: Direct field search with LIKE/ILIKE
WHERE title ILIKE '%trump%' AND description ILIKE '%melania%'
```

### Elasticsearch
Native support through field-specific queries:
```json
{
  "bool": {
    "must": [
      { "match": { "title": "Trump" }},
      { "match": { "description": "Melania" }}
    ]
  }
}
```

**Use Cases**:
- Searching for specific authors
- Title-only searches for better precision
- Category-based filtering with text search
- Multi-field coordinated queries

---

## Fuzzy and Approximate Search

**Status**: Planned

**Definition**: Character-level matching techniques that handle typos, spelling variations, and approximate string matching.

### Fuzzy Search (Edit Distance)

**Concept**: Measures similarity based on minimum number of character edits (insertions, deletions, substitutions) needed to transform one string to another.

**Levenshtein Distance**: Most common edit distance metric
- Distance of 0 = exact match
- Distance of 1 = one character different
- Distance of 2 = two characters different

**Example**:
```
Query: "Obema"
Matches: "Obama" (distance=1, substitution: e→a)
```

### PostgreSQL - pg_trgm (Trigram Similarity)

**What is pg_trgm?**
Extension that provides trigram-based text similarity and pattern matching.

**Trigram**: A group of three consecutive characters
```
"cat" → {c,ca,cat,at,t}
"cart" → {c,ca,car,art,rt,t}
```

**Key Functions**:
- `similarity(text1, text2)`: Returns similarity score 0.0-1.0
- `word_similarity(text1, text2)`: Word-level similarity
- `%` operator: Similarity match operator
- `<->` operator: Distance operator for sorting

**Index Types**:
- **GIN (Generalized Inverted Index)**: Fast lookups, slower updates
- **GiST (Generalized Search Tree)**: Balanced read/write performance

**Example Usage**:
```sql
-- Enable extension
CREATE EXTENSION pg_trgm;

-- Create GIN index
CREATE INDEX idx_article_title_trgm ON articles USING GIN (title gin_trgm_ops);

-- Similarity search
SELECT title, similarity(title, 'Obema') as sim
FROM articles
WHERE title % 'Obema'  -- % is similarity operator
ORDER BY sim DESC;

-- Adjust similarity threshold
SET pg_trgm.similarity_threshold = 0.3;
```

**Configuration**:
- `pg_trgm.similarity_threshold`: Default 0.3 (adjust for stricter/looser matching)
- Lower threshold = more matches, higher recall
- Higher threshold = fewer matches, higher precision

### Elasticsearch Fuzzy Search

**Built-in Fuzzy Matching**:
```json
{
  "match": {
    "title": {
      "query": "Obema",
      "fuzziness": "AUTO"
    }
  }
}
```

**Fuzziness Levels**:
- `0`: Exact match only
- `1`: One edit allowed
- `2`: Two edits allowed
- `AUTO`: Automatic based on term length
  - 1-2 chars: exact match
  - 3-5 chars: 1 edit
  - >5 chars: 2 edits

**Algorithms**:
- **Damerau-Levenshtein**: Includes transpositions
- **N-gram tokenization**: Similar to trigrams

### Use Cases
- Handling user typos ("Obema" → "Obama")
- Name variations ("Smith" → "Smyth")
- Autocomplete and suggestions
- Tolerant search for user input
- Cross-language transliteration errors

---

## Vector and Semantic Search

**Status**: Planned

**Definition**: Search based on semantic meaning using dense vector representations (embeddings) rather than exact keyword matches.

### Core Concepts

**Embeddings**: Dense numerical vectors that capture semantic meaning
- Dimensionality: Typically 384, 768, or 1536 dimensions
- Similar concepts have similar vector representations
- Generated by neural networks (transformers)

**Semantic Similarity**: Measures how close two concepts are in meaning
```
Vector("king") - Vector("man") + Vector("woman") ≈ Vector("queen")
```

**Distance Metrics**:
- **Cosine Similarity**: Angle between vectors (most common)
  - Range: -1 to 1 (1 = identical, 0 = orthogonal, -1 = opposite)
- **Euclidean Distance (L2)**: Straight-line distance
- **Dot Product**: Inner product of vectors

### PostgreSQL - pgvector

**What is pgvector?**
Extension that adds vector data type and similarity search operations to PostgreSQL.

**Installation**:
```sql
CREATE EXTENSION vector;
```

**Vector Operations**:
```sql
-- Create table with vector column
CREATE TABLE articles (
    id UUID PRIMARY KEY,
    title TEXT,
    content TEXT,
    embedding vector(768)  -- 768-dimensional vector
);

-- Cosine similarity (<=>)
SELECT title, 1 - (embedding <=> query_vector) as similarity
FROM articles
ORDER BY embedding <=> query_vector
LIMIT 10;

-- L2 distance (<->)
SELECT title, embedding <-> query_vector as distance
FROM articles
ORDER BY embedding <-> query_vector
LIMIT 10;

-- Inner product (<#>)
SELECT title, (embedding <#> query_vector) * -1 as score
FROM articles
ORDER BY embedding <#> query_vector
LIMIT 10;
```

**Index Types for ANN (Approximate Nearest Neighbors)**:
- **IVFFlat**: Inverted file with flat compression
  - Faster indexing, moderate search speed
  - Good for smaller datasets (<1M vectors)
- **HNSW**: Hierarchical Navigable Small World
  - Slower indexing, faster search
  - Better for larger datasets
  - Higher accuracy

**Index Creation**:
```sql
-- IVFFlat index
CREATE INDEX ON articles USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);

-- HNSW index
CREATE INDEX ON articles USING hnsw (embedding vector_cosine_ops);
```

### Elasticsearch - Dense Vector Search

**Vector Field Type**:
```json
{
  "mappings": {
    "properties": {
      "title_vector": {
        "type": "dense_vector",
        "dims": 768,
        "index": true,
        "similarity": "cosine"
      }
    }
  }
}
```

**kNN Search**:
```json
{
  "knn": {
    "field": "title_vector",
    "query_vector": [...],
    "k": 10,
    "num_candidates": 100
  }
}
```

### Embedding Generation

**Popular Models**:
- **OpenAI**: text-embedding-3-small (1536 dims), text-embedding-3-large (3072 dims)
- **Sentence Transformers**: all-MiniLM-L6-v2 (384 dims), all-mpnet-base-v2 (768 dims)
- **Cohere**: embed-multilingual-v3.0
- **BGE**: BAAI/bge-large-en-v1.5

**Trade-offs**:
- Higher dimensions → Better accuracy, more storage, slower search
- Lower dimensions → Faster search, less storage, potentially lower accuracy

### Use Cases
- "Find similar articles" functionality
- Semantic search ("renewable energy" matches "solar power", "wind turbines")
- Cross-lingual search (query in English, find articles in other languages)
- Question answering (match query to relevant documents)
- Content recommendation

---

## Hybrid Search

**Status**: Planned

**Definition**: Combines multiple search approaches (typically lexical + semantic) to leverage strengths of each method and improve overall search quality.

### Why Hybrid?

**Lexical Search Strengths**:
- Exact keyword matching
- Fast and efficient
- Good for specific terms (names, IDs, technical terms)
- Interpretable results

**Lexical Search Weaknesses**:
- Vocabulary mismatch problem
- No semantic understanding
- Misses synonyms and related concepts

**Semantic Search Strengths**:
- Understands meaning and context
- Handles synonyms and paraphrasing
- Cross-lingual capabilities
- Better for conceptual queries

**Semantic Search Weaknesses**:
- Computationally expensive
- May miss exact term matches
- Less interpretable
- Requires embedding generation

### Score Fusion Methods

#### 1. Reciprocal Rank Fusion (RRF)
Most popular fusion technique. Combines rankings rather than raw scores.

**Formula**:
```
RRF(d) = Σ (1 / (k + rank_i(d)))
```
where:
- `d` = document
- `k` = constant (typically 60)
- `rank_i(d)` = rank of document in result set i

**Advantages**:
- No need to normalize scores across different systems
- Robust to score scale differences
- Simple to implement

**Example**:
```
Lexical results:  [doc1(rank=1), doc2(rank=2), doc3(rank=3)]
Semantic results: [doc2(rank=1), doc4(rank=2), doc1(rank=3)]

RRF scores (k=60):
doc1: 1/(60+1) + 1/(60+3) = 0.0164 + 0.0159 = 0.0323
doc2: 1/(60+2) + 1/(60+1) = 0.0161 + 0.0164 = 0.0325
doc3: 1/(60+3) + 0         = 0.0159
doc4: 0         + 1/(60+2) = 0.0161

Final ranking: doc2, doc1, doc4, doc3
```

#### 2. Weighted Score Fusion
Combines normalized scores with weights.

**Formula**:
```
score(d) = α * score_lexical(d) + β * score_semantic(d)
```
where α + β = 1

**Challenges**:
- Requires score normalization
- Need to tune weights (α, β)
- Sensitive to score distributions

#### 3. Convex Combination
Linear combination of scores after normalization.

### Implementation Approaches

#### PostgreSQL
```sql
-- Separate queries, combine in application
WITH lexical AS (
    SELECT id, ts_rank(...) as lex_score,
           row_number() OVER (ORDER BY ts_rank(...) DESC) as lex_rank
    FROM articles
    WHERE search_vector @@ plainto_tsquery(...)
),
semantic AS (
    SELECT id, 1 - (embedding <=> $1) as sem_score,
           row_number() OVER (ORDER BY embedding <=> $1) as sem_rank
    FROM articles
    ORDER BY embedding <=> $1
    LIMIT 100
)
SELECT
    COALESCE(l.id, s.id) as id,
    (1.0/(60 + COALESCE(l.lex_rank, 999)) + 1.0/(60 + COALESCE(s.sem_rank, 999))) as rrf_score
FROM lexical l
FULL OUTER JOIN semantic s ON l.id = s.id
ORDER BY rrf_score DESC
LIMIT 20;
```

#### Elasticsearch
```json
{
  "query": {
    "hybrid": {
      "queries": [
        {
          "multi_match": {
            "query": "renewable energy",
            "fields": ["title", "content"]
          }
        },
        {
          "knn": {
            "field": "embedding",
            "query_vector": [...],
            "k": 100
          }
        }
      ]
    }
  },
  "rank": {
    "rrf": {
      "window_size": 100,
      "rank_constant": 60
    }
  }
}
```

### Use Cases
- Best overall search quality
- E-commerce product search
- Academic paper retrieval
- Knowledge base search
- Any high-stakes search application

---

## Core Search Concepts

### Full-Text Search vs Keyword Search

| Aspect | Full-Text Search | Keyword Search |
|--------|------------------|----------------|
| **Matching** | Analyzed, tokenized, fuzzy | Exact, case-sensitive |
| **Use Case** | Content search, articles | Filters, IDs, tags |
| **Field Type** | `text` (ES), `tsvector` (PG) | `keyword` (ES) |
| **Example** | "Climate Change" matches "climatic" | author.keyword = "John Smith" |

### Relevance Ranking

Both implementations calculate relevance scores:
- **PostgreSQL**: `ts_rank()` returns 0-1 scores
- **Elasticsearch**: BM25 algorithm returns scores typically 1-30+

**Note:** Scores are not directly comparable between backends. See `docs/FUTURE_WORK.md` for ranking unification roadmap.

### Field Boosting
#### Elasticsearch
- Different fields have different importance weights:
```
title^3        → Title matches are 3x more important
description^2  → Description matches are 2x more important
content        → Content matches have base weight
```
#### PostgreSQL
- Similar effect achieved by combining multiple `tsvector` fields with weighted contributions in the ranking function:
```
ts_rank(
  setweight(to_tsvector(coalesce(title, '')), 'A') ||
  setweight(to_tsvector(coalesce(description, '')), 'B') ||
  setweight(to_tsvector(coalesce(content, '')), 'C'),
  plainto_tsquery('search terms')
)
```

---

## PostgreSQL-Specific Terms

### Text Search Data Types

**tsvector**: Document representation for full-text search
- Sorted list of distinct lexemes (normalized words)
- Removes duplicates, stop words
- Stores position information for phrase search
- Example: `'cat':1,5 'dog':2 'run':3,7`

**tsquery**: Query representation for full-text search
- Boolean combinations of lexemes
- Operators: `&` (AND), `|` (OR), `!` (NOT), `<->` (phrase)
- Example: `'climate' & 'change'`

### Text Search Functions

**to_tsvector(config, document)**: Converts text to tsvector
```sql
to_tsvector('english', 'The cats are running')
-- Result: 'cat':2 'run':4
```

**to_tsquery(config, query)**: Parses query with operators
```sql
to_tsquery('english', 'cats & dogs')
-- Result: 'cat' & 'dog'
```

**plainto_tsquery(config, query)**: Plain text to tsquery (ANDs all terms)
```sql
plainto_tsquery('english', 'cats dogs')
-- Result: 'cat' & 'dog'
```

**phraseto_tsquery(config, query)**: Phrase query
```sql
phraseto_tsquery('english', 'climate change')
-- Result: 'climate' <-> 'change'
```

**websearch_to_tsquery(config, query)**: Web-style syntax
```sql
websearch_to_tsquery('english', 'cats OR dogs -birds')
-- Result: 'cat' | 'dog' & !'bird'
```

### Ranking Functions

**ts_rank(tsvector, tsquery)**: Basic ranking
- Returns 0.0 to 1.0
- Based on term frequency
- Position independent

**ts_rank_cd(tsvector, tsquery)**: Cover density ranking
- Takes position into account
- Rewards proximity of terms
- Better for phrase searches

**Ranking Normalization Flags**:
- `0`: Default (no normalization)
- `1`: Divide by (1 + log(length))
- `2`: Divide by length
- `4`: Divide by mean harmonic distance
- `8`: Divide by number of unique words
- `16`: Divide by (1 + log(unique words))
- `32`: Divide by (rank + 1)

```sql
ts_rank(search_vector, query, 1)  -- Normalized by document length
```

### Text Search Operators

- `@@`: Match operator
- `||`: tsvector concatenation
- `&&`: tsquery AND
- `||`: tsquery OR
- `!!`: tsquery NOT
- `<->`: Phrase operator (adjacent words)
- `<N>`: Phrase operator with distance N

```sql
-- Match operator
WHERE search_vector @@ plainto_tsquery('climate change')

-- Phrase search with distance
WHERE search_vector @@ to_tsquery('climate <2> change')
```

### Text Search Configurations

**Configuration**: Language-specific rules for text processing
- Stop words (common words to ignore)
- Stemming rules (running → run)
- Dictionary mappings

**Built-in Configurations**:
- `simple`: No stemming, minimal processing
- `english`: English stemming and stop words
- `spanish`, `french`, `german`, etc.

```sql
-- Set default configuration
SET default_text_search_config = 'pg_catalog.english';

-- Use specific configuration
SELECT to_tsvector('spanish', 'corriendo rápidamente');
```

### Weights and Boosting

**setweight()**: Assign importance weights to tsvector
- Weights: `A` (highest), `B`, `C`, `D` (lowest)
- Used in ranking calculations

```sql
SELECT setweight(to_tsvector('title text'), 'A') ||
       setweight(to_tsvector('body text'), 'C');
```

### Index Types

**GIN (Generalized Inverted Index)**:
- Optimized for full-text search
- Fast lookups, slower updates
- Larger index size
- Better for read-heavy workloads

```sql
CREATE INDEX idx_search ON articles USING GIN(search_vector);
```

**GiST (Generalized Search Tree)**:
- Smaller index, balanced performance
- Faster updates, slower lookups
- Better for write-heavy workloads

```sql
CREATE INDEX idx_search ON articles USING GiST(search_vector);
```

### Extensions

**pg_trgm**: Trigram matching and similarity
- Functions: `similarity()`, `word_similarity()`
- Operators: `%` (similar), `<->` (distance)
- Index operators: `gin_trgm_ops`, `gist_trgm_ops`

**pgvector**: Vector similarity search
- Vector type: `vector(N)` where N is dimensions
- Operators: `<=>` (cosine), `<->` (L2), `<#>` (inner product)
- Indexes: IVFFlat, HNSW

**fuzzystrmatch**: Fuzzy string matching
- `levenshtein(text1, text2)`: Edit distance
- `soundex(text)`: Phonetic matching
- `metaphone(text)`: Phonetic algorithm

---

## Elasticsearch-Specific Terms

### Core Concepts

**Document**: JSON object stored and searchable
**Index**: Collection of documents with similar characteristics
**Shard**: Subdivision of an index for horizontal scaling
**Replica**: Copy of a shard for redundancy and performance
**Node**: Single server instance in ES cluster
**Cluster**: Collection of nodes

### Field Types

**text**: Full-text analyzed field
- Tokenized and analyzed
- Used for search
- Not suitable for sorting/aggregations

**keyword**: Exact value field
- Not analyzed
- Used for filtering, sorting, aggregations
- Example: IDs, tags, categories

**dense_vector**: Vector embeddings
- Numerical array for semantic search
- Supports kNN queries

### Analyzers

**Analyzer**: Text processing pipeline
1. **Character Filters**: Pre-processing (HTML stripping, etc.)
2. **Tokenizer**: Splits text into tokens
3. **Token Filters**: Post-processing (lowercase, stemming, etc.)

**Built-in Analyzers**:
- `standard`: Default, splits on word boundaries
- `simple`: Lowercase letters only
- `whitespace`: Splits on whitespace
- `english`: English stemming and stop words
- `keyword`: No analysis, exact match

```json
{
  "analyzer": {
    "custom_analyzer": {
      "tokenizer": "standard",
      "filter": ["lowercase", "stop", "snowball"]
    }
  }
}
```

### Query Types

**Match Query**: Full-text search with analysis
```json
{"match": {"title": "climate change"}}
```

**Multi-Match Query**: Search across multiple fields
```json
{
  "multi_match": {
    "query": "climate change",
    "fields": ["title^3", "description^2", "content"]
  }
}
```

**Term Query**: Exact match without analysis
```json
{"term": {"status": "published"}}
```

**Bool Query**: Combine multiple queries
```json
{
  "bool": {
    "must": [],     // AND
    "should": [],   // OR
    "must_not": [], // NOT
    "filter": []    // AND without scoring
  }
}
```

**Query String Query**: Parse complex query syntax
```json
{
  "query_string": {
    "query": "title:Trump AND description:Melania"
  }
}
```

**Fuzzy Query**: Typo tolerance
```json
{
  "fuzzy": {
    "title": {
      "value": "Obema",
      "fuzziness": "AUTO"
    }
  }
}
```

**kNN Query**: Vector similarity search
```json
{
  "knn": {
    "field": "embedding",
    "query_vector": [...],
    "k": 10,
    "num_candidates": 100
  }
}
```

### Scoring Algorithms

**BM25 (Best Match 25)**: Default scoring algorithm
- Improved version of TF-IDF
- Parameters:
  - `k1`: Term saturation (default 1.2)
  - `b`: Length normalization (default 0.75)

**TF-IDF (Term Frequency - Inverse Document Frequency)**:
- TF: How often term appears in document
- IDF: How rare term is across all documents
- Score = TF × IDF

**DFR (Divergence from Randomness)**: Probabilistic scoring
**DFI (Divergence from Independence)**: Alternative probabilistic model
**IB (Information-Based)**: Information theory based

### Inverted Index

**Structure**: Maps terms to documents containing them
```
Term → [doc1, doc2, doc3, ...]
"climate" → [1, 3, 7, 15, ...]
"change" → [1, 2, 7, 20, ...]
```

**Components**:
- Term dictionary: All unique terms
- Postings list: Documents containing each term
- Term frequencies: Occurrences per document
- Positions: Word positions for phrase queries

### Search Features

**Highlighting**: Mark matched terms in results
**Aggregations**: Analytics and faceting
**Suggesters**: Autocomplete and suggestions
**Percolation**: Reverse search (which queries match this doc?)
**Scroll API**: Efficient pagination for large result sets
**Search After**: Cursor-based pagination
**Point in Time (PIT)**: Consistent view for pagination

---

## Performance and Benchmarking

### Benchmark Metrics

#### 1. Query Latency
**Definition**: Time to execute search query and return results

**Measurements**:
- **p50 (Median)**: 50% of queries faster than this
- **p95**: 95% of queries faster than this
- **p99**: 99% of queries faster than this
- **Mean**: Average response time
- **Max**: Worst case latency

**Typical Targets**:
- Web search: <100ms p95
- Interactive: <500ms p95
- Background: <2s mean

#### 2. Throughput (QPS - Queries Per Second)
**Definition**: Number of concurrent queries system can handle

**Measurements**:
- Queries per second under load
- Sustained vs burst throughput
- Throughput at different latency percentiles

#### 3. Indexing Performance
**Definition**: Speed of adding/updating documents

**Measurements**:
- Documents per second
- Bulk indexing throughput
- Index build time for full dataset
- Update latency

#### 4. Resource Utilization
**CPU Usage**: Percentage during search operations
**Memory**: RAM consumption
- PostgreSQL: Shared buffers, work_mem
- Elasticsearch: Heap size, filesystem cache

**Disk I/O**: Read/write operations
**Storage**: Index size on disk

#### 5. Relevance Metrics (IR Metrics)

**Precision**: Proportion of retrieved documents that are relevant
```
Precision = Relevant Retrieved / Total Retrieved
```

**Recall**: Proportion of relevant documents that are retrieved
```
Recall = Relevant Retrieved / Total Relevant
```

**F1 Score**: Harmonic mean of precision and recall
```
F1 = 2 × (Precision × Recall) / (Precision + Recall)
```

**Mean Average Precision (MAP)**: Average precision across queries
**Normalized Discounted Cumulative Gain (NDCG)**: Ranking quality metric
- Considers position of relevant results
- Higher scores for relevant results at top positions

**Mean Reciprocal Rank (MRR)**: Position of first relevant result
```
MRR = 1 / rank of first relevant result
```

### Benchmark Scenarios

#### 1. Single-Term Queries
Simple queries with one keyword
```
"Trump"
"climate"
"election"
```

#### 2. Multi-Term Queries
Multiple keywords (AND semantics)
```
"climate change"
"Trump election fraud"
"renewable energy policy"
```

#### 3. Boolean Queries
Complex logic with operators
```
"(climate OR weather) AND change"
"Trump AND NOT Biden"
```

#### 4. Field-Specific Queries
Targeted field searches
```
title:"election" AND content:"fraud"
author:"John Smith"
```

#### 5. Fuzzy Queries
Typo-tolerant searches
```
"Obema" → "Obama"
"Trum" → "Trump"
```

#### 6. Phrase Queries
Exact phrase matching
```
"climate change"
"renewable energy"
```

#### 7. Long Queries
Complex, multi-term queries (5-10+ terms)

#### 8. Rare Terms vs Common Terms
- Rare: Very few matching documents
- Common: Many matching documents

### Dataset Characteristics

**Size Variations**:
- Small: 10K documents
- Medium: 100K documents
- Large: 1M documents
- XLarge: 10M+ documents

**Document Characteristics**:
- Short documents (headlines, tweets)
- Medium documents (news articles)
- Long documents (research papers, books)

**Language Distribution**:
- Monolingual vs multilingual
- Language-specific challenges

**Data Distribution**:
- Uniform: Even distribution of terms
- Skewed: Power law distribution (realistic)

### PostgreSQL Tuning Parameters

```sql
-- Full-text search
default_text_search_config = 'pg_catalog.english'
ts_rank_normalization = 1

-- Memory
shared_buffers = 25% of RAM
work_mem = 50MB  -- Per operation
maintenance_work_mem = 1GB
effective_cache_size = 75% of RAM

-- Query planning
random_page_cost = 1.1  -- For SSD
effective_io_concurrency = 200

-- Parallel query
max_parallel_workers_per_gather = 4
max_worker_processes = 8
```

### Elasticsearch Tuning Parameters

```json
{
  "index.number_of_shards": 1,
  "index.number_of_replicas": 1,
  "index.refresh_interval": "30s",
  "index.max_result_window": 10000,

  "index.search.slowlog.threshold.query.warn": "10s",
  "index.search.slowlog.threshold.fetch.warn": "1s"
}
```

**JVM Heap**: 50% of RAM, max 31GB
**Thread Pools**: Configure for workload
**Circuit Breakers**: Prevent OOM

### Benchmark Tools

**Apache JMeter**: Load testing
**Gatling**: Performance testing with Scala
**wrk/wrk2**: HTTP benchmarking
**pgbench**: PostgreSQL benchmarking
**esrally**: Elasticsearch benchmarking
**Custom Go benchmark harness**: Application-specific

### Comparison Dimensions

1. **Query Performance**: Latency and throughput
2. **Indexing Speed**: Bulk insert performance
3. **Storage Efficiency**: Index size vs data size
4. **Relevance Quality**: Precision, recall, NDCG
5. **Feature Completeness**: Boolean, fuzzy, vector, hybrid
6. **Operational Complexity**: Setup, maintenance, tuning
7. **Resource Usage**: CPU, memory, disk
8. **Scalability**: Performance under increasing load/data
9. **Cost**: Hardware, licensing, operational costs

---

## Research Status Summary

### Implemented ✓
- **Lexical Search (Full-Text)**: PostgreSQL (tsvector + ts_rank) vs Elasticsearch (MultiMatch + BM25)
- Cursor-based pagination
- Normalized scoring
- Multi-field search

### Planned
- **Boolean Search**: AND, OR, NOT operators
- **Field-Level Search**: field:value syntax
- **Fuzzy Search**: pg_trgm + Levenshtein
- **Vector Search**: pgvector + Elasticsearch kNN
- **Hybrid Search**: RRF score fusion
- **Comprehensive Benchmarks**: Performance and relevance evaluation

---

*Last updated: 2025-11-08*
*Master Thesis: "PostgreSQL as a Search Engine"*
