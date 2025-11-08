# AGENTS.md

This file provides guidance for AI agents working with the News Hunter codebase.

## Project Context

**News Hunter** is a full-text search engine for multilingual news articles, built in Go 1.24+. This is a master thesis research project evaluating PostgreSQL as a search engine compared to Elasticsearch.

**Research Focus**: PostgreSQL vs Elasticsearch performance across multiple search paradigms (full-text, boolean, fuzzy, vector, hybrid search).

## Architecture Overview

### Core Structure
- **cmd/**: Application entry points
  - `data_import/`: Dataset import pipeline
  - `news_search/`: HTTP API server  
  - `schemagen/`: Schema generation utilities

- **internal/**: Business logic
  - `domain/`: Core entities (Article, Query, Score)
  - `storage/`: Storage abstractions with pg/es/in_mem implementations
  - `reader/`: CSV/YAML data processing pipeline
  - `collector/`: Data collection orchestration
  - `processor/`: Article processing logic
  - `router/`: HTTP route handlers
  - `server/`: HTTP server configuration
  - `middleware/`: HTTP middleware (logging, etc.)

- **pkg/**: Shared utilities
  - `apis/datamapping/`: Data mapping types
  - `pagination/`: Cursor/offset pagination
  - `schema/`: Schema generation
  - `utils/`: Common utilities

## Key Development Patterns

### Storage Layer
- Interface-based design with `Reader` and `Storer` contracts
- Factory pattern for storage creation (`storage/factory`)
- PostgreSQL implementation uses tsvector/ts_rank for full-text search
- Elasticsearch implementation with multilingual support
- Bulk operations support for performance

### Data Pipeline
1. Reader loads source data (CSV, JSON)
2. YAML config maps fields to internal Article structure
3. Collector orchestrates processing
4. Storage persists with bulk operations

### Configuration Management
- Environment-based configuration via `.env` files
- Separate configs for different storage backends
- YAML-based data mapping configurations

## Essential Commands

### Build & Run
```bash
# Build all commands
make build-all

# Build specific components
make build-data-import
make build-news-search
make build-schemagen

# Run services
make run-import-pg    # Import data to PostgreSQL
make run-import-es     # Import data to Elasticsearch
make run-search-pg     # Run search API with PostgreSQL
make run-search-es     # Run search API with Elasticsearch
```

### Development
```bash
make dev              # Format, vet, test, build
make test             # Run all tests
make fmt              # Format code
make vet              # Static analysis
make schema-gen       # Generate API schemas
```

### Database
```bash
make dc-up            # Start PostgreSQL via Docker
make migrate-up       # Run database migrations
```

## Testing Strategy

- Unit tests alongside source files (`*_test.go`)
- Integration tests for storage implementations
- Use `go test ./...` for full test suite
- Package-specific testing with `go test ./internal/reader`

## Code Conventions

### Go Standards
- Follow standard Go formatting (`go fmt`)
- Use `go vet` for static analysis
- Interface-based design for extensibility
- Error handling with explicit error returns

### Project-Specific
- Domain entities in `internal/domain/`
- Storage implementations in `storage/` subpackages
- Configuration via environment variables
- YAML-based data mapping configurations
- UUID-based primary keys for articles

## Key Files to Understand

### Core Domain
- `internal/domain/article.go`: Article entity definition
- `internal/domain/query.go`: Search query structures
- `internal/domain/score.go`: Relevance scoring

### Storage Layer
- `storage/reader.go`: Reader interface
- `storage/storer.go`: Storer interface
- `storage/factory/factory.go`: Storage factory
- `storage/pg/pg_storer.go`: PostgreSQL implementation
- `storage/es/es_storer.go`: Elasticsearch implementation

### Configuration
- `cmd/data_import/.env.example`: Data import config template
- `cmd/news_search/.env.example`: Search API config template
- `configs/mappings/gl_news_data_mapping.yaml`: Data mapping example

## Environment Variables

### Data Import
- `STORAGE_TYPE`: pg, es, or in_mem
- `MAPPING_CONFIG_PATH`: YAML mapping file path
- `DATASET_PATH`: Source dataset file path
- `PG_CONNECTION_STRING`: PostgreSQL connection
- `ES_ADDRESSES`: Elasticsearch cluster addresses
- `BULK_ENABLED`: Enable bulk operations
- `BULK_SIZE`: Bulk operation batch size

### Search API
- `PORT`: HTTP server port (default: 8080)
- `USE_HTTP2`: HTTP/2 support
- `CORS_ORIGINS`: Allowed CORS origins

## Database Schema

- PostgreSQL on port 54320
- Database: `news_db`
- User: `news_user`
- Password: `news_password`
- Migrations in `db/migrations/`
- Full-text search with tsvector columns

## Search Capabilities

### PostgreSQL Features
- tsvector/ts_rank for relevance scoring
- Full-text search with multiple languages
- Trigram similarity (pg_trgm) for fuzzy matching
- Vector search with pgvector extension

### Elasticsearch Features
- Multilingual search with analyzers
- Multi-match queries with field boosting
- Advanced indexing and relevance tuning

## Common Tasks

### Adding New Storage Backend
1. Implement `Reader` and `Storer` interfaces in `storage/newbackend/`
2. Add factory creation logic in `storage/factory/`
3. Add environment configuration
4. Update documentation

### Adding New Data Source
1. Create YAML mapping configuration in `configs/mappings/`
2. Add reader implementation if needed
3. Update data import pipeline
4. Test with sample data

### Modifying Search Features
1. Update domain query structures
2. Modify storage implementations
3. Update HTTP API handlers
4. Add tests for new functionality

## Performance Considerations

- Use bulk operations for data import
- Implement proper database indexing
- Consider connection pooling for PostgreSQL
- Monitor Elasticsearch cluster health
- Use pagination for large result sets

## Debugging Tips

- Check environment configuration in `.env` files
- Verify database connectivity with health checks
- Use verbose logging for debugging data pipeline
- Test storage implementations independently
- Monitor resource usage during bulk operations
