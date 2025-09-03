package factory

import (
	"context"
	"fmt"

	"github.com/DjordjeVuckovic/news-hunter/internal/storage"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage/es"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage/pg"
)

// NewStorer creates a new storage.Storer based on the storage type
func NewStorer(ctx context.Context, cfg StorageConfig) (storage.Storer, error) {
	switch cfg.Type {
	case storage.PG:
		pgConfig := *cfg.Pg

		pool, err := pg.NewConnectionPool(ctx, pgConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create PostgreSQL connection pool: %w", err)
		}

		return pg.NewStorer(pool)

	case storage.ES:
		esConfig := *cfg.Es

		return es.NewStorer(ctx, esConfig)

	case storage.InMem:
		return storage.NewInMemStorer(), nil

	default:
		return nil, fmt.Errorf(string(storage.ErrUnsupportedStorer), cfg.Type)
	}
}

// NewReader creates a new storage.Reader based on the storage type
func NewReader(ctx context.Context, cfg StorageConfig) (storage.Reader, error) {
	switch cfg.Type {
	case storage.PG:
		pgConfig := *cfg.Pg

		pool, err := pg.NewConnectionPool(ctx, pgConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create PostgreSQL connection pool: %w", err)
		}

		return pg.NewReader(pool)

	case storage.ES:
		esConfig := *cfg.Es

		return es.NewReader(esConfig)

	case storage.InMem:
		// TODO: Implement InMem Reader when needed
		return nil, fmt.Errorf("inmem reader not yet implemented")

	default:
		return nil, fmt.Errorf(string(storage.ErrUnsupportedStorer), cfg.Type)
	}
}
