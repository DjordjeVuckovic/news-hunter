package domain

// QueryType represents the search paradigm to use
type QueryType string

const (
	// QueryTypeFullText: Token-based full-text search with relevance ranking
	QueryTypeFullText QueryType = "full_text"

	// QueryTypeBoolean: Structured queries with logical operators (AND, OR, NOT)
	QueryTypeBoolean QueryType = "boolean"
)

// SearchQuery is the top-level query container
// Only one query field should be non-nil based on Type
type SearchQuery struct {
	Type     QueryType      `json:"type"`
	FullText *FullTextQuery `json:"full_text,omitempty"`
	Boolean  *BooleanQuery  `json:"boolean,omitempty"`
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
	// "and" - All terms must match (higher precision, lower recall)
	// "or" - Any term can match (lower precision, higher recall)
	// PostgreSQL: Controls whether to use plainto_tsquery (AND) vs OR combination
	// Elasticsearch: Sets the "operator" parameter in multi_match query
	// Default: "or"
	Operator string `json:"operator,omitempty"`
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

	// DefaultFieldWeights are the default field weights (equal weighting)
	DefaultFieldWeights = map[string]float64{
		"title":       1.0,
		"description": 1.0,
		"content":     1.0,
	}

	// RecommendedFieldWeights are recommended weights based on field importance
	// Title is most important, description is medium, content is base
	RecommendedFieldWeights = map[string]float64{
		"title":       3.0,
		"description": 2.0,
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
