# Complete Elasticsearch Search Paradigms Guide
*Beyond Lexical and Semantic/Vector Search*

## Overview of All Search Paradigms

Elasticsearch supports **7 major search paradigms**, each optimized for different use cases:

1. **Lexical/Full-Text Search** - Traditional text search with analysis
2. **Semantic/Vector Search** - Similarity search using embeddings
3. **Geospatial Search** - Location-based queries
4. **Structured/Term-Level Search** - Exact value matching
5. **Nested/Hierarchical Search** - Querying complex object relationships
6. **Percolate/Reverse Search** - Match documents against stored queries
7. **Graph/Join Search** - Parent-child relationships

---

## 1. Lexical/Full-Text Search
*(Already covered in previous guide)*

**What it is**: Text analysis with tokenization, stemming, and relevance scoring.

**Key queries**: `match`, `multi_match`, `match_phrase`, `query_string`

**Use cases**: Search engines, content discovery, article search

---

## 2. Semantic/Vector Search
*(Already discussed)*

**What it is**: Dense vector similarity search using ML embeddings (kNN).

**Key features**:
- `dense_vector` field type
- `knn` query
- Cosine similarity, dot product, L2 norm

**Use cases**: Recommendation systems, similar document search, image search

---

## 3. Geospatial Search

### Overview
Search based on geographic coordinates, shapes, and distances. Critical for location-based services.

### Field Types

#### geo_point
Represents latitude/longitude coordinates.

```json
PUT /locations
{
  "mappings": {
    "properties": {
      "name": { "type": "text" },
      "location": { "type": "geo_point" }
    }
  }
}
```

**Indexing geo_point:**
```json
// Method 1: Object with lat/lon
PUT /locations/_doc/1
{
  "name": "Central Park",
  "location": {
    "lat": 40.785091,
    "lon": -73.968285
  }
}

// Method 2: String format
PUT /locations/_doc/2
{
  "name": "Times Square",
  "location": "40.758896, -73.985130"
}

// Method 3: Array [lon, lat] - NOTE: longitude first!
PUT /locations/_doc/3
{
  "name": "Brooklyn Bridge",
  "location": [-73.996864, 40.706086]
}

// Method 4: Geohash
PUT /locations/_doc/4
{
  "name": "Statue of Liberty",
  "location": "dr5regw2z"
}
```

#### geo_shape
Represents complex shapes (polygons, lines, circles, etc.)

```json
PUT /zones
{
  "mappings": {
    "properties": {
      "name": { "type": "text" },
      "area": { "type": "geo_shape" }
    }
  }
}
```

### Geospatial Queries

#### 1. geo_distance (radius search)
Find all points within a certain distance.

```json
GET /locations/_search
{
  "query": {
    "bool": {
      "filter": {
        "geo_distance": {
          "distance": "5km",
          "location": {
            "lat": 40.7589,
            "lon": -73.9851
          }
        }
      }
    }
  }
}
```

**Distance units**: `km`, `mi` (miles), `m` (meters), `ft` (feet)

#### 2. geo_bounding_box
Find points within a rectangular area.

```json
GET /locations/_search
{
  "query": {
    "bool": {
      "filter": {
        "geo_bounding_box": {
          "location": {
            "top_left": {
              "lat": 40.73,
              "lon": -74.1
            },
            "bottom_right": {
              "lat": 40.01,
              "lon": -71.12
            }
          }
        }
      }
    }
  }
}
```

#### 3. geo_polygon
Find points within an arbitrary polygon.

```json
GET /locations/_search
{
  "query": {
    "bool": {
      "filter": {
        "geo_polygon": {
          "location": {
            "points": [
              {"lat": 40.73, "lon": -74.1},
              {"lat": 40.01, "lon": -71.12},
              {"lat": 39.5, "lon": -73.5}
            ]
          }
        }
      }
    }
  }
}
```

#### 4. geo_shape Query
Search for documents with shapes that intersect, are within, or contain a given shape.

```json
GET /zones/_search
{
  "query": {
    "bool": {
      "filter": {
        "geo_shape": {
          "area": {
            "shape": {
              "type": "circle",
              "coordinates": [-73.985130, 40.758896],
              "radius": "1km"
            },
            "relation": "intersects"  // or "within", "contains", "disjoint"
          }
        }
      }
    }
  }
}
```

