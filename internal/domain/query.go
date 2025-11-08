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
	// Map of field name to boost multiplier
	// Example: {"title": 3.0, "description": 2.0, "content": 1.0}
	// PostgreSQL: Uses setweight() with A, B, C, D based on weight values
	// Elasticsearch: Uses field^boost syntax in multi_match query
	// If nil/empty, uses default equal weighting across all fields
	FieldWeights map[string]float64 `json:"field_weights,omitempty"`

	// Language: Text search language configuration
	// PostgreSQL: Maps to text search configuration (e.g., 'english', 'spanish', 'french')
	// Elasticsearch: Maps to analyzer (e.g., 'english', 'spanish')
	// Default: 'english' if not specified
	Language string `json:"language,omitempty"`

	// Fields: Which fields to search
	// Example: ["title", "description", "content"]
	// If nil/empty, searches all text fields (title, description, content)
	Fields []string `json:"fields,omitempty"`

	// Operator: How to combine multiple terms (AND vs OR behavior)
	// "and" - All terms must match (higher precision, lower recall)
	// "or" - Any term can match (lower precision, higher recall)
	// PostgreSQL: Controls whether to use plainto_tsquery (AND) vs OR combination
	// Elasticsearch: Sets the "operator" parameter in multi_match query
	// Default: "or"
	Operator string `json:"operator,omitempty"`

	// MinimumShouldMatch: Minimum number/percentage of terms that should match
	// Examples: "2", "75%", "3<90%"
	// Elasticsearch: Direct mapping to minimum_should_match parameter
	// PostgreSQL: Custom implementation with partial matching
	// Default: nil (uses operator setting)
	MinimumShouldMatch *string `json:"minimum_should_match,omitempty"`
}

// BooleanQuery: Structured queries using logical operators
// Allows complex logic with AND, OR, NOT operators and grouping with parentheses
//
// Example: "climate AND (change OR warming) AND NOT politics"
// PostgreSQL: Converts to tsquery syntax: 'climate' & ('change' | 'warming') & !'politics'
// Elasticsearch: Converts to bool query with must, should, must_not clauses
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

// Default values for full-text search
const (
	DefaultLanguage = "english"
)

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

// WithDefaults returns a copy of FullTextQuery with default values applied
func (q *FullTextQuery) WithDefaults() *FullTextQuery {
	result := &FullTextQuery{
		Text:         q.Text,
		FieldWeights: q.FieldWeights,
		Language:     q.Language,
		Fields:       q.Fields,
	}

	// Set default language
	if result.Language == "" {
		result.Language = DefaultLanguage
	}

	// Set default fields
	if len(result.Fields) == 0 {
		result.Fields = DefaultFields
	}

	// Set default field weights (equal weighting if not specified)
	if len(result.FieldWeights) == 0 {
		result.FieldWeights = make(map[string]float64)
		for _, field := range result.Fields {
			result.FieldWeights[field] = 1.0
		}
	}

	return result
}

// GetLanguage returns the language with default fallback
func (q *FullTextQuery) GetLanguage() string {
	if q.Language == "" {
		return DefaultLanguage
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
