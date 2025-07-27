package storage

import (
	"context"
	"github.com/DjordjeVuckovic/news-hunter/internal/domain"
)

type Reader interface {
	SearchBasic(ctx context.Context, query string, page int, size int) ([]domain.Article, error)
}
