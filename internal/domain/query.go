package domain

import "github.com/DjordjeVuckovic/news-hunter/internal/domain/operator"

// QueryType represents the search paradigm to use
type QueryType string

const (
	// QueryTypeFullText: Token-based full-text search with relevance ranking
	QueryTypeFullText QueryType = "full_text"

	// QueryTypeMatch: Single-field match query (Elasticsearch terminology)
	// ES: match query on single field
	// PG: tsvector search on single field
	QueryTypeMatch QueryType = "match"

	// QueryTypeMultiMatch: Multi-field match query (Elasticsearch terminology)
	// ES: multi_match query with field boosting
	// PG: weighted tsvector search across multiple fields
	QueryTypeMultiMatch QueryType = "multi_match"

	// QueryTypeBoolean: Structured queries with logical operators (AND, OR, NOT)
	QueryTypeBoolean QueryType = "boolean"
)

// SearchQuery is the top-level query container
// Only one query field should be non-nil based on Type
type SearchQuery struct {
	Type       QueryType        `json:"type"`
	FullText   *FullTextQuery   `json:"full_text,omitempty"`
	Match      *MatchQuery      `json:"match,omitempty"`
	MultiMatch *MultiMatchQuery `json:"multi_match,omitempty"`
	Boolean    *BooleanQuery    `json:"boolean,omitempty"`
}

// FullTextQuery: Token-based full-text search with relevance ranking
// Analyzes and tokenizes text, performs stemming, handles stop words
type FullTextQuery struct {
	Text string `json:"text" validate:"required,min=1"`

	// FieldWeights: Optional field-specific boosting/weights
	FieldWeights map[string]float64 `json:"field_weights,omitempty"`

	// Language: Text search language configuration
	Language SearchLanguage `json:"language,omitempty"`

	// Fields: Which fields to search
	Fields []string `json:"fields,omitempty"`

	// Operator: How to combine multiple terms (AND vs OR behavior)
	// Default: operator.Or
	Operator operator.Operator `json:"operator,omitempty"`
}

// BooleanQuery: Structured queries using logical operators
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

	// FieldWeights are the default field weights (equal weighting)
	FieldWeights = map[string]float64{
		"title":       1.0,
		"description": 1.0,
		"content":     1.0,
	}
)

type FullTextQueryOption func(q *FullTextQuery)

func NewFullTextQuery(text string, opts ...FullTextQueryOption) *FullTextQuery {
	q := &FullTextQuery{
		Text: text,
	}

	qBase := q.WithDefaults()

	for _, opt := range opts {
		opt(qBase)
	}

	return qBase
}

// WithDefaults returns a copy of FullTextQuery with default values applied
func (q *FullTextQuery) WithDefaults() *FullTextQuery {
	result := &FullTextQuery{
		Text:         q.Text,
		FieldWeights: q.FieldWeights,
		Language:     q.Language,
		Fields:       q.Fields,
	}

	if result.Language == "" {
		result.Language = DefaultSearchLanguage
	}

	if len(result.Fields) == 0 {
		result.Fields = DefaultFields
	}

	if len(result.FieldWeights) == 0 {
		result.FieldWeights = make(map[string]float64)
		for _, field := range result.Fields {
			result.FieldWeights[field] = 1.0
		}
	}

	return result
}

// GetLanguage returns the language with default fallback
func (q *FullTextQuery) GetLanguage() SearchLanguage {
	if q.Language == "" {
		return DefaultSearchLanguage
	}
	return q.Language
}

// GetFields returns the fields with default fallback
func (q *FullTextQuery) GetFields() []string {
	if len(q.Fields) == 0 {
		return DefaultFields
	}
	return q.Fields
}

// GetFieldWeight returns the weight for a specific field, or 1.0 if not specified
func (q *FullTextQuery) GetFieldWeight(field string) float64 {
	if len(q.FieldWeights) == 0 {
		return 1.0
	}
	if weight, ok := q.FieldWeights[field]; ok {
		return weight
	}
	return 1.0
}