**Supported shapes**: `point`, `linestring`, `polygon`, `multipoint`, `multilinestring`, `multipolygon`, `geometrycollection`, `envelope`, `circle`

### Geospatial Aggregations

#### geo_distance Aggregation
Group results by distance buckets.

```json
GET /locations/_search
{
  "size": 0,
  "aggs": {
    "rings_around_times_square": {
      "geo_distance": {
        "field": "location",
        "origin": "40.758896, -73.985130",
        "unit": "km",
        "ranges": [
          { "to": 1 },
          { "from": 1, "to": 5 },
          { "from": 5, "to": 10 },
          { "from": 10 }
        ]
      }
    }
  }
}
```

#### geo_bounds Aggregation
Calculate bounding box of all geo_points.

```json
GET /locations/_search
{
  "size": 0,
  "aggs": {
    "viewport": {
      "geo_bounds": {
        "field": "location"
      }
    }
  }
}
```

#### geohash_grid Aggregation
Group points into geohash buckets (for heatmaps).

```json
GET /locations/_search
{
  "size": 0,
  "aggs": {
    "location_heatmap": {
      "geohash_grid": {
        "field": "location",
        "precision": 5  // 1-12, higher = more granular
      }
    }
  }
}
```

### Sorting by Distance

```json
GET /locations/_search
{
  "query": {
    "match_all": {}
  },
  "sort": [
    {
      "_geo_distance": {
        "location": {
          "lat": 40.7589,
          "lon": -73.9851
        },
        "order": "asc",
        "unit": "km",
        "distance_type": "arc"  // or "plane" (faster, less accurate)
      }
    }
  ]
}
```

### Use Cases
- Store locator (find nearest stores)
- Ride-sharing apps (find nearby drivers)
- Real estate search (find properties in area)
- Delivery zones (check if address is in coverage area)
- Geofencing (trigger events when entering/leaving areas)
- Travel/tourism (find attractions near user)

---

## 4. Structured/Term-Level Search
*(Partially covered in previous guide)*

**What it is**: Exact value matching without text analysis - like SQL WHERE clauses.

**Key difference from lexical**: No tokenization, stemming, or analysis. Searches exact values.

### Primary Queries
- `term` - exact match on single value
- `terms` - exact match on multiple values (SQL IN)
- `range` - numeric/date ranges
- `exists` - field has value
- `prefix` - starts with
- `wildcard` - pattern matching
- `regexp` - regex matching
- `ids` - match by document IDs

**Best practices**:
- Always use `.keyword` fields for text
- Put in `filter` context (cached, no scoring)
- Ideal for faceted search, filtering

---

## 5. Nested & Hierarchical Search

### The Problem
Arrays of objects in Elasticsearch are flattened, losing relationships.

**Example:**
```json
PUT /users/_doc/1
{
  "name": "John",
  "comments": [
    { "author": "Alice", "rating": 5 },
    { "author": "Bob", "rating": 2 }
  ]
}
```

**Flattened internally to:**
```json
{
  "name": "John",
  "comments.author": ["Alice", "Bob"],
  "comments.rating": [5, 2]
}
```

**Problem**: Query for `author:Alice AND rating:2` would incorrectly match because Elasticsearch lost the object boundaries!

### Solution 1: Nested Objects

#### Mapping
```json
PUT /blog
{
  "mappings": {
    "properties": {
      "title": { "type": "text" },
      "comments": {
        "type": "nested",
        "properties": {
          "username": { "type": "keyword" },
          "comment": { "type": "text" },
          "rating": { "type": "integer" },
          "created_at": { "type": "date" }
        }
      }
    }
  }
}
```

#### Indexing
```json
PUT /blog/_doc/1
{
  "title": "Elasticsearch Guide",
  "comments": [
    {
      "username": "alice",
      "comment": "Great tutorial!",
      "rating": 5,
      "created_at": "2024-01-15"
    },
    {
      "username": "bob",
      "comment": "Needs more examples",
      "rating": 3,
      "created_at": "2024-01-16"
    }
  ]
}
```

#### Nested Query
```json
GET /blog/_search
{
  "query": {
    "nested": {
      "path": "comments",
      "query": {
        "bool": {
          "must": [
            { "match": { "comments.username": "alice" } },
            { "range": { "comments.rating": { "gte": 4 } } }
          ]
        }
      }
    }
  }
}
```

