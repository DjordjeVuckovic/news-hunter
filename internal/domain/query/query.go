package query

import (
	"github.com/DjordjeVuckovic/news-hunter/internal/domain/operator"
)

// Type QueryType represents the search paradigm to use
type Type string

const (

	// QueryStringType QueryTypeQueryString: Simple text-based search query (application-optimized)
	QueryStringType Type = "query_string"

	// MatchType MatchType: Single-field match query (Elasticsearch terminology)
	// ES: match query on single field
	// PG: tsvector search on single field
	MatchType Type = "match"

	// MultiMatchType MultiMatchType: Multi-field match query (Elasticsearch terminology)
	// ES: multi_match query with field boosting
	// PG: weighted tsvector search across multiple fields
	MultiMatchType Type = "multi_match"

	// BooleanType BooleanType: Structured queries with logical operators (AND, OR, NOT)
	BooleanType Type = "boolean"
)

// SearchQuery is the top-level query container
// Only one query field should be non-nil based on Type
type SearchQuery struct {
	Type        Type          `json:"type"`
	QueryString *QueryString  `json:"query_string,omitempty"`
	Match       *Match        `json:"match,omitempty"`
	MultiMatch  *MultiMatch   `json:"multi_match,omitempty"`
	Boolean     *BooleanQuery `json:"boolean,omitempty"`
}

// QueryString represents a simple text-based search query
// The application parses the query string and determines optimal search strategy
// based on index configuration, content type, and query analysis.
//
// This is the primary search API for end-user queries (e.g., search box input).
// The application handles field selection, weighting, and query optimization.
//
// Inspired by Elasticsearch's query_string query.
//
// Examples:
//
//	"climate change"           → Multi-field text search with default operator
//	"renewable energy"         → Analyzed and tokenized across configured fields
type QueryString struct {
	// Query: The search text to query
	Query string `json:"query" validate:"required,min=1"`

	// Language: Text analysis language configuration
	Language Language `json:"language,omitempty"`

	// DefaultOperator: How to combine terms when no explicit operator specified
	// "climate change" with OR → "climate OR change"
	// "climate change" with AND → "climate AND change"
	// Default: operator.Or
	DefaultOperator operator.Operator `json:"default_operator,omitempty"`
}

// BooleanQuery BooleanQuery: Structured queries using logical operators
type BooleanQuery struct {
	// Expression: Boolean query string with operators
	// Supported operators:
	//   - AND (&): All terms must be present
	//   - OR (|): At least one term must be present
	//   - NOT (!): Term must not be present
	//   - (): Grouping for precedence
	//
	// Examples:
	//   "climate AND change"
	//   "(renewable OR sustainable) AND energy"
	//   "Trump AND NOT biden"
	//   "(climate OR weather) AND (change OR warming)"
	Expression string `json:"expression" validate:"required,min=1"`
}

var (
	// DefaultFields are the default fields to search when not specified
	DefaultFields = []string{"title", "description", "content"}

	// DefaultFieldWeights are the default field weights (equal weighting)
	DefaultFieldWeights = map[string]float64{
		"title":       1.0,
		"description": 1.0,
		"content":     1.0,
	}

	RecommendedFieldWeights = map[string]float64{
		"title":       3.0,
		"description": 2.0,
		"content":     1.0,
	}
)

type QueryStringOption func(q *QueryString)

// NewQueryString creates a new QueryString query with sensible defaults
func NewQueryString(query string, opts ...QueryStringOption) *QueryString {
	q := &QueryString{
		Query:           query,
		Language:        DefaultLanguage,
		DefaultOperator: operator.Or,
	}

	for _, opt := range opts {
		opt(q)
	}

	return q
}

// WithQueryStringLanguage sets the language for QueryString
func WithQueryStringLanguage(lang Language) QueryStringOption {
	return func(q *QueryString) {
		q.Language = lang
	}
}

// WithQueryStringOperator sets the default operator for QueryString
func WithQueryStringOperator(op operator.Operator) QueryStringOption {
	return func(q *QueryString) {
		q.DefaultOperator = op
	}
}

// GetLanguage returns the language with default fallback
func (q *QueryString) GetLanguage() Language {
	if q.Language == "" {
		return DefaultLanguage
	}
	return q.Language
}

// GetDefaultOperator returns the default operator with fallback
func (q *QueryString) GetDefaultOperator() operator.Operator {
	if q.DefaultOperator == "" {
		return operator.Or
	}
	return q.DefaultOperator
}

// Match Match: Single-field match query
// Performs analyzed full-text search on a single field with relevance scoring
// Elasticsearch: Translates to {"match": {"field": {"query": "text"}}}
// PostgreSQL: Uses weighted tsvector on single field
// Example:
//
//	{"field": "title", "query": "climate change", "operator": "and"}
type Match struct {
	// Query: The text to search for (analyzed and tokenized)
	Query string `json:"query" validate:"required,min=1"`

	// Field: The single field to search in
	Field string `json:"field" validate:"required"`

	// Language: Text search language configuration
	Language Language `json:"language,omitempty"`

	// Operator: How to combine multiple terms
	// Default: operator.Or
	Operator operator.Operator `json:"operator,omitempty"`

	// Fuzziness: Typo tolerance (general search concept)
	// "AUTO", "0", "1", "2" - Levenshtein edit distance
	// Elasticsearch: Native support via fuzziness parameter
	// PostgreSQL: Ignored (would require pg_trgm extension)
	Fuzziness string `json:"fuzziness,omitempty"`
}