// MatchQuery: Single-field match query (Elasticsearch terminology)
// Performs analyzed full-text search on a single field with relevance scoring
//
// Elasticsearch: Translates to {"match": {"field": {"query": "text"}}}
// PostgreSQL: Uses weighted tsvector on single field
//
// Example:
//
//	{"field": "title", "query": "climate change", "operator": "and"}
type MatchQuery struct {
	// Query: The text to search for (analyzed and tokenized)
	Query string `json:"query" validate:"required,min=1"`

	// Field: The single field to search in
	Field string `json:"field" validate:"required"`

	// Language: Text search language configuration
	Language SearchLanguage `json:"language,omitempty"`

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
func (q *MatchQuery) GetLanguage() SearchLanguage {
	if q.Language == "" {
		return DefaultSearchLanguage
	}
	return q.Language
}

// GetOperator returns the operator with default fallback
func (q *MatchQuery) GetOperator() operator.Operator {
	if q.Operator == "" {
		return operator.Default
	}
	return q.Operator
}

type MatchQueryOption func(q *MatchQuery)

func NewMatchQuery(field, query string, opts ...MatchQueryOption) *MatchQuery {
	q := &MatchQuery{
		Field:    field,
		Query:    query,
		Language: DefaultSearchLanguage,
		Operator: operator.Default,
	}

	for _, opt := range opts {
		opt(q)
	}

	return q
}

// MultiMatchQuery: Multi-field match query (Elasticsearch terminology)
// Performs analyzed full-text search across multiple fields with per-field boosting
//
// Elasticsearch: Translates to {"multi_match": {"query": "text", "fields": ["title^3", "content"]}}
// PostgreSQL: Uses weighted tsvector with custom field weights
//
// Example:
//
//	{"query": "climate change", "fields": ["title", "content"], "field_weights": {"title": 3.0, "content": 1.0}}
type MultiMatchQuery struct {
	// Query: The text to search for (analyzed and tokenized)
	Query string `json:"query" validate:"required,min=1"`

	// Fields: Which fields to search (required for multi_match)
	Fields []string `json:"fields" validate:"required,min=1"`

	// FieldWeights: Field-specific boosting (Elasticsearch terminology: boost)
	// Maps field names to boost multipliers
	// Example: {"title": 3.0, "description": 2.0, "content": 1.0}
	FieldWeights map[string]float64 `json:"field_weights,omitempty"`

	// Language: Text search language configuration
	Language SearchLanguage `json:"language,omitempty"`

	// Operator: How to combine multiple terms
	// Default: operator.Or
	Operator operator.Operator `json:"operator,omitempty"`
}
type MultiMatchQueryOption func(q *MultiMatchQuery)

func WithMultiMatchFieldWeights(weights map[string]float64) MultiMatchQueryOption {
	return func(q *MultiMatchQuery) {
		q.FieldWeights = weights
	}
}
func NewMultiMatchQuery(query string, fields []string, opts ...MultiMatchQueryOption) *MultiMatchQuery {
	q := &MultiMatchQuery{
		Query:        query,
		Fields:       fields,
		Language:     DefaultSearchLanguage,
		Operator:     operator.Default,
		FieldWeights: make(map[string]float64),
	}

	for _, opt := range opts {
		opt(q)
	}

	return q
}

// GetLanguage returns the language with default fallback
func (q *MultiMatchQuery) GetLanguage() SearchLanguage {
	if q.Language == "" {
		return DefaultSearchLanguage
	}
	return q.Language
}

// GetFields returns the fields with default fallback
func (q *MultiMatchQuery) GetFields() []string {
	if len(q.Fields) == 0 {
		return DefaultFields
	}
	return q.Fields
}

// GetFieldWeight returns the weight for a specific field, or 1.0 if not specified
func (q *MultiMatchQuery) GetFieldWeight(field string) float64 {
	if len(q.FieldWeights) == 0 {
		return 1.0
	}
	if weight, ok := q.FieldWeights[field]; ok {
		return weight
	}
	return 1.0
}

// GetOperator returns the operator with default fallback
func (q *MultiMatchQuery) GetOperator() operator.Operator {
	if q.Operator == "" {
		return operator.Default
	}
	return q.Operator
}