**This correctly returns only posts where Alice gave rating >= 4!**

#### Nested Query with Highlighting
```json
GET /blog/_search
{
  "query": {
    "nested": {
      "path": "comments",
      "query": {
        "match": { "comments.comment": "tutorial" }
      },
      "inner_hits": {
        "highlight": {
          "fields": {
            "comments.comment": {}
          }
        }
      }
    }
  }
}
```

`inner_hits` returns which nested documents matched.

#### Nested Aggregations
```json
GET /blog/_search
{
  "size": 0,
  "aggs": {
    "comments": {
      "nested": {
        "path": "comments"
      },
      "aggs": {
        "avg_rating": {
          "avg": {
            "field": "comments.rating"
          }
        },
        "top_commenters": {
          "terms": {
            "field": "comments.username"
          }
        }
      }
    }
  }
}
```

#### Multiple Levels of Nesting
```json
PUT /company
{
  "mappings": {
    "properties": {
      "name": { "type": "text" },
      "departments": {
        "type": "nested",
        "properties": {
          "name": { "type": "keyword" },
          "employees": {
            "type": "nested",
            "properties": {
              "name": { "type": "keyword" },
              "skills": { "type": "keyword" }
            }
          }
        }
      }
    }
  }
}
```

**Query nested inside nested:**
```json
GET /company/_search
{
  "query": {
    "nested": {
      "path": "departments",
      "query": {
        "nested": {
          "path": "departments.employees",
          "query": {
            "term": {
              "departments.employees.skills": "elasticsearch"
            }
          }
        }
      }
    }
  }
}
```

#### Reverse Nested Aggregation
Go back from nested to parent.

```json
GET /blog/_search
{
  "size": 0,
  "aggs": {
    "comments": {
      "nested": {
        "path": "comments"
      },
      "aggs": {
        "top_commenters": {
          "terms": {
            "field": "comments.username"
          },
          "aggs": {
            "back_to_posts": {
              "reverse_nested": {},
              "aggs": {
                "top_post_titles": {
                  "terms": {
                    "field": "title.keyword"
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}
```

### Solution 2: Parent-Child Relationships

**When to use**:
- Many children per parent (nested has performance issues >10K children)
- Children updated frequently (nested requires reindexing entire parent)
- Need to search/aggregate on parents and children independently

#### Mapping with Join Field
```json
PUT /company
{
  "mappings": {
    "properties": {
      "my_join_field": {
        "type": "join",
        "relations": {
          "company": "employee"  // company is parent of employee
        }
      },
      "name": { "type": "text" },
      "title": { "type": "text" }
    }
  }
}
```

#### Indexing Parent
```json
PUT /company/_doc/1?routing=1
{
  "name": "Acme Corp",
  "my_join_field": {
    "name": "company"
  }
}
```

**Important**: Use `routing` to ensure parent and children are on same shard!

#### Indexing Children
```json
PUT /company/_doc/101?routing=1
{
  "name": "Alice",
  "title": "Engineer",
  "my_join_field": {
    "name": "employee",
    "parent": "1"
  }
}

PUT /company/_doc/102?routing=1
{
  "name": "Bob",
  "title": "Manager",
  "my_join_field": {
    "name": "employee",
    "parent": "1"
  }
}
```

#### has_child Query
Find parents that have children matching criteria.

```json
GET /company/_search
{
  "query": {
    "has_child": {
      "type": "employee",
      "query": {
        "match": {
          "title": "engineer"
        }
      }
    }
  }
}
```

Returns company documents that have at least one engineer employee.

#### has_parent Query
Find children whose parents match criteria.

```json
GET /company/_search
{
  "query": {
    "has_parent": {
      "parent_type": "company",
      "query": {
        "match": {
          "name": "Acme"
        }
      }
    }
  }
}
```

Returns employee documents whose company name matches "Acme".

#### Multi-level Parent-Child
```json
PUT /relations
{
  "mappings": {
    "properties": {
      "my_join_field": {
        "type": "join",
        "relations": {
          "company": "department",
          "department": "employee"  // department is child of company, parent of employee
        }
      }
    }
  }
}
```

### Nested vs Parent-Child Comparison

