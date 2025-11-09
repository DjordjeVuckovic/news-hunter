package dto

type SearchQuery interface {
	isQuery()
}

type SearchRequest struct {
	Query  SearchQuery `json:"query"`
	Size   int         `json:"size,omitempty"`
	Cursor string      `json:"cursor,omitempty"`
}

type MatchQueryWrapper struct {
	Match map[string]MatchQueryParams `json:"match"`
}

type MatchQueryParams struct {
	Query    string `json:"query" validate:"required,min=1"`
	Operator string `json:"operator,omitempty"`
}

func (MatchQueryWrapper) isQuery() {}

type MultiMatchQueryWrapper struct {
	MultiMatch MultiMatchQueryParams `json:"multi_match"`
}

func (MultiMatchQueryWrapper) isQuery() {}

type BoolQueryWrapper struct {
	Bool BoolQueryParams `json:"bool"`
}

func (BoolQueryWrapper) isQuery() {}

type MultiMatchQueryParams struct {
	Query        string             `json:"query" validate:"required,min=1"`
	Fields       []string           `json:"fields" validate:"required,min=1"`
	FieldWeights map[string]float64 `json:"field_weights,omitempty"`
	Operator     string             `json:"operator,omitempty"`
}

type BoolQueryParams struct {
	Must    []any `json:"must,omitempty"`
	Should  []any `json:"should,omitempty"`
	MustNot []any `json:"must_not,omitempty"`
	Filter  []any `json:"filter,omitempty"`
}

type QueryConverter struct{}

//func (c *QueryConverter) ToDomain(wrapper SearchQuery) (interface{}, error) {
//	switch q := wrapper.(type) {
//	case *MatchQueryWrapper:
//		return c.convertMatchQuery(q)
//	case *MultiMatchQueryWrapper:
//		return c.convertMultiMatchQuery(q)
//	default:
//		return nil, fmt.Errorf("unknown query type: %T", wrapper)
//	}
//}
//
//func (c *QueryConverter) convertMatchQuery(wrapper *MatchQueryWrapper) (*query.Match, error) {
//	// Extract field and params from the map
//	if len(wrapper.Match) != 1 {
//		return nil, fmt.Errorf("match query must have exactly one field")
//	}
//
//	var field string
//	var params MatchQueryParams
//	for f, p := range wrapper.Match {
//		field = f
//		params = p
//	}
//
//	// Convert operator string to domain type
//	op, err := operator.Parse(params.Operator)
//	if err != nil {
//		return nil, fmt.Errorf("invalid operator: %w", err)
//	}
//}
//
//func (c *QueryConverter) convertMultiMatchQuery(wrapper *MultiMatchQueryWrapper) (*MultiMatchQuery, error) {
//	params := wrapper.MultiMatch
//
//	// Convert operator string to domain type
//	op, err := operator.Parse(params.Operator)
//	if err != nil {
//		return nil, fmt.Errorf("invalid operator: %w", err)
//	}
//
//	return &MultiMatchQuery{
//		Query:        params.Query,
//		Fields:       params.Fields,
//		FieldWeights: params.FieldWeights,
//		Operator:     op,
//		Language:     DefaultSearchLanguage,
//	}, nil
//}
