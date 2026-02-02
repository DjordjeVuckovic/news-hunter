# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

News Hunter is a full-text search engine for exploring multilingual news headlines and articles. The project uses Go 1.25 and focuses on importing, processing, and storing news data from various sources (currently Kaggle datasets).

## Research Goals

**Master Thesis**: "PostgreSQL as a Search Engine"

This project is part of a master thesis research that explores PostgreSQL's comprehensive search capabilities and compares its performance with Elasticsearch through extensive benchmarks across multiple search paradigms.

**Search Paradigms to Explore**:
1. **Full-Text Search**: Traditional keyword-based full-text search with relevance ranking
2. **Structured Search**: Boolean operators and field-specific queries
3. **Fuzzy/Approximate Search**: Trigram similarity (pg_trgm), Levenshtein distance for typo tolerance
4. **Semantic Search**: Vector-based similarity search using embeddings
5. **Hybrid Search**: Combining multiple search approaches for optimal results

**Key Goal**: Evaluate PostgreSQL as a comprehensive alternative to dedicated search engines like Elasticsearch, exploring capabilities beyond basic exact matching and simple filtering.


## Design Principles
- You MUST write clean, idiomatic Go code following best practices.
- You MUST organize code into clear packages with single responsibility.
- You MUST NOT write unnecessary comments; code should be self-explanatory.
- You MUST write unit tests for all core logic with high coverage.

## Architecture

The project follows a layered architecture pattern:

- **cmd/**: Entry points for different operations
  - `ds_ingest/`: Imports News datasets into the database
  - `news_api/`: HTTP API server for search functionality
  - `schemagen/`: Schema generation utilities
  - `benchmark/`: Search engine benchmarking tools

- **internal/**: Core business logic organized by domain
  - `types/`: Core type definitions organized by bounded contexts
    - `document/`: Document types (Article, ArticleMetadata, WeightedDocument)
    - `query/`: Query types (search query types, language, scoring, cursor)
    - `operator/`: Operator value object (AND/OR logic)
  - `ingest/`: Data ingestion pipeline
    - `reader/`: CSV/YAML data reading and parsing
    - `mapper/`: Data mapping according to configuration
    - `collector/`: Orchestration of data ingestion process
    - `processor/`: Article processing logic
  - `api/`: API layer for HTTP server
  - `storage/`: Storage abstractions and implementations
    - `factory/`: Storage factory for creating storage instances
    - `pg/`: PostgreSQL storage implementation with full-text search
    - `es/`: Elasticsearch storage implementation
  - `dto/`: Data transfer objects for API layer

- **pkg/**: Shared packages and APIs
  - `apis/datamapping/`: Data mapping type definitions
  - `schema/`: Schema generation utilities

- **api/**: API schemas and examples
  - Data mapping configuration examples and JSON schemas


- **configs/**: Configuration files
  - `mappings/`: YAML configuration files for data field mappings
  - `elasticsearch/`: Elasticsearch configuration (index templates, ILM policies)
  - `benchmark/`: Benchmarking configuration files
- **db/**: Database-related files
  - `migrations/`: SQL migration files for database schema
  - `query/`: SQL query files for database operations
- **dataset/**: Sample datasets and documentation
- **scripts/**: Utility scripts for setup and maintenance

## Key Components

### Type System Organization

The type system is organized into bounded contexts, each with its own package:

**Benefits of this organization**:
- Clear separation of concerns between document and query types
- Reduced import cycles through proper package boundaries
- Clean namespacing (e.g., `query.QueryString`, `document.Article`)
- Easier to understand and navigate type definitions
- Follows DDD bounded context principles
- Query API design follows industry standards (Elasticsearch `query_string` terminology)

### Data Mapping System
The project uses YAML configuration files to map source data fields to internal Article structure. Configuration files follow the DataMapping schema with fieldMappings that specify source/target fields and their types.

### Storage Layer
Follows idiomatic Go patterns with sub-package organization:
- **Interfaces**: Storage contracts define clear responsibilities
  - `FTSSearcher`: Full-text search operations
  - `Indexer`: Document indexing operations
  - `Reader`: Document retrieval operations
- **Factory Pattern**: `storage/factory` package provides centralized creation logic
- **PostgreSQL**: `storage/pg` - Full-text search with tsvector, ranking, and pagination
- **Elasticsearch**: `storage/es` - Multilingual search with advanced indexing
- **In-memory**: Built-in implementation for development/testing

**Key Features**:
- SearchResult with pagination metadata (total, hasMore, page info)
- PostgreSQL uses native tsvector with ts_rank for relevance scoring
- Factory pattern avoids import cycles while maintaining clean separation
- Separate interfaces for search and indexing operations
- Clean file naming without redundant prefixes

### Pipeline Architecture
Uses a pipeline pattern for data processing with common interfaces:
1. Reader loads and parses source data
2. Mapper transforms data according to configuration
3. Collector orchestrates the process
4. Factory creates storage instances based on configuration
5. Storage persists the articles with bulk operations support

### HTTP API Server
Built with Echo framework providing:
- **Search API**: Comprehensive search capabilities with multiple query types
- **Health Checks**: Database connectivity and service health monitoring
- **Middleware**: CORS support, request logging, and error recovery
- **Configuration**: Environment-based config with validation
- **Graceful Shutdown**: Proper resource cleanup on termination

**Search API Endpoints**:

1. **Simple Query String API** (`GET /v1/articles/search`)
   - Simple, Google-like search experience
   - Application determines optimal fields and weights
   - Query parameter: `?query=climate change`
   - Best for: End-user search boxes, simple text queries

2. **Structured Match API** (`POST /v1/articles/_search/match`)
   - Explicit single-field control
   - User specifies field, operator (AND/OR), fuzziness
   - Request body:
     ```json
     {
       "field": "title",
       "query": "climate change",
       "operator": "AND",
       "fuzziness": 1
     }
     ```
   - Best for: Advanced users needing precise control

**Search Features**:
- Full-text search with relevance ranking
- Single-field and multi-field search support
- Cursor-based pagination with total count and hasMore indicators
- Field-level weight control for relevance tuning
- Operator control (AND/OR) for term combination
- Fuzziness support for typo tolerance (Elasticsearch)
- Multi-language support (English, Serbian)
- Input validation and comprehensive error handling
- PostgreSQL: tsvector with ts_rank scoring
- Elasticsearch: match/multi_match queries with BM25 scoring

**API Design Pattern**:
- **Simple API**: Application-optimized (QueryString)
- **Explicit API**: User-controlled (Match, MultiMatch)
- **DTO Layer**: Clean separation between API and domain
- **Optional Interfaces**: Storage backends can optionally implement Match/MultiMatch
- **Extensible**: Easy to add new query types in the future

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
go build -o bin/data-import ./cmd/ds_ingest

# Run with environment variables
go run ./cmd/ds_ingest

# Build other commands
go build -o bin/schemagen ./cmd/schemagen
go build -o bin/news-search ./cmd/news_api

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

**Type System Purity**:
- Type definitions should represent search concepts, not implementation details
- Storage layer translates query types to engine-specific queries
- Type validation ensures type safety before reaching storage layer