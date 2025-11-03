package storage

import (
	"context"

	"github.com/DjordjeVuckovic/news-hunter/internal/dto"
)

// SearchResult represents search results with cursor-based pagination
// Contains domain objects - no encoding/decoding at this layer
type SearchResult struct {
	Items      []dto.ArticleSearchResult `json:"items"`
	NextCursor *dto.Cursor               `json:"-"` // Domain cursor, not encoded
	HasMore    bool                      `json:"has_more"`
}

type Reader interface {
	// SearchFullText performs full-text search with cursor-based pagination
	// cursor: optional decoded cursor from previous result (nil for first page)
	// size: number of results to return per page
	// Returns domain objects with domain cursor (not encoded string)
	SearchFullText(ctx context.Context, query string, cursor *dto.Cursor, size int) (*SearchResult, error)
}
