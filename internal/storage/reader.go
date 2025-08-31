package storage

import (
	"context"

	"github.com/DjordjeVuckovic/news-hunter/internal/dto"
)

type SearchResult struct {
	Articles []dto.ArticleSearchResult `json:"articles"`
	Total    int64                     `json:"total"`
	Page     int                       `json:"page"`
	Size     int                       `json:"size"`
	HasMore  bool                      `json:"has_more"`
}

type Reader interface {
	SearchBasic(ctx context.Context, query string, page int, size int) (*SearchResult, error)
}
