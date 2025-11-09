# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

News Hunter is a full-text search engine for exploring multilingual news headlines and articles. The project uses Go 1.25 and focuses on importing, processing, and storing news data from various sources (currently Kaggle datasets).

## Research Goals

**Master Thesis**: "PostgreSQL as a Search Engine"

This project is part of a master thesis research that explores PostgreSQL's comprehensive search capabilities and compares its performance with Elasticsearch through extensive benchmarks across multiple search paradigms.

**Research Focus**:
- **Full-Text Search**: Advanced features (tsvector, ts_rank, text search configurations, ranking, relevance)
- **Boolean Search**: Complex query operators (AND, OR, NOT) within full-text search context
- **Field-Level Search**: Targeted searches on specific fields (e.g., title:"Trump" & description:"Melania")
- **Fuzzy Search**: Typo tolerance and approximate matching using Levenshtein distance and similarity algorithms
- **Trigram Similarity (pg_trgm)**: Character-level matching for handling spelling variations, typos, and partial matches
- **Vector Search**: Semantic/similarity search using pgvector extension for embedding-based retrieval
- **Performance Benchmarking**: Query response time, indexing performance, resource utilization
- **Relevance Evaluation**: Comparing search quality and ranking algorithms between PostgreSQL and Elasticsearch
- **Multilingual Support**: Text search across different languages and character sets

**Search Paradigms to Explore**:
1. **Full-Text Search**: Traditional keyword-based full-text search with relevance ranking
2. **Structured Search**: Boolean operators and field-specific queries
3. **Fuzzy/Approximate Search**: Trigram similarity (pg_trgm), Levenshtein distance for typo tolerance
4. **Semantic Search**: Vector-based similarity search using embeddings
5. **Hybrid Search**: Combining multiple search approaches for optimal results

**Key Goal**: Evaluate PostgreSQL as a comprehensive alternative to dedicated search engines like Elasticsearch, exploring capabilities beyond basic exact matching and simple filtering.

## Architecture

The project follows a layered architecture pattern:

