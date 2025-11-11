package query

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/DjordjeVuckovic/news-hunter/internal/domain/operator"
)

// Type QueryType represents the search paradigm to use
type Type string

const (

	// QueryStringType QueryTypeQueryString: Simple text-based search query (application-optimized)
	QueryStringType Type = "query_string"

	// MatchType: Single-field match query (Elasticsearch terminology)
	// ES: match query on single field
	// PG: tsvector search on single field
	MatchType Type = "match"

	// MultiMatchType: Multi-field match query (Elasticsearch terminology)
	// ES: multi_match query with field boosting
	// PG: weighted tsvector search across multiple fields
	MultiMatchType Type = "multi_match"

	// BooleanType: Structured queries with logical operators (AND, OR, NOT)
	BooleanType Type = "boolean"
)

// SearchQuery is the top-level query container
// Only one query field should be non-nil based on Type
type SearchQuery struct {
	Type        Type          `json:"type"`
	QueryString *String       `json:"query_string,omitempty"`
	Match       *Match        `json:"match,omitempty"`
	MultiMatch  *MultiMatch   `json:"multi_match,omitempty"`
	Boolean     *BooleanQuery `json:"boolean,omitempty"`
}

// String represents a simple text-based search query
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
type String struct {
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

type StringOption func(q *String)

// NewQueryString creates a new QueryString query with sensible defaults
func NewQueryString(query string, opts ...StringOption) *String {
	q := &String{
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
func WithQueryStringLanguage(lang Language) StringOption {
	return func(q *String) {
		q.Language = lang
	}
}

// WithQueryStringOperator sets the default operator for QueryString
func WithQueryStringOperator(op operator.Operator) StringOption {
	return func(q *String) {
		q.DefaultOperator = op
	}
}

// GetLanguage returns the language with default fallback
func (q *String) GetLanguage() Language {
	if q.Language == "" {
		return DefaultLanguage
	}
	return q.Language
}

// GetDefaultOperator returns the default operator with fallback
func (q *String) GetDefaultOperator() operator.Operator {
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

type MultiMatchField struct {
	Name   string
	Weight float64
}

func NewMultiMatchField(name string) MultiMatchField {
	return MultiMatchField{
		Name:   name,
		Weight: 1.0,
	}
}

func NewMultiMatchBoostedField(name string, boost float64) MultiMatchField {
	return MultiMatchField{
		Name:   name,
		Weight: boost,
	}
}

type MultiMatchStrategy string

const (
	// MultiMatchBestFields MultiMatchTypeBestFields: Finds documents that match any field, but uses the _best_ matching field to score each document.
	// For now, only this strategy is supported.
	MultiMatchBestFields MultiMatchStrategy = "best_fields"
)

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

	Fields []MultiMatchField `json:"fields,omitempty"`

	// Language: Text search language configuration
	Language Language `json:"language,omitempty"`

	// Operator: How to combine multiple terms
	// Default: operator.Or
	Operator operator.Operator `json:"operator,omitempty"`

	MatchStrategy MultiMatchStrategy `json:"match_strategy,omitempty"`
}
type MultiMatchQueryOption func(q *MultiMatch)

func NewMultiMatchQuery(query string, fields []string, opts ...MultiMatchQueryOption) (*MultiMatch, error) {
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if len(fields) == 0 {
		return nil, fmt.Errorf("fields are required")
	}

	q := &MultiMatch{
		Query:         query,
		Language:      DefaultLanguage,
		Operator:      operator.Default,
		MatchStrategy: MultiMatchBestFields,
		Fields:        newMultiMatchNewFields(fields),
	}

	for _, opt := range opts {
		opt(q)
	}

	return q, nil
}

func WithMultiMatchLanguage(lang Language) MultiMatchQueryOption {
	return func(q *MultiMatch) {
		q.Language = lang
	}
}

func WithMultiMatchOperator(op operator.Operator) MultiMatchQueryOption {
	return func(q *MultiMatch) {
		q.Operator = op
	}
}

func newMultiMatchNewFields(fields []string) []MultiMatchField {
	parsedFields := make([]MultiMatchField, 0, len(fields))

	for _, field := range fields {
		fieldParts := strings.Split(strings.TrimSpace(field), "^")
		switch len(fieldParts) {
		case 1:
			parsedFields = append(parsedFields, NewMultiMatchField(fieldParts[0]))
		case 2:
			weight, err := strconv.ParseFloat(fieldParts[1], 64)
			if err != nil {
				slog.Info("Invalid weight value in MultiMatchNewFields, defaulting to 1.0", "field", field, "error", err)
				weight = 1.0
			}
			parsedFields = append(parsedFields, NewMultiMatchBoostedField(fieldParts[0], weight))
		default:
			slog.Info("Invalid field format in MultiMatchNewFields", "field", field)
		}
	}

	return parsedFields
}

func (q *MultiMatch) GetLanguage() Language {
	if q.Language == "" {
		return DefaultLanguage
	}
	return q.Language
}

func (q *MultiMatch) GetFields() []MultiMatchField {
	return q.Fields
}

func (q *MultiMatch) GetOperator() operator.Operator {
	if q.Operator == "" {
		return operator.Default
	}
	return q.Operator
}