| Feature | Nested | Parent-Child |
|---------|--------|--------------|
| **Stored** | Together (same Lucene doc) | Separately (different docs) |
| **Query speed** | Fast (single doc) | Slower (join at query time) |
| **Index speed** | Slower (reindex whole doc) | Faster (independent) |
| **Updates** | Must update parent | Can update child only |
| **Best for** | <100 nested docs per parent | Many children (>1000s) |
| **Memory** | Lower | Higher (maintains join map) |
| **Use when** | Rarely updated, small arrays | Frequently updated, large arrays |

### Use Cases
- **Nested**: Blog posts with comments, products with reviews, orders with line items
- **Parent-Child**: Companies with employees, forums with posts and replies, hierarchical categories

---

## 6. Percolate / Reverse Search

### Concept
**Normal search**: Store documents, query to find matching documents.
**Percolate**: Store queries, send document to find matching queries.

**Use case**: "Tell me when something matching MY criteria appears"

### Real-World Examples
1. **Alerting**: User creates alert "notify me when iPhone price < $800" → Store as query → Percolate new products
2. **Content routing**: News articles → Find which users' interest queries they match → Send notifications
3. **Classification**: Document comes in → Match against classification rules (queries) → Tag automatically
4. **Monitoring**: Log entry arrives → Check against error pattern queries → Trigger alerts

### Setup

#### 1. Create Index with Percolator Field
```json
PUT /alerts
{
  "mappings": {
    "properties": {
      "alert_name": { "type": "text" },
      "user_id": { "type": "keyword" },
      "query": { "type": "percolator" },
      "product": { "type": "keyword" },
      "price": { "type": "float" }
    }
  }
}
```

**Key**: `query` field is type `percolator` - stores Elasticsearch queries!

#### 2. Index Queries (User Alerts)
```json
// User 1 wants iPhone under $800
PUT /alerts/_doc/1
{
  "alert_name": "iPhone Deal",
  "user_id": "user_001",
  "query": {
    "bool": {
      "must": [
        { "term": { "product": "iPhone" } },
        { "range": { "price": { "lte": 800 } } }
      ]
    }
  }
}

// User 2 wants any Apple product under $500
PUT /alerts/_doc/2
{
  "alert_name": "Apple Budget",
  "user_id": "user_002",
  "query": {
    "bool": {
      "must": [
        { "match": { "brand": "Apple" } },
        { "range": { "price": { "lte": 500 } } }
      ]
    }
  }
}
```

#### 3. Percolate New Documents
```json
GET /alerts/_search
{
  "query": {
    "percolate": {
      "field": "query",
      "document": {
        "product": "iPhone",
        "brand": "Apple",
        "price": 750
      }
    }
  }
}
```

