package factory

import (
	"context"
	"fmt"

	"github.com/DjordjeVuckovic/news-hunter/internal/embedding"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage/es"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage/in_mem"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage/pg"
	"github.com/DjordjeVuckovic/news-hunter/internal/storage/pg/native"
)

// TODO: accept pool as param and reuse it across indexer and searcher when using PG, to avoid creating multiple pools

// NewIndexer creates a new storage.Indexer based on the storage type
func NewIndexer(ctx context.Context, cfg StorageConfig) (storage.Indexer, error) {
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

func NewEmbedderIndexer(ctx context.Context, cfg StorageConfig) (storage.EmbedIndexer, error) {
	switch cfg.Type {
	case storage.PG:
		pgConfig := pg.PoolConfig{
			ConnStr:          cfg.Pg.ConnStr,
			RegisterVecTypes: true,
		}

		pool, err := pg.NewConnectionPool(ctx, pgConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create PostgreSQL connection pool: %w", err)
		}

		return pg.NewEmbedder(pool), nil

	case storage.ES:
		return nil, fmt.Errorf("elasticsearch embedder not yet implemented")
	}
	return nil, fmt.Errorf(string(storage.ErrUnsupportedStorer), cfg.Type)
}

// NewSearcher creates a new storage.FtsSearcher based on the storage type
func NewSearcher(ctx context.Context, cfg StorageConfig) (storage.FtsSearcher, error) {
	switch cfg.Type {
	case storage.PG:
		pgConfig := *cfg.Pg

		pool, err := pg.NewConnectionPool(ctx, pgConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create PostgreSQL connection pool: %w", err)
		}

		return native.NewReader(pool)

	case storage.ES:
		esConfig := *cfg.Es

		return es.NewSearcher(esConfig)

	case storage.Solr:
		return nil, fmt.Errorf("solr reader not yet implemented")

	case storage.InMem:
		// TODO: Implement InMem when needed
		return nil, fmt.Errorf("inmem reader not yet implemented")

	default:
		return nil, fmt.Errorf(string(storage.ErrUnsupportedStorer), cfg.Type)
	}
}

func NewSemanticSearcher(ctx context.Context, cfg StorageConfig, client embedding.Client) (storage.SemanticSearcher, error) {
	switch cfg.Type {
	case storage.PG:
		pgConfig := pg.PoolConfig{
			ConnStr:          cfg.Pg.ConnStr,
			RegisterVecTypes: true,
		}

		pool, err := pg.NewConnectionPool(ctx, pgConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create PostgreSQL connection pool: %w", err)
		}

		embedder := embedding.NewEmbedder(client, embedding.WithExecutorMaxLength(1024))

		return pg.NewSemanticSearcher(embedder, pool), nil

	case storage.ES:
		return nil, fmt.Errorf("elasticsearch semantic searcher not yet implemented")

	case storage.Solr:
		return nil, fmt.Errorf("solr semantic searcher not yet implemented")

	case storage.InMem:
		return nil, fmt.Errorf("inmem semantic searcher not yet implemented")

	default:
		return nil, fmt.Errorf(string(storage.ErrUnsupportedStorer), cfg.Type)
	}
}
