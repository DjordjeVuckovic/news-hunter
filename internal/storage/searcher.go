package storage

import (
	"context"

	"github.com/DjordjeVuckovic/news-hunter/internal/dto"
	"github.com/DjordjeVuckovic/news-hunter/internal/types/query"
)

// SearchResult represents search results with cursor-based pagination
// Contains types objects - no encoding/decoding at this layer
type SearchResult struct {
	Hits         []dto.ArticleSearchResult `json:"hits"`
	NextCursor   *dto.Cursor               `json:"-"`
	HasMore      bool                      `json:"has_more"`
	MaxScore     float64                   `json:"max_score"`
	PageMaxScore float64                   `json:"page_max_score,omitempty"`
	TotalMatches int64                     `json:"total_matches,omitempty"`
}

// Searcher is the base interface that ALL storage backends must implement
// Provides full-text search capability
type Searcher interface {
	// SearchQueryString performs simple string-based search with application-optimized settings
	// The storage implementation determines optimal fields, weights, and search strategy
	// based on index configuration and content type.
	//
	// cursor: optional decoded cursor from previous result (nil for first page)
	// size: number of results to return per page
	// Returns types objects with types cursor (not encoded string)
	SearchQueryString(ctx context.Context, query *query.String, cursor *dto.Cursor, size int) (*SearchResult, error)
	// SearchMatch performs single-field match query with relevance scoring
	SearchMatch(ctx context.Context, query *query.Match, cursor *dto.Cursor, size int) (*SearchResult, error)
	// SearchMultiMatch performs multi-field match query with per-field weighting
	SearchMultiMatch(ctx context.Context, query *query.MultiMatch, cursor *dto.Cursor, size int) (*SearchResult, error)
}

type PhraseSearcher interface {
	// SearchPhrase performs phrase search on specified field with slop support
	// Elasticsearch: Uses match_phrase query with slop parameter
	// PostgreSQL: Uses phraseto_tsquery with positional matching
	SearchPhrase(ctx context.Context, query *query.Phrase, cursor *dto.Cursor, size int) (*SearchResult, error)
	SearchBoolean(ctx context.Context, query *query.Boolean, cursor *dto.Cursor, size int) (*SearchResult, error)
}