**Result**: Returns alert_id 1 (matches user_001's criteria)!

### Percolate Existing Document
```json
// Percolate document that's already indexed elsewhere
GET /alerts/_search
{
  "query": {
    "percolate": {
      "field": "query",
      "index": "products",
      "id": "123"
    }
  }
}
```

### Percolate Multiple Documents
```json
GET /alerts/_search
{
  "query": {
    "percolate": {
      "field": "query",
      "documents": [
        { "product": "iPhone", "price": 750 },
        { "product": "iPad", "price": 400 },
        { "product": "MacBook", "price": 1200 }
      ]
    }
  }
}
```

### Combining with Other Queries
Filter which percolator queries to check:

```json
GET /alerts/_search
{
  "query": {
    "bool": {
      "must": [
        {
          "percolate": {
            "field": "query",
            "document": {
              "product": "iPhone",
              "price": 750
            }
          }
        }
      ],
      "filter": [
        { "term": { "user_id": "user_001" } }  // Only check user_001's alerts
      ]
    }
  }
}
```

### Highlighting Matched Fields
```json
GET /alerts/_search
{
  "query": {
    "percolate": {
      "field": "query",
      "document": {
        "description": "Brand new iPhone 15 on sale",
        "price": 799
      }
    }
  },
  "highlight": {
    "fields": {
      "description": {}
    }
  }
}
```

### Performance Tips
1. **Use filters in percolator queries** - cached and faster
2. **Limit complexity** - Simple queries percolate faster
3. **Use terms over match** - More efficient
4. **Monitor query count** - 10,000+ queries can be slow
5. **Consider routing** - Group related queries

### Advanced: Classification Example
```json
PUT /document-classifier
{
  "mappings": {
    "properties": {
      "tag": { "type": "keyword" },
      "description": { "type": "text" },
      "query": { "type": "percolator" }
    }
  }
}

// Index classification rules
PUT /document-classifier/_doc/tech
{
  "tag": "Technology",
  "query": {
    "match": {
      "content": "software hardware computer AI"
    }
  }
}

PUT /document-classifier/_doc/finance
{
  "tag": "Finance",
  "query": {
    "match": {
      "content": "stock market investment banking"
    }
  }
}

// Classify incoming document
GET /document-classifier/_search
{
  "query": {
    "percolate": {
      "field": "query",
      "document": {
        "content": "New AI software revolutionizes stock market predictions"
      }
    }
  }
}
// Returns both "Technology" and "Finance" tags!
```

### Use Cases
- **User alerting** - Price drops, job postings, news mentions
- **Content routing** - Route content to interested users
- **Auto-tagging** - Classify documents automatically
- **Monitoring** - Log error patterns, security threats
- **Real-time matching** - Dating apps, job matching
- **Search-as-you-type** - Suggest matching saved searches

---

## 7. Graph / Join Search

Already covered in Parent-Child section (#5), but here's the distinction:

**Parent-Child** (covered above):
- Explicit join field in mapping
- `has_parent` and `has_child` queries
- Good for hierarchical data

**Application-side Joins**:
- Store IDs, join in application
- Multiple queries
- More flexible but slower

**Denormalization**:
- Duplicate data
- Faster queries, harder updates

Elasticsearch is NOT a relational database - for complex joins, consider:
1. Denormalize (duplicate data)
2. Nested objects (for contained relationships)
3. Parent-child (for true separate entities)
4. Application-side joins (for complex relationships)

---

## Summary Table: When to Use Each Search Type

| Search Type | Best For | Speed | Complexity |
|-------------|----------|-------|------------|
| **Lexical** | Text search, articles, logs | Fast | Medium |
| **Semantic/Vector** | Recommendations, similarity | Medium | High |
| **Geospatial** | Location-based, maps | Fast | Medium |
| **Structured** | Filters, exact matches | Very Fast | Low |
| **Nested** | Small arrays of objects | Fast | Medium |
| **Parent-Child** | Large hierarchies, frequently updated | Medium | High |
| **Percolate** | Alerting, classification | Medium | High |

---

## Performance Comparison for Benchmarking

### Query Speed (Fastest → Slowest)
1. **Structured/Term in filter** (cached)
2. **Geospatial distance** (optimized with geo-hash)
3. **Lexical match** (inverted index)
4. **Nested queries** (same doc)
5. **Semantic/Vector** (ANN, not exact)
6. **Parent-child** (query-time join)
7. **Percolate** (many queries to check)

### Index Speed (Fastest → Slowest)
1. **Structured/Term** (no analysis)
2. **Lexical** (analyzed but simple)
3. **Geospatial** (additional indexing)
4. **Parent-child** (separate docs)
5. **Vector** (dimension calculation)
6. **Nested** (multiple hidden docs)
7. **Percolate** (query compilation)

### PostgreSQL Comparison Suggestions

1. **Lexical** ↔ PostgreSQL full-text search (tsvector)
2. **Structured** ↔ Regular WHERE clauses with indexes
3. **Nested** ↔ JSONB queries
4. **Parent-Child** ↔ JOINs
5. **Geospatial** ↔ PostGIS extension
6. **Vector** ↔ pgvector extension
7. **Percolate** ↔ No direct equivalent (application logic)

---

## Quick Reference: Query Syntax Cheat Sheet

```json
// Lexical
{ "match": { "field": "value" } }

// Vector
{ "knn": { "field": "embedding", "query_vector": [...], "k": 10 } }

// Geospatial
{ "geo_distance": { "distance": "5km", "location": { "lat": 40, "lon": -73 } } }

// Structured
{ "term": { "status.keyword": "active" } }

// Nested
{ "nested": { "path": "comments", "query": { "match": { "comments.text": "great" } } } }

// Parent-Child
{ "has_child": { "type": "employee", "query": { "match": { "name": "Alice" } } } }

// Percolate
{ "percolate": { "field": "query", "document": { "text": "..." } } }
```

---

Good luck with your comprehensive benchmarking! 🚀