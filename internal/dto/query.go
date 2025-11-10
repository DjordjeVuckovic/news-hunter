package dto

import (
	"encoding/json"
	"fmt"

	"github.com/DjordjeVuckovic/news-hunter/internal/domain/operator"
	"github.com/DjordjeVuckovic/news-hunter/internal/domain/query"
)

// SearchRequest represents the base search request with unified structure
// All search queries follow the pattern: {"size": N, "cursor": "...", "query": {"query_type": {...}}}
//
// Example Match:
//
//	{
//	  "size": 10,
//	  "cursor": "base64...",
//	  "query": {
//	    "match": {
//	      "field": "title",
//	      "query": "climate change",
//	      "operator": "and",
//	      "fuzziness": "AUTO",
//	      "language": "english"
//	    }
//	  }
//	}
//
// Example MultiMatch:
//
//	{
//	  "size": 10,
//	  "query": {
//	    "multi_match": {
//	      "query": "climate change",
//	      "fields": ["title", "description", "content"],
//	      "field_weights": {
//	        "title": 3.0,
//	        "description": 2.0,
//	        "content": 1.0
//	      },
//	      "operator": "or",
//	      "language": "english"
//	    }
//	  }
//	}
type SearchRequest struct {
	Size   int          `json:"size,omitempty"`
	Cursor string       `json:"cursor,omitempty"`
	Query  QueryWrapper `json:"query"`
}

// SearchResponse represents the API response for full-text search
// This is a concrete type for Swagger documentation (swag doesn't support generics yet)
type SearchResponse struct {
	NextCursor   *string               `json:"next_cursor,omitempty"`
	HasMore      bool                  `json:"has_more"`
	MaxScore     float64               `json:"max_score,omitempty"`
	PageMaxScore float64               `json:"page_max_score,omitempty"`
	TotalMatches int64                 `json:"total_matches,omitempty"`
	Hits         []ArticleSearchResult `json:"hits"`
}

// QueryWrapper wraps the actual query type
// Only one query field should be non-nil
type QueryWrapper struct {
	Match      *MatchParams      `json:"match,omitempty"`
	MultiMatch *MultiMatchParams `json:"multi_match,omitempty"`
}

// MatchParams represents match query parameters (maps directly to domain)
type MatchParams struct {
	Query     string `json:"query"`
	Field     string `json:"field"`
	Operator  string `json:"operator,omitempty"`
	Fuzziness string `json:"fuzziness,omitempty"`
	Language  string `json:"language,omitempty"`
}

// MultiMatchParams represents multi_match query parameters (maps directly to domain)
type MultiMatchParams struct {
	Query        string             `json:"query"`
	Fields       []string           `json:"fields"`
	FieldWeights map[string]float64 `json:"field_weights,omitempty"`
	Operator     string             `json:"operator,omitempty"`
	Language     string             `json:"language,omitempty"`
}

func (p *MatchParams) ToDomain() (*query.Match, error) {
	if p.Query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if p.Field == "" {
		return nil, fmt.Errorf("field is required")
	}

	var opts []query.MatchQueryOption

	if p.Operator != "" {
		op, err := operator.Parse(p.Operator)
		if err != nil {
			return nil, fmt.Errorf("invalid operator: %w", err)
		}
		opts = append(opts, query.WithMatchOperator(op))
	}

	if p.Fuzziness != "" {
		opts = append(opts, query.WithMatchFuzziness(p.Fuzziness))
	}

	if p.Language != "" {
		lang := query.Language(p.Language)
		if !query.SupportedLanguages[lang] {
			return nil, fmt.Errorf("unsupported language: %s", p.Language)
		}
		opts = append(opts, query.WithMatchLanguage(lang))
	}

	return query.NewMatch(p.Field, p.Query, opts...), nil
}

func (p *MultiMatchParams) ToDomain() (*query.MultiMatch, error) {
	if p.Query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if len(p.Fields) == 0 {
		return nil, fmt.Errorf("fields are required")
	}

	var opts []query.MultiMatchQueryOption

	if len(p.FieldWeights) > 0 {
		opts = append(opts, query.WithMultiMatchFieldWeights(p.FieldWeights))
	}

	if p.Operator != "" {
		op, err := operator.Parse(p.Operator)
		if err != nil {
			return nil, fmt.Errorf("invalid operator: %w", err)
		}
		opts = append(opts, query.WithMultiMatchOperator(op))
	}

	if p.Language != "" {
		lang := query.Language(p.Language)
		if !query.SupportedLanguages[lang] {
			return nil, fmt.Errorf("unsupported language: %s", p.Language)
		}
		opts = append(opts, query.WithMultiMatchLanguage(lang))
	}

	return query.NewMultiMatchQuery(p.Query, p.Fields, opts...), nil
}

// GetQueryType returns the type of query in the wrapper
func (q *QueryWrapper) GetQueryType() query.Type {
	if q.Match != nil {
		return query.MatchType
	}
	if q.MultiMatch != nil {
		return query.MultiMatchType
	}
	return ""
}

// UnmarshalJSON implements custom JSON unmarshaling with validation
func (q *QueryWrapper) UnmarshalJSON(data []byte) error {
	type Alias QueryWrapper
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(q),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Validate that exactly one query type is provided
	count := 0
	if q.Match != nil {
		count++
	}
	if q.MultiMatch != nil {
		count++
	}

	if count == 0 {
		return fmt.Errorf("query must specify one of: match, multi_match")
	}
	if count > 1 {
		return fmt.Errorf("query must specify only one query type")
	}

	return nil
}