- **cmd/**: Entry points for different operations
  - `data_import/`: Imports News datasets into the database
  - `news_search/`: HTTP API server for search functionality
  - `schemagen/`: Schema generation utilities

- **internal/**: Core business logic organized by domain
  - `domain/`: Core data structures (Article, ArticleMetadata, Query)
  - `reader/`: CSV reading and YAML configuration mapping
  - `collector/`: Article collection orchestration
  - `processor/`: Article processing logic
  - `storage/`: Storage abstractions and implementations
    - `factory/`: Storage factory for creating storage instances
    - `pg/`: PostgreSQL storage implementation with full-text search
    - `es/`: Elasticsearch storage implementation
  - `server/`: HTTP server configuration and setup
  - `router/`: HTTP route handlers
  - `middleware/`: HTTP middleware (logging, etc.)

- **pkg/**: Shared packages and APIs
  - `apis/datamapping/`: Data mapping type definitions
  - `schema/`: Schema generation utilities

- **api/**: API schemas and examples
  - Data mapping configuration examples and JSON schemas


- **configs/**: Configuration files
  - `mappings/`: YAML configuration files for data field mappings
  - `elasticsearch/`: Elasticsearch configuration (index templates, ILM policies)
- **db/**: Database-related files
  - `migrations/`: SQL migration files for database schema
  - `query/`: SQL query files for database operations
- **dataset/**: Sample datasets and documentation
- **scripts/**: Utility scripts for setup and maintenance

## Key Components

### Data Mapping System
The project uses YAML configuration files to map source data fields to internal Article structure. Configuration files follow the DataMapping schema with fieldMappings that specify source/target fields and their types.

### Storage Layer
Follows idiomatic Go patterns with sub-package organization:
- **Interfaces**: `Reader` and `Storer` interfaces define storage contracts
- **Factory Pattern**: `storage/factory` package provides centralized creation logic
- **PostgreSQL**: `storage/pg` - Full-text search with tsvector, ranking, and pagination
- **Elasticsearch**: `storage/es` - Multilingual search with advanced indexing
- **In-memory**: Built-in implementation for development/testing

**Key Features**:
- SearchResult with pagination metadata (total, hasMore, page info)
- PostgreSQL uses native tsvector with ts_rank for relevance scoring
- Factory pattern avoids import cycles while maintaining clean separation

### Pipeline Architecture
Uses a pipeline pattern for data processing with common interfaces:
1. Reader loads and parses source data
2. Mapper transforms data according to configuration
3. Collector orchestrates the process
4. Factory creates storage instances based on configuration
5. Storage persists the articles with bulk operations support

### HTTP API Server
Built with Echo framework providing:
- **Search API**: Full-text search with pagination, ranking, and filtering
- **Health Checks**: Database connectivity and service health monitoring
- **Middleware**: CORS support, request logging, and error recovery
- **Configuration**: Environment-based config with validation
- **Graceful Shutdown**: Proper resource cleanup on termination

**Search Features**:
- Full-text search with relevance ranking
- Multi-field search across title, description, and content
- Pagination with total count and hasMore indicators
- Input validation and error handling
- PostgreSQL tsvector with ts_rank scoring
- Elasticsearch multi-match with field boosting

## Development Commands

### Database
```bash
# Start PostgreSQL container
docker-compose up -d

# Database runs on port 54320 with:
# - Database: news_db
# - User: news_user
# - Password: news_password
```

### Building and Running
```bash
# Build specific command
go build -o bin/data-import ./cmd/data_import

# Run with environment variables
go run ./cmd/data_import

# Build other commands
go build -o bin/schemagen ./cmd/schemagen
go build -o bin/news-search ./cmd/news_search

# Run tests
go test ./...

# Run tests for specific package
go test ./internal/reader

# Format code
go fmt ./...

# Vet code
go vet ./...
```

### Environment Setup
Commands expect environment variables (typically in `.env` files):

**Data Import (`cmd/data_import/.env`)**:
- `STORAGE_TYPE`: Storage backend (`pg`, `es`, `in_mem`)
- `MAPPING_CONFIG_PATH`: Path to YAML mapping configuration
- `DATASET_PATH`: Path to source dataset file
- `PG_CONNECTION_STRING`: PostgreSQL connection string
- `ES_ADDRESSES`: Elasticsearch cluster addresses (comma-separated)
- `ES_INDEX_NAME`: Elasticsearch index name
- `BULK_ENABLED`: Enable bulk operations (`true`/`false`)
- `BULK_SIZE`: Bulk operation batch size

**Search API (`cmd/news_search/.env`)**:
- `PORT`: HTTP server port (default: 8080)
- `USE_HTTP2`: Enable HTTP/2 support (`true`/`false`)
- `CORS_ORIGINS`: Allowed CORS origins (comma-separated)

## Testing
Tests are located alongside source files with `_test.go` suffix. Use standard Go testing patterns:
- `go test ./...` - Run all tests
- `go test -v ./internal/reader` - Run specific package tests with verbose output

## Design Principles

### Domain-Driven Design (DDD)

The project follows Domain-Driven Design principles where valuable and appropriate:

**Value Objects**:
- Use value objects for domain concepts with rich behavior and validation
- Example: `operator.Operator` - encapsulates AND/OR logic with validation and behavior
- Value objects are immutable and validated at creation
- Provide meaningful methods like `IsAnd()`, `IsOr()` rather than string comparisons
- **Idiomatic Go**: Place value objects in dedicated packages for clean namespacing

**When to use Value Objects**:
- Domain concepts with validation rules (e.g., operator.Operator, domain.SearchLanguage)
- Concepts that should be immutable once created
- Types that benefit from type safety over primitive types (no string mistakes)
- Concepts with behavior that belongs to the value itself

**Package Structure for Value Objects**:
Follow Go's idiomatic approach of using packages as namespaces:
- `internal/domain/operator/` - Contains `operator.Operator` type
- Usage: `operator.And`, `operator.Or` (not `OperatorAnd`, `OperatorOr`)
- Mirrors Go stdlib (`time.Monday`) and popular libraries (`http.MethodGet`)

**Example - Operator Value Object**:
```go
// ✅ Good: Dedicated package with clean namespacing
// File: internal/domain/operator/operator.go
package operator

type Operator string

const (
    And Operator = "and"  // Clean: operator.And
    Or  Operator = "or"   // Clean: operator.Or
)

func New(op string) (Operator, error) {
    switch op {
    case "and", "AND":
        return And, nil
    case "or", "OR", "":
        return Or, nil
    default:
        return "", fmt.Errorf("invalid operator: %q", op)
    }
}

func (o Operator) IsAnd() bool { return o == And }
func (o Operator) IsOr() bool { return o == Or }

// Usage in domain types:
query.Operator = operator.And  // ✅ Clean and idiomatic

// ❌ Bad: All in domain package with redundant prefixes
const OperatorAnd = "and"  // Would be domain.OperatorAnd - redundant!
```

**Ubiquitous Language**:
- Use Elasticsearch terminology as the ubiquitous language for search concepts
- Rationale: ES is the industry standard, aids benchmarking clarity, helps with adoption
- General search concepts (fuzziness, operators) are preferred over engine-specific features
- Document engine-specific limitations (e.g., "PostgreSQL: Ignored", "Elasticsearch: Full support")

**Engine Neutrality**:
- Avoid exposing engine-specific features in domain layer when they don't translate
- Example: Removed `tie_breaker` (ES-only) from MultiMatchQuery domain model
- Keep general concepts like `fuzziness` with documentation of engine support
- Let storage implementations handle engine-specific optimizations

**Domain Layer Purity**:
- Domain types should represent search concepts, not implementation details
- Storage layer translates domain queries to engine-specific queries
- Domain validation ensures type safety before reaching storage layer