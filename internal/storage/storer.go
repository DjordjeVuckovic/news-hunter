package storage

import (
	"context"
	"github.com/DjordjeVuckovic/news-hunter/internal/domain"
	"github.com/google/uuid"
)

type Storer interface {
	Save(ctx context.Context, article domain.Article) (uuid.UUID, error)
	SaveBulk(ctx context.Context, articles []domain.Article) error
}
