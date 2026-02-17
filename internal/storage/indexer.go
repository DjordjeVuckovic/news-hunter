package storage

import (
	"context"

	"github.com/DjordjeVuckovic/news-hunter/internal/embedding"
	"github.com/DjordjeVuckovic/news-hunter/internal/types/document"
	"github.com/google/uuid"
)

type Indexer interface {
	Save(ctx context.Context, article document.Article) (uuid.UUID, error)
	SaveBulk(ctx context.Context, articles []document.Article) error
}

type Type string

const (
	ES    Type = "es"
	PG         = "pg"
	Solr       = "solr"
	InMem      = "in_mem"
)

type StorerError string

const (
	ErrUnsupportedStorer StorerError = "unsupported storer type: %s"
)

func (e StorerError) Error() string {
	return string(e)
}

type EmbedIndexer interface {
	Save(ctx context.Context, article *embedding.Vec) (uuid.UUID, error)
	SaveBulk(ctx context.Context, article []*embedding.Vec) error
}
