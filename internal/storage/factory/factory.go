package factory

import (
	"context"
	"fmt"

	"github.com/DjordjeVuckovic/news-hunter/internal/storage"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage/es"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage/pg"
)

// NewStorer creates a new storage.Storer based on the storage type
func NewStorer(storageType storage.Type, ctx context.Context, cfg interface{}) (storage.Storer, error) {
	switch storageType {
	case storage.PG:
		pgConfig, ok := cfg.(pg.Config)
		if !ok {
			return nil, fmt.Errorf("invalid config type for PostgreSQL storage: expected pg.ClientConfig")
		}

		pool, err := pg.NewConnectionPool(ctx, pgConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create PostgreSQL connection pool: %w", err)
		}

		return pg.NewStorer(pool)

	case storage.ES:
		esConfig, ok := cfg.(es.ClientConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config type for Elasticsearch storage: expected es.ClientConfig")
		}

		return es.NewStorer(ctx, esConfig)

	case storage.InMem:
		return storage.NewInMemStorer(), nil

	default:
		return nil, fmt.Errorf(string(storage.ErrUnsupportedStorer), storageType)
	}
}

// NewReader creates a new storage.Reader based on the storage type
func NewReader(storageType storage.Type, ctx context.Context, cfg interface{}) (storage.Reader, error) {
	switch storageType {
	case storage.PG:
		pgConfig, ok := cfg.(pg.Config)
		if !ok {
			return nil, fmt.Errorf("invalid config type for PostgreSQL storage: expected pg.ClientConfig")
		}

		pool, err := pg.NewConnectionPool(ctx, pgConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create PostgreSQL connection pool: %w", err)
		}

		return pg.NewReader(pool)

	case storage.ES:
		esConfig, ok := cfg.(es.ClientConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config type for Elasticsearch storage: expected es.ClientConfig")
		}

		return es.NewReader(esConfig)

	case storage.InMem:
		// TODO: Implement InMem Reader when needed
		return nil, fmt.Errorf("inmem reader not yet implemented")

	default:
		return nil, fmt.Errorf(string(storage.ErrUnsupportedStorer), storageType)
	}
}
