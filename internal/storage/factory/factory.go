package factory

import (
	"context"
	"fmt"

	"github.com/DjordjeVuckovic/news-hunter/internal/storage"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage/es"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage/in_mem"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage/pg"
)

// NewStorer creates a new storage.Indexer based on the storage type
func NewStorer(ctx context.Context, cfg StorageConfig) (storage.Indexer, error) {
	switch cfg.Type {
	case storage.PG:
		pgConfig := *cfg.Pg

		pool, err := pg.NewConnectionPool(ctx, pgConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create PostgreSQL connection pool: %w", err)
		}

		return pg.NewIndexer(pool)

	case storage.ES:
		esConfig := *cfg.Es

		return es.NewIndexer(ctx, esConfig)

	case storage.Solr:
		return nil, fmt.Errorf("solr storer not yet implemented")

	case storage.InMem:
		return in_mem.NewInMemIndexer(), nil

	default:
		return nil, fmt.Errorf(string(storage.ErrUnsupportedStorer), cfg.Type)
	}
}

// NewReader creates a new storage.FTSSearcher based on the storage type
func NewReader(ctx context.Context, cfg StorageConfig) (storage.FTSSearcher, error) {
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

		return es.NewSeacher(esConfig)

	case storage.Solr:
		return nil, fmt.Errorf("solr reader not yet implemented")

	case storage.InMem:
		// TODO: Implement InMem FTSSearcher when needed
		return nil, fmt.Errorf("inmem reader not yet implemented")

	default:
		return nil, fmt.Errorf(string(storage.ErrUnsupportedStorer), cfg.Type)
	}
}
