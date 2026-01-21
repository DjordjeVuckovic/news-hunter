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
	NextCursor   *query.Cursor             `json:"-"`
	HasMore      bool                      `json:"has_more"`
	MaxScore     float64                   `json:"max_score"`
	PageMaxScore float64                   `json:"page_max_score,omitempty"`
	TotalMatches int64                     `json:"total_matches,omitempty"`
}

// Searcher is the structured API interface
// Provides full-text search capability
type Searcher interface {
	// Search performs simple string-based search with application-optimized settings
	// The storage implementation determines optimal fields, weights, and search strategy
	// based on index configuration and content type.
	// size: number of results to return per page
	// Returns types objects with types cursor (not encoded string)
	Search(ctx context.Context, query *query.String, baseOpts *query.BaseOptions) (*SearchResult, error)
	// SearchField performs single-field match query with relevance scoring
	SearchField(ctx context.Context, query *query.Match, baseOpts *query.BaseOptions) (*SearchResult, error)
	// SearchFields performs multi-field match query with per-field weighting
	SearchFields(ctx context.Context, query *query.MultiMatch, baseOpts *query.BaseOptions) (*SearchResult, error)
	// SearchPhrase performs phrase search on specified fields with slop support
	// Elasticsearch: Uses match_phrase query with slop parameter
	// PostgreSQL: Uses phraseto_tsquery (slop=0) or to_tsquery with distance operators (slop>0)
	SearchPhrase(ctx context.Context, query *query.Phrase, baseOpts *query.BaseOptions) (*SearchResult, error)
}
