# Plan: Test Container Helpers in `pkg/testing`

## Goal

Create PostgreSQL and Elasticsearch test container helpers using `testcontainers-go` for integration tests.

## Files to Create

### 1. `pkg/testing/pgcontainer.go`

PostgreSQL container helper that:
- Starts a `postgres:17.5` container (matching docker-compose)
- Auto-runs all migrations from `db/migrations/` (001 extensions, 002 articles table, 003 search vector trigger)
- Registers `t.Cleanup` for automatic container termination
- Returns a `PGContainer` struct with the connection string

```go
type PGContainer struct {
    ConnString string
}

func NewPGContainer(ctx context.Context, tb testing.TB) *PGContainer
```

Key implementation details:
- Use `postgres.Run()` from `testcontainers-go/modules/postgres`
- Use `postgres.WithDatabase("news_test_db")`, `postgres.WithUsername("test")`, `postgres.WithPassword("test")`
- Use `postgres.BasicWaitStrategies()` for readiness
- Run migrations by reading SQL files from `db/migrations/` sorted by filename (using `postgres.WithInitScripts` or executing via `container.Exec`)
- Since migrations use `DO $$ ... $$` blocks and extensions, execute them via `psql` inside the container using `ctr.Exec()`
- Connection string includes `sslmode=disable`
- `tb.Cleanup()` calls `testcontainers.TerminateContainer()`

### 2. `pkg/testing/escontainer.go`

Elasticsearch container helper that:
- Starts a `docker.elastic.co/elasticsearch/elasticsearch:8.12.0` container (matching docker-compose)
- Disables security (xpack.security.enabled=false)
- Registers `t.Cleanup` for automatic container termination
- Returns an `ESContainer` struct with the address

```go
type ESContainer struct {
    Address string
}

func NewESContainer(ctx context.Context, tb testing.TB) *ESContainer
```

Key implementation details:
- Use `elasticsearch.Run()` from `testcontainers-go/modules/elasticsearch`
- Configure single-node discovery and disable security via environment variables
- Use `elasticsearch.WithPassword("")` or env overrides to disable auth
- Get address from `elasticsearchContainer.Settings.Address`
- `tb.Cleanup()` calls `testcontainers.TerminateContainer()`

## Dependencies to Add

```
go get github.com/testcontainers/testcontainers-go/modules/postgres
go get github.com/testcontainers/testcontainers-go/modules/elasticsearch
```

## Migration Strategy

The migrations in `db/migrations/` are:
- `001_init.up.sql`: `CREATE EXTENSION uuid-ossp, pg_trgm, unaccent`
- `002_create_articles_table.up.sql`: `CREATE TABLE articles` with all columns + GIN index
- `003_add_search_vector_trigger.up.sql`: Trigger function for weighted tsvector

These will be run in order via `psql -f` commands inside the container using `ctr.Exec()`. The migration path will be resolved relative to the project root using `runtime.Caller` to locate the `db/migrations/` directory, and files will be copied into the container or read and executed.

Approach: Use `postgres.WithInitScripts()` which copies SQL files into the container's `/docker-entrypoint-initdb.d/` and runs them on startup. This requires providing the host paths to the migration files.

## Verification

1. Run `go build ./pkg/testing/...` to verify compilation
2. Run `go vet ./pkg/testing/...`
3. Verify both containers can start by running a simple test
