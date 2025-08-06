package storage

import (
	"context"
	"github.com/DjordjeVuckovic/news-hunter/internal/domain"
)

type SearchResult struct {
	Articles []domain.Article `json:"articles"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	Size     int              `json:"size"`
	HasMore  bool             `json:"has_more"`
}

type Reader interface {
	SearchBasic(ctx context.Context, query string, page int, size int) (*SearchResult, error)
}
