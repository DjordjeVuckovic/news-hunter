# Elasticsearch Search Operators & Functions Guide
*For PostgreSQL vs Elasticsearch Benchmarking Research*

## Table of Contents
1. [Query Structure Basics](#query-structure-basics)
2. [Full-Text Search Queries](#full-text-search-queries)
3. [Term-Level Queries (Exact Matching)](#term-level-queries)
4. [Compound Queries (Bool)](#compound-queries)
5. [Range & Numeric Queries](#range-queries)
6. [Fuzzy & Wildcard Queries](#fuzzy-queries)
7. [Aggregations](#aggregations)
8. [Query Context vs Filter Context](#query-filter-context)
9. [Performance Tips](#performance-tips)

---
## Introduction
### Full-Text Search (FTS) vs Term-Level Queries
ES Full-Text Query Family:
- match - Single field FTS
- multi_match - Multiple fields FTS
- match_phrase - Phrase matching
- match_phrase_prefix - Prefix phrase matching
- query_string - Advanced with operators
- simple_query_string - Simplified query_string

NOT considered FTS (term-level):
- term, terms - Exact matching (no analysis)
- range, exists - Filtering
- prefix, wildcard, regexp - Pattern matching
- fuzzy - Typo tolerance

## Query Structure Basics

### Basic Query Structure
```json
GET /index_name/_search
{
  "query": {
    // Your query here
  },
  "size": 10,        // Number of results (default: 10)
  "from": 0,         // Offset for pagination
  "aggs": {          // Optional aggregations
    // Aggregation definitions
  }
}
```

### Match All Query
Returns all documents with a score of 1.0.
```json
GET /products/_search
{
  "query": {
    "match_all": {}
  }
}
```

---

## Full-Text Search Queries

### 1. Match Query
**Use case**: Basic full-text search with analysis (tokenization, stemming, etc.)

```json
GET /articles/_search
{
  "query": {
    "match": {
      "title": "elasticsearch tutorial"
    }
  }
}
```

**With options:**
```json
GET /articles/_search
{
  "query": {
    "match": {
      "title": {
        "query": "elasticsearch tutorial",
        "operator": "and",           // "or" (default) or "and"
        "fuzziness": "AUTO",         // Allow typos
        "minimum_should_match": 2    // Min terms that must match
      }
    }
  }
}
```

### 2. Multi-Match Query
**Use case**: Search across multiple fields

```json
GET /articles/_search
{
  "query": {
    "multi_match": {
      "query": "elasticsearch",
      "fields": ["title^3", "content", "tags"],  // ^3 = 3x boost
      "type": "best_fields"                      // or "most_fields", "cross_fields"
    }
  }
}
```

### 3. Match Phrase Query
**Use case**: Exact phrase matching (order matters)

```json
GET /articles/_search
{
  "query": {
    "match_phrase": {
      "content": "machine learning algorithms"
    }
  }
}
```

**With slop** (allows words in between):
```json
GET /articles/_search
{
  "query": {
    "match_phrase": {
      "content": {
        "query": "machine algorithms",
        "slop": 2  // Allows up to 2 words between
      }
    }
  }
}
```

### 4. Query String
**Use case**: Advanced queries with operators (AND, OR, NOT, wildcards)

```json
GET /articles/_search
{
  "query": {
    "query_string": {
      "query": "(elasticsearch OR solr) AND search -deprecated",
      "fields": ["title", "content"],
      "default_operator": "AND"
    }
  }
}
```

**Operators available:**
- `AND`, `OR`, `NOT` (or `+`, `-`)
- `field:value` - search specific field
- `"exact phrase"` - phrase search
- `*` and `?` - wildcards
- `~` - fuzzy search
- `field:[1 TO 100]` - range query

### 5. Simple Query String
**Use case**: Simpler version of query_string, doesn't throw errors on invalid syntax

```json
GET /articles/_search
{
  "query": {
    "simple_query_string": {
      "query": "elasticsearch + tutorial | guide",
      "fields": ["title", "body"],
      "default_operator": "and"
    }
  }
}
```

---

## Term-Level Queries (Exact Matching)

**Important**: These queries work on exact values, not analyzed text. Use `.keyword` fields for text.

### 1. Term Query
**Use case**: Exact match on a single value

```json
GET /products/_search
{
  "query": {
    "term": {
      "status.keyword": "active"
    }
  }
}
```

### 2. Terms Query
**Use case**: Match any of multiple exact values (like SQL IN)

```json
GET /products/_search
{
  "query": {
    "terms": {
      "category.keyword": ["electronics", "computers", "phones"]
    }
  }
}
```

### 3. Exists Query
**Use case**: Check if field exists and is not null

```json
GET /products/_search
{
  "query": {
    "exists": {
      "field": "discount_price"
    }
  }
}
```

### 4. IDs Query
**Use case**: Fetch documents by their _id

```json
GET /products/_search
{
  "query": {
    "ids": {
      "values": ["1", "2", "3"]
    }
  }
}
```

---

## Compound Queries (Bool)

**The most important query type** - combines multiple queries with boolean logic.

### Bool Query Structure
```json
{
  "query": {
    "bool": {
      "must": [],      // AND - must match, affects score
      "filter": [],    // AND - must match, NO scoring (faster, cached)
      "should": [],    // OR - should match, affects score
      "must_not": []   // NOT - must not match, NO scoring
    }
  }
}
```

### Complete Example
```json
GET /products/_search
{
  "query": {
    "bool": {
      "must": [
        {
          "match": {
            "title": "laptop"
          }
        }
      ],
      "filter": [
        {
          "term": {
            "status": "published"
          }
        },
        {
          "range": {
            "price": {
              "gte": 500,
              "lte": 2000
            }
          }
        }
      ],
      "should": [
        {
          "term": {
            "brand.keyword": "Dell"
          }
        },
        {
          "term": {
            "brand.keyword": "HP"
          }
        }
      ],
      "must_not": [
        {
          "term": {
            "discontinued": true
          }
        }
      ],
      "minimum_should_match": 1
    }
  }
}
```

**Key differences:**
- `must` vs `filter`: Both are AND conditions, but `filter` is faster (no scoring, cached)
- `should`: Optional by default, but use `minimum_should_match` to require N matches
- `must_not`: Excludes documents, no scoring

### Nested Bool Queries
```json
GET /products/_search
{
  "query": {
    "bool": {
      "should": [
        {
          "bool": {
            "must": [
              {"match": {"brand": "Apple"}},
              {"match": {"category": "phones"}}
            ]
          }
        },
        {
          "bool": {
            "must": [
              {"match": {"brand": "Samsung"}},
              {"match": {"category": "tablets"}}
            ]
          }
        }
      ]
    }
  }
}
```
*Returns: (Apple AND phones) OR (Samsung AND tablets)*

---

## Range Queries

### Numeric Range
```json
GET /products/_search
{
  "query": {
    "range": {
      "price": {
        "gte": 100,   // greater than or equal
        "lte": 500,   // less than or equal
        "gt": 100,    // greater than
        "lt": 500     // less than
      }
    }
  }
}
```

### Date Range
```json
GET /logs/_search
{
  "query": {
    "range": {
      "timestamp": {
        "gte": "2024-01-01",
        "lte": "2024-12-31",
        "format": "yyyy-MM-dd"
      }
    }
  }
}
```

**Date math:**
```json
{
  "range": {
    "timestamp": {
      "gte": "now-7d/d",      // 7 days ago, rounded to day
      "lte": "now/d"          // Today, rounded to day
    }
  }
}
```

---

## Fuzzy & Wildcard Queries

### 1. Fuzzy Query
**Use case**: Handle typos (Levenshtein distance)

```json
GET /products/_search
{
  "query": {
    "fuzzy": {
      "name": {
        "value": "laptpo",         // Typo for "laptop"
        "fuzziness": "AUTO",       // AUTO, 0, 1, 2
        "prefix_length": 0,        // Characters at start that must match exactly
        "max_expansions": 50       // Max terms to match
      }
    }
  }
}
```

### 2. Wildcard Query
**Use case**: Pattern matching with `*` (0+ chars) and `?` (1 char)

```json
GET /products/_search
{
  "query": {
    "wildcard": {
      "product_code.keyword": "LAP*2024"
    }
  }
}
```

**Warning**: Avoid leading wildcards (`*something`) - very slow!

### 3. Prefix Query
**Use case**: Match terms that start with prefix

```json
GET /products/_search
{
  "query": {
    "prefix": {
      "name.keyword": "lap"
    }
  }
}
```

### 4. Regexp Query
**Use case**: Regular expression matching

```json
GET /products/_search
{
  "query": {
    "regexp": {
      "product_code.keyword": "LAP[0-9]{4}"
    }
  }
}
```

---

## Aggregations

Aggregations = SQL GROUP BY + aggregate functions, but much more powerful.

### Types of Aggregations
1. **Bucket**: Group documents (like GROUP BY)
2. **Metric**: Calculate values (like SUM, AVG, COUNT)
3. **Pipeline**: Aggregate on aggregation results

### 1. Terms Aggregation (Bucket)
**Use case**: Group by field values

```json
GET /sales/_search
{
  "size": 0,  // Don't return documents, just aggregations
  "aggs": {
    "products_by_category": {
      "terms": {
        "field": "category.keyword",
        "size": 10,           // Top 10 categories
        "order": {
          "_count": "desc"    // or "_key": "asc"
        }
      }
    }
  }
}
```

**Response:**
```json
{
  "aggregations": {
    "products_by_category": {
      "buckets": [
        {
          "key": "electronics",
          "doc_count": 150
        },
        {
          "key": "clothing",
          "doc_count": 98
        }
      ]
    }
  }
}
```

### 2. Histogram Aggregation (Bucket)
**Use case**: Group numeric values into ranges

```json
GET /products/_search
{
  "size": 0,
  "aggs": {
    "price_ranges": {
      "histogram": {
        "field": "price",
        "interval": 100  // Buckets: 0-100, 100-200, 200-300, etc.
      }
    }
  }
}
```

### 3. Date Histogram Aggregation (Bucket)
**Use case**: Time-series analysis

```json
GET /logs/_search
{
  "size": 0,
  "aggs": {
    "events_over_time": {
      "date_histogram": {
        "field": "timestamp",
        "calendar_interval": "day"  // or "month", "week", "hour"
        // fixed_interval": "12h"   // alternative: fixed intervals
      }
    }
  }
}
```

### 4. Range Aggregation (Bucket)
**Use case**: Custom ranges

```json
GET /products/_search
{
  "size": 0,
  "aggs": {
    "price_tiers": {
      "range": {
        "field": "price",
        "ranges": [
          { "key": "cheap", "to": 50 },
          { "key": "medium", "from": 50, "to": 200 },
          { "key": "expensive", "from": 200 }
        ]
      }
    }
  }
}
```

### 5. Metric Aggregations
**Common metrics:**

```json
GET /sales/_search
{
  "size": 0,
  "aggs": {
    "avg_price": {
      "avg": { "field": "price" }
    },
    "max_price": {
      "max": { "field": "price" }
    },
    "min_price": {
      "min": { "field": "price" }
    },
    "total_revenue": {
      "sum": { "field": "price" }
    },
    "unique_customers": {
      "cardinality": { "field": "customer_id" }  // Count distinct
    },
    "price_stats": {
      "stats": { "field": "price" }  // min, max, avg, sum, count together
    },
    "extended_stats": {
      "extended_stats": { "field": "price" }  // + std deviation, variance
    }
  }
}
```

### 6. Sub-Aggregations (Nested)
**Use case**: Aggregate within aggregations

```json
GET /sales/_search
{
  "size": 0,
  "aggs": {
    "by_category": {
      "terms": {
        "field": "category.keyword"
      },
      "aggs": {
        "avg_price": {
          "avg": { "field": "price" }
        },
        "total_revenue": {
          "sum": { "field": "price" }
        },
        "top_products": {
          "terms": {
            "field": "product_name.keyword",
            "size": 3
          }
        }
      }
    }
  }
}
```

**Response structure:**
```json
{
  "aggregations": {
    "by_category": {
      "buckets": [
        {
          "key": "electronics",
          "doc_count": 150,
          "avg_price": { "value": 299.5 },
          "total_revenue": { "value": 44925 },
          "top_products": {
            "buckets": [
              { "key": "Laptop Pro", "doc_count": 45 }
            ]
          }
        }
      ]
    }
  }
}
```

### 7. Filter Aggregation (Bucket)
**Use case**: Aggregate only on filtered subset

```json
GET /sales/_search
{
  "size": 0,
  "aggs": {
    "expensive_products": {
      "filter": {
        "range": { "price": { "gte": 1000 } }
      },
      "aggs": {
        "avg_price": {
          "avg": { "field": "price" }
        }
      }
    }
  }
}
```

### 8. Pipeline Aggregations
**Use case**: Aggregate on aggregation results

```json
GET /sales/_search
{
  "size": 0,
  "aggs": {
    "sales_per_month": {
      "date_histogram": {
        "field": "date",
        "calendar_interval": "month"
      },
      "aggs": {
        "monthly_revenue": {
          "sum": { "field": "price" }
        }
      }
    },
    "avg_monthly_revenue": {
      "avg_bucket": {
        "buckets_path": "sales_per_month>monthly_revenue"
      }
    },
    "revenue_derivative": {
      "derivative": {
        "buckets_path": "sales_per_month>monthly_revenue"
      }
    }
  }
}
```

---

## Query Context vs Filter Context

### Query Context
- Calculates relevance score (`_score`)
- Used in `query` parameter
- More CPU intensive
- Results NOT cached

```json
{
  "query": {
    "match": { "title": "elasticsearch" }
  }
}
```

### Filter Context
- Binary yes/no (no scoring)
- Used in `filter`, `must_not` within `bool`
- Faster execution
- **Results are cached** (huge performance benefit)

```json
{
  "query": {
    "bool": {
      "filter": [
        { "term": { "status": "published" } },
        { "range": { "date": { "gte": "2024-01-01" } } }
      ]
    }
  }
}
```

### Combined Example
```json
GET /articles/_search
{
  "query": {
    "bool": {
      "must": [
        { "match": { "title": "elasticsearch" } }  // Query context (scored)
      ],
      "filter": [
        { "term": { "status": "published" } },     // Filter context (cached)
        { "range": { "views": { "gte": 1000 } } }  // Filter context (cached)
      ]
    }
  }
}
```

---

## Performance Tips

### 1. Use Filter Context When Possible
```json
// ❌ Slower (scoring not needed)
{
  "query": {
    "bool": {
      "must": [
        { "term": { "status": "active" } }
      ]
    }
  }
}

// ✅ Faster (cached, no scoring)
{
  "query": {
    "bool": {
      "filter": [
        { "term": { "status": "active" } }
      ]
    }
  }
}
```

### 2. Avoid Leading Wildcards
```json
// ❌ Very slow
{ "wildcard": { "name": "*phone" } }

// ✅ Much faster
{ "wildcard": { "name": "phone*" } }
```

### 3. Use `.keyword` for Exact Matching
```json
// ❌ Wrong - analyzed field
{ "term": { "category": "Electronics" } }

// ✅ Correct - keyword field
{ "term": { "category.keyword": "Electronics" } }
```

### 4. Limit Aggregation Size
```json
{
  "aggs": {
    "categories": {
      "terms": {
        "field": "category.keyword",
        "size": 10  // Limit buckets returned
      }
    }
  }
}
```

### 5. Use `size: 0` for Aggregation-Only Queries
```json
GET /products/_search
{
  "size": 0,  // Don't return documents, just aggregations
  "aggs": { ... }
}
```

---

## Common Query Patterns for Benchmarking

### 1. E-commerce Product Search
```json
GET /products/_search
{
  "query": {
    "bool": {
      "must": [
        {
          "multi_match": {
            "query": "wireless headphones",
            "fields": ["name^3", "description", "brand^2"]
          }
        }
      ],
      "filter": [
        { "range": { "price": { "gte": 50, "lte": 200 } } },
        { "term": { "in_stock": true } },
        { "terms": { "category.keyword": ["electronics", "audio"] } }
      ],
      "should": [
        { "term": { "featured": true } },
        { "range": { "rating": { "gte": 4.0 } } }
      ],
      "minimum_should_match": 1
    }
  },
  "aggs": {
    "price_ranges": {
      "range": {
        "field": "price",
        "ranges": [
          { "to": 50 },
          { "from": 50, "to": 100 },
          { "from": 100, "to": 200 },
          { "from": 200 }
        ]
      }
    },
    "brands": {
      "terms": {
        "field": "brand.keyword",
        "size": 10
      }
    }
  }
}
```

### 2. Log Analysis
```json
GET /logs/_search
{
  "query": {
    "bool": {
      "must": [
        { "match": { "message": "error exception" } }
      ],
      "filter": [
        { "term": { "log_level": "ERROR" } },
        { "range": { "timestamp": { "gte": "now-24h" } } }
      ]
    }
  },
  "aggs": {
    "errors_over_time": {
      "date_histogram": {
        "field": "timestamp",
        "fixed_interval": "1h"
      },
      "aggs": {
        "error_types": {
          "terms": {
            "field": "error_type.keyword"
          }
        }
      }
    }
  }
}
```

### 3. Geospatial Search
```json
GET /stores/_search
{
  "query": {
    "bool": {
      "filter": [
        {
          "geo_distance": {
            "distance": "10km",
            "location": {
              "lat": 40.7128,
              "lon": -74.0060
            }
          }
        },
        { "term": { "status": "open" } }
      ]
    }
  }
}
```

---

## Quick Reference: Query Types Comparison

| Query Type | Use Case | Analyzed? | Scoring? | Cacheable? |
|------------|----------|-----------|----------|------------|
| `match` | Full-text search | Yes | Yes | No |
| `term` | Exact value | No | Yes | No |
| `filter` (in bool) | Exact value | No | No | Yes |
| `range` | Numeric/date ranges | No | Depends | In filter: Yes |
| `wildcard` | Pattern matching | No | Yes | No |
| `fuzzy` | Typo tolerance | Yes | Yes | No |
| `prefix` | Starts with | No | Yes | No |
| `exists` | Field exists | No | Yes | In filter: Yes |

---

## Scoring and Relevance

### Boost Queries
```json
{
  "query": {
    "bool": {
      "should": [
        {
          "match": {
            "title": {
              "query": "elasticsearch",
              "boost": 3  // 3x importance
            }
          }
        },
        {
          "match": {
            "content": "elasticsearch"  // Normal boost (1.0)
          }
        }
      ]
    }
  }
}
```

### Function Score Query
Advanced scoring manipulation:
```json
{
  "query": {
    "function_score": {
      "query": { "match": { "title": "laptop" } },
      "functions": [
        {
          "filter": { "term": { "featured": true } },
          "weight": 2  // Boost featured items
        },
        {
          "field_value_factor": {
            "field": "popularity",
            "factor": 1.2,
            "modifier": "sqrt"
          }
        }
      ],
      "boost_mode": "multiply"
    }
  }
}
```

---

## Summary for Benchmarking

### Most Common Query Patterns to Test:
1. **Full-text search**: `match`, `multi_match`
2. **Exact matching**: `term`, `terms` (in filter context)
3. **Range queries**: Price ranges, date ranges
4. **Boolean combinations**: Complex `bool` queries
5. **Fuzzy/wildcard**: Typo tolerance, pattern matching
6. **Aggregations**: Terms, histogram, date histogram with metrics
7. **Nested aggregations**: Multi-level grouping

### Performance Characteristics:
- **Fastest**: Filter context queries (term, range in filter)
- **Fast**: Match queries on properly analyzed fields
- **Moderate**: Bool queries with multiple clauses
- **Slower**: Wildcard (especially leading), regexp, fuzzy
- **Slowest**: Script queries, complex nested aggregations

### For Your PostgreSQL Comparison:
- Compare `match` queries to PostgreSQL full-text search
- Compare `term` queries in filter to PostgreSQL exact WHERE clauses
- Compare aggregations to PostgreSQL GROUP BY with aggregate functions
- Compare bool queries to complex WHERE clauses with AND/OR/NOT
- Test performance on: small datasets (1K docs), medium (100K), large (1M+)

