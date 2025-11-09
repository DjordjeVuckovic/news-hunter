package storage

import (
	"context"

	"github.com/DjordjeVuckovic/news-hunter/internal/domain"
	"github.com/DjordjeVuckovic/news-hunter/internal/dto"
)

// SearchResult represents search results with cursor-based pagination
// Contains domain objects - no encoding/decoding at this layer
type SearchResult struct {
	Hits         []dto.ArticleSearchResult `json:"hits"`
	NextCursor   *dto.Cursor               `json:"-"`
	HasMore      bool                      `json:"has_more"`
	MaxScore     float64                   `json:"max_score"`
	PageMaxScore float64                   `json:"page_max_score,omitempty"`
	TotalMatches int64                     `json:"total_matches,omitempty"`
}

// Reader is the base interface that ALL storage backends must implement
// Provides full-text search capability
type Reader interface {
	// SearchFullText performs token-based full-text search with relevance ranking
	// cursor: optional decoded cursor from previous result (nil for first page)
	// size: number of results to return per page
	// Returns domain objects with domain cursor (not encoded string)
	SearchFullText(ctx context.Context, query *domain.FullTextQuery, cursor *dto.Cursor, size int) (*SearchResult, error)
}

// BooleanSearcher is an optional interface for boolean search capabilities
// Storage backends that support structured queries with AND, OR, NOT operators should implement this
type BooleanSearcher interface {
	// SearchBoolean performs boolean search with logical operators
	// Supports AND, OR, NOT operators with grouping via parentheses
	// Example: "climate AND (change OR warming) AND NOT politics"
	SearchBoolean(ctx context.Context, query *domain.BooleanQuery, cursor *dto.Cursor, size int) (*SearchResult, error)
}
