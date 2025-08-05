package storage

import (
	"context"
	"github.com/DjordjeVuckovic/news-hunter/internal/es"
	"github.com/DjordjeVuckovic/news-hunter/internal/pg"
)

func NewStorer(storerType Type, ctx context.Context, cfg interface{}) (Storer, error) {
	var storer Storer
	var err error

	switch storerType {
	case PG:
		pool, err := pg.NewConnectionPool(ctx, cfg.(pg.Config))
		if err != nil {
			return nil, err
		}

		storer, err = pg.NewStorer(pool)
		if err != nil {
			return nil, err
		}
	case ES:
		storer, err = es.NewStorer(ctx, cfg.(es.Config))
		if err != nil {
			return nil, err
		}
	case InMem:
		return nil, ErrUnsupportedStorer
	default:
		return nil, ErrUnsupportedStorer
	}

	return storer, nil
}