// GetLanguage returns the language with default fallback
func (q *Match) GetLanguage() Language {
	if q.Language == "" {
		return DefaultLanguage
	}
	return q.Language
}

// GetOperator returns the operator with default fallback
func (q *Match) GetOperator() operator.Operator {
	if q.Operator == "" {
		return operator.Default
	}
	return q.Operator
}

type MatchQueryOption func(q *Match)

func NewMatch(field, query string, opts ...MatchQueryOption) *Match {
	q := &Match{
		Field:    field,
		Query:    query,
		Language: DefaultLanguage,
		Operator: operator.Default,
	}

	for _, opt := range opts {
		opt(q)
	}

	return q
}

// WithMatchLanguage sets the language for Match query
func WithMatchLanguage(lang Language) MatchQueryOption {
	return func(q *Match) {
		q.Language = lang
	}
}

// WithMatchOperator sets the operator for Match query
func WithMatchOperator(op operator.Operator) MatchQueryOption {
	return func(q *Match) {
		q.Operator = op
	}
}

// WithMatchFuzziness sets the fuzziness for Match query
func WithMatchFuzziness(fuzziness string) MatchQueryOption {
	return func(q *Match) {
		q.Fuzziness = fuzziness
	}
}

// MultiMatch MultiMatch: Multi-field match query (Elasticsearch terminology)
// Performs analyzed full-text search across multiple fields with per-field boosting
//
// Elasticsearch: Translates to {"multi_match": {"query": "text", "fields": ["title^3", "content"]}}
// PostgreSQL: Uses weighted tsvector with custom field weights
//
// Example:
//
//	{"query": "climate change", "fields": ["title", "content"], "field_weights": {"title": 3.0, "content": 1.0}}
type MultiMatch struct {
	// Query: The text to search for (analyzed and tokenized)
	Query string `json:"query" validate:"required,min=1"`

	// Fields: Which fields to search (required for multi_match)
	Fields []string `json:"fields" validate:"required,min=1"`

	// FieldWeights: Field-specific boosting (Elasticsearch terminology: boost)
	// Maps field names to boost multipliers
	// Example: {"title": 3.0, "description": 2.0, "content": 1.0}
	FieldWeights map[string]float64 `json:"field_weights,omitempty"`

	// Language: Text search language configuration
	Language Language `json:"language,omitempty"`

	// Operator: How to combine multiple terms
	// Default: operator.Or
	Operator operator.Operator `json:"operator,omitempty"`
}
type MultiMatchQueryOption func(q *MultiMatch)

// WithMultiMatchFieldWeights sets field weights for MultiMatch query
func WithMultiMatchFieldWeights(weights map[string]float64) MultiMatchQueryOption {
	return func(q *MultiMatch) {
		q.FieldWeights = weights
	}
}

// WithMultiMatchLanguage sets the language for MultiMatch query
func WithMultiMatchLanguage(lang Language) MultiMatchQueryOption {
	return func(q *MultiMatch) {
		q.Language = lang
	}
}

// WithMultiMatchOperator sets the operator for MultiMatch query
func WithMultiMatchOperator(op operator.Operator) MultiMatchQueryOption {
	return func(q *MultiMatch) {
		q.Operator = op
	}
}

func NewMultiMatchQuery(query string, fields []string, opts ...MultiMatchQueryOption) *MultiMatch {
	q := &MultiMatch{
		Query:        query,
		Fields:       fields,
		Language:     DefaultLanguage,
		Operator:     operator.Default,
		FieldWeights: make(map[string]float64),
	}

	for _, opt := range opts {
		opt(q)
	}

	return q
}

// GetLanguage returns the language with default fallback
func (q *MultiMatch) GetLanguage() Language {
	if q.Language == "" {
		return DefaultLanguage
	}
	return q.Language
}

// GetFields returns the fields with default fallback
func (q *MultiMatch) GetFields() []string {
	if len(q.Fields) == 0 {
		return DefaultFields
	}
	return q.Fields
}

// GetFieldWeight returns the weight for a specific field, or 1.0 if not specified
func (q *MultiMatch) GetFieldWeight(field string) float64 {
	if len(q.FieldWeights) == 0 {
		return 1.0
	}
	if weight, ok := q.FieldWeights[field]; ok {
		return weight
	}
	return 1.0
}

// GetOperator returns the operator with default fallback
func (q *MultiMatch) GetOperator() operator.Operator {
	if q.Operator == "" {
		return operator.Default
	}
	return q.Operator
}
